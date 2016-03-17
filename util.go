package main

import (
	"unicode/utf8"
)

func count(s string) int {
	return utf8.RuneCountInString(s)
}

func numOccurences(s string, c byte) int {
	var n int
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			n++
		}
	}
	return n
}

func emptyString(n int) string {
	var str string
	for i := 0; i < n; i++ {
		str += " "
	}
	return str
}
