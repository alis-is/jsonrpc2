package endpoints

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/alis-is/jsonrpc2/rpc"
)

type HttpClientEndpoint struct {
	*http.Client

	pendingMutex sync.Mutex
	pending      map[interface{}]chan rpc.Message

	url    string
	logger ILogger
}

func NewHttpClientEndpoint(baseUrl string, client *http.Client) *HttpClientEndpoint {
	if client == nil {
		client = http.DefaultClient
	}

	return &HttpClientEndpoint{
		Client:       client,
		url:          baseUrl,
		pendingMutex: sync.Mutex{},
		pending:      make(map[interface{}]chan rpc.Message, 1),
		logger:       &DefaultLogger{},
	}
}

func (c *HttpClientEndpoint) UseLogger(logger ILogger) {
	if logger == nil {
		c.logger.Debugf("ignored nil logger")
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

func (c *HttpClientEndpoint) RegisterPendingRequest(requestID interface{}) <-chan rpc.Message {
	responseChan := make(chan rpc.Message, 1)
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
	c.logger.Debugf("sending request to %s: %s\n", c.url, string(requestBody))
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
		return rpc.ErrEmptyResponse
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("jsonrpc2: http error: %s", string(body))
	}

	var rpcObj rpc.Object
	err = json.Unmarshal(body, &rpcObj)
	c.logger.Debugf("jsonrpc2: received message: %s\n", string(body))
	if err != nil {
		return err
	}
	messages := rpcObj.GetMessages()
	for _, rpcMsg := range messages {
		kind, err := rpcMsg.GetKind()
		switch kind {
		case rpc.REQUEST_KIND:
			return fmt.Errorf("jsonrpc2: request received on client endpoint")
		case rpc.NOTIFICATION_KIND:
			return fmt.Errorf("jsonrpc2: notification received on client endpoint")
		case rpc.SUCCESS_RESPONSE_KIND:
			fallthrough
		case rpc.ERROR_RESPONSE_KIND:
			// this is just shim to allow make common methods callable on http client
			pendingChannel, ok := c.pending[rpcMsg.Id]
			if !ok {
				c.logger.Debugf("jsonrpc2: ignoring response #%s with no corresponding request\n", rpcMsg.Id)
				continue
			}
			pendingChannel <- rpcMsg
		case rpc.INVALID_KIND:
			return fmt.Errorf("jsonrpc2: invalid message received: %s", err.Error())
		default:
			return fmt.Errorf("jsonrpc2: unknown message kind received: %s", kind)
		}
	}
	return nil
}
