package endpoints

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/alis-is/jsonrpc2/rpc"
	"github.com/alis-is/jsonrpc2/test"
	"github.com/stretchr/testify/assert"
)

func getWritableSideOfPipe() net.Conn {
	connA, connB := net.Pipe()
	go func() {
		b := make([]byte, 1024)
		for _, ok := connA.Read(b); ok == nil; {
		}
	}()
	return connB
}

func TestNewStreamEndpoint(t *testing.T) {
	assert := assert.New(t)
	conn := getWritableSideOfPipe()
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(conn))
	assert.NotNil(c)
	assert.Nil(c.WriteObject("test"))
}

func TestStreamClose(t *testing.T) {
	assert := assert.New(t)
	conn := getWritableSideOfPipe()
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(conn))
	assert.NotNil(c)
	c.Close()
	assert.True(c.IsClosed())
	assert.NotNil(c.WriteObject("test"))
}

func TestStreamClose2(t *testing.T) {
	assert := assert.New(t)
	conn := getWritableSideOfPipe()
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(conn))
	assert.NotNil(c)
	assert.Nil(c.Close())
	assert.EqualError(c.Close(), ErrStreamClosed.Error())
}

func TestStreamcloseWithCause(t *testing.T) {
	assert := assert.New(t)
	conn := getWritableSideOfPipe()
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(conn))
	testLogger := test.NewLogger()
	c.UseLogger(testLogger)
	c.close(fmt.Errorf("testCause"))
	assert.Contains(testLogger.LastLog(), "stream closing, reason")
	assert.Contains(testLogger.LastLog(), "testCause")
}

func TestStreamClosePending(t *testing.T) {
	assert := assert.New(t)
	conn := getWritableSideOfPipe()
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(conn))
	ch := c.RegisterPendingRequest("testRequest")
	c.Close()
	assert.True(c.IsClosed())
	_, ok := <-ch
	assert.False(ok)
}

func TestStreamCloseNotify(t *testing.T) {
	assert := assert.New(t)
	conn := getWritableSideOfPipe()
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(conn))
	assert.NotNil(c)
	c.Close()
	select {
	case <-c.GetOnCloseListener():
	// case timeout 5s
	case <-time.After(5 * time.Second):
		assert.Fail("no disconnect notification")
	}
}

func TestStreamUseLogger(t *testing.T) {
	assert := assert.New(t)
	conn := getWritableSideOfPipe()
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(conn))
	assert.NotNil(c)
	c.UseLogger(nil)
	assert.NotNil(c.logger)
	oldLogger := c.logger
	l := test.NewLogger()
	c.UseLogger(l)
	assert.NotEqual(oldLogger, c.logger)
}

func TestStreamListMethods(t *testing.T) {
	assert := assert.New(t)
	conn := getWritableSideOfPipe()
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(conn))
	assert.Len(c.ListMethods(), 0)
	RegisterEndpointMethod(c, "test", func(ctx context.Context, data string) (string, *rpc.Error) {
		return "hello " + data, nil
	})
	assert.Len(c.ListMethods(), 1)
}

func TestStreamRequestNonExistingMethod(t *testing.T) {
	assert := assert.New(t)
	connA, connB := net.Pipe()
	s := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connA))
	testLogger := test.NewLogger()
	s.UseLogger(testLogger)
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connB))

	response, err := Request[string, string](context.Background(), c, "test", "test data")
	assert.Nil(err)
	assert.NotNil(response.Error)

	_, err = response.Unwrap()
	assert.Contains(err.Error(), "Method not found")
}

func TestStreamRequest(t *testing.T) {
	assert := assert.New(t)
	connA, connB := net.Pipe()
	s := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connA))
	testLogger := test.NewLogger()
	s.UseLogger(testLogger)
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connB))
	RegisterEndpointMethod(s, "test", func(ctx context.Context, data string) (string, *rpc.Error) {
		return "hello " + data, nil
	})

	response, err := Request[string, string](context.Background(), c, "test", "world")
	assert.Nil(err)
	assert.Nil(response.Error)

	res, err := response.Unwrap()
	assert.Nil(err)
	assert.Equal(res, "hello world")
}

func TestStreamNotifyNonExistingMethod(t *testing.T) {
	assert := assert.New(t)
	connA, connB := net.Pipe()
	s := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connA))
	testLogger := test.NewLogger()
	s.UseLogger(testLogger)
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connB))

	err := Notify(context.Background(), c, "test", "test data")
	assert.Nil(err)
}

