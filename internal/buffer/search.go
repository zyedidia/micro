package buffer

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/zyedidia/micro/v2/internal/util"
)

// RegexpGroup combines a Regexp with padded versions.
type RegexpData struct {
	// We want "^" and "$" to match only the beginning/end of a line, not that
	// of the search region somewhere in the middle of a line. In that case we
	// use padded regexps to require a rune before or after the match. (This
	// also affects other empty-string patters like "\\b".)
	regex [4]*regexp.Regexp
}

// Regexp returns the Regexp determining the RegexpData
func (redata *RegexpData) Regexp() *regexp.Regexp {
	return redata.regex[0]
}

const (
	padStart = 1 << iota
	padEnd
)

// NewRegexpData creates RegexpData from a string
func NewRegexpData(s string) (*RegexpData, error) {
	var regex [4]*regexp.Regexp
	var err error
	regex[0], err = regexp.Compile(s)
	if err == nil {
		s_e := s + `\E`
		_, err_e := regexp.Compile(s_e)
		if err_e == nil {
			s = s_e
		}
		regex[padStart] = regexp.MustCompile(".(?:" + s + ")")
		regex[padEnd] = regexp.MustCompile("(?:" + s + ").")
		regex[padStart|padEnd] = regexp.MustCompile(".(?:" + s + ").")
	}
	return &RegexpData{regex}, err
}

func regexpData(re any) (*RegexpData, error) {
	switch re := re.(type) {
	case *RegexpData:
		return re, nil
	case string:
		return NewRegexpData(re)
	default:
		return &RegexpData{}, fmt.Errorf(`cannot convert "%v" (of type %[1]T) to type RegexpData`, re)
	}
}

type bytesFind func(*regexp.Regexp, []byte) []int

func (b *Buffer) findDownFunc(redata *RegexpData, start, end Loc, find bytesFind) []Loc {
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
		match := find(redata.regex[padMode], s)

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
					if j%2 == 0 {
						r, _ := utf8.DecodeRune(s[pos:])
						if util.IsMark(r) {
							x--
						}
					}
					return Loc{x, i}
				} else { // start or end of unused submatch
					return Loc{-1, -1}
				}
			})
		}
	}
	return nil
}

type bufferFind func(*Buffer, *RegexpData, Loc, Loc) []Loc

// FindDown returns a slice containing the start and end positions
// of the first match of `re` between `start` and `end` plus those
// of all submatches (capturing groups), or nil if no match exists.
// The start and end positions of an unused submatch are invalid.
func (b *Buffer) FindDown(re string, start, end Loc) ([]Loc, error) {
	redata, err := NewRegexpData(re)
	if err != nil {
		return nil, err
	}
	return b.FindRegexpDown(redata, start, end), nil
}

func (b *Buffer) FindRegexpDown(redata *RegexpData, start, end Loc) []Loc {
	if start.GreaterThan(end) {
		return nil
	}
	return b.findDownFunc(redata, start, end, (*regexp.Regexp).FindSubmatchIndex)
}

// FindUp returns a slice containing the start and end positions
// of the last match of `re` between `start` and `end` plus those
// of all submatches (capturing groups), or nil if no match exists.
// The start and end positions of an unused submatch are invalid.
func (b *Buffer) FindUp(re string, start, end Loc) ([]Loc, error) {
	redata, err := NewRegexpData(re)
	if err != nil {
		return nil, err
	}
	return b.FindRegexpUp(redata, start, end), nil
}

func (b *Buffer) FindRegexpUp(redata *RegexpData, start, end Loc) []Loc {
	if start.GreaterThan(end) {
		return nil
	}

	var locs []Loc
	for i := end.Y; i >= start.Y; i-- {
		charCount := util.CharacterCount(b.LineBytes(i))
		from := Loc{0, i}.Clamp(start, end)
		to := Loc{charCount, i}.Clamp(start, end)

		b.findAllFuncFunc(redata, from, to, func(b *Buffer, redata *RegexpData, start, end Loc) []Loc {
			return b.findDownFunc(redata, start, end, func(r *regexp.Regexp, l []byte) []int {
				allMatches := r.FindAllSubmatchIndex(l, -1)
				if allMatches != nil {
					return allMatches[len(allMatches)-1]
				} else {
					return nil
				}
			})
		}, func(match []Loc) {
			locs = match
		})

		if locs != nil {
			return locs
		}
	}
	return nil
}

