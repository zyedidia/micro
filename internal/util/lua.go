package util

import (
	"unicode/utf8"
)

// LuaRuneAt is a helper function for lua plugins to return the rune
// at an index within a string
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

// LuaGetLeadingWhitespace returns the leading whitespace of a string (used by lua plugins)
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

// LuaIsWordChar returns true if the first rune in a string is a word character
func LuaIsWordChar(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return IsWordChar(r)
}
