package strings

import (
	gostr "strings"
	"unicode/utf8"
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

func IndexAll(s, sep string) []int {
	var pos []int
	// special cases
	switch {
	case len(sep) == 0:
		pos = append(pos, utf8.RuneCountInString(s)+1)
		return pos
	case len(sep) == 1:
		// special case worth making fast
		c := sep[0]
		for i := 0; i < len(s); i++ {
			if s[i] == c {
				pos = append(pos, i)
			}
		}
		return pos
	case len(sep) > len(s):
		return pos
	case len(sep) == len(s):
		if sep == s {
			pos = append(pos, 0)
			return pos
		}
		return pos
	}
	// Rabin-Karp search
	hashsep, pow := hashStr(sep)
	h := uint32(0)
	for i := 0; i < len(sep); i++ {
		h = h*primeRK + uint32(s[i])
	}
	lastmatch := 0
	if h == hashsep && s[:len(sep)] == sep { // 一开始匹配上
		pos = append(pos, 0)
		lastmatch = len(sep)
	}
	for i := len(sep); i < len(s); {
		h *= primeRK
		h += uint32(s[i])
		h -= pow * uint32(s[i-len(sep)])
		i++
		if h == hashsep && lastmatch <= i-len(sep) && s[i-len(sep):i] == sep {
			pos = append(pos, i-len(sep))
			lastmatch = i
		}
	}
	return pos
}

// primeRK is the prime base used in Rabin-Karp algorithm.
const primeRK = 16777619

// hashStr returns the hash and the appropriate multiplicative
// factor for use in Rabin-Karp algorithm.
func hashStr(sep string) (uint32, uint32) {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*primeRK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, primeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}
