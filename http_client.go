package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/alis-is/jsonrpc2/types"
)

type HttpClientEndpoint struct {
	*http.Client

	pendingMutex sync.Mutex
	pending      map[interface{}]chan types.Message

	url    string
	logger *slog.Logger
}

func NewHttpClientEndpoint(baseUrl string, client *http.Client) *HttpClientEndpoint {
	if client == nil {
		client = http.DefaultClient
	}

	return &HttpClientEndpoint{
		Client:       client,
		url:          baseUrl,
		pendingMutex: sync.Mutex{},
		pending:      make(map[interface{}]chan types.Message, 1),
		logger:       slog.Default(),
	}
}

func (c *HttpClientEndpoint) UseLogger(logger *slog.Logger) {
	if logger == nil {
		c.logger.Debug("ignored nil logger")
		return
	}
	c.logger = logger
}

func (c *HttpClientEndpoint) Close() error {
	return nil
}

func (c *HttpClientEndpoint) IsClosed() bool {
	return false
}

func (c *HttpClientEndpoint) RegisterPendingRequest(requestID interface{}) <-chan types.Message {
	responseChan := make(chan types.Message, 1)
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()
	c.pending[requestID] = responseChan
	return responseChan
}

func (c *HttpClientEndpoint) UnregisterPendingRequest(requestID interface{}) {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()
	delete(c.pending, requestID)
}

func (c *HttpClientEndpoint) WriteObject(object interface{}) error {
	requestBody, err := json.Marshal(object)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.url, strings.NewReader(string(requestBody)))
	if err != nil {
		return err
	}
	c.logger.Debug("sending request", "to", c.url, "request", string(requestBody))
	req.Header.Add("Content-Type", "application/json")

	response, err := c.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return types.ErrEmptyResponse
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("jsonrpc2: http error: %s", string(body))
	}

	var rpcObj types.Object
	err = json.Unmarshal(body, &rpcObj)
	c.logger.Debug("jsonrpc2: received message", "message", string(body))
	if err != nil {
		return err
	}
	messages := rpcObj.GetMessages()
	for _, rpcMsg := range messages {
		kind, err := rpcMsg.GetKind()
		switch kind {
		case types.REQUEST_KIND:
			return fmt.Errorf("jsonrpc2: request received on client endpoint")
		case types.NOTIFICATION_KIND:
			return fmt.Errorf("jsonrpc2: notification received on client endpoint")
		case types.SUCCESS_RESPONSE_KIND:
			fallthrough
		case types.ERROR_RESPONSE_KIND:
			// this is just shim to allow make common methods callable on http client
			pendingChannel, ok := c.pending[rpcMsg.Id]
			if !ok {
				c.logger.Debug("jsonrpc2: ignoring response with no corresponding request", "response_id", rpcMsg.Id)
				continue
			}
			pendingChannel <- rpcMsg
		case types.INVALID_KIND:
			return fmt.Errorf("jsonrpc2: invalid message received: %s", err.Error())
		default:
			return fmt.Errorf("jsonrpc2: unknown message kind received: %s", kind)
		}
	}
	return nil
}
