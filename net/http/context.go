package http

type Context struct {
	Req     *Request
	Resp    *Response
	ReqAddr *Host
	//	elapse time.Duration
}

func NewContext() *Context {
	ctx := new(Context)
	return ctx
}
