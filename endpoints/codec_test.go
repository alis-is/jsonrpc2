package endpoints

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferedStreamWithPlainCodec(t *testing.T) {
	assert := assert.New(t)
	connA, connB := net.Pipe()
	a := NewBufferedStream(connA, PlainObjectCodec{})
	b := NewBufferedStream(connB, PlainObjectCodec{})

	go a.WriteObject("test")
	var bObj string
	assert.Nil(b.ReadObject(&bObj))
	assert.Equal("test", bObj)

	assert.NotNil(a.WriteObject(func() {}))
}
func TestVSCodeObjectCodec(t *testing.T) {
	assert := assert.New(t)
	connA, connB := net.Pipe()
	a := NewBufferedStream(connA, VSCodeObjectCodec{}).(*bufferedObjectStream)
	b := NewBufferedStream(connB, VSCodeObjectCodec{}).(*bufferedObjectStream)

	go a.WriteObject("test")
	var bObj string
	assert.Nil(b.ReadObject(&bObj))
	assert.Equal("test", bObj)

	assert.NotNil(a.WriteObject(func() {}))

	go func() {
		data := []byte("\"abcd\"")
		fmt.Fprintf(a.w, "Content-Length: %d\r\r\n", len(data))
		a.w.Flush()
	}()

	bObj = ""
	assert.ErrorContains(b.ReadObject(&bObj), "line endings must be")
	assert.Equal("", bObj)

	go func() {
		a.Close()
	}()
	assert.ErrorContains(b.ReadObject(&bObj), "EOF")
	assert.ErrorContains(b.WriteObject("test"), "closed pipe")
	assert.Equal("", bObj)
}

func TestVarintObjectCodec(t *testing.T) {
	assert := assert.New(t)
	connA, connB := net.Pipe()
	a := NewBufferedStream(connA, VarintObjectCodec{}).(*bufferedObjectStream)
	b := NewBufferedStream(connB, VarintObjectCodec{}).(*bufferedObjectStream)

	go a.WriteObject("test")
	var bObj string
	assert.Nil(b.ReadObject(&bObj))
	assert.Equal("test", bObj)

	assert.NotNil(a.WriteObject(func() {}))

	bObj = ""
	go func() {
		a.Close()
	}()
	assert.ErrorContains(b.ReadObject(&bObj), "EOF")
	assert.ErrorContains(b.WriteObject("test"), "closed pipe")
	assert.Equal("", bObj)
}
