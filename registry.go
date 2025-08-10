package jsonrpc2

import (
	"context"
)

type RpcMethod[TParam Params, TResult Result] func(ctx context.Context, p TParam) (TResult, *Error)
type RpcHandler func(ctx context.Context, rpcMessage *message) interface{}
type RpcMethodRegistry map[string]RpcHandler

func NewMethodRegistry() RpcMethodRegistry {
	return make(RpcMethodRegistry)
}

func getMethodHandler(reg RpcMethodRegistry, rpcMsg *message) (RpcHandler, *errorResponse) {
	if rpcMsg == nil {
		return nil, NewInvalidRequestWithData(ErrInternalInvalidJsonRpcMessage.Error()).ToResponse(nil)
	}
	if !rpcMsg.IsRequest() {
		return nil, NewInvalidRequestWithData(ErrInternalNotRequest.Error()).ToResponse(rpcMsg.Id)
	}
	handler, ok := reg[rpcMsg.Method]
	if !ok {
		return nil, NewMethodNotFound().ToResponse(rpcMsg.Id)
	}
	return handler, nil
}

func ProcessRpcRequest(ctx context.Context, reg RpcMethodRegistry, rpcMsg *message) interface{} {
	handler, errResponse := getMethodHandler(reg, rpcMsg)
	if errResponse != nil {
		return errResponse
	}
	return handler(ctx, rpcMsg)
}

func RegisterMethod[TParam Params, TResult Result](reg RpcMethodRegistry, method string, handler RpcMethod[TParam, TResult]) {
	reg[method] = func(ctx context.Context, rpcMsg *message) interface{} {
		request := MessageToRequest[TParam](rpcMsg)
		result, jsonRpcErr := handler(ctx, request.Params)
		if jsonRpcErr != nil {
			response := jsonRpcErr.ToResponse(request.Id)
			return response
		}
		response := NewSuccessResponseI(request.Id, result)
		return response
	}
}
