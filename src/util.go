package main

import (
	"unicode/utf8"
)

// Util.go is a collection of utility functions that are used throughout
// the program

// Count returns the length of a string in runes
// This is exactly equivalent to utf8.RuneCountInString(), just less characters
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

// Spaces returns a string with n spaces
func Spaces(n int) string {
	var str string
	for i := 0; i < n; i++ {
		str += " "
	}
	return str
}

// Min takes the min of two ints
func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// Max takes the max of two ints
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// IsWordChar returns whether or not the string is a 'word character'
// If it is a unicode character, then it does not match
// Word characters are defined as [A-Za-z0-9_]
func IsWordChar(str string) bool {
	if len(str) > 1 {
		// Unicode
		return false
	}
	c := str[0]
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c == '_')
}

// Contains returns whether or not a string array contains a given string
func Contains(list []string, a string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