func TestStreamNotify(t *testing.T) {
	assert := assert.New(t)
	connA, connB := net.Pipe()
	s := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connA))
	testLogger := test.NewLogger()
	s.UseLogger(testLogger)
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connB))

	signaled := make(chan bool, 1)
	RegisterEndpointMethod(s, "test", func(ctx context.Context, data string) (string, *rpc.Error) {
		if data == "world" {
			signaled <- true
		}
		return "hello " + data, nil
	})

	err := Notify(context.Background(), c, "test", "world")
	assert.Nil(err)
	select {
	case <-signaled:
	case <-time.After(5 * time.Second):
		assert.Fail("no notification")
	}
}

func TestStreamBatchNonExistingMethod(t *testing.T) {
	assert := assert.New(t)
	connA, connB := net.Pipe()
	s := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connA))
	testLogger := test.NewLogger()
	s.UseLogger(testLogger)
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connB))

	requests := []RequestInfo[string]{
		{Method: "test", Params: "world"},
		{Method: "test", Params: "universe"},
	}

	responses, err := Batch[string, string](context.Background(), c, requests)
	assert.Nil(err)
	for _, response := range responses {
		assert.NotNil(response.Error)
		_, err = response.Unwrap()
		assert.Contains(err.Error(), "Method not found")
	}
}

func TestStreamBatch(t *testing.T) {
	assert := assert.New(t)
	connA, connB := net.Pipe()
	s := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connA))
	testLogger := test.NewLogger()
	s.UseLogger(testLogger)
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connB))
	RegisterEndpointMethod(s, "test", func(ctx context.Context, data string) (string, *rpc.Error) {
		return "hello " + data, nil
	})

	requests := []RequestInfo[string]{
		{Method: "test", Params: "world"},
		{Method: "test", Params: "universe"},
	}
	responses, err := Batch[string, string](context.Background(), c, requests)
	assert.Nil(err)
	for _, response := range responses {
		assert.Nil(response.Error)
	}

	res, err := responses[0].Unwrap()
	assert.Nil(err)
	assert.Equal(res, "hello world")
	res, err = responses[1].Unwrap()
	assert.Nil(err)
	assert.Equal(res, "hello universe")
}

func TestReadMessagesInvalidMessage(t *testing.T) {
	assert := assert.New(t)
	connA, connB := net.Pipe()
	s := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connA))
	testLogger := test.NewLogger()
	s.UseLogger(testLogger)
	connB.Write([]byte("{ \"id\": 1, \"method\": \"test\", \"params\": \"hello world\", \"error\": {}, \"result\": \"hello world\" }"))
	connB.Close()
	time.Sleep(2 * time.Second) // wait for logs to be written
	var lastLog string
	func() {
		// skip to last element in channel
		for {
			select {
			case lastLog = <-testLogger.LogChannel():
			default:
				return
			}
		}
	}()
	assert.Contains(lastLog, "ignoring invalid message")
}

func TestReadMessagesContextClosed(t *testing.T) {
	assert := assert.New(t)
	connA, connB := net.Pipe()
	ctx, cancelCtx := context.WithCancel(context.Background())
	s := NewStreamEndpoint(ctx, NewPlainObjectStream(connA))
	testLogger := test.NewLogger()
	s.UseLogger(testLogger)
	cancelCtx()
	connB.Write([]byte("hello world"))

	assert.True(s.IsClosed())
}

func TestStreamRequestNotRegisteredAsPending(t *testing.T) {
	assert := assert.New(t)
	connA, connB := net.Pipe()
	s := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connA))
	testLogger := test.NewLogger()
	c := NewStreamEndpoint(context.Background(), NewPlainObjectStream(connB))
	c.UseLogger(testLogger)
	RegisterEndpointMethod(s, "test", func(ctx context.Context, data string) (string, *rpc.Error) {
		return "hello " + data, nil
	})

	c.WriteObject(rpc.NewRequest("testId", "test", "world"))
	time.Sleep(2 * time.Second) // wait for logs to be written
	var lastLog string
	func() {
		// skip to last element in channel
		for {
			select {
			case lastLog = <-testLogger.LogChannel():
			default:
				return
			}
		}
	}()
	assert.Contains(lastLog, "ignoring response")
}
