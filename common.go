package jsonrpc2

import (
	"context"

	"github.com/google/uuid"
)

// register method to server endpoint
func RegisterEndpointMethod[TParam Params, TResult Result](c EndpointServer, method string, handler RpcMethod[TParam, TResult]) {
	if c == nil {
		return
	}
	RegisterMethod(c.GetMethods(), method, handler)
}

// request
func Request[TParams Params, TResult Result](ctx context.Context, c EndpointClient, method string, params TParams) (*Response[TResult], error) {
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

	request := NewRequest(requestId, method, params)
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

		response, err := MessageToResponse[TResult](&responseMsg)
		return response, err
	}
}

func RequestTo[TParams Params, TResult Result](ctx context.Context, c EndpointClient, method string, params TParams, result *Response[TResult]) error {
	response, err := Request[TParams, TResult](ctx, c, method, params)
	if err != nil {
		return err
	}
	*result = *response
	return nil
}

// notify
func Notify[TParams Params](ctx context.Context, c EndpointClient, method string, params TParams) error {
	if c == nil {
		return ErrInvalidEndpoint
	}
	if c.IsClosed() {
		return ErrStreamClosed
	}
	notification := NewNotification(method, params)
	if err := c.WriteObject(notification); err != nil {
		if err == ErrEmptyResponse { // we do not expect response here
			return nil
		}
		return err
	}
	return nil
}

type RequestInfo[TParams Params] struct {
	Method         string
	Params         TParams
	IsNotification bool
}

// Batch
func Batch[TParams Params, TResult Result](ctx context.Context, c EndpointClient, requests []RequestInfo[TParams]) ([]*Response[TResult], error) {
	if c == nil {
		return nil, ErrInvalidEndpoint
	}
	rpcRequests := make([]*request[TParams], 0, len(requests))
	resultChannels := make([]<-chan message, 0, len(requests))
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
			rpcRequests = append(rpcRequests, NewNotification(request.Method, request.Params))
			continue
		}
		resultChannels = append(resultChannels, c.RegisterPendingRequest(requestId))
		defer c.UnregisterPendingRequest(requestId)

		rpcRequests = append(rpcRequests, NewRequest(requestId, request.Method, request.Params))
	}

	if err := c.WriteObject(rpcRequests); err != nil {
		return nil, err
	}
	results := make([]*Response[TResult], 0, len(requests))
	for _, ch := range resultChannels {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case responseMsg, ok := <-ch:
			if !ok {
				return nil, ErrStreamClosed
			}

			response, err := MessageToResponse[TResult](&responseMsg)
			if err != nil {
				return nil, err
			}
			results = append(results, response)
		}
	}
	return results, nil
}

func BatchTo[TParams Params, TResult Result](ctx context.Context, c EndpointClient, requests []RequestInfo[TParams], results []*Response[TResult]) error {
	r, err := Batch[TParams, TResult](ctx, c, requests)
	if err != nil {
		return err
	}
	copy(results, r)
	return nil
}
