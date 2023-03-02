package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/alis-is/jsonrpc2/rpc"
)

func writeJsonResponse(w http.ResponseWriter, response []byte, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

type EndpointRegistry map[string]RpcMethodRegistry

type ServerMux struct {
	http.ServeMux
	endpoints EndpointRegistry

	logger ILogger
}

func (mux *ServerMux) RegisterEndpoint(path string) {
	if _, ok := mux.endpoints[path]; ok {
		return
	}
	mux.HandleFunc(path, createHandler(mux, path))
}

func NewServerMux() *ServerMux {
	result := &ServerMux{
		ServeMux:  http.ServeMux{},
		endpoints: make(EndpointRegistry, 1),
		logger:    &DefaultLogger{},
	}

	result.RegisterEndpoint("/")
	return result
}

func (mux *ServerMux) GetMethods() RpcMethodRegistry {
	return mux.endpoints["/"]
}

func (mux *ServerMux) UseLogger(logger ILogger) {
	if logger == nil {
		mux.logger.Tracef("ignored nil logger")
		return
	}
	mux.logger = logger
}

func (mux *ServerMux) GetEndpoints() EndpointRegistry {
	return mux.endpoints
}

func RegisterServerMuxEndpointMethod[TParam rpc.ParamsType, TResult rpc.ResultType](mux *ServerMux, endpoint string, method string, handler RpcMethod[TParam, TResult]) {
	mux.RegisterEndpoint(endpoint)
	RegisterMethod(mux.endpoints[endpoint], method, handler)
}

func createHandler(mux *ServerMux, path string) func(w http.ResponseWriter, r *http.Request) {
	var reg RpcMethodRegistry
	var ok bool
	if reg, ok = mux.endpoints[path]; !ok {
		reg = make(RpcMethodRegistry, 1)
		mux.endpoints[path] = reg
		mux.logger.Tracef("registered new endpoint: %s", path)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		contentLengthHeader := r.Header.Get("Content-Length")
		switch contentType {
		case "application/jsonrequest":
			fallthrough
		case "application/json-rpc":
			fallthrough
		case "application/json":
		default:
			mux.logger.Tracef("got request with unsupported content type: %s", contentType)
			writeJsonResponse(w, rpc.NewInvalidRequestWithData(fmt.Sprintf("unsupported content type: %s", contentType)).ToResponseBytes(nil), http.StatusUnsupportedMediaType)
			return
		}

		contentLength, err := strconv.Atoi(contentLengthHeader)
		if contentLengthHeader == "" || err != nil {
			mux.logger.Tracef("got request with invalid content length: %s", contentLengthHeader)
			writeJsonResponse(w, rpc.NewInvalidRequestWithData(fmt.Sprintf("invalid content length: %s", contentLengthHeader)).ToResponseBytes(nil), http.StatusUnsupportedMediaType)
			return
		}

		msg := make([]byte, contentLength)
		defer r.Body.Close()
		_, err = io.ReadFull(r.Body, msg)
		if err != nil {
			mux.logger.Tracef("failed to read request body: %s", err.Error())
			writeJsonResponse(w, rpc.NewInvalidRequestWithData("invalid request body").ToResponseBytes(nil), http.StatusUnsupportedMediaType)
			return
		}

		var rpcObj rpc.Object
		err = json.Unmarshal(msg, &rpcObj)
		if err != nil {
			mux.logger.Tracef("failed to parse request body: %s", err.Error())
			writeJsonResponse(w, rpc.NewParseErrorWithData(err.Error()).ToResponseBytes(nil), http.StatusUnsupportedMediaType)
			return
		}

		messages := rpcObj.GetMessages()
		results := make([]interface{}, 0, len(messages))
		for _, rpcMsg := range messages {
			kind, err := rpcMsg.GetKind()
			switch kind {
			case rpc.REQUEST_KIND:
				results = append(results, ProcessRpcRequest(context.Background(), reg, &rpcMsg))
			case rpc.NOTIFICATION_KIND:
				_ = ProcessRpcRequest(context.Background(), reg, &rpcMsg)
			case rpc.SUCCESS_RESPONSE_KIND:
				fallthrough
			case rpc.ERROR_RESPONSE_KIND:
				mux.logger.Tracef("ignoring response message: %v", rpcMsg)
			default:
				mux.logger.Tracef("got invalid message: %v", rpcMsg)
				results = append(results, rpc.NewInvalidRequestWithData(err.Error()).ToResponse(rpcMsg.Id))
			}
		}

		nonEmptyResults := make([]interface{}, 0, len(results))
		for _, result := range results {
			if result != nil {
				nonEmptyResults = append(nonEmptyResults, result)
			}
		}

		if len(nonEmptyResults) == 0 {
			return
		}

		if !rpcObj.IsBatch() {
			mux.logger.Tracef("sending single response: %v", nonEmptyResults[0])
			responseBody, err := json.Marshal(nonEmptyResults[0])
			if err != nil {
				mux.logger.Tracef("failed to marshal response: %s", err.Error())
				writeJsonResponse(w, rpc.NewInternalErrorWithData(fmt.Sprintf("failed to marshal response: %s", err.Error())).ToResponseBytes(nil), http.StatusInternalServerError)
				return
			}
			if _, isErrorResponse := nonEmptyResults[0].(*rpc.ErrorResponse); isErrorResponse {
				writeJsonResponse(w, responseBody, http.StatusBadRequest)
			} else {
				writeJsonResponse(w, responseBody, http.StatusOK)
			}
			return
		}
		mux.logger.Tracef("sending batch response: %v", nonEmptyResults)
		responseBody, err := json.Marshal(nonEmptyResults)
		if err != nil {
			mux.logger.Tracef("failed to marshal response: %s", err.Error())
			writeJsonResponse(w, rpc.NewInternalErrorWithData(err.Error()).ToResponseBytes(nil), http.StatusInternalServerError)
			return
		}
		// there is no information in the spec about how to handle multiple responses
		// whether to return any other status code than 200 if there is error in one of the responses or all of them
		// so we just return 200 and let the client handle the responses
		writeJsonResponse(w, responseBody, http.StatusOK)
	}
}
