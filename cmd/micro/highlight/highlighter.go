package highlight

import (
	"regexp"
	"strings"
)

func combineLineMatch(src, dst LineMatch) LineMatch {
	for k, v := range src {
		if g, ok := dst[k]; ok {
			if g == 0 {
				dst[k] = v
			}
		} else {
			dst[k] = v
		}
	}
	return dst
}

// A State represents the region at the end of a line
type State *Region

// LineStates is an interface for a buffer-like object which can also store the states and matches for every line
type LineStates interface {
	LineData() [][]byte
	State(lineN int) State
	SetState(lineN int, s State)
	SetMatch(lineN int, m LineMatch)
}

// A Highlighter contains the information needed to highlight a string
type Highlighter struct {
	lastRegion *Region
	def        *Def
}

// NewHighlighter returns a new highlighter from the given syntax definition
func NewHighlighter(def *Def) *Highlighter {
	h := new(Highlighter)
	h.def = def
	return h
}

// LineMatch represents the syntax highlighting matches for one line. Each index where the coloring is changed is marked with that
// color's group (represented as one byte)
type LineMatch map[int]uint8

func findIndex(regex *regexp.Regexp, str []byte, canMatchStart, canMatchEnd bool) []int {
	regexStr := regex.String()
	if strings.Contains(regexStr, "^") {
		if !canMatchStart {
			return nil
		}
	}
	if strings.Contains(regexStr, "$") {
		if !canMatchEnd {
			return nil
		}
	}
	return regex.FindIndex(str)
}

func findAllIndex(regex *regexp.Regexp, str []byte, canMatchStart, canMatchEnd bool) [][]int {
	regexStr := regex.String()
	if strings.Contains(regexStr, "^") {
		if !canMatchStart {
			return nil
		}
	}
	if strings.Contains(regexStr, "$") {
		if !canMatchEnd {
			return nil
		}
	}
	return regex.FindAllIndex(str, -1)
}

func (h *Highlighter) highlightRegion(start int, canMatchEnd bool, lineNum int, line []byte, region *Region) LineMatch {
	fullHighlights := make([]uint8, len(line))
	for i := 0; i < len(fullHighlights); i++ {
		fullHighlights[i] = region.group
	}

	highlights := make(LineMatch)

	if start == 0 {
		highlights[0] = region.group
	}

	loc := findIndex(region.end, line, start == 0, canMatchEnd)
	if loc != nil {
		if region.parent == nil {
			highlights[start+loc[1]] = 0
			return combineLineMatch(highlights,
				combineLineMatch(h.highlightRegion(start, false, lineNum, line[:loc[0]], region),
					h.highlightEmptyRegion(start+loc[1], canMatchEnd, lineNum, line[loc[1]:])))
		}
		highlights[start+loc[1]] = region.parent.group
		return combineLineMatch(highlights,
			combineLineMatch(h.highlightRegion(start, false, lineNum, line[:loc[0]], region),
				h.highlightRegion(start+loc[1], canMatchEnd, lineNum, line[loc[1]:], region.parent)))
	}

	if len(line) == 0 {
		if canMatchEnd {
			h.lastRegion = region
		}

		return highlights
	}

	firstLoc := []int{len(line), 0}
	var firstRegion *Region
	for _, r := range region.rules.regions {
		loc := findIndex(r.start, line, start == 0, canMatchEnd)
		if loc != nil {
			if loc[0] < firstLoc[0] {
				firstLoc = loc
				firstRegion = r
			}
		}
	}
	if firstLoc[0] != len(line) {
		highlights[start+firstLoc[0]] = firstRegion.group
		return combineLineMatch(highlights,
			combineLineMatch(h.highlightRegion(start, false, lineNum, line[:firstLoc[0]], region),
				h.highlightRegion(start+firstLoc[1], canMatchEnd, lineNum, line[firstLoc[1]:], firstRegion)))
	}

	for _, p := range region.rules.patterns {
		matches := findAllIndex(p.regex, line, start == 0, canMatchEnd)
		for _, m := range matches {
			for i := m[0]; i < m[1]; i++ {
				fullHighlights[i] = p.group
			}
		}
	}
	for i, h := range fullHighlights {
		if i == 0 || h != fullHighlights[i-1] {
			if _, ok := highlights[start+i]; !ok {
				highlights[start+i] = h
			}
		}
	}

	if canMatchEnd {
		h.lastRegion = region
	}

	return highlights
}

