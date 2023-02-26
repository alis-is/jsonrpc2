package rpc

import (
	"fmt"
)

type Response[TResult ResultType] struct {
	MessageBase
	Id     interface{} `json:"id"`
	Result TResult     `json:"result,omitempty"`
	Error  *ErrorObj   `json:"error,omitempty"`
}

func NewResponseI[TResult ResultType](id interface{}, result TResult, err *ErrorObj) *Response[TResult] {
	return &Response[TResult]{
		MessageBase{Version: jsonRpcVersion},
		id,
		result,
		err,
	}
}

func NewResponse[TResult ResultType](id interface{}, result TResult, err *ErrorObj) *Response[TResult] {
	var zero TResult
	return NewResponseI(id, zero, nil)
}

func (r *Response[TResult]) IsSuccess() bool {
	return r.Error == nil
}

func (r *Response[TResult]) IsError() bool {
	return r.Error != nil
}

func (r *Response[TResult]) Unwrap() (TResult, error) {
	var zero TResult
	if r.IsSuccess() {
		return r.Result, nil
	} else {
		if r.Error.Data != nil {
			return zero, fmt.Errorf("rpc error: %s (code: %d, data: %s)", r.Error.Message, r.Error.Code, string(*r.Error.Data))
		}
		return zero, fmt.Errorf("rpc error: %s (code: %d)", r.Error.Message, r.Error.Code)
	}
}

type SuccessResponse[TResult ResultType] Response[TResult]

func NewSuccessResponseI[TResult ResultType](id interface{}, result TResult) *SuccessResponse[TResult] {
	return &SuccessResponse[TResult]{
		MessageBase{Version: jsonRpcVersion},
		id,
		result,
		nil,
	}
}

func NewSuccessResponse[TId IdType, TResult ResultType](id TId, result TResult) *SuccessResponse[TResult] {
	return NewSuccessResponseI((interface{})(id), result)
}

type ErrorResponse Response[interface{}]

func NewErrorResponseI(id interface{}, err *ErrorObj) *ErrorResponse {
	return &ErrorResponse{
		MessageBase{
			jsonRpcVersion,
		},
		id,
		nil,
		err,
	}
}

func NewErrorResponse[TId IdType](id TId, err *ErrorObj) *ErrorResponse {
	return NewErrorResponseI((interface{})(id), err)
}
