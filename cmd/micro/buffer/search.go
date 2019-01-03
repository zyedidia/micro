package buffer

import (
	"regexp"
	"unicode/utf8"

	"github.com/zyedidia/micro/cmd/micro/util"
)

func (b *Buffer) findDown(r *regexp.Regexp, start, end Loc) ([2]Loc, bool) {
	start.Y = util.Clamp(start.Y, 0, b.LinesNum()-1)

	for i := start.Y; i <= end.Y; i++ {
		l := b.LineBytes(i)
		charpos := 0

		if i == start.Y {
			nchars := utf8.RuneCount(l)
			start.X = util.Clamp(start.X, 0, nchars-1)
			l = util.SliceEnd(l, start.X)
			charpos = start.X
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

	for i := start.Y; i >= end.Y; i-- {
		l := b.LineBytes(i)

		if i == start.Y {
			nchars := utf8.RuneCount(l)
			start.X = util.Clamp(start.X, 0, nchars-1)
			l = util.SliceStart(l, start.X)
		}

		match := r.FindIndex(l)

		if match != nil {
			start := Loc{util.RunePos(l, match[0]), i}
			end := Loc{util.RunePos(l, match[1]), i}
			return [2]Loc{start, end}, true
		}
	}
	return [2]Loc{}, false
}

// FindNext finds the next occurrence of a given string in the buffer
// It returns the start and end location of the match (if found) and
// a boolean indicating if it was found
// May also return an error if the search regex is invalid
func (b *Buffer) FindNext(s string, from Loc, down bool) ([2]Loc, bool, error) {
	if s == "" {
		return [2]Loc{}, false, nil
	}

	var r *regexp.Regexp
	var err error
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
		l, found = b.findDown(r, from, b.End())
		if !found {
			l, found = b.findDown(r, b.Start(), from)
		}
	} else {
		l, found = b.findUp(r, from, b.Start())
		if !found {
			l, found = b.findUp(r, b.End(), from)
		}
	}
	return l, found, nil
}
