package buffer

import (
	"regexp"
	"unicode/utf8"

	"github.com/zyedidia/micro/v2/internal/util"
)

// RegexpGroup combines a Regexp with padded versions.
// We want "^" and "$" to match only the beginning/end of a line, not that
// of the search region somewhere in the middle of a line. In that case we
// use padded regexps to require a rune before or after the match. (This
// also affects other empty-string patters like "\\b".)
type RegexpGroup [4]*regexp.Regexp

const (
	padStart = 1 << iota
	padEnd
)

// NewRegexpGroup creates a RegexpGroup from a string
func NewRegexpGroup(s string) (RegexpGroup, error) {
	var rgrp RegexpGroup
	var err error
	rgrp[0], err = regexp.Compile(s)
	if err == nil {
		rgrp[padStart] = regexp.MustCompile(".(?:" + s + ")")
		rgrp[padEnd] = regexp.MustCompile("(?:" + s + ").")
		rgrp[padStart|padEnd] = regexp.MustCompile(".(?:" + s + ").")
	}
	return rgrp, err
}

type bytesFind func(*regexp.Regexp, []byte) []int

func (b *Buffer) findDownFunc(rgrp RegexpGroup, start, end Loc, find bytesFind) []Loc {
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
		from, to := 0, len(l)
		padMode := 0

		if i == end.Y {
			nchars := util.CharacterCount(l)
			end.X = util.Clamp(end.X, 0, nchars)
			if end.X < nchars {
				padMode |= padEnd
				to = util.NextRunePos(l, util.BytePosFromCharPos(l, end.X))
			}
		}

		if i == start.Y {
			nchars := util.CharacterCount(l)
			start.X = util.Clamp(start.X, 0, nchars)
			if start.X > 0 {
				padMode |= padStart
				from = util.PreviousRunePos(l, util.BytePosFromCharPos(l, start.X))
			}
		}

		s := l[from:to]
		match := find(rgrp[padMode], s)

		if match != nil {
			if padMode&padStart != 0 {
				match[0] = util.NextRunePos(s, match[0])
			}
			if padMode&padEnd != 0 {
				match[1] = util.PreviousRunePos(s, match[1])
			}
			return util.RangeMap(match, func(j, pos int) Loc {
				if pos >= 0 {
					x := util.CharacterCount(l[:from+pos])
					if j % 2 == 0 {
						r, _ := utf8.DecodeRune(s[pos:])
						if util.IsMark(r) {
							x--
						}
					}
					return Loc{x, i}
				} else { // start or end of unused submatch
					return LocVoid()
				}
			})
		}
	}
	return nil
}

type bufferFind func(*Buffer, RegexpGroup, Loc, Loc) []Loc

// FindDown returns a slice containing the start and end positions
// of the first match of `rgrp` between `start` and `end`, or nil
// if no match exists.
func (b *Buffer) FindDown(rgrp RegexpGroup, start, end Loc) []Loc {
	return b.findDownFunc(rgrp, start, end, (*regexp.Regexp).FindIndex)
}

// FindDownSubmatch returns a slice containing the start and end positions
// of the first match of `rgrp` between `start` and `end` plus those
// of all submatches (capturing groups), or nil if no match exists.
// The start and end positions of an unused submatch are void.
func (b *Buffer) FindDownSubmatch(rgrp RegexpGroup, start, end Loc) []Loc {
	return b.findDownFunc(rgrp, start, end, (*regexp.Regexp).FindSubmatchIndex)
}

func (b *Buffer) findUpFunc(rgrp RegexpGroup, start, end Loc, find bytesFind) []Loc {
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

	var locs []Loc
	for i := end.Y; i >= start.Y; i-- {
		charCount := util.CharacterCount(b.LineBytes(i))
		from := Loc{0, i}.Clamp(start, end)
		to := Loc{charCount, i}.Clamp(start, end)

		b.findAllFuncFunc(rgrp, from, to, func(b *Buffer, rgrp RegexpGroup, start, end Loc) []Loc {
			return b.findDownFunc(rgrp, start, end, find)
		}, func(match []Loc) {
			locs = match
		})

		if locs != nil {
			return locs
		}
	}
	return nil
}

