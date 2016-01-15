package http

import (
	"strings"

	cystr "github.com/mnhkahn/cygo/strings"
)

type URL struct {
	Scheme   string
	Opaque   string    // encoded opaque data
	User     *Userinfo // username and password information
	Host     *Host     // host or host:port
	Path     string
	RawPath  string // encoded path hint (Go 1.5 and later only; see EscapedPath method)
	RawQuery string // encoded query values, without '?'
	Fragment string // fragment for references, without '#'
}

func ParseUrl(raw string) (*URL, error) {
	var err error
	u := new(URL)
	u.RawPath = raw

	// schema
	index := strings.Index(raw, "://")
	if index != -1 {
		u.Scheme = raw[:index]
		if index+3 <= len(raw) {
			raw = raw[index+3:]
		}
	}

	// userinfo
	index = strings.Index(raw, "@")
	if index != -1 {
		u.User = new(Userinfo)
		i := strings.Index(raw[:index], ":")
		if i != -1 {
			u.User.username, u.User.password = raw[:index][:i], raw[:index][i+1:]
		} else {
			u.User.username = raw[:index]
		}
		if index+1 <= len(raw) {
			raw = raw[index+1:]
		}
	}

	// host
	index = strings.Index(raw, "/")
	if index != -1 {
		u.Host = new(Host)
		i := strings.Index(raw[:index], ":")
		if i != -1 {
			u.Host.host, u.Host.port = raw[:index][:i], raw[:index][i+1:]
		} else {
			u.Host.host = raw[:index]
		}
		raw = raw[index:]
	}

	// path
	index = strings.Index(raw, "?")
	if index != -1 {
		u.Path = raw[:index]
		if index+1 <= len(raw) {
			raw = raw[index+1:]
		}
	} else {
		u.Path = raw
		return u, err
	}

	// query
	index = strings.Index(raw, "#")
	if index != -1 {
		u.RawQuery = raw[:index]
		if index+1 <= len(raw) {
			raw = raw[index+1:]
		}
	} else {
		u.RawQuery = raw
		return u, err
	}

	// fragment
	u.Fragment = raw

	return u, err
}

func (this *URL) Query() Values {
	m := make(Values)
	var k, v string
	var queryKeyValue = func(q string) {
		k, v = "", ""
		i := strings.Index(q, "=")
		if i != -1 {
			k = q[:i]
			if i+1 <= len(q) {
				v = q[i+1:]
			}
			m.Add(k, v)
		}
	}

	cystr.SplitEachAfter(this.RawQuery, "&", queryKeyValue)

	return m
}

type Userinfo struct {
	username string
	password string
}

func (this *Userinfo) Username() string {
	return this.username
}

func (this *Userinfo) Password() string {
	return this.password
}

func (this *Userinfo) String() string {
	return this.username + ":" + this.password
}

type Host struct {
	host string
	port string
}

func ParseHost(host string) *Host {
	h := new(Host)
	i := strings.Index(host, ":")
	if i != -1 {
		h.host, h.port = host[:i], host[i+1:]
	} else {
		h.host = host
	}
	return h
}

func (this *Host) Address() string {
	return this.host
}

func (this *Host) Port() string {
	return this.port
}

func (this *Host) String() string {
	return this.host + ":" + this.port
}

type Values map[string][]string

func (this Values) Add(k, v string) {
	this[k] = append(this[k], v)
}

func (this Values) Get(param string) []string {
	return this[param]
}
