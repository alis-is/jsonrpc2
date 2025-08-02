package jsonrpc2

import (
	"context"

	types "github.com/alis-is/jsonrpc2/types"
)

type RpcMethod[TParam types.ParamsType, TResult types.ResultType] func(ctx context.Context, p TParam) (TResult, *types.Error)
type RpcHandler func(ctx context.Context, rpcMessage *types.Message) interface{}
type RpcMethodRegistry map[string]RpcHandler

func NewMethodRegistry() RpcMethodRegistry {
	return make(RpcMethodRegistry)
}

func getMethodHandler(reg RpcMethodRegistry, rpcMsg *types.Message) (RpcHandler, *types.ErrorResponse) {
	if rpcMsg == nil {
		return nil, types.NewInvalidRequestWithData(types.ErrInternalInvalidJsonRpcMessage.Error()).ToResponse(nil)
	}
	if !rpcMsg.IsRequest() {
		return nil, types.NewInvalidRequestWithData(types.ErrInternalNotRequest.Error()).ToResponse(rpcMsg.Id)
	}
	handler, ok := reg[rpcMsg.Method]
	if !ok {
		return nil, types.NewMethodNotFound().ToResponse(rpcMsg.Id)
	}
	return handler, nil
}

func ProcessRpcRequest(ctx context.Context, reg RpcMethodRegistry, rpcMsg *types.Message) interface{} {
	handler, errResponse := getMethodHandler(reg, rpcMsg)
	if errResponse != nil {
		return errResponse
	}
	return handler(ctx, rpcMsg)
}

func RegisterMethod[TParam types.ParamsType, TResult types.ResultType](reg RpcMethodRegistry, method string, handler RpcMethod[TParam, TResult]) {
	reg[method] = func(ctx context.Context, rpcMsg *types.Message) interface{} {
		request := types.MessageToRequest[TParam](rpcMsg)
		result, jsonRpcErr := handler(ctx, request.Params)
		if jsonRpcErr != nil {
			response := jsonRpcErr.ToResponse(request.Id)
			return response
		}
		response := types.NewSuccessResponseI(request.Id, result)
		return response
	}
}
