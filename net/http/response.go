package http

type Response struct {
	StatusCode int

	Proto string // e.g. "HTTP/1.0"

	Headers Header

	Body string
}

func NewResponse() *Response {
	resp := new(Response)
	resp.StatusCode = StatusOK
	resp.Headers = NewHeader()

	return resp
}
