package endpoints

import (
	"fmt"

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

func (d *DefaultLogger) log(level string, f string, v ...interface{}) {
	if f[len(f)-1] != '\n' {
		f += "\n"
	}
	fmt.Printf(level+": "+f, v...)
}

func (d *DefaultLogger) Tracef(f string, v ...interface{}) {
	d.log("TRACE", f, v...)
}

func (d *DefaultLogger) Debugf(f string, v ...interface{}) {
	d.log("DEBUG", f, v...)
}

func (d *DefaultLogger) Infof(f string, v ...interface{}) {
	d.log("INFO", f, v...)
}

func (d *DefaultLogger) Warnf(f string, v ...interface{}) {
	d.log("WARN", f, v...)
}

func (d *DefaultLogger) Errorf(f string, v ...interface{}) {
	d.log("ERROR", f, v...)
}

func (d *DefaultLogger) Fatalf(f string, v ...interface{}) {
	d.log("FATAL", f, v...)
}
