package endpoints

import (
	"context"

	"github.com/alis-is/jsonrpc2/rpc"
	"github.com/google/uuid"
)

// register method to server endpoint
func RegisterEndpointMethod[TParam rpc.ParamsType, TResult rpc.ResultType](c IEndpointServer, method string, handler RpcMethod[TParam, TResult]) {
	if c == nil {
		return
	}
	RegisterMethod(c.GetMethods(), method, handler)
}

// request
func Request[TParams rpc.ParamsType, TResult rpc.ResultType](ctx context.Context, c IEndpointClient, method string, params TParams) (*rpc.Response[TResult], error) {
	if c == nil {
		return nil, ErrInvalidEndpoint
	}
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	requestId := uuid.String()
	if c.IsClosed() {
		return nil, ErrStreamClosed
	}

	ch := c.RegisterPendingRequest(requestId)
	defer c.UnregisterPendingRequest(requestId)

	request := rpc.NewRequest(requestId, method, params)
	if err := c.WriteObject(request); err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case responseMsg, ok := <-ch:
		if !ok {
			return nil, ErrStreamClosed
		}

		response, err := rpc.MessageToResponse[TResult](&responseMsg)
		return response, err
	}
}

func RequestTo[TParams rpc.ParamsType, TResult rpc.ResultType](ctx context.Context, c IEndpointClient, method string, params TParams, result *rpc.Response[TResult]) error {
	response, err := Request[TParams, TResult](ctx, c, method, params)
	if err != nil {
		return err
	}
	*result = *response
	return nil
}

// notify
func Notify[TParams rpc.ParamsType](ctx context.Context, c IEndpointClient, method string, params TParams) error {
	if c == nil {
		return ErrInvalidEndpoint
	}
	if c.IsClosed() {
		return ErrStreamClosed
	}
	notification := rpc.NewNotification(method, params)
	if err := c.WriteObject(notification); err != nil {
		if err == rpc.ErrEmptyResponse { // we do not expect response here
			return nil
		}
		return err
	}
	return nil
}

type RequestInfo[TParams rpc.ParamsType] struct {
	Method         string
	Params         TParams
	IsNotification bool
}

// Batch
func Batch[TParams rpc.ParamsType, TResult rpc.ResultType](ctx context.Context, c IEndpointClient, requests []RequestInfo[TParams]) ([]*rpc.Response[TResult], error) {
	if c == nil {
		return nil, ErrInvalidEndpoint
	}
	rpcRequests := make([]*rpc.Request[TParams], 0, len(requests))
	resultChannels := make([]<-chan rpc.Message, 0, len(requests))
	for _, request := range requests {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
		requestId := uuid.String()
		if c.IsClosed() {
			return nil, ErrStreamClosed
		}
		if request.IsNotification {
			rpcRequests = append(rpcRequests, rpc.NewNotification(request.Method, request.Params))
			continue
		}
		resultChannels = append(resultChannels, c.RegisterPendingRequest(requestId))
		defer c.UnregisterPendingRequest(requestId)

		rpcRequests = append(rpcRequests, rpc.NewRequest(requestId, request.Method, request.Params))
	}

	if err := c.WriteObject(rpcRequests); err != nil {
		return nil, err
	}
	results := make([]*rpc.Response[TResult], 0, len(requests))
	for _, ch := range resultChannels {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case responseMsg, ok := <-ch:
			if !ok {
				return nil, ErrStreamClosed
			}

			response, err := rpc.MessageToResponse[TResult](&responseMsg)
			if err != nil {
				return nil, err
			}
			results = append(results, response)
		}
	}
	return results, nil
}

func BatchTo[TParams rpc.ParamsType, TResult rpc.ResultType](ctx context.Context, c IEndpointClient, requests []RequestInfo[TParams], results []*rpc.Response[TResult]) error {
	r, err := Batch[TParams, TResult](ctx, c, requests)
	if err != nil {
		return err
	}
	copy(results, r)
	return nil
}
