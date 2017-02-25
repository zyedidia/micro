package highlight

import (
	"regexp"
	"strings"
)

func combineLineMatch(src, dst LineMatch) LineMatch {
	for k, v := range src {
		if g, ok := dst[k]; ok {
			if g == "" {
				dst[k] = v
			}
		} else {
			dst[k] = v
		}
	}
	return dst
}

type State *Region

type LineStates interface {
	LineData() [][]byte
	State(lineN int) State
	SetState(lineN int, s State)
	SetMatch(lineN int, m LineMatch)
}

type Highlighter struct {
	lastRegion *Region
	def        *Def
}

func NewHighlighter(def *Def) *Highlighter {
	h := new(Highlighter)
	h.def = def
	return h
}

type LineMatch map[int]string

func FindIndex(regex *regexp.Regexp, str []byte, canMatchStart, canMatchEnd bool) []int {
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

func FindAllIndex(regex *regexp.Regexp, str []byte, canMatchStart, canMatchEnd bool) [][]int {
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
	highlights := make(LineMatch)

	if len(line) == 0 {
		if canMatchEnd {
			h.lastRegion = region
		}

		return highlights
	}

	loc := FindIndex(region.end, line, start == 0, canMatchEnd)
	if loc != nil {
		if region.parent == nil {
			highlights[start+loc[1]] = ""
			return combineLineMatch(highlights,
				combineLineMatch(h.highlightRegion(start, false, lineNum, line[:loc[0]], region),
					h.highlightEmptyRegion(start+loc[1], canMatchEnd, lineNum, line[loc[1]:])))
		}
		highlights[start+loc[1]] = region.parent.group
		return combineLineMatch(highlights,
			combineLineMatch(h.highlightRegion(start, false, lineNum, line[:loc[0]], region),
				h.highlightRegion(start+loc[1], canMatchEnd, lineNum, line[loc[1]:], region.parent)))
	}

	firstLoc := []int{len(line), 0}
	var firstRegion *Region
	for _, r := range region.rules.regions {
		loc := FindIndex(r.start, line, start == 0, canMatchEnd)
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
		matches := FindAllIndex(p.regex, line, start == 0, canMatchEnd)
		for _, m := range matches {
			highlights[start+m[0]] = p.group
			if _, ok := highlights[start+m[1]]; !ok {
				highlights[start+m[1]] = region.group
			}
		}
	}

	if canMatchEnd {
		h.lastRegion = region
	}

	return highlights
}

func (h *Highlighter) highlightEmptyRegion(start int, canMatchEnd bool, lineNum int, line []byte) LineMatch {
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
		loc := FindIndex(r.start, line, start == 0, canMatchEnd)
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
		matches := FindAllIndex(p.regex, line, start == 0, canMatchEnd)
		for _, m := range matches {
			highlights[start+m[0]] = p.group
			if _, ok := highlights[start+m[1]]; !ok {
				highlights[start+m[1]] = ""
			}
		}
	}

	if canMatchEnd {
		h.lastRegion = nil
	}

	return highlights
}

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

func (h *Highlighter) Highlight(input LineStates, startline int) {
	lines := input.LineData()

	for i := startline; i < len(lines); i++ {
		line := []byte(lines[i])

		var match LineMatch
		if i == 0 || h.lastRegion == nil {
			match = h.highlightEmptyRegion(0, true, i, line)
		} else {
			match = h.highlightRegion(0, true, i, line, h.lastRegion)
		}

		curState := h.lastRegion

		input.SetMatch(i, match)
		input.SetState(i, curState)
	}
}

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

func (h *Highlighter) ReHighlight(input LineStates, startline int) {
	lines := input.LineData()

	h.lastRegion = nil
	if startline > 0 {
		h.lastRegion = input.State(startline - 1)
	}
	for i := startline; i < len(lines); i++ {
		line := []byte(lines[i])

		var match LineMatch
		if i == 0 || h.lastRegion == nil {
			match = h.highlightEmptyRegion(0, true, i, line)
		} else {
			match = h.highlightRegion(0, true, i, line, h.lastRegion)
		}
		curState := h.lastRegion
		lastState := input.State(i)

		input.SetMatch(i, match)
		input.SetState(i, curState)

		if curState == lastState {
			break
		}
	}
}