// FindUp returns a slice containing the start and end positions
// of the last match of `rgrp` between `start` and `end`, or nil
// if no match exists.
func (b *Buffer) FindUp(rgrp RegexpGroup, start, end Loc) []Loc {
	return b.findUpFunc(rgrp, start, end, func(re *regexp.Regexp, l []byte) []int {
		allMatches := re.FindAllIndex(l, -1)
		if allMatches != nil {
			return allMatches[len(allMatches)-1]
		} else {
			return nil
		}
	})
}

// FindUpSubmatch returns a slice containing the start and end positions
// of the last match of `rgrp` between `start` and `end` plus those
// of all submatches (capturing groups), or nil if no match exists.
// The start and end positions of an unused submatch are void.
func (b *Buffer) FindUpSubmatch(rgrp RegexpGroup, start, end Loc) []Loc {
	return b.findUpFunc(rgrp, start, end, func(re *regexp.Regexp, l []byte) []int {
		allMatches := re.FindAllSubmatchIndex(l, -1)
		if allMatches != nil {
			return allMatches[len(allMatches)-1]
		} else {
			return nil
		}
	})
}

func (b *Buffer) findAllFuncFunc(r RegexpGroup, start, end Loc, find bufferFind, f func([]Loc)) int {
	n := 0
	loc := start
	for {
		match := find(b, r, loc, end)
		if match == nil {
			break
		}
		n++
		f(match)
		if match[0] != match[1] {
			loc = match[1]
		} else if match[1] != end {
			loc = match[1].Move(1, b)
		} else {
			break
		}
	}
	return n
}

// FindAllFunc calls the function `f` once for each match between `start`
// and `end` of the regexp given by `s`. The argument of `f` is the slice
// containing the start and end positions of the match. FindAllFunc returns
// the number of matches plus any error that occured when compiling the regexp.
func (b *Buffer) FindAllFunc(s string, start, end Loc, f func([]Loc)) (int, error) {
	rgrp, err := NewRegexpGroup(s)
	if err == nil {
		return b.findAllFuncFunc(rgrp, start, end, (*Buffer).FindDown, f), nil
	} else {
		return -1, err
	}
}

// FindAll returns a slice containing the start and end positions of all
// matches between `start` and `end` of the regexp given by `s`, plus any
// error that occured when compiling the regexp. If no match is found, the
// slice returned is nil.
func (b *Buffer) FindAll(s string, start, end Loc) ([][]Loc, error) {
	var matches [][]Loc
	_, err := b.FindAllFunc(s, start, end, func(match []Loc) {
		matches = append(matches, match)
	})
	return matches, err
}

// FindAllSubmatchFunc calls the function `f` once for each match between
// `start` and `end` of the regexp given by `s`. The argument of `f` is the
// slice containing the start and end positions of the match and all submatches
// (capturing groups). FindAllSubmatch Func returns the number of matches plus
// any error that occured when compiling the regexp.
func (b *Buffer) FindAllSubmatchFunc(s string, start, end Loc, f func([]Loc)) (int, error) {
	rgrp, err := NewRegexpGroup(s)
	if err == nil {
		return b.findAllFuncFunc(rgrp, start, end, (*Buffer).FindDownSubmatch, f), nil
	} else {
		return -1, err
	}
}

// FindAllSubmatch returns a slice containing the start and end positions of
// all matches and all submatches (capturing groups) between `start` and `end`
// of the regexp given by `s`, plus any error that occured when compiling
// the regexp. If no match is found, the slice returned is nil.
func (b *Buffer) FindAllSubmatch(s string, start, end Loc) ([][]Loc, error) {
	var matches [][]Loc
	_, err := b.FindAllSubmatchFunc(s, start, end, func(match []Loc) {
		matches = append(matches, match)
	})
	return matches, err
}

