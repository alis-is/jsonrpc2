package types

type Request[TParam ParamsType] struct {
	MessageBase
	Id     interface{} `json:"id,omitempty"`
	Method string      `json:"method,omitempty"`
	Params TParam      `json:"params,omitempty"`
}

func NewRequest[TId IdType, TParam ParamsType](id TId, method string, params TParam) (req *Request[TParam]) {
	return &Request[TParam]{
		MessageBase: MessageBase{Version: jsonRpcVersion},
		Id:          (interface{})(id),
		Method:      method,
		Params:      params,
	}
}

func NewNotification[TParam ParamsType](method string, params TParam) (req *Request[TParam]) {
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
