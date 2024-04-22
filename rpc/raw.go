package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
)

type MessageKind string

const (
	INVALID_KIND          MessageKind = "invalid"
	NOTIFICATION_KIND     MessageKind = "notification"
	REQUEST_KIND          MessageKind = "request"
	ERROR_RESPONSE_KIND   MessageKind = "error"
	SUCCESS_RESPONSE_KIND MessageKind = "success"

	jsonRpcVersion = "2.0"
)

type IdType interface {
	~string | ~int64 | ~int32 | ~int16 | ~int8 | ~int | ~uint64 | ~uint32 | ~uint16 | ~uint8 | ~uint
}
type ResultType interface{ any }
type ParamsType interface {
	any | []any
}
type RawMessageType interface{ string | ~[]byte }

type MessageBase struct {
	Version string `json:"jsonrpc"`
}

type Message struct {
	MessageBase
	Id     interface{}     `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *ErrorObj       `json:"error,omitempty"`
}

func (r *Message) IsRequest() bool {
	return r.Method != ""
}

func (r *Message) IsSuccessResponse() bool {
	return r.Result != nil
}

func (r *Message) IsErrorResponse() bool {
	return r.Error != nil
}

func (r *Message) isNonDeterminableKind() bool {
	matchedKinds := []bool{
		r.IsRequest(),
		r.IsSuccessResponse(),
		r.IsErrorResponse(),
	}
	truthy := 0
	for _, matched := range matchedKinds {
		if matched {
			truthy++
		}
	}
	return truthy != 1
}

func (r *Message) GetKind() (MessageKind, error) {
	if r.Version != jsonRpcVersion {
		return INVALID_KIND, fmt.Errorf("invalid jsonrpc version: %s", r.Version)
	}
	if r.isNonDeterminableKind() {
		return INVALID_KIND, ErrInternalInvalidMessageStructure
	}
	if r.IsRequest() {
		if r.Method == "" {
			return INVALID_KIND, ErrInternalMethodRequired
		}
		if r.Id == nil {
			return NOTIFICATION_KIND, nil
		}
		return REQUEST_KIND, nil
	}
	if r.IsSuccessResponse() || r.IsErrorResponse() {
		if r.Id == nil {
			return INVALID_KIND, errors.New("id is required")
		}
		switch r.Id.(type) {
		case string, int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint:
		default:
			return INVALID_KIND, fmt.Errorf("invalid id type: %T", r.Id)
		}
	}
	if r.IsSuccessResponse() {
		return SUCCESS_RESPONSE_KIND, nil
	}
	if r.IsErrorResponse() {
		return ERROR_RESPONSE_KIND, nil
	}

	return INVALID_KIND, ErrInternalInvalidMessageStructure
}

func MessageToRequest[TParam ParamsType](r *Message) *Request[TParam] {
	method := r.Method
	var params TParam
	if r.Params != nil {
		err := json.Unmarshal(r.Params, &params)
		if err != nil {
			return nil
		}
	}

	return &Request[TParam]{
		MessageBase: MessageBase{Version: jsonRpcVersion},
		Id:          r.Id,
		Method:      method,
		Params:      params,
	}
}

func MessageToSuccessResponse[TResult ResultType](rpc *Message) (*SuccessResponse[TResult], error) {
	if !rpc.IsSuccessResponse() {
		return nil, errors.New("invalid rpc message type - not a success response")
	}
	var result TResult
	if rpc.Result != nil {
		err := json.Unmarshal(rpc.Result, &result)
		if err != nil {
			return nil, err
		}
	}

	return &SuccessResponse[TResult]{
		MessageBase{Version: jsonRpcVersion},
		rpc.Id,
		result,
		nil,
	}, nil
}

func MessageToErrorResponse(rpc *Message) (*ErrorResponse, error) {
	if !rpc.IsSuccessResponse() {
		return nil, errors.New("invalid rpc message type - not a success response")
	}

	return &ErrorResponse{
		MessageBase{Version: jsonRpcVersion},
		rpc.Id,
		nil,
		rpc.Error,
	}, nil
}

func MessageToResponse[TResult ResultType](rpc *Message) (*Response[TResult], error) {
	kind, err := rpc.GetKind()
	if err != nil {
		return nil, err
	}

	response := &Response[TResult]{
		MessageBase: MessageBase{Version: jsonRpcVersion},
		Id:          rpc.Id,
	}

	switch kind {
	case SUCCESS_RESPONSE_KIND:
		var result TResult
		err := json.Unmarshal(rpc.Result, &result)
		if err != nil {
			return nil, err
		}
		response.Result = result
		return response, nil
	case ERROR_RESPONSE_KIND:
		response.Error = rpc.Error
		return response, nil
	default:
		return nil, fmt.Errorf("invalid message kind: %s", kind)
	}
}

// Object is a wrapper for a single or batch of rpc messages
type Object struct {
	messages []Message
	isBatch  bool
}

func (r *Object) IsBatch() bool {
	return r.isBatch
}

func (r *Object) GetMessages() []Message {
	return r.messages
}

func (r *Object) GetSingleMessage() *Message {
	if r.isBatch {
		return nil
	}
	return &r.messages[0]
}

func (r *Object) UnmarshalJSON(data []byte) error {
	if data[0] == '[' {
		r.isBatch = true
		return json.Unmarshal(data, &r.messages)
	}
	r.isBatch = false
	r.messages = make([]Message, 1)
	return json.Unmarshal(data, &r.messages[0])
}

func (r *Object) MarshalJSON() ([]byte, error) {
	if r.isBatch {
		return json.Marshal(r.messages)
	}
	return json.Marshal(r.messages[0])
}
