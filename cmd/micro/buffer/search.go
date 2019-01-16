package buffer

import (
	"regexp"
	"unicode/utf8"

	"github.com/zyedidia/micro/cmd/micro/util"
)

func (b *Buffer) findDown(r *regexp.Regexp, start, end Loc) ([2]Loc, bool) {
	start.Y = util.Clamp(start.Y, 0, b.LinesNum()-1)
	end.Y = util.Clamp(end.Y, 0, b.LinesNum()-1)

	if start.GreaterThan(end) {
		start, end = end, start
	}

	for i := start.Y; i <= end.Y; i++ {
		l := b.LineBytes(i)
		charpos := 0

		if i == start.Y && start.Y == end.Y {
			nchars := utf8.RuneCount(l)
			start.X = util.Clamp(start.X, 0, nchars-1)
			end.X = util.Clamp(end.X, 0, nchars-1)
			l = util.SliceStart(l, end.X)
			l = util.SliceEnd(l, start.X)
			charpos = start.X
		} else if i == start.Y {
			nchars := utf8.RuneCount(l)
			start.X = util.Clamp(start.X, 0, nchars-1)
			l = util.SliceEnd(l, start.X)
			charpos = start.X
		} else if i == end.Y {
			nchars := utf8.RuneCount(l)
			end.X = util.Clamp(end.X, 0, nchars-1)
			l = util.SliceStart(l, end.X)
		}

		match := r.FindIndex(l)

		if match != nil {
			start := Loc{charpos + util.RunePos(l, match[0]), i}
			end := Loc{charpos + util.RunePos(l, match[1]), i}
			return [2]Loc{start, end}, true
		}
	}
	return [2]Loc{}, false
}

func (b *Buffer) findUp(r *regexp.Regexp, start, end Loc) ([2]Loc, bool) {
	start.Y = util.Clamp(start.Y, 0, b.LinesNum()-1)
	end.Y = util.Clamp(end.Y, 0, b.LinesNum()-1)

	if start.GreaterThan(end) {
		start, end = end, start
	}

	for i := end.Y; i >= start.Y; i-- {
		l := b.LineBytes(i)
		charpos := 0

		if i == start.Y && start.Y == end.Y {
			nchars := utf8.RuneCount(l)
			start.X = util.Clamp(start.X, 0, nchars-1)
			end.X = util.Clamp(end.X, 0, nchars-1)
			l = util.SliceStart(l, end.X)
			l = util.SliceEnd(l, start.X)
			charpos = start.X
		} else if i == start.Y {
			nchars := utf8.RuneCount(l)
			start.X = util.Clamp(start.X, 0, nchars-1)
			l = util.SliceEnd(l, start.X)
			charpos = start.X
		} else if i == end.Y {
			nchars := utf8.RuneCount(l)
			end.X = util.Clamp(end.X, 0, nchars-1)
			l = util.SliceStart(l, end.X)
		}

		match := r.FindIndex(l)

		if match != nil {
			start := Loc{charpos + util.RunePos(l, match[0]), i}
			end := Loc{charpos + util.RunePos(l, match[1]), i}
			return [2]Loc{start, end}, true
		}
	}
	return [2]Loc{}, false
}

// FindNext finds the next occurrence of a given string in the buffer
// It returns the start and end location of the match (if found) and
// a boolean indicating if it was found
// May also return an error if the search regex is invalid
func (b *Buffer) FindNext(s string, start, end, from Loc, down bool, useRegex bool) ([2]Loc, bool, error) {
	if s == "" {
		return [2]Loc{}, false, nil
	}

	var r *regexp.Regexp
	var err error

	if !useRegex {
		s = regexp.QuoteMeta(s)
	}

	if b.Settings["ignorecase"].(bool) {
		r, err = regexp.Compile("(?i)" + s)
	} else {
		r, err = regexp.Compile(s)
	}

	if err != nil {
		return [2]Loc{}, false, err
	}

	found := false
	var l [2]Loc
	if down {
		l, found = b.findDown(r, from, end)
		if !found {
			l, found = b.findDown(r, start, from)
		}
	} else {
		l, found = b.findUp(r, from, start)
		if !found {
			l, found = b.findUp(r, end, from)
		}
	}
	return l, found, nil
}

// ReplaceRegex replaces all occurrences of 'search' with 'replace' in the given area
// and returns the number of replacements made
func (b *Buffer) ReplaceRegex(start, end Loc, search *regexp.Regexp, replace []byte) int {
	if start.GreaterThan(end) {
		start, end = end, start
	}

	found := 0
	var deltas []Delta
	for i := start.Y; i <= end.Y; i++ {
		l := b.lines[i].data
		charpos := 0

		// TODO: replace within X coords of selection
		if start.Y == end.Y && i == start.Y {
			l = util.SliceStart(l, end.X)
			l = util.SliceEnd(l, start.X)
			charpos = start.X
		} else if i == start.Y {
			l = util.SliceEnd(l, start.X)
			charpos = start.X
		} else if i == end.Y {
			l = util.SliceStart(l, end.X)
		}
		newText := search.ReplaceAllFunc(l, func(in []byte) []byte {
			found++
			return replace
		})

		from := Loc{charpos, i}
		to := Loc{charpos + utf8.RuneCount(l), i}

		deltas = append(deltas, Delta{newText, from, to})
	}
	b.MultipleReplace(deltas)

	return found
}
