package main

import (
	"unicode/utf8"
)

// Count returns the length of a string in runes
func Count(s string) int {
	return utf8.RuneCountInString(s)
}

// NumOccurences counts the number of occurences of a byte in a string
func NumOccurences(s string, c byte) int {
	var n int
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			n++
		}
	}
	return n
}

// EmptyString returns an empty string n spaces long
func EmptyString(n int) string {
	var str string
	for i := 0; i < n; i++ {
		str += " "
	}
	return str
}
