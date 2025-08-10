package jsonrpc2

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/alis-is/jsonrpc2/test"
	"github.com/stretchr/testify/assert"
)

func createHttpServer(mux *ServerMux, port int) *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port)}
	srv.Handler = mux
	return srv
}

func TestHttpClientUseLogger(t *testing.T) {
	assert := assert.New(t)
	c := NewHttpClientEndpoint("127.0.0.1", nil)
	assert.NotNil(c)
	c.UseLogger(nil)
	assert.NotNil(c.logger)
	oldLogger := c.logger
	l := test.NewLogger()
	c.UseLogger(l.Logger)
	assert.NotEqual(oldLogger, c.logger)
}

func TestHttpClientClose(t *testing.T) {
	assert := assert.New(t)
	c := NewHttpClientEndpoint("127.0.0.1", nil)
	assert.NotNil(c)
	c.Close()
	assert.False(c.IsClosed())
}

func TestServerMuxUseLogger(t *testing.T) {
	assert := assert.New(t)
	c := NewServerMux()
	assert.NotNil(c)
	c.UseLogger(nil)
	assert.NotNil(c.logger)
	oldLogger := c.logger
	l := test.NewLogger()
	c.UseLogger(l.Logger)
	assert.NotEqual(oldLogger, c.logger)
}

func TestServerMuxGetEndpoints(t *testing.T) {
	assert := assert.New(t)
	c := NewServerMux()
	assert.NotNil(c)
	assert.Equal(1, len(c.GetEndpoints()))
	c.RegisterEndpoint("/test")
	assert.Equal(2, len(c.GetEndpoints()))
}

func TestServerMuxGetMethods(t *testing.T) {
	assert := assert.New(t)
	c := NewServerMux()
	assert.NotNil(c)
	assert.Equal(0, len(c.GetMethods()))
	c.RegisterEndpoint("/test")
	assert.Equal(0, len(c.GetMethods()))
	RegisterEndpointMethod(c, "test", func(ctx context.Context, name string) (string, *Error) {
		return "Hello " + name, nil
	})
	assert.Equal(1, len(c.GetMethods()))
}

func TestHttpRequest(t *testing.T) {
	assert := assert.New(t)
	c1 := NewHttpClientEndpoint("http://127.0.0.1:8080/hello", nil)
	assert.NotNil(c1)
	c2 := NewHttpClientEndpoint("http://127.0.0.1:8080/bye", nil)
	assert.NotNil(c2)

	mux := NewServerMux()
	s := createHttpServer(mux, 8080)
	RegisterServerMuxEndpointMethod(mux, "/bye", "bye", func(ctx context.Context, name string) (string, *Error) {
		return "Bye " + name, nil
	})
	RegisterServerMuxEndpointMethod(mux, "/hello", "hello", func(ctx context.Context, name string) (string, *Error) {
		return "Hello " + name, nil
	})
	go s.ListenAndServe()
	defer s.Close()
	time.Sleep(1 * time.Second)
	r, e := Request[string, string](context.Background(), c1, "hello", "World")
	assert.Nil(e)
	result, e := r.Unwrap()
	assert.Nil(e)
	assert.Equal("Hello World", result)
	r, e = Request[string, string](context.Background(), c2, "bye", "World")
	assert.Nil(e)
	result, e = r.Unwrap()
	assert.Nil(e)
	assert.Equal("Bye World", result)
}

// TODO: fix this test, timeouts in github
func TestHttpBatchRequest(t *testing.T) {
	assert := assert.New(t)
	c1 := NewHttpClientEndpoint("http://127.0.0.1:8081/hello", nil)
	assert.NotNil(c1)
	c2 := NewHttpClientEndpoint("http://127.0.0.1:8081/bye", nil)
	assert.NotNil(c2)

	mux := NewServerMux()
	s := createHttpServer(mux, 8081)
	RegisterServerMuxEndpointMethod(mux, "/", "bye", func(ctx context.Context, name string) (string, *Error) {
		return "Bye " + name, nil
	})
	RegisterServerMuxEndpointMethod(mux, "/", "hello", func(ctx context.Context, name string) (string, *Error) {
		return "Hello " + name, nil
	})
	go s.ListenAndServe()
	defer s.Close()
	time.Sleep(3 * time.Second)
	rs, e := Batch[string, string](context.Background(), c1, []RequestInfo[string]{
		{"hello", "World", false},
		{"bye", "World", false},
	})
	assert.Nil(e)
	assert.Equal(2, len(rs))
	result, e := rs[0].Unwrap()
	assert.Nil(e)
	assert.Equal("Hello World", result)
	result, e = rs[1].Unwrap()
	assert.Nil(e)
	assert.Equal("Bye World", result)
}

func TestHttpNotification(t *testing.T) {
	assert := assert.New(t)
	c1 := NewHttpClientEndpoint("http://127.0.0.1:8082/hello", nil)
	assert.NotNil(c1)
	c2 := NewHttpClientEndpoint("http://127.0.0.1:8082/bye", nil)
	assert.NotNil(c2)

	mux := NewServerMux()
	s := createHttpServer(mux, 8082)
	signaled := make(chan bool, 1)
	signaled2 := make(chan bool, 1)
	RegisterServerMuxEndpointMethod(mux, "/bye", "bye", func(ctx context.Context, name string) (string, *Error) {
		signaled <- true
		return "Bye " + name, nil
	})
	RegisterServerMuxEndpointMethod(mux, "/hello", "hello", func(ctx context.Context, name string) (string, *Error) {
		signaled2 <- true
		return "Hello " + name, nil
	})
	go s.ListenAndServe()
	defer s.Close()
	time.Sleep(1 * time.Second)
	e := Notify(context.Background(), c1, "hello", "World")
	assert.Nil(e)
	e = Notify(context.Background(), c2, "bye", "World")
	assert.Nil(e)
	select {
	case <-signaled:
	case <-time.After(5 * time.Second):
		assert.Fail("Timeout")
	}
	select {
	case <-signaled2:
	case <-time.After(5 * time.Second):
		assert.Fail("Timeout")
	}
}

func TestSendServerResponse(t *testing.T) {
	assert := assert.New(t)
	c1 := NewHttpClientEndpoint("http://127.0.0.1:8083/hello", nil)
	assert.NotNil(c1)
	mux := NewServerMux()
	testLogger := test.NewLogger()
	mux.UseLogger(testLogger.Logger)
	s := createHttpServer(mux, 8083)
	go s.ListenAndServe()
	defer s.Close()

	time.Sleep(2 * time.Second)

	c1.WriteObject(NewSuccessResponse("test", "test data"))
	assert.Contains(<-testLogger.LogChannel(), "ignoring response")
	c1.WriteObject(NewUnknownError().ToResponse("test"))
	assert.Contains(<-testLogger.LogChannel(), "ignoring response")
}
