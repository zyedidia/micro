package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// Util.go is a collection of utility functions that are used throughout
// the program

// Count returns the length of a string in runes
// This is exactly equivalent to utf8.RuneCountInString(), just less characters
func Count(s string) int {
	return utf8.RuneCountInString(s)
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
		return true
	}
	c := str[0]
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c == '_')
}

// IsWhitespace returns true if the given rune is a space, tab, or newline
func IsWhitespace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n'
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
func IsSpaces(str string) bool {
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

// FuncName returns the name of a given function object
func FuncName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// SplitCommandArgs separates multiple command arguments which may be quoted.
// The returned slice contains at least one string
func SplitCommandArgs(input string) []string {
	var result []string
	var curQuote *bytes.Buffer

	curArg := new(bytes.Buffer)
	escape := false

	finishQuote := func() {
		if curQuote == nil {
			return
		}
		str := curQuote.String()
		if unquoted, err := strconv.Unquote(str); err == nil {
			str = unquoted
		}
		curArg.WriteString(str)
		curQuote = nil
	}

	appendResult := func() {
		finishQuote()
		escape = false

		str := curArg.String()
		result = append(result, str)
		curArg.Reset()
	}

	for _, r := range input {
		if r == ' ' && curQuote == nil {
			appendResult()
		} else {
			runeHandled := false
			appendRuneToBuff := func() {
				if curQuote != nil {
					curQuote.WriteRune(r)
				} else {
					curArg.WriteRune(r)
				}
				runeHandled = true
			}

			if r == '"' && curQuote == nil {
				curQuote = new(bytes.Buffer)
				appendRuneToBuff()
			} else {
				if curQuote != nil && !escape {
					if r == '"' {
						appendRuneToBuff()
						finishQuote()
					} else if r == '\\' {
						appendRuneToBuff()
						escape = true
						continue
					}
				}
			}
			if !runeHandled {
				appendRuneToBuff()
			}
		}

		escape = false
	}
	appendResult()
	return result
}

// JoinCommandArgs joins multiple command arguments and quote the strings if needed.
func JoinCommandArgs(args ...string) string {
	buf := new(bytes.Buffer)
	first := true
	for _, arg := range args {
		if first {
			first = false
		} else {
			buf.WriteRune(' ')
		}
		quoted := strconv.Quote(arg)
		if quoted[1:len(quoted)-1] != arg || strings.ContainsRune(arg, ' ') {
			buf.WriteString(quoted)
		} else {
			buf.WriteString(arg)
		}
	}

	return buf.String()
}
