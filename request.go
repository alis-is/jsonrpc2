package jsonrpc2

type request[TParam Params] struct {
	messageBase
	Id     interface{} `json:"id,omitempty"`
	Method string      `json:"method,omitempty"`
	Params TParam      `json:"params,omitempty"`
}

func NewRequest[TId Id, TParam Params](id TId, method string, params TParam) (req *request[TParam]) {
	return &request[TParam]{
		messageBase: messageBase{Version: jsonRpcVersion},
		Id:          (interface{})(id),
		Method:      method,
		Params:      params,
	}
}

func NewNotification[TParam Params](method string, params TParam) (req *request[TParam]) {
	return &request[TParam]{
		messageBase: messageBase{
			Version: jsonRpcVersion,
		},
		Method: method,
		Params: params,
	}
}

func (r *request[TParam]) IsNotification() bool {
	return r.Id == nil
}

func (r *request[TParam]) IsRequest() bool {
	return r.Id != nil
}

func (r *request[TParam]) Kind() MessageKind {
	if r.IsNotification() {
		return NOTIFICATION_KIND
	}
	return REQUEST_KIND
}
