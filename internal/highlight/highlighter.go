package highlight

import (
	"regexp"
	"strings"

	"github.com/zyedidia/micro/v2/internal/util"
)

// A State represents the region at the end of a line
type State *region

// LineStates is an interface for a buffer-like object which can also store the states and matches for every line
type LineStates interface {
	LineBytes(n int) []byte
	LinesNum() int
	State(lineN int) State
	SetState(lineN int, s State)
	SetMatch(lineN int, m LineMatch)
	Lock()
	Unlock()
}

// A Highlighter contains the information needed to highlight a string
type Highlighter struct {
	lastRegion *region
	Def        *Def
}

// NewHighlighter returns a new highlighter from the given syntax definition
func NewHighlighter(def *Def) *Highlighter {
	h := new(Highlighter)
	h.Def = def
	return h
}

// LineMatch represents the syntax highlighting matches for one line. Each index where the coloring is changed is marked with that
// color's group (represented as one byte)
type LineMatch map[int]Group

func findIndex(regex *regexp.Regexp, skip *regexp.Regexp, str []byte) []int {
	var strbytes []byte
	if skip != nil {
		strbytes = skip.ReplaceAllFunc(str, func(match []byte) []byte {
			res := make([]byte, util.CharacterCount(match))
			return res
		})
	} else {
		strbytes = str
	}

	match := regex.FindIndex(strbytes)
	if match == nil {
		return nil
	}
	// return []int{match.Index, match.Index + match.Length}
	return []int{util.RunePos(str, match[0]), util.RunePos(str, match[1])}
}

func findAllIndex(regex *regexp.Regexp, str []byte) [][]int {
	matches := regex.FindAllIndex(str, -1)
	for i, m := range matches {
		matches[i][0] = util.RunePos(str, m[0])
		matches[i][1] = util.RunePos(str, m[1])
	}
	return matches
}

func (h *Highlighter) highlightRegion(highlights LineMatch, start int, lineNum int, line []byte, curRegion *region, statesOnly bool) LineMatch {
	lineLen := util.CharacterCount(line)
	if start == 0 {
		if !statesOnly {
			if _, ok := highlights[0]; !ok {
				highlights[0] = curRegion.group
			}
		}
	}

	var nestedRegion *region
	nestedLoc := []int{lineLen, 0}
	searchNesting := true
	endLoc := findIndex(curRegion.end, curRegion.skip, line)
	if endLoc != nil {
		if start == endLoc[0] {
			searchNesting = false
		} else {
			nestedLoc = endLoc
		}
	}
	if searchNesting {
		for _, r := range curRegion.rules.regions {
			loc := findIndex(r.start, r.skip, line)
			if loc != nil {
				if loc[0] < nestedLoc[0] {
					nestedLoc = loc
					nestedRegion = r
				}
			}
		}
	}
	if nestedRegion != nil && nestedLoc[0] != lineLen {
		if !statesOnly {
			highlights[start+nestedLoc[0]] = nestedRegion.limitGroup
		}
		slice := util.SliceEnd(line, nestedLoc[1])
		h.highlightEmptyRegion(highlights, start+nestedLoc[1], lineNum, slice, statesOnly)
		h.highlightRegion(highlights, start+nestedLoc[1], lineNum, slice, nestedRegion, statesOnly)
		return highlights
	}

	if !statesOnly {
		fullHighlights := make([]Group, lineLen)
		for i := 0; i < len(fullHighlights); i++ {
			fullHighlights[i] = curRegion.group
		}

		if searchNesting {
			for _, p := range curRegion.rules.patterns {
				if curRegion.group == curRegion.limitGroup || p.group == curRegion.limitGroup {
					matches := findAllIndex(p.regex, line)
					for _, m := range matches {
						if (endLoc == nil) || (m[0] < endLoc[0]) {
							for i := m[0]; i < m[1]; i++ {
								fullHighlights[i] = p.group
							}
						}
					}
				}
			}
		}
		for i, h := range fullHighlights {
			if i == 0 || h != fullHighlights[i-1] {
				highlights[start+i] = h
			}
		}
	}

	loc := endLoc
	if loc != nil {
		if !statesOnly {
			highlights[start+loc[0]] = curRegion.limitGroup
		}
		if curRegion.parent == nil {
			if !statesOnly {
				highlights[start+loc[1]] = 0
			}
			h.highlightEmptyRegion(highlights, start+loc[1], lineNum, util.SliceEnd(line, loc[1]), statesOnly)
			return highlights
		}
		if !statesOnly {
			highlights[start+loc[1]] = curRegion.parent.group
		}
		h.highlightRegion(highlights, start+loc[1], lineNum, util.SliceEnd(line, loc[1]), curRegion.parent, statesOnly)
		return highlights
	}

	h.lastRegion = curRegion

	return highlights
}

