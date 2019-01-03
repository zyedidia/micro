package util

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	runewidth "github.com/mattn/go-runewidth"
)

// SliceEnd returns a byte slice where the index is a rune index
// Slices off the start of the slice
func SliceEnd(slc []byte, index int) []byte {
	len := len(slc)
	i := 0
	totalSize := 0
	for totalSize < len {
		if i >= index {
			return slc[totalSize:]
		}

		_, size := utf8.DecodeRune(slc[totalSize:])
		totalSize += size
		i++
	}

	return slc[totalSize:]
}

// SliceStart returns a byte slice where the index is a rune index
// Slices off the end of the slice
func SliceStart(slc []byte, index int) []byte {
	len := len(slc)
	i := 0
	totalSize := 0
	for totalSize < len {
		if i >= index {
			return slc[:totalSize]
		}

		_, size := utf8.DecodeRune(slc[totalSize:])
		totalSize += size
		i++
	}

	return slc[:totalSize]
}

// SliceVisualEnd will take a byte slice and slice off the start
// up to a given visual index. If the index is in the middle of a
// rune the number of visual columns into the rune will be returned
// It will also return the char pos of the first character of the slice
func SliceVisualEnd(b []byte, n, tabsize int) ([]byte, int, int) {
	width := 0
	i := 0
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)

		w := 0
		switch r {
		case '\t':
			ts := tabsize - (width % tabsize)
			w = ts
		default:
			w = runewidth.RuneWidth(r)
		}
		if width+w > n {
			return b, n - width, i
		}
		width += w
		b = b[size:]
		i++
	}
	return b, n - width, i
}

// Abs is a simple absolute value function for ints
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// StringWidth returns the visual width of a byte array indexed from 0 to n (rune index)
// with a given tabsize
func StringWidth(b []byte, n, tabsize int) int {
	if n <= 0 {
		return 0
	}
	i := 0
	width := 0
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		b = b[size:]

		switch r {
		case '\t':
			ts := tabsize - (width % tabsize)
			width += ts
		default:
			width += runewidth.RuneWidth(r)
		}

		i++

		if i == n {
			return width
		}
	}
	return width
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

// FSize gets the size of a file
func FSize(f *os.File) int64 {
	fi, _ := f.Stat()
	return fi.Size()
}

// IsWordChar returns whether or not the string is a 'word character'
// If it is a unicode character, then it does not match
// Word characters are defined as [A-Za-z0-9_]
func IsWordChar(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r == '_')
}

// Spaces returns a string with n spaces
func Spaces(n int) string {
	return strings.Repeat(" ", n)
}

// IsSpaces checks if a given string is only spaces
func IsSpaces(str []byte) bool {
	for _, c := range str {
		if c != ' ' {
			return false
		}
	}

	return true
}

// IsSpacesOrTabs checks if a given string contains only spaces and tabs
func IsSpacesOrTabs(str []byte) bool {
	for _, c := range str {
		if c != ' ' && c != '\t' {
			return false
		}
	}

	return true
}

// IsWhitespace returns true if the given rune is a space, tab, or newline
func IsWhitespace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n'
}

// IsStrWhitespace returns true if the given string is all whitespace
func IsStrWhitespace(str string) bool {
	// Range loop for unicode correctness
	for _, c := range str {
		if !IsWhitespace(c) {
			return false
		}
	}
	return true
}

// RunePos returns the rune index of a given byte index
// Make sure the byte index is not between code points
func RunePos(b []byte, i int) int {
	return utf8.RuneCount(b[:i])
}

// TODO: consider changing because of snap segfault
// ReplaceHome takes a path as input and replaces ~ at the start of the path with the user's
// home directory. Does nothing if the path does not start with '~'.
func ReplaceHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	var userData *user.User
	var err error

	homeString := strings.Split(path, "/")[0]
	if homeString == "~" {
		userData, err = user.Current()
		if err != nil {
			return "", errors.New("Could not find user: " + err.Error())
		}
	} else {
		userData, err = user.Lookup(homeString[1:])
		if err != nil {
			return "", errors.New("Could not find user: " + err.Error())
		}
	}

	home := userData.HomeDir

	return strings.Replace(path, homeString, home, 1), nil
}

// GetPathAndCursorPosition returns a filename without everything following a `:`
// This is used for opening files like util.go:10:5 to specify a line and column
// Special cases like Windows Absolute path (C:\myfile.txt:10:5) are handled correctly.
func GetPathAndCursorPosition(path string) (string, []string) {
	re := regexp.MustCompile(`([\s\S]+?)(?::(\d+))(?::(\d+))?`)
	match := re.FindStringSubmatch(path)
	// no lines/columns were specified in the path, return just the path with no cursor location
	if len(match) == 0 {
		return path, nil
	} else if match[len(match)-1] != "" {
		// if the last capture group match isn't empty then both line and column were provided
		return match[1], match[2:]
	}
	// if it was empty, then only a line was provided, so default to column 0
	return match[1], []string{match[2], "0"}
}

// GetModTime returns the last modification time for a given file
func GetModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Now(), err
	}
	return info.ModTime(), nil
}

// EscapePath replaces every path separator in a given path with a %
func EscapePath(path string) string {
	path = filepath.ToSlash(path)
	return strings.Replace(path, "/", "%", -1)
}

// GetLeadingWhitespace returns the leading whitespace of the given byte array
func GetLeadingWhitespace(b []byte) []byte {
	ws := []byte{}
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if r == ' ' || r == '\t' {
			ws = append(ws, byte(r))
		} else {
			break
		}

		b = b[size:]
	}
	return ws
}

// IntOpt turns a float64 setting to an int
func IntOpt(opt interface{}) int {
	return int(opt.(float64))
}

// GetCharPosInLine gets the char position of a visual x y
// coordinate (this is necessary because tabs are 1 char but
// 4 visual spaces)
func GetCharPosInLine(b []byte, visualPos int, tabsize int) int {

	// Scan rune by rune until we exceed the visual width that we are
	// looking for. Then we can return the character position we have found
	i := 0     // char pos
	width := 0 // string visual width
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		b = b[size:]

		switch r {
		case '\t':
			ts := tabsize - (width % tabsize)
			width += ts
		default:
			width += runewidth.RuneWidth(r)
		}

		if width >= visualPos {
			if width == visualPos {
				i++
			}
			break
		}
		i++
	}

	return i
}

func Clamp(val, min, max int) int {
	if val < min {
		val = min
	} else if val > max {
		val = max
	}
	return val
}
