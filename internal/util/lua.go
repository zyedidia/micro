package util

import (
	"unicode/utf8"
)

func LuaRuneAt(str string, runeidx int) string {
	i := 0
	for len(str) > 0 {
		r, size := utf8.DecodeRuneInString(str)

		str = str[size:]

		if i == runeidx {
			return string(r)
		}

		i++
	}
	return ""
}

func LuaGetLeadingWhitespace(s string) string {
	ws := []byte{}
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if r == ' ' || r == '\t' {
			ws = append(ws, byte(r))
		} else {
			break
		}

		s = s[size:]
	}
	return string(ws)
}

func LuaIsWordChar(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return IsWordChar(r)
}
