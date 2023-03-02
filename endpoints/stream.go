package endpoints

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/alis-is/jsonrpc2/rpc"
)

var (
	ErrInvalidEndpoint = errors.New("invalid endpoint")
	ErrStreamClosed    = errors.New("stream closed")
)

// StreamEndpoint is a endpoint that implements both client and server side of jsonrpc over a stream.
// Usually used over tcp or stdio streams.
type StreamEndpoint struct {
	stream ObjectStream

	pendingMutex sync.Mutex
	closed       bool
	pending      map[interface{}]chan rpc.Message

	writeMutex sync.Mutex

	closeNotify chan struct{}

	logger ILogger
	// Set by ConnOpt funcs.
	methodRegistry RpcMethodRegistry
}

func NewStreamEndpoint(ctx context.Context, stream ObjectStream) *StreamEndpoint {
	c := &StreamEndpoint{
		stream:         stream,
		pending:        make(map[interface{}]chan rpc.Message, 1),
		closeNotify:    make(chan struct{}),
		methodRegistry: NewMethodRegistry(),
		logger:         &DefaultLogger{},
	}
	go c.readMessages(ctx)
	return c
}

// returns a channel that will be closed when the connection is closed
func (c *StreamEndpoint) GetOnCloseListener() <-chan struct{} {
	return c.closeNotify
}

func (c *StreamEndpoint) close(cause error) error {
	c.writeMutex.Lock()
	c.pendingMutex.Lock()
	defer c.writeMutex.Unlock()
	defer c.pendingMutex.Unlock()

	if c.closed {
		return ErrStreamClosed
	}

	for _, pendingChannel := range c.pending {
		close(pendingChannel)
	}

	if cause != nil && cause != io.EOF && cause != io.ErrUnexpectedEOF {
		c.logger.Tracef("stream closing, reason: %v\n", cause)
	}

	close(c.closeNotify)
	c.closed = true
	return c.stream.Close()
}

func (c *StreamEndpoint) Close() error {
	return c.close(nil)
}

func (c *StreamEndpoint) GetMethods() RpcMethodRegistry {
	return c.methodRegistry
}

func (c *StreamEndpoint) ListMethods() []string {
	methods := make([]string, 0, len(c.methodRegistry))
	for method := range c.methodRegistry {
		methods = append(methods, method)
	}
	return methods
}

func (c *StreamEndpoint) UseLogger(logger ILogger) {
	if logger == nil {
		c.logger.Tracef("ignored nil logger")
		return
	}
	c.logger = logger
}

func (c *StreamEndpoint) readMessages(ctx context.Context) {
	var err error
	for err == nil {
		if ctx.Err() != nil {
			c.logger.Tracef("jsonrpc2: context closed")
			break
		}
		var rpcObj rpc.Object
		err = c.stream.ReadObject(&rpcObj)
		if err != nil {
			c.logger.Tracef("jsonrpc2: error reading message: %s", err.Error())
			break
		}
		c.logger.Tracef(fmt.Sprintf("jsonrpc2: received message: %v", rpcObj))
		go func() {
			messages := rpcObj.GetMessages()
			results := make([]interface{}, 0, len(messages))
			for _, rpcMsg := range messages {
				kind, err := rpcMsg.GetKind()
				switch kind {
				case rpc.REQUEST_KIND:
					results = append(results, ProcessRpcRequest(ctx, c.methodRegistry, &rpcMsg))
				case rpc.NOTIFICATION_KIND:
					_ = ProcessRpcRequest(ctx, c.methodRegistry, &rpcMsg)
				case rpc.SUCCESS_RESPONSE_KIND:
					fallthrough
				case rpc.ERROR_RESPONSE_KIND:
					pendingChannel, ok := c.pending[rpcMsg.Id]
					if !ok {
						c.logger.Tracef("jsonrpc2: ignoring response #%s with no corresponding request", rpcMsg.Id)
						continue
					}
					pendingChannel <- rpcMsg
				default:
					// ignore invalid messages to prevent DoS
					c.logger.Tracef("jsonrpc2: ignoring invalid message: %v", err)
					continue
				}
			}

			if len(results) == 0 {
				return
			}

			c.writeMutex.Lock()
			defer c.writeMutex.Unlock()
			if rpcObj.IsBatch() {
				c.logger.Tracef(fmt.Sprintf("jsonrpc2: sending batch response: %v", results))
				c.stream.WriteObject(results)
				return
			}
			c.logger.Tracef(fmt.Sprintf("jsonrpc2: sending response: %v", results[0]))
			c.stream.WriteObject(results[0])
		}()
	}
	c.close(err)
}

func (c *StreamEndpoint) WriteObject(obj interface{}) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	return c.stream.WriteObject(obj)
}

func (c *StreamEndpoint) RegisterPendingRequest(requestId interface{}) <-chan rpc.Message {
	ch := make(chan rpc.Message, 1)
	c.pendingMutex.Lock()
	c.pending[requestId] = ch
	c.pendingMutex.Unlock()
	return ch
}

func (c *StreamEndpoint) UnregisterPendingRequest(requestId interface{}) {
	c.pendingMutex.Lock()
	delete(c.pending, requestId)
	c.pendingMutex.Unlock()
}

func (c *StreamEndpoint) IsClosed() bool {
	return c.closed
}
