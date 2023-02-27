package endpoints

import (
	"github.com/alis-is/jsonrpc2/rpc"
)

type IEndpointClient interface {
	WriteObject(obj interface{}) error
	RegisterPendingRequest(id interface{}) <-chan rpc.Message
	UnregisterPendingRequest(id interface{})
	Close() error
	IsClosed() bool
	UseLogger(logger ILogger)
}

type IEndpointServer interface {
	GetMethods() RpcMethodRegistry
	UseLogger(logger ILogger)
}

type ILogger interface {
	Tracef(f string, v ...interface{})
	Debugf(f string, v ...interface{})
	Infof(f string, v ...interface{})
	Warnf(f string, v ...interface{})
	Errorf(f string, v ...interface{})
	Fatalf(f string, v ...interface{})
}

type DefaultLogger struct {
}

func (d *DefaultLogger) Tracef(f string, v ...interface{}) {
}

func (d *DefaultLogger) Debugf(f string, v ...interface{}) {
}

func (d *DefaultLogger) Infof(f string, v ...interface{}) {
}

func (d *DefaultLogger) Warnf(f string, v ...interface{}) {
}

func (d *DefaultLogger) Errorf(f string, v ...interface{}) {
}

func (d *DefaultLogger) Fatalf(f string, v ...interface{}) {
}
