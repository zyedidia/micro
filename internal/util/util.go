package util

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/blang/semver"
	runewidth "github.com/mattn/go-runewidth"
)

var (
	// These variables should be set by the linker when compiling

	// Version is the version number or commit hash
	Version = "0.0.0-unknown"
	// Semantic version
	SemVersion semver.Version
	// CommitHash is the commit this version was built on
	CommitHash = "Unknown"
	// CompileDate is the date this binary was compiled on
	CompileDate = "Unknown"
	// Debug logging
	Debug = "OFF"
	// FakeCursor is used to disable the terminal cursor and have micro
	// draw its own (enabled for windows consoles where the cursor is slow)
	FakeCursor = false

	// Stdout is a buffer that is written to stdout when micro closes
	Stdout *bytes.Buffer
)

func init() {
	var err error
	SemVersion, err = semver.Make(Version)
	if err != nil {
		fmt.Println("Invalid version: ", Version, err)
	}

	if runtime.GOOS == "windows" {
		FakeCursor = true
	}
	Stdout = new(bytes.Buffer)
}

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

		_, _, size := DecodeCharacter(slc[totalSize:])
		totalSize += size
		i++
	}

	return slc[totalSize:]
}

// SliceEndStr is the same as SliceEnd but for strings
func SliceEndStr(str string, index int) string {
	len := len(str)
	i := 0
	totalSize := 0
	for totalSize < len {
		if i >= index {
			return str[totalSize:]
		}

		_, _, size := DecodeCharacterInString(str[totalSize:])
		totalSize += size
		i++
	}

	return str[totalSize:]
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

		_, _, size := DecodeCharacter(slc[totalSize:])
		totalSize += size
		i++
	}

	return slc[:totalSize]
}

// SliceStartStr is the same as SliceStart but for strings
func SliceStartStr(str string, index int) string {
	len := len(str)
	i := 0
	totalSize := 0
	for totalSize < len {
		if i >= index {
			return str[:totalSize]
		}

		_, _, size := DecodeCharacterInString(str[totalSize:])
		totalSize += size
		i++
	}

	return str[:totalSize]
}

// SliceVisualEnd will take a byte slice and slice off the start
// up to a given visual index. If the index is in the middle of a
// rune the number of visual columns into the rune will be returned
// It will also return the char pos of the first character of the slice
func SliceVisualEnd(b []byte, n, tabsize int) ([]byte, int, int) {
	width := 0
	i := 0
	for len(b) > 0 {
		r, _, size := DecodeCharacter(b)

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
		r, _, size := DecodeCharacter(b)
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
// Word characters are defined as numbers, letters, or '_'
func IsWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
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
	return unicode.IsSpace(c)
}

// IsBytesWhitespace returns true if the given bytes are all whitespace
func IsBytesWhitespace(b []byte) bool {
	for _, c := range b {
		if !IsWhitespace(rune(c)) {
			return false
		}
	}
	return true
}

// RunePos returns the rune index of a given byte index
// Make sure the byte index is not between code points
func RunePos(b []byte, i int) int {
	return CharacterCount(b[:i])
}

// MakeRelative will attempt to make a relative path between path and base
func MakeRelative(path, base string) (string, error) {
	if len(path) > 0 {
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return path, err
		}
		return rel, nil
	}
	return path, nil
}

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
		r, _, size := DecodeCharacter(b)
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
		r, _, size := DecodeCharacter(b)
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

// ParseBool is almost exactly like strconv.ParseBool, except it also accepts 'on' and 'off'
// as 'true' and 'false' respectively
func ParseBool(str string) (bool, error) {
	if str == "on" {
		return true, nil
	}
	if str == "off" {
		return false, nil
	}
	return strconv.ParseBool(str)
}

// Clamp clamps a value between min and max
func Clamp(val, min, max int) int {
	if val < min {
		val = min
	} else if val > max {
		val = max
	}
	return val
}

func IsNonAlphaNumeric(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '_'
}

func ParseSpecial(s string) string {
	return strings.Replace(s, "\\t", "\t", -1)
}

// String converts a byte array to a string (for lua plugins)
func String(s []byte) string {
	return string(s)
}
