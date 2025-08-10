package types

type Request[TParam Params] struct {
	MessageBase
	Id     interface{} `json:"id,omitempty"`
	Method string      `json:"method,omitempty"`
	Params TParam      `json:"params,omitempty"`
}

func NewRequest[TId Id, TParam Params](id TId, method string, params TParam) (req *Request[TParam]) {
	return &Request[TParam]{
		MessageBase: MessageBase{Version: jsonRpcVersion},
		Id:          (interface{})(id),
		Method:      method,
		Params:      params,
	}
}

func NewNotification[TParam Params](method string, params TParam) (req *Request[TParam]) {
	return &Request[TParam]{
		MessageBase: MessageBase{
			Version: jsonRpcVersion,
		},
		Method: method,
		Params: params,
	}
}

func (r *Request[TParam]) IsNotification() bool {
	return r.Id == nil
}

func (r *Request[TParam]) IsRequest() bool {
	return r.Id != nil
}

func (r *Request[TParam]) Kind() MessageKind {
	if r.IsNotification() {
		return NOTIFICATION_KIND
	}
	return REQUEST_KIND
}
