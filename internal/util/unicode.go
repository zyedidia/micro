package util

import (
	"unicode"
	"unicode/utf8"
)

// Unicode is annoying. A "code point" (rune in Go-speak) may need up to
// 4 bytes to represent it. In general, a code point will represent a
// complete character, but this is not always the case. A character with
// accents may be made up of multiple code points (the code point for the
// original character, and additional code points for each accent/marking).
// The functions below are meant to help deal with these additional "combining"
// code points. In underlying operations (search, replace, etc...), micro will
// treat a character with combining code points as just the original code point.
// For rendering, micro will display the combining characters. It's not perfect
// but it's pretty good.

// DecodeCharacter returns the next character from an array of bytes
// A character is a rune along with any accompanying combining runes
func DecodeCharacter(b []byte) (rune, []rune, int) {
	r, size := utf8.DecodeRune(b)
	b = b[size:]
	c, s := utf8.DecodeRune(b)

	var combc []rune
	for unicode.In(c, unicode.Mark) {
		combc = append(combc, c)
		size += s

		b = b[s:]
		c, s = utf8.DecodeRune(b)
	}

	return r, combc, size
}

// DecodeCharacterInString returns the next character from a string
// A character is a rune along with any accompanying combining runes
func DecodeCharacterInString(str string) (rune, []rune, int) {
	r, size := utf8.DecodeRuneInString(str)
	str = str[size:]
	c, s := utf8.DecodeRuneInString(str)

	var combc []rune
	for unicode.In(c, unicode.Mark) {
		combc = append(combc, c)
		size += s

		str = str[s:]
		c, s = utf8.DecodeRuneInString(str)
	}

	return r, combc, size
}

// CharacterCount returns the number of characters in a byte array
// Similar to utf8.RuneCount but for unicode characters
func CharacterCount(b []byte) int {
	s := 0

	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if !unicode.In(r, unicode.Mark) {
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
		if !unicode.In(r, unicode.Mark) {
			s++
		}
	}

	return s
}
