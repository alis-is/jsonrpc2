package endpoints

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alis-is/jsonrpc2/rpc"
	"github.com/stretchr/testify/assert"
)

func TestGetMethodHandler(t *testing.T) {
	assert := assert.New(t)
	reg := NewMethodRegistry()

	_, e := getMethodHandler(reg, nil)
	assert.Contains(string(*e.Error.Data), rpc.ErrInternalInvalidJsonRpcMessage.Error())
	_, e = getMethodHandler(reg, &rpc.Message{})
	assert.Contains(string(*e.Error.Data), rpc.ErrInternalNotRequest.Error())
}

func TestRegisterMethod(t *testing.T) {
	assert := assert.New(t)
	reg := NewMethodRegistry()

	RegisterMethod(reg, "test", func(ctx context.Context, p string) (string, *rpc.Error) {
		return "", rpc.NewInternalError()
	})
	handler, ok := reg["test"]
	assert.True(ok)
	var p json.RawMessage = []byte("\"test\"")
	r := handler(context.Background(), &rpc.Message{
		Params: &p,
	})
	_, ok = r.(*rpc.ErrorResponse)
	assert.True(ok)
}
