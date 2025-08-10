package jsonrpc2

import (
	"fmt"
)

type response[TResult Result] struct {
	messageBase
	Id     interface{} `json:"id"`
	Result TResult     `json:"result,omitempty"`
	Error  *ErrorObj   `json:"error,omitempty"`
}

func NewResponseI[TResult Result](id interface{}, result TResult, err *ErrorObj) *response[TResult] {
	return &response[TResult]{
		messageBase{Version: jsonRpcVersion},
		id,
		result,
		err,
	}
}

func NewResponse[TResult Result](id interface{}, result TResult, err *ErrorObj) *response[TResult] {
	var zero TResult
	return NewResponseI(id, zero, nil)
}

func (r *response[TResult]) IsSuccess() bool {
	return r.Error == nil
}

func (r *response[TResult]) IsError() bool {
	return r.Error != nil
}

func (r *response[TResult]) Unwrap() (TResult, error) {
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

type successResponse[TResult Result] response[TResult]

func NewSuccessResponseI[TResult Result](id interface{}, result TResult) *successResponse[TResult] {
	return &successResponse[TResult]{
		messageBase{Version: jsonRpcVersion},
		id,
		result,
		nil,
	}
}

func NewSuccessResponse[TId Id, TResult Result](id TId, result TResult) *successResponse[TResult] {
	return NewSuccessResponseI((interface{})(id), result)
}

type errorResponse response[interface{}]

func NewErrorResponseI(id interface{}, err *ErrorObj) *errorResponse {
	return &errorResponse{
		messageBase{
			jsonRpcVersion,
		},
		id,
		nil,
		err,
	}
}

func NewErrorResponse[TId Id](id TId, err *ErrorObj) *errorResponse {
	return NewErrorResponseI((interface{})(id), err)
}
