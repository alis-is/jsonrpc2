package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alis-is/jsonrpc2/types"
	"github.com/stretchr/testify/assert"
)

func TestGetMethodHandler(t *testing.T) {
	assert := assert.New(t)
	reg := NewMethodRegistry()

	_, e := getMethodHandler(reg, nil)
	assert.Contains(string(*e.Error.Data), types.ErrInternalInvalidJsonRpcMessage.Error())
	_, e = getMethodHandler(reg, &types.Message{})
	assert.Contains(string(*e.Error.Data), types.ErrInternalNotRequest.Error())
}

func TestRegisterMethod(t *testing.T) {
	assert := assert.New(t)
	reg := NewMethodRegistry()

	RegisterMethod(reg, "test", func(ctx context.Context, p string) (string, *types.Error) {
		return "", types.NewInternalError()
	})
	handler, ok := reg["test"]
	assert.True(ok)
	var p json.RawMessage = []byte("\"test\"")
	r := handler(context.Background(), &types.Message{
		Params: p,
	})
	_, ok = r.(*types.ErrorResponse)
	assert.True(ok)
}