// MatchedStrings converts a slice containing start and end positions of
// matches or submatches to a slice containing the corresponding strings.
// Unused submatches are converted to empty strings.
func (b *Buffer) MatchedStrings(locs []Loc) []string {
	strs := make([]string, len(locs)/2)
	for i := 0; 2*i < len(locs); i += 2 {
		if !locs[2*i].IsVoid() {
			strs[i] = string(b.Substr(locs[2*i], locs[2*i+1]))
		}
	}
	return strs
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

	rgrp, err := NewRegexpGroup(s)
	if err != nil {
		return [2]Loc{}, false, err
	}

	var match []Loc
	if down {
		match = b.FindDown(rgrp, from, end)
		if match == nil {
			match = b.FindDown(rgrp, start, end)
		}
	} else {
		match = b.FindUp(rgrp, from, start)
		if match == nil {
			match = b.FindUp(rgrp, end, start)
		}
	}
	if match != nil {
		return [2]Loc{match[0], match[1]}, true, nil
	} else {
		return [2]Loc{}, false, nil
	}
}

func (b *Buffer) replaceAllFuncFunc(s string, start, end Loc, find bufferFind, repl func(match []Loc) []byte) (int, Loc, error) {
	rgrp, err := NewRegexpGroup(s)
	if err != nil {
		return -1, LocVoid(), err
	}

	charsEnd := util.CharacterCount(b.LineBytes(end.Y))
	var deltas []Delta

	n := b.findAllFuncFunc(rgrp, start, end, find, func(match []Loc) {
		deltas = append(deltas, Delta{repl(match), match[0], match[1]})
	})
	b.MultipleReplace(deltas)

	deltaX := util.CharacterCount(b.LineBytes(end.Y)) - charsEnd
	return n, Loc{end.X + deltaX, end.Y}, nil
}

// ReplaceAll replaces all matches of the regexp `s` in the given area. The
// new text is obtained from `template` by replacing each variable with the
// corresponding submatch as in `Regexp.Expand`. The function returns the
// number of replacements made, the new end position and any error that
// occured during regexp compilation
func (b *Buffer) ReplaceAll(s string, start, end Loc, template []byte) (int, Loc, error) {
	var replace []byte

	find := func(b *Buffer, r RegexpGroup, start, end Loc) []Loc {
		return b.findDownFunc(r, start, end, func(re *regexp.Regexp, l []byte) []int {
			match := re.FindSubmatchIndex(l)
			if match == nil {
				return nil
			}
			replace = re.Expand(nil, template, l, match)
			return match[:2] // this way match[2:] is not transformed to Loc's
		})
	}

	return b.replaceAllFuncFunc(s, start, end, find, func(match []Loc) []byte {
		return replace
	})
}

// ReplaceAllLiteral replaces all matches of the regexp `s` with `repl` in
// the given area. The function returns the number of replacements made, the
// new end position and any error that occured during regexp compilation
func (b *Buffer) ReplaceAllLiteral(s string, start, end Loc, repl []byte) (int, Loc, error) {
	return b.ReplaceAllFunc(s, start, end, func([]Loc) []byte {
		return repl
	})
}

// ReplaceAllFunc replaces all matches of the regexp `s` with `repl(match)`
// in the given area, where `match` is the slice containing start and end
// positions of the match. The function returns the number of replacements
// made, the new end position and any error that occured during regexp
// compilation
func (b *Buffer) ReplaceAllFunc(s string, start, end Loc, repl func(match []Loc) []byte) (int, Loc, error) {
	return b.replaceAllFuncFunc(s, start, end, (*Buffer).FindDown, repl)
}

// ReplaceAllSubmatchFunc replaces all matches of the regexp `s` with
// `repl(match)` in the given area, where `match` is the slice containing
// start and end positions of the match and all submatches. The function
// returns the number of replacements made, the new end position and any
// error that occured during regexp compilation
func (b *Buffer) ReplaceAllSubmatchFunc(s string, start, end Loc, repl func(match []Loc) []byte) (int, Loc, error) {
	return b.replaceAllFuncFunc(s, start, end, (*Buffer).FindDownSubmatch, repl)
}
