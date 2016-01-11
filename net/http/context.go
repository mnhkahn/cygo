package http

type Context struct {
	Req     *Request
	Resp    *Response
	ReqAddr *Address
	//	elapse time.Duration
}

func NewContext() *Context {
	ctx := new(Context)
	return ctx
}
