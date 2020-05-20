package util

import (
	"unicode"
	"unicode/utf8"
)

// combining character range table
var combining = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x0300, 0x036f, 1}, // combining diacritical marks
		{0x1ab0, 0x1aff, 1}, // combining diacritical marks extended
		{0x1dc0, 0x1dff, 1}, // combining diacritical marks supplement
		{0x20d0, 0x20ff, 1}, // combining diacritical marks for symbols
		{0xfe20, 0xfe2f, 1}, // combining half marks
	},
}

// DecodeCharacter returns the next character from an array of bytes
// A character is a rune along with any accompanying combining runes
func DecodeCharacter(b []byte) (rune, []rune, int) {
	r, size := utf8.DecodeRune(b)
	b = b[size:]
	c, s := utf8.DecodeRune(b)

	var combc []rune
	for unicode.In(c, combining) {
		combc = append(combc, c)
		size += s

		b = b[s:]
		c, s = utf8.DecodeRune(b)
	}

	return r, combc, size
}

// CharacterCount returns the number of characters in a byte array
// Similar to utf8.RuneCount but for unicode characters
func CharacterCount(b []byte) int {
	s := 0

	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if !unicode.In(r, combining) {
			s++
		}

		b = b[size:]
	}

	return s
}

// CharacterCount returns the number of characters in a string
// Similar to utf8.RuneCountInString but for unicode characters
func CharacterCountInString(str string) int {
	s := 0

	for _, r := range str {
		if !unicode.In(r, combining) {
			s++
		}
	}

	return s
}
