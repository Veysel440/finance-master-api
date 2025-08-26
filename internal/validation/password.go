package validation

import "unicode"

type simpleErr string

func (e simpleErr) Error() string { return string(e) }

func errStr(s string) error { return simpleErr(s) }

func containsFold(s, sub string) bool {
	rs, rsub := []rune(s), []rune(sub)
	n, m := len(rs), len(rsub)
	for i := 0; i+m <= n; i++ {
		ok := true
		for j := 0; j < m; j++ {
			if unicode.ToLower(rs[i+j]) != unicode.ToLower(rsub[j]) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}
func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
