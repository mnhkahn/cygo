package http

import (
	"bytes"
	"encoding/base64"
	"strings"

	cyurl "github.com/mnhkahn/cygo/net/url"
)

type Request struct {
	Method string

	Url *cyurl.URL

	Proto string

	UserAgent string

	Host string

	Headers Header

	Body string

	Raw bytes.Buffer
}

func NewRequest() *Request {
	r := new(Request)
	r.Headers = make(map[string][]string, 0)
	r.Url = new(cyurl.URL)
	return r
}

func (this *Request) ParseStartLine(line []byte) {
	startLine := strings.Split(string(line), " ")
	if len(startLine) == 3 {
		var path string
		this.Method, path, this.Proto = startLine[0], startLine[1], startLine[2]
		this.Url, _ = cyurl.ParseUrl(path)
	}
}

func (this *Request) ParseHeader(line []byte) {
	l := string(line)
	k, v := l[:strings.Index(l, ":")], l[strings.Index(l, ":")+1:]
	k, v = strings.TrimSpace(k), strings.TrimSpace(v)
	if k == HTTP_HEAD_USERAGENT {
		this.UserAgent = v
	} else if k == HTTP_HEAD_HOST {
		this.Host = v
	} else {
		this.Headers[k] = append(this.Headers[k], v)
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
