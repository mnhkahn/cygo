package http

type Header map[string][]string

func NewHeader() Header {
	h := make(Header)
	h.Add(HTTP_HEAD_SERVER, "Cyeam")
	return h
}

func (this Header) Add(k, v string) {
	this[k] = append(this[k], v)
}

func (this Header) Get(k string) string {
	if len(this[k]) > 0 {
		return this[k][0]
	}
	return ""
}

const (
	HTTP_HEAD_USERAGENT        = "User-Agent"
	HTTP_HEAD_HOST             = "Host"
	HTTP_HEAD_LOCATION         = "Location"
	HTTP_HEAD_SERVER           = "Server"
	HTTP_HEAD_CONTENTTYPE      = "Content-Type"
	HTTP_HEAD_CONTENTLENGTH    = "Content-Length"
	HTTP_HEAD_DATE             = "Date"
	HTTP_HEAD_ALLOW            = "Allow"
	HTTP_HEAD_FORM             = "Form" // 用户的Email地址，如果是爬虫，最好有这个头
	HTTP_HEAD_X_FORWARDED_FOR  = "X-Forwarded-For"
	HTTP_HEAD_WWW_AUTHENTICATE = "WWW-authenticate"
	HTTP_HEAD_AUTHORIZATION    = "Authorization"
)