func (h *Highlighter) highlightEmptyRegion(highlights LineMatch, start int, lineNum int, line []byte, statesOnly bool) LineMatch {
	lineLen := util.CharacterCount(line)
	if lineLen == 0 {
		h.lastRegion = nil
		return highlights
	}

	var firstRegion *region
	firstLoc := []int{lineLen, 0}
	for _, r := range h.Def.rules.regions {
		loc := findIndex(r.start, r.skip, line)
		if loc != nil {
			if loc[0] < firstLoc[0] {
				firstLoc = loc
				firstRegion = r
			}
		}
	}
	if firstRegion != nil && firstLoc[0] != lineLen {
		if !statesOnly {
			highlights[start+firstLoc[0]] = firstRegion.limitGroup
		}
		h.highlightEmptyRegion(highlights, start, lineNum, util.SliceStart(line, firstLoc[0]), statesOnly)
		h.highlightRegion(highlights, start+firstLoc[1], lineNum, util.SliceEnd(line, firstLoc[1]), firstRegion, statesOnly)
		return highlights
	}

	if statesOnly {
		return highlights
	}

	fullHighlights := make([]Group, len(line))
	for _, p := range h.Def.rules.patterns {
		matches := findAllIndex(p.regex, line)
		for _, m := range matches {
			for i := m[0]; i < m[1]; i++ {
				fullHighlights[i] = p.group
			}
		}
	}
	for i, h := range fullHighlights {
		if i == 0 || h != fullHighlights[i-1] {
			// if _, ok := highlights[start+i]; !ok {
			highlights[start+i] = h
			// }
		}
	}

	return highlights
}

// HighlightString syntax highlights a string
// Use this function for simple syntax highlighting and use the other functions for
// more advanced syntax highlighting. They are optimized for quick rehighlighting of the same
// text with minor changes made
func (h *Highlighter) HighlightString(input string) []LineMatch {
	lines := strings.Split(input, "\n")
	var lineMatches []LineMatch

	for i := 0; i < len(lines); i++ {
		line := []byte(lines[i])
		highlights := make(LineMatch)

		if i == 0 || h.lastRegion == nil {
			lineMatches = append(lineMatches, h.highlightEmptyRegion(highlights, 0, i, line, false))
		} else {
			lineMatches = append(lineMatches, h.highlightRegion(highlights, 0, i, line, h.lastRegion, false))
		}
	}

	return lineMatches
}

// Highlight sets the state and matches for each line from startline to endline
// It sets all other matches in the buffer to nil to conserve memory
func (h *Highlighter) Highlight(input LineStates, startline, endline int) {
	h.lastRegion = nil
	if startline > 0 {
		input.Lock()
		if startline-1 < input.LinesNum() {
			h.lastRegion = input.State(startline - 1)
		}
		input.Unlock()
	}

	for i := startline; i <= endline; i++ {
		input.Lock()
		if i >= input.LinesNum() {
			input.Unlock()
			break
		}

		line := input.LineBytes(i)
		highlights := make(LineMatch)

		var match LineMatch
		if i == 0 || h.lastRegion == nil {
			match = h.highlightEmptyRegion(highlights, 0, i, line, false)
		} else {
			match = h.highlightRegion(highlights, 0, i, line, h.lastRegion, false)
		}

		input.SetState(i, h.lastRegion)
		input.SetMatch(i, match)
		input.Unlock()
	}
}

// ReHighlightLine will rehighlight the state and match for a single line
func (h *Highlighter) ReHighlightLine(input LineStates, lineN int) {
	input.Lock()
	defer input.Unlock()

	line := input.LineBytes(lineN)
	highlights := make(LineMatch)

	h.lastRegion = nil
	if lineN > 0 {
		h.lastRegion = input.State(lineN - 1)
	}

	var match LineMatch
	if lineN == 0 || h.lastRegion == nil {
		match = h.highlightEmptyRegion(highlights, 0, lineN, line, false)
	} else {
		match = h.highlightRegion(highlights, 0, lineN, line, h.lastRegion, false)
	}

	input.SetState(lineN, h.lastRegion)
	input.SetMatch(lineN, match)
}
