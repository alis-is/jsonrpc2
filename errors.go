package jsonrpc2

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

var (
	ErrInternalInvalidJsonRpcMessage   = errors.New("invalid jsonrpc 2.0 message")
	ErrInternalNotRequest              = errors.New("not a request")
	ErrInternalMethodRequired          = errors.New("method is required")
	ErrInternalInvalidMessageStructure = errors.New("invalid message structure")
	ErrInternalUnsupportedMessageKind  = errors.New("unsupported message kind")
	ErrEmptyResponse                   = errors.New("empty response")
)

type ErrorObj struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    *json.RawMessage `json:"data,omitempty"`
}

func (e *ErrorObj) ToErrorResponse(id interface{}) *errorResponse {
	return NewErrorResponseI(id, e)
}

func (e *ErrorObj) ToResponse(id interface{}) *Response[interface{}] {
	return NewResponseI[interface{}](id, nil, e)
}

func (e *ErrorObj) ToErrorResponseBytes(id interface{}) []byte {
	result, err := json.Marshal(NewErrorResponseI(id, e))
	if err != nil {
		result, _ = json.Marshal(NewErrorResponseI(id, NewInternalErrorWithData(err.Error()).toErrorObj()))
	}
	return result
}

type ErrorKind string

const (
	ParseErrorKind     ErrorKind = "Parse error"
	InvalidRequestKind ErrorKind = "Invalid Request"
	MethodNotFoundKind ErrorKind = "Method not found"
	InvalidParamsKind  ErrorKind = "Invalid params"
	InternalErrorKind  ErrorKind = "Internal error"
	UnknownErrorKind   ErrorKind = "Unknown error"
	ServerErrorKind    ErrorKind = "Server error"
)

type Error struct {
	code int
	Kind ErrorKind
	Data interface{}
}

func NewParseError() *Error {
	return &Error{Kind: ParseErrorKind, code: -32700}
}

func NewParseErrorWithData[T any](data T) *Error {
	result := NewParseError()
	result.Data = data
	return result
}

func NewInvalidRequest() *Error {
	return &Error{Kind: InvalidRequestKind, code: -32600}
}

func NewInvalidRequestWithData[T any](data T) *Error {
	result := NewInvalidRequest()
	result.Data = data
	return result
}

func NewMethodNotFound() *Error {
	return &Error{Kind: MethodNotFoundKind, code: -32601}
}

func NewMethodNotFoundWithData[T any](data T) *Error {
	result := NewMethodNotFound()
	result.Data = data
	return result
}

func NewInvalidParams() *Error {
	return &Error{Kind: InvalidParamsKind, code: -32602}
}

func NewInvalidParamsWithData[T any](data T) *Error {
	result := NewInvalidParams()
	result.Data = data
	return result
}

func NewInternalError() *Error {
	return &Error{Kind: InternalErrorKind, code: -32603}
}

func NewInternalErrorWithData[T any](data T) *Error {
	result := NewInternalError()
	result.Data = data
	return result
}

func NewUnknownError() *Error {
	return &Error{Kind: UnknownErrorKind, code: -32000}
}

func NewUnknownErrorWithData[T any](data T) *Error {
	result := NewUnknownError()
	result.Data = data
	return result
}

// code has to be in range -32099 to -32000
func NewServerError(code int) *Error {
	return &Error{Kind: InternalErrorKind}
}

func NewServerErrorWithData[T any](code int, data T) *Error {
	result := NewServerError(code)
	result.Data = data
	return result
}

func (e *Error) Error() string {
	return strings.ToLower(string(e.Kind))
}

func (e *Error) toErrorObj() *ErrorObj {
	var data *json.RawMessage
	if e.Data != nil {
		data = new(json.RawMessage)
		*data, _ = json.Marshal(e.Data)
	}

	switch e.Kind {
	case ParseErrorKind:
		return &ErrorObj{Code: e.code, Message: string(e.Kind), Data: data}
	case InvalidRequestKind:
		return &ErrorObj{Code: e.code, Message: string(e.Kind), Data: data}
	case MethodNotFoundKind:
		return &ErrorObj{Code: e.code, Message: string(e.Kind), Data: data}
	case InvalidParamsKind:
		return &ErrorObj{Code: e.code, Message: string(e.Kind), Data: data}
	case InternalErrorKind:
		return &ErrorObj{Code: e.code, Message: string(e.Kind), Data: data}
	case ServerErrorKind:
		return &ErrorObj{Code: e.code, Message: string(e.Kind), Data: data}
	default:
		return &ErrorObj{Code: e.code, Message: string(e.Kind), Data: data}
	}
}

func (e *Error) ToResponse(id interface{}) *errorResponse {
	return NewErrorResponseI(id, e.toErrorObj())
}

func (e *Error) ToResponseBytes(id interface{}) []byte {
	return e.toErrorObj().ToErrorResponseBytes(id)
}

func (e *Error) ToHttpError() int {
	switch e.Kind {
	case ParseErrorKind:
		return http.StatusInternalServerError
	case InvalidRequestKind:
		return http.StatusBadRequest
	case MethodNotFoundKind:
		return http.StatusNotFound
	case InvalidParamsKind:
		return http.StatusInternalServerError
	case InternalErrorKind:
		return http.StatusInternalServerError
	case ServerErrorKind:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func ResponseFromError[TId Id](id TId, err *Error) *errorResponse {
	return NewErrorResponse(id, err.toErrorObj())
}
