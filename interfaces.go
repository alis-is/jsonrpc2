package main

import (
	"log/slog"

	"github.com/alis-is/jsonrpc2/types"
)

type EndpointClient interface {
	WriteObject(obj interface{}) error
	RegisterPendingRequest(id interface{}) <-chan types.Message
	UnregisterPendingRequest(id interface{})
	Close() error
	IsClosed() bool
	UseLogger(logger *slog.Logger)
}

type EndpointServer interface {
	GetMethods() RpcMethodRegistry
	UseLogger(logger *slog.Logger)
}