func (h *Highlighter) highlightEmptyRegion(start int, canMatchEnd bool, lineNum int, line []byte) LineMatch {
	fullHighlights := make([]uint8, len(line))
	highlights := make(LineMatch)
	if len(line) == 0 {
		if canMatchEnd {
			h.lastRegion = nil
		}
		return highlights
	}

	firstLoc := []int{len(line), 0}
	var firstRegion *Region
	for _, r := range h.def.rules.regions {
		loc := findIndex(r.start, line, start == 0, canMatchEnd)
		if loc != nil {
			if loc[0] < firstLoc[0] {
				firstLoc = loc
				firstRegion = r
			}
		}
	}
	if firstLoc[0] != len(line) {
		highlights[start+firstLoc[0]] = firstRegion.group
		return combineLineMatch(highlights,
			combineLineMatch(h.highlightEmptyRegion(start, false, lineNum, line[:firstLoc[0]]),
				h.highlightRegion(start+firstLoc[1], canMatchEnd, lineNum, line[firstLoc[1]:], firstRegion)))
	}

	for _, p := range h.def.rules.patterns {
		matches := findAllIndex(p.regex, line, start == 0, canMatchEnd)
		for _, m := range matches {
			for i := m[0]; i < m[1]; i++ {
				fullHighlights[i] = p.group
			}
		}
	}
	for i, h := range fullHighlights {
		if i == 0 || h != fullHighlights[i-1] {
			if _, ok := highlights[start+i]; !ok {
				highlights[start+i] = h
			}
		}
	}

	if canMatchEnd {
		h.lastRegion = nil
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

		if i == 0 || h.lastRegion == nil {
			lineMatches = append(lineMatches, h.highlightEmptyRegion(0, true, i, line))
		} else {
			lineMatches = append(lineMatches, h.highlightRegion(0, true, i, line, h.lastRegion))
		}
	}

	return lineMatches
}

// HighlightStates correctly sets all states for the buffer
func (h *Highlighter) HighlightStates(input LineStates) {
	lines := input.LineData()

	for i := 0; i < len(lines); i++ {
		line := []byte(lines[i])

		if i == 0 || h.lastRegion == nil {
			h.highlightEmptyRegion(0, true, i, line)
		} else {
			h.highlightRegion(0, true, i, line, h.lastRegion)
		}

		curState := h.lastRegion

		input.SetState(i, curState)
	}
}

// HighlightMatches sets the matches for each line in between startline and endline
// It sets all other matches in the buffer to nil to conserve memory
// This assumes that all the states are set correctly
func (h *Highlighter) HighlightMatches(input LineStates, startline, endline int) {
	lines := input.LineData()

	for i := 0; i < len(lines); i++ {
		if i >= startline && i < endline {
			line := []byte(lines[i])

			var match LineMatch
			if i == 0 || input.State(i-1) == nil {
				match = h.highlightEmptyRegion(0, true, i, line)
			} else {
				match = h.highlightRegion(0, true, i, line, input.State(i-1))
			}

			input.SetMatch(i, match)
		} else {
			input.SetMatch(i, nil)
		}
	}
}

// ReHighlightStates will scan down from `startline` and set the appropriate end of line state
// for each line until it comes across the same state in two consecutive lines
func (h *Highlighter) ReHighlightStates(input LineStates, startline int) {
	lines := input.LineData()

	h.lastRegion = nil
	if startline > 0 {
		h.lastRegion = input.State(startline - 1)
	}
	for i := startline; i < len(lines); i++ {
		line := []byte(lines[i])

		// var match LineMatch
		if i == 0 || h.lastRegion == nil {
			h.highlightEmptyRegion(0, true, i, line)
		} else {
			h.highlightRegion(0, true, i, line, h.lastRegion)
		}
		curState := h.lastRegion
		lastState := input.State(i)

		input.SetState(i, curState)

		if curState == lastState {
			break
		}
	}
}

// ReHighlightLine will rehighlight the state and match for a single line
func (h *Highlighter) ReHighlightLine(input LineStates, lineN int) {
	lines := input.LineData()

	line := []byte(lines[lineN])

	h.lastRegion = nil
	if lineN > 0 {
		h.lastRegion = input.State(lineN - 1)
	}

	var match LineMatch
	if lineN == 0 || h.lastRegion == nil {
		match = h.highlightEmptyRegion(0, true, lineN, line)
	} else {
		match = h.highlightRegion(0, true, lineN, line, h.lastRegion)
	}
	curState := h.lastRegion

	input.SetMatch(lineN, match)
	input.SetState(lineN, curState)
}
