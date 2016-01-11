package http

import (
	"bytes"
	"encoding/base64"
	"strings"
)

type Request struct {
	Method string

	Url *URL

	Proto string

	UserAgent string

	Host string

	Headers Header

	Body string

	Raw bytes.Buffer
}

func NewRequest() *Request {
	r := new(Request)
	return r
}

func (this *Request) Init() {
	b := string(this.Raw.Bytes())
	if len(b) == 0 {
		return
	}

	this.Headers = make(map[string][]string, 0)

	this.Url = new(URL)

	startLine := strings.Split(b[:strings.Index(b, CRLF)], " ")
	if len(startLine) == 3 {
		this.Method, this.Url.RawPath, this.Proto = startLine[0], startLine[1], startLine[2]

		i := strings.Index(this.Url.RawPath, "?")
		if i != -1 {
			this.Url.Path = this.Url.RawPath[:i]
		} else {
			this.Url.Path = this.Url.RawPath
		}

		raw := this.Url.RawPath[i+1:]
		i = strings.Index(raw, "#")
		if i != -1 {
			this.Url.RawQuery = raw[:i]
		} else {
			this.Url.RawQuery = raw
		}
		this.Url.Fragment = raw[i+1:]
	}
	b = b[strings.Index(b, CRLF)+len(CRLF):]

	if strings.LastIndex(b, CRLF+CRLF) != -1 {
		this.Body = b[strings.LastIndex(b, CRLF+CRLF):]
		b = b[:strings.LastIndex(b, CRLF+CRLF)-2]
	}
	b = strings.TrimSpace(b)

	for _, line := range strings.Split(b, CRLF) {
		k, v := line[:strings.Index(line, ":")], line[strings.Index(line, ":")+1:]
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if k == HTTP_HEAD_USERAGENT {
			this.UserAgent = v
		} else if k == HTTP_HEAD_HOST {
			this.Host = v
		} else {
			this.Headers[k] = append(this.Headers[k], v)
		}
	}
}

func (this *Request) Authorization() (username, password string, ok bool) {
	if this.Headers.Get(HTTP_HEAD_AUTHORIZATION) != "" {
		if strings.HasPrefix(this.Headers.Get(HTTP_HEAD_AUTHORIZATION), AUTH_BASIC) {
			authorization, err := base64.StdEncoding.DecodeString(this.Headers.Get(HTTP_HEAD_AUTHORIZATION)[len(AUTH_BASIC):])
			if err != nil {
				return
			}
			s := strings.IndexByte(string(authorization), ':')
			if s < 0 {
				return
			}
			return string(authorization[:s]), string(authorization[s+1:]), true
		}
	}
	return
}
