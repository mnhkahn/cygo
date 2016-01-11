package http

import (
	"strings"
)

type URL struct {
	Scheme   string
	Opaque   string    // encoded opaque data
	User     *Userinfo // username and password information
	Host     string    // host or host:port
	Path     string
	RawPath  string // encoded path hint (Go 1.5 and later only; see EscapedPath method)
	RawQuery string // encoded query values, without '?'
	Fragment string // fragment for references, without '#'
}

func ParseUrl(raw string) (*URL, error) {
	var err error
	u := new(URL)
	u.RawPath = raw

	return u, err
}

func (this *URL) Query() Values {
	m := make(Values)

	qs := strings.Split(this.RawQuery, "&")
	for _, q := range qs {
		qss := strings.Split(q, "=")
		if len(qss) == 2 {
			m.Add(qss[0], qss[1])
		}
	}
	return m
}

type Userinfo struct {
	// contains filtered or unexported fields
}

func (this *Userinfo) String() string {
	return ""
}

type Values map[string][]string

func (this Values) Add(k, v string) {
	this[k] = append(this[k], v)
}

func (this Values) Get(param string) []string {
	return this[param]
}
