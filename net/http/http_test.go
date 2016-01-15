package http

import "testing"

func TestUrl(t *testing.T) {
	raw := "rtsp://admin:123@www.cyeam.com:80/hello?debug=yes#fuuuuuuuuuuuck"
	u, err := ParseUrl(raw)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(u.Scheme)
		if u.Scheme != "rtsp" {
			t.Error("schema")
		}

		t.Log(u.User)
		if u.User.Username() != "admin" || u.User.Password() != "123" {
			t.Error("userinfo")
		}

		t.Log(u.Host)
		if u.Host.String() != "www.cyeam.com:80" {
			t.Error("host")
		}

		t.Log(u.Path)
		if u.Path != "/hello" {
			t.Error("path")
		}

		t.Log(u.RawQuery)
		if u.RawQuery != "debug=yes" {
			t.Error("query")
		}

		t.Log(u.Fragment)
		if u.Fragment != "fuuuuuuuuuuuck" {
			t.Error("fragment")
		}
	}
}
