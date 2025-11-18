package buffer

import (
	"regexp"
	"regexp/syntax"
	"unicode/utf8"

	"github.com/zyedidia/micro/v2/internal/util"
)

// RegexpGroup combines a Regexp with padded versions.
type RegexpGroup struct {
	// TODO: This is currently not intended to be used in plugins, and a better
	// abstract API should be provided

	// We want "^" and "$" to match only the beginning/end of a line, not that
	// of the search region somewhere in the middle of a line. In that case we
	// use padded regexps to require a rune before or after the match. (This
	// also affects other empty-string patterns like "\\b".)
	//
	// Although some are not needed in all cases, they are checked because
	// not returning an error for a regex in some cases would be an
	// implementation detail that could cause confusion.
	regex [4]*regexp.Regexp
}

const (
	padStart = 1 << iota
	padEnd
)

func findLineParams(b *Buffer, start, end Loc, i int) ([]byte, int, int) {
	l := b.LineBytes(i)
	charpos := 0
	padMode := 0

	if i == end.Y {
		nchars := util.CharacterCount(l)
		end.X = util.Clamp(end.X, 0, nchars)
		if end.X < nchars {
			l = util.SliceStart(l, end.X+1)
			padMode |= padEnd
		}
	}

	if i == start.Y {
		nchars := util.CharacterCount(l)
		start.X = util.Clamp(start.X, 0, nchars)
		if start.X > 0 {
			charpos = start.X - 1
			l = util.SliceEnd(l, charpos)
			padMode |= padStart
		}
	}

	return l, charpos, padMode
}

// NewRegexpGroup creates a RegexpGroup from a string
func NewRegexpGroup(s string) (*RegexpGroup, error) {
	var regex [4]*regexp.Regexp
	var err error
	regex[0], err = regexp.Compile(s)
	if err == nil {
		regex[padStart], err = regexp.Compile(".(?:" + s + ")")
		if err == nil {
			regex[padEnd] = regexp.MustCompile("(?:" + s + ").")
			regex[padStart|padEnd] = regexp.MustCompile(".(?:" + s + ").")
		} else {
			err = &syntax.Error{syntax.ErrorCode(`possibly \Q without \E`), s}
		}
	}
	return &RegexpGroup{regex}, err
}

func (b *Buffer) findDown(rg *RegexpGroup, start, end Loc) ([2]Loc, bool) {
	lastcn := util.CharacterCount(b.LineBytes(b.LinesNum() - 1))
	if start.Y > b.LinesNum()-1 {
		start.X = lastcn - 1
	}
	if end.Y > b.LinesNum()-1 {
		end.X = lastcn
	}
	start.Y = util.Clamp(start.Y, 0, b.LinesNum()-1)
	end.Y = util.Clamp(end.Y, 0, b.LinesNum()-1)

	if start.GreaterThan(end) {
		start, end = end, start
	}

	for i := start.Y; i <= end.Y; i++ {
		l, charpos, padMode := findLineParams(b, start, end, i)

		match := rg.regex[padMode].FindIndex(l)

		if match != nil {
			if padMode&padStart != 0 {
				_, size := utf8.DecodeRune(l[match[0]:])
				match[0] += size
			}
			if padMode&padEnd != 0 {
				_, size := utf8.DecodeLastRune(l[:match[1]])
				match[1] -= size
			}
			start := Loc{charpos + util.RunePos(l, match[0]), i}
			end := Loc{charpos + util.RunePos(l, match[1]), i}
			return [2]Loc{start, end}, true
		}
	}
	return [2]Loc{}, false
}

func (b *Buffer) findUp(rg *RegexpGroup, start, end Loc) ([2]Loc, bool) {
	lastcn := util.CharacterCount(b.LineBytes(b.LinesNum() - 1))
	if start.Y > b.LinesNum()-1 {
		start.X = lastcn - 1
	}
	if end.Y > b.LinesNum()-1 {
		end.X = lastcn
	}
	start.Y = util.Clamp(start.Y, 0, b.LinesNum()-1)
	end.Y = util.Clamp(end.Y, 0, b.LinesNum()-1)

	if start.GreaterThan(end) {
		start, end = end, start
	}

	for i := end.Y; i >= start.Y; i-- {
		charCount := util.CharacterCount(b.LineBytes(i))
		from := Loc{0, i}.Clamp(start, end)
		to := Loc{charCount, i}.Clamp(start, end)

		allMatches := b.findAll(rg, from, to)
		if allMatches != nil {
			match := allMatches[len(allMatches)-1]
			return [2]Loc{match[0], match[1]}, true
		}
	}
	return [2]Loc{}, false
}