func (b *Buffer) findAllFuncFunc(redata *RegexpData, start, end Loc, find bufferFind, f func([]Loc)) int {
	n := 0
	loc := start
	for {
		match := find(b, redata, loc, end)
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

// FindAllFunc calls the function `f` once for each match between
// `start` and `end` of the regexp given by `re`. The argument of `f` is the
// slice containing the start and end positions of the match and all submatches
// (capturing groups). FindAllFunc returns the number of matches plus
// any error that occured when compiling the regexp.
func (b *Buffer) FindAllFunc(re string, start, end Loc, f func([]Loc)) (int, error) {
	redata, err := NewRegexpData(re)
	if err != nil {
		return -1, err
	}
	return b.findAllFuncFunc(redata, start, end, (*Buffer).FindRegexpDown, f), nil
}

// FindAll returns a slice containing the start and end positions of
// all matches and all submatches (capturing groups) between `start` and `end`
// of the regexp given by `re`, plus any error that occured when compiling
// the regexp. If no match is found, the slice returned is nil.
func (b *Buffer) FindAll(re string, start, end Loc) ([][]Loc, error) {
	var matches [][]Loc
	_, err := b.FindAllFunc(re, start, end, func(match []Loc) {
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
		if locs[2*i].IsValid() {
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

	redata, err := NewRegexpData(s)
	if err != nil {
		return [2]Loc{}, false, err
	}

	if start.GreaterThan(end) {
		start, end = end, start
	}

	var match []Loc
	if down {
		match = b.FindRegexpDown(redata, from, end)
		if match == nil {
			match = b.FindRegexpDown(redata, start, end)
		}
	} else {
		match = b.FindRegexpUp(redata, start, from)
		if match == nil {
			match = b.FindRegexpUp(redata, start, end)
		}
	}
	if match != nil {
		return [2]Loc{match[0], match[1]}, true, nil
	} else {
		return [2]Loc{}, false, nil
	}
}

// Expand returns the template, with variables replaced by submatches.
// It is analogous to `(*regexp.Regexp).Expand`
func (b *Buffer) Expand(re any, template []byte, match []Loc) (string, error) {
	redata, err := regexpData(re)
	if err != nil {
		return "", err
	}
	l := b.LineBytes(match[0].Y)
	m := util.RangeMap(match, func(_ int, pos Loc) int {
		return util.BytePosFromCharPos(l, pos.X)
	})
	return string(redata.Regexp().Expand(nil, template, l, m)), nil
}

func (b *Buffer) replaceAllFuncFunc(re string, start, end Loc, find bufferFind, repl func(match []Loc) []byte) (int, Loc, error) {
	redata, err := NewRegexpData(re)
	if err != nil {
		return -1, Loc{-1, -1}, err
	}

	charsEnd := util.CharacterCount(b.LineBytes(end.Y))
	var deltas []Delta

	n := b.findAllFuncFunc(redata, start, end, find, func(match []Loc) {
		deltas = append(deltas, Delta{repl(match), match[0], match[1]})
	})

	b.MultipleReplace(deltas)

	deltaX := util.CharacterCount(b.LineBytes(end.Y)) - charsEnd
	return n, Loc{end.X + deltaX, end.Y}, nil
}

// ReplaceAll replaces all matches of the regexp `re` in the given area. The
// new text is obtained from `template` by replacing each variable with the
// corresponding submatch as in `(*regexp.Regexp).Expand`. The function
// returns the number of replacements made, the new end position and any
// error that occured during regexp compilation
func (b *Buffer) ReplaceAll(re string, start, end Loc, template []byte) (int, Loc, error) {
	var replace []byte

	find := func(b *Buffer, redata *RegexpData, start, end Loc) []Loc {
		return b.findDownFunc(redata, start, end, func(re *regexp.Regexp, l []byte) []int {
			match := re.FindSubmatchIndex(l)
			if match == nil {
				return nil
			}
			replace = re.Expand(nil, template, l, match)
			return match[:2] // this way match[2:] is not transformed to Loc's
		})
	}

	return b.replaceAllFuncFunc(re, start, end, find, func(match []Loc) []byte {
		return replace
	})
}

// ReplaceAllLiteral replaces all matches of the regexp `re` with `repl` in
// the given area. The function returns the number of replacements made, the
// new end position and any error that occured during regexp compilation
func (b *Buffer) ReplaceAllLiteral(re string, start, end Loc, repl []byte) (int, Loc, error) {
	return b.ReplaceAllFunc(re, start, end, func([]Loc) []byte {
		return repl
	})
}

// ReplaceAllFunc replaces all matches of the regexp `re` with
// `repl(match)` in the given area, where `match` is the slice containing
// start and end positions of the match and all submatches. The function
// returns the number of replacements made, the new end position and any
// error that occured during regexp compilation
func (b *Buffer) ReplaceAllFunc(re string, start, end Loc, repl func(match []Loc) []byte) (int, Loc, error) {
	return b.replaceAllFuncFunc(re, start, end, (*Buffer).FindRegexpDown, repl)
}
