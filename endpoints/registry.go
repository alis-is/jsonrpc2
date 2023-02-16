package endpoints

import (
	"context"

	"github.com/alis-is/jsonrpc2/rpc"
)

type RpcMethod[TParam rpc.ParamsType, TResult rpc.ResultType] func(ctx context.Context, p TParam) (TResult, *rpc.Error)
type RpcHandler func(ctx context.Context, rpcMessage *rpc.Message) interface{}
type RpcMethodRegistry map[string]RpcHandler

func NewMethodRegistry() RpcMethodRegistry {
	return make(RpcMethodRegistry)
}

func getMethodHandler(reg RpcMethodRegistry, rpcMsg *rpc.Message) (RpcHandler, *rpc.ErrorResponse) {
	if rpcMsg == nil {
		return nil, rpc.NewInvalidRequestWithData(rpc.ErrInternalInvalidJsonRpcMessage.Error()).ToResponse(nil)
	}
	if !rpcMsg.IsRequest() {
		return nil, rpc.NewInvalidRequestWithData(rpc.ErrInternalNotRequest.Error()).ToResponse(rpcMsg.Id)
	}
	handler, ok := reg[*rpcMsg.Method]
	if !ok {
		return nil, rpc.NewMethodNotFound().ToResponse(rpcMsg.Id)
	}
	return handler, nil
}

func ProcessRpcRequest(ctx context.Context, reg RpcMethodRegistry, rpcMsg *rpc.Message) interface{} {
	handler, errResponse := getMethodHandler(reg, rpcMsg)
	if errResponse != nil {
		return errResponse
	}
	return handler(ctx, rpcMsg)
}

func RegisterMethod[TParam rpc.ParamsType, TResult rpc.ResultType](reg RpcMethodRegistry, method string, handler RpcMethod[TParam, TResult]) {
	reg[method] = func(ctx context.Context, rpcMsg *rpc.Message) interface{} {
		request := rpc.MessageToRequest[TParam](rpcMsg)
		result, jsonRpcErr := handler(ctx, request.Params)
		if jsonRpcErr != nil {
			response := jsonRpcErr.ToResponse(request.Id)
			return response
		}
		response := rpc.NewSuccessResponseI(request.Id, result)
		return response
	}
}
