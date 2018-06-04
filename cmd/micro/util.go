package main

import (
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"regexp"
	"github.com/go-errors/errors"
)

// Util.go is a collection of utility functions that are used throughout
// the program

// Count returns the length of a string in runes
// This is exactly equivalent to utf8.RuneCountInString(), just less characters
func Count(s string) int {
	return utf8.RuneCountInString(s)
}

// Convert byte array to rune array
func toRunes(b []byte) []rune {
	runes := make([]rune, 0, utf8.RuneCount(b))

	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		runes = append(runes, r)

		b = b[size:]
	}

	return runes
}

func sliceStart(slc []byte, index int) []byte {
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

func sliceEnd(slc []byte, index int) []byte {
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

// NumOccurrences counts the number of occurrences of a byte in a string
func NumOccurrences(s string, c byte) int {
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
	return strings.Repeat(" ", n)
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
	// get the size
	return fi.Size()
}

// IsWordChar returns whether or not the string is a 'word character'
// If it is a unicode character, then it does not match
// Word characters are defined as [A-Za-z0-9_]
func IsWordChar(str string) bool {
	if len(str) > 1 {
		// Unicode
		return true
	}
	c := str[0]
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c == '_')
}

// IsWhitespace returns true if the given rune is a space, tab, or newline
func IsWhitespace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n'
}

// IsStrWhitespace returns true if the given string is all whitespace
func IsStrWhitespace(str string) bool {
	for _, c := range str {
		if !IsWhitespace(c) {
			return false
		}
	}
	return true
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

// Insert makes a simple insert into a string at the given position
func Insert(str string, pos int, value string) string {
	return string([]rune(str)[:pos]) + value + string([]rune(str)[pos:])
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

// GetLeadingWhitespace returns the leading whitespace of the given string
func GetLeadingWhitespace(str string) string {
	ws := ""
	for _, c := range str {
		if c == ' ' || c == '\t' {
			ws += string(c)
		} else {
			break
		}
	}
	return ws
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
func IsSpacesOrTabs(str string) bool {
	for _, c := range str {
		if c != ' ' && c != '\t' {
			return false
		}
	}

	return true
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

// EscapePath replaces every path separator in a given path with a %
func EscapePath(path string) string {
	path = filepath.ToSlash(path)
	return strings.Replace(path, "/", "%", -1)
}

// GetModTime returns the last modification time for a given file
// It also returns a boolean if there was a problem accessing the file
func GetModTime(path string) (time.Time, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Now(), false
	}
	return info.ModTime(), true
}

// StringWidth returns the width of a string where tabs count as `tabsize` width
func StringWidth(str string, tabsize int) int {
	sw := runewidth.StringWidth(str)
	lineIdx := 0
	for _, ch := range str {
		switch ch {
		case '\t':
			ts := tabsize - (lineIdx % tabsize)
			sw += ts
			lineIdx += ts
		case '\n':
			lineIdx = 0
		default:
			lineIdx++
		}
	}
	return sw
}

// WidthOfLargeRunes searches all the runes in a string and counts up all the widths of runes
// that have a width larger than 1 (this also counts tabs as `tabsize` width)
func WidthOfLargeRunes(str string, tabsize int) int {
	count := 0
	lineIdx := 0
	for _, ch := range str {
		var w int
		if ch == '\t' {
			w = tabsize - (lineIdx % tabsize)
		} else {
			w = runewidth.RuneWidth(ch)
		}
		if w > 1 {
			count += (w - 1)
		}
		if ch == '\n' {
			lineIdx = 0
		} else {
			lineIdx += w
		}
	}
	return count
}

// RunePos returns the rune index of a given byte index
// This could cause problems if the byte index is between code points
func runePos(p int, str string) int {
	return utf8.RuneCountInString(str[:p])
}

func lcs(a, b string) string {
	arunes := []rune(a)
	brunes := []rune(b)

	lcs := ""
	for i, r := range arunes {
		if i >= len(brunes) {
			break
		}
		if r == brunes[i] {
			lcs += string(r)
		} else {
			break
		}
	}
	return lcs
}

// CommonSubstring gets a common substring among the inputs
func CommonSubstring(arr ...string) string {
	commonStr := arr[0]

	for _, str := range arr[1:] {
		commonStr = lcs(commonStr, str)
	}

	return commonStr
}

// Abs is a simple absolute value function for ints
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// FuncName returns the full name of a given function object
func FuncName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// ShortFuncName returns the name only of a given function object
func ShortFuncName(i interface{}) string {
	return strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name(), "main.(*View).")
}

// ReplaceHome takes a path as input and replaces ~ at the start of the path with the user's
// home directory. Does nothing if the path does not start with '~'.
func ReplaceHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	var userData *user.User
	var err error

	homeString := strings.Split(path, "/")[0]
	if homeString == "~" {
		userData, err = user.Current()
		if err != nil {
			messenger.Error("Could not find user: ", err)
		}
	} else {
		userData, err = user.Lookup(homeString[1:])
		if err != nil {
			if messenger != nil {
				messenger.Error("Could not find user: ", err)
			} else {
				TermMessage("Could not find user: ", err)
			}
			return ""
		}
	}

	home := userData.HomeDir

	return strings.Replace(path, homeString, home, 1)
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

func ParseCursorLocation(cursorPositions []string) (Loc, error) {
	startpos := Loc{0, 0}
	var err error

	// if no positions are available exit early
	if cursorPositions == nil {
		return startpos, errors.New("No cursor positions were provided.")
	}

	startpos.Y, err = strconv.Atoi(cursorPositions[0])
	if err != nil {
		messenger.Error("Error parsing cursor position: ", err)
	} else {
		if len(cursorPositions) > 1 {
			startpos.X, err = strconv.Atoi(cursorPositions[1])
			if err != nil {
				messenger.Error("Error parsing cursor position: ", err)
			}
		}
	}

	return startpos, err
}
