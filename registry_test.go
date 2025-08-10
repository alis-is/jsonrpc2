package jsonrpc2

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMethodHandler(t *testing.T) {
	assert := assert.New(t)
	reg := NewMethodRegistry()

	_, e := getMethodHandler(reg, nil)
	assert.Contains(string(*e.Error.Data), ErrInternalInvalidJsonRpcMessage.Error())
	_, e = getMethodHandler(reg, &message{})
	assert.Contains(string(*e.Error.Data), ErrInternalNotRequest.Error())
}

func TestRegisterMethod(t *testing.T) {
	assert := assert.New(t)
	reg := NewMethodRegistry()

	RegisterMethod(reg, "test", func(ctx context.Context, p string) (string, *Error) {
		return "", NewInternalError()
	})
	handler, ok := reg["test"]
	assert.True(ok)
	var p json.RawMessage = []byte("\"test\"")
	r := handler(context.Background(), &message{
		Params: p,
	})
	_, ok = r.(*errorResponse)
	assert.True(ok)
}