func (b *Buffer) findAll(rg *RegexpGroup, start, end Loc) [][2]Loc {
	var matches [][2]Loc
	loc := start
	for {
		match, found := b.findDown(rg, loc, end)
		if !found {
			break
		}
		matches = append(matches, match)
		if match[0] != match[1] {
			loc = match[1]
		} else if match[1] != end {
			loc = match[1].Move(1, b)
		} else {
			break
		}
	}
	return matches
}

// FindNext finds the next occurrence of a given string in the buffer
// It returns the start and end location of the match (if found) and
// a boolean indicating if it was found
// May also return an error if the search regex is invalid
func (b *Buffer) FindNext(s string, start, end, from Loc, down bool, useRegex bool) ([2]Loc, bool, error) {
	if s == "" {
		return [2]Loc{}, false, nil
	}

	if !useRegex {
		s = regexp.QuoteMeta(s)
	}

	if b.Settings["ignorecase"].(bool) {
		s = "(?i)" + s
	}

	rg, err := NewRegexpGroup(s)
	if err != nil {
		return [2]Loc{}, false, err
	}

	var found bool
	var l [2]Loc
	if down {
		l, found = b.findDown(rg, from, end)
		if !found {
			l, found = b.findDown(rg, start, end)
		}
	} else {
		l, found = b.findUp(rg, from, start)
		if !found {
			l, found = b.findUp(rg, end, start)
		}
	}
	return l, found, nil
}

// ReplaceRegex replaces all occurrences of 'search' with 'replace' in the given area
// and returns the number of replacements made and the number of characters
// added or removed on the last line of the range
func (b *Buffer) ReplaceRegex(start, end Loc, search *regexp.Regexp, replace []byte, captureGroups bool) (int, int) {
	rg, err := NewRegexpGroup(search.String())
	if err != nil {
		return 0, 0
	}
	return b.ReplaceRegexGroup(start, end, rg, replace, captureGroups)
}

func (b *Buffer) ReplaceRegexGroup(start, end Loc, rg *RegexpGroup, replace []byte, captureGroups bool) (int, int) {
	if start.GreaterThan(end) {
		start, end = end, start
	}

	charsEnd := util.CharacterCount(b.LineBytes(end.Y))
	found := 0
	var deltas []Delta

	search := rg.regex[0]

	for i := start.Y; i <= end.Y; i++ {
		l := b.LineBytes(i)
		charCount := util.CharacterCount(l)
		if (i == start.Y && start.X > 0) || (i == end.Y && end.X < charCount) {
			// This replacement code works in general, but it creates a separate
			// modification for each match. We only use it for the first and last
			// lines, which may use padded regexps

			from := Loc{0, i}.Clamp(start, end)
			to := Loc{charCount, i}.Clamp(start, end)
			matches := b.findAll(rg, from, to)
			found += len(matches)

			for j := len(matches) - 1; j >= 0; j-- {
				// if we counted upwards, the different deltas would interfere
				match := matches[j]
				var newText []byte
				if captureGroups {
					newText = search.ReplaceAll(b.Substr(match[0], match[1]), replace)
				} else {
					newText = replace
				}
				deltas = append(deltas, Delta{newText, match[0], match[1]})
			}
		} else {
			newLine := search.ReplaceAllFunc(l, func(in []byte) []byte {
				found++
				var result []byte
				if captureGroups {
					match := search.FindSubmatchIndex(in)
					result = search.Expand(result, replace, in, match)
				} else {
					result = replace
				}
				return result
			})
			deltas = append(deltas, Delta{newLine, Loc{0, i}, Loc{charCount, i}})
		}
	}

	b.MultipleReplace(deltas)

	return found, util.CharacterCount(b.LineBytes(end.Y)) - charsEnd
}
