package strings

import (
	gostr "strings"
)

const (
	DEFAULT_SEP = ","
)

func SplitEachAfter(s, sep string, f func(string)) {
	if sep == "" {
		sep = DEFAULT_SEP
	}
	for i := gostr.Index(s, sep); i >= 0; i = gostr.Index(s, sep) {
		f(s[:i])
		if i+len(sep) <= len(s) {
			s = s[i+len(sep):]
		} else {
			break
		}
	}
	f(s)
}

func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
