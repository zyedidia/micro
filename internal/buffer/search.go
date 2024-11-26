package buffer

import (
	"regexp"

	"github.com/zyedidia/micro/v2/internal/util"
)

func (b *Buffer) findDown(r *regexp.Regexp, start, end Loc) ([]Loc, bool) {
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
		l := b.LineBytes(i)
		charpos := 0

		if i == start.Y && start.Y == end.Y {
			nchars := util.CharacterCount(l)
			start.X = util.Clamp(start.X, 0, nchars)
			end.X = util.Clamp(end.X, 0, nchars)
			l = util.SliceStart(l, end.X)
			l = util.SliceEnd(l, start.X)
			charpos = start.X
		} else if i == start.Y {
			nchars := util.CharacterCount(l)
			start.X = util.Clamp(start.X, 0, nchars)
			l = util.SliceEnd(l, start.X)
			charpos = start.X
		} else if i == end.Y {
			nchars := util.CharacterCount(l)
			end.X = util.Clamp(end.X, 0, nchars)
			l = util.SliceStart(l, end.X)
		}

		match := r.FindSubmatchIndex(l)

		if match != nil {
			loc := make([]Loc, len(match))
			for j, x := range match {
				if x == -1 { // submatch doesn't occur in match
					loc[j] = Loc{-1, -1}
				} else {
					loc[j] = Loc{charpos + util.RunePos(l, x), i}
				}
			}
			return loc, true
		}
	}
	return []Loc{}, false
}

func (b *Buffer) findUp(r *regexp.Regexp, start, end Loc) ([]Loc, bool) {
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
		l := b.LineBytes(i)
		charpos := 0

		if i == start.Y && start.Y == end.Y {
			nchars := util.CharacterCount(l)
			start.X = util.Clamp(start.X, 0, nchars)
			end.X = util.Clamp(end.X, 0, nchars)
			l = util.SliceStart(l, end.X)
			l = util.SliceEnd(l, start.X)
			charpos = start.X
		} else if i == start.Y {
			nchars := util.CharacterCount(l)
			start.X = util.Clamp(start.X, 0, nchars)
			l = util.SliceEnd(l, start.X)
			charpos = start.X
		} else if i == end.Y {
			nchars := util.CharacterCount(l)
			end.X = util.Clamp(end.X, 0, nchars)
			l = util.SliceStart(l, end.X)
		}

		allMatches := r.FindAllSubmatchIndex(l, -1)

		if allMatches != nil {
			match := allMatches[len(allMatches)-1]
			loc := make([]Loc, len(match))
			for j, x := range match {
				if x == -1 { // submatch doesn't occur in match
					loc[j] = Loc{-1, -1}
				} else {
					loc[j] = Loc{charpos + util.RunePos(l, x), i}
				}
			}
			return loc, true
		}
	}
	return []Loc{}, false
}

// FindNextSubmatch finds the next occurrence of a given string in the
// buffer. It returns the start and end location of the match and of
// all submatches (if found) and a boolean indicating if it was found.
// May also return an error if the search regex is invalid
func (b *Buffer) FindNextSubmatch(s string, start, end, from Loc, down bool) ([]Loc, bool, error) {
	if s == "" {
		return []Loc{}, false, nil
	}

	var r *regexp.Regexp
	var err error

	if b.Settings["ignorecase"].(bool) {
		r, err = regexp.Compile("(?i)" + s)
	} else {
		r, err = regexp.Compile(s)
	}

	if err != nil {
		return []Loc{}, false, err
	}

	var found bool
	var l []Loc
	if down {
		l, found = b.findDown(r, from, end)
		if !found {
			l, found = b.findDown(r, start, end)
		}
	} else {
		l, found = b.findUp(r, from, start)
		if !found {
			l, found = b.findUp(r, end, start)
		}
	}
	return l, found, nil
}

// FindNext finds the next occurrence of a given string in the buffer
// It returns the start and end location of the match (if found) and
// a boolean indicating if it was found
// May also return an error if the search regex is invalid
func (b *Buffer) FindNext(s string, start, end, from Loc, down bool, useRegex bool) ([2]Loc, bool, error) {
	if !useRegex {
		s = regexp.QuoteMeta(s)
	}
	l, found, err := b.FindNextSubmatch(s, start, end, from, down)
	if found {
		return [2]Loc{l[0], l[1]}, found, err
	} else {
		return [2]Loc{}, found, err
	}
}

// ReplaceRegex replaces all occurrences of 'search' with 'replace' in the given area
// and returns the number of replacements made and the number of runes
// added or removed on the last line of the range
func (b *Buffer) ReplaceRegex(start, end Loc, search *regexp.Regexp, replace []byte, captureGroups bool) (int, int) {
	if start.GreaterThan(end) {
		start, end = end, start
	}

	netrunes := 0

	found := 0
	var deltas []Delta
	for i := start.Y; i <= end.Y; i++ {
		l := b.lines[i].data
		charpos := 0

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
			var result []byte
			if captureGroups {
				for _, submatches := range search.FindAllSubmatchIndex(in, -1) {
					result = search.Expand(result, replace, in, submatches)
				}
			} else {
				result = replace
			}
			found++
			if i == end.Y {
				netrunes += util.CharacterCount(result) - util.CharacterCount(in)
			}
			return result
		})

		from := Loc{charpos, i}
		to := Loc{charpos + util.CharacterCount(l), i}

		deltas = append(deltas, Delta{newText, from, to})
	}
	b.MultipleReplace(deltas)

	return found, netrunes
}
