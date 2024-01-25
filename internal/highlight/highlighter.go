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
	lastStart  int
	lastEnd    int
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

func findAllIndex(regex *regexp.Regexp, skip *regexp.Regexp, str []byte) [][]int {
	var strbytes []byte
	if skip != nil {
		strbytes = skip.ReplaceAllFunc(str, func(match []byte) []byte {
			res := make([]byte, util.CharacterCount(match))
			return res
		})
	} else {
		strbytes = str
	}

	matches := regex.FindAllIndex(strbytes, -1)
	for i, m := range matches {
		matches[i][0] = util.RunePos(str, m[0])
		matches[i][1] = util.RunePos(str, m[1])
	}
	return matches
}

func (h *Highlighter) highlightRange(fullHighlights []Group, start int, end int, group Group) {
	if start <= end && end <= len(fullHighlights) {
		for i := start; i < end; i++ {
			fullHighlights[i] = group
		}
	}
}

func (h *Highlighter) highlightPatterns(fullHighlights []Group, start int, lineNum int, line []byte, curRegion *region) {
	lineLen := util.CharacterCount(line)
	if lineLen == 0 {
		return
	}

	var patterns []*pattern
	if curRegion == nil {
		patterns = h.Def.rules.patterns
	} else {
		patterns = curRegion.rules.patterns
	}

	for _, p := range patterns {
		matches := findAllIndex(p.regex, nil, line)
		for _, m := range matches {
			h.highlightRange(fullHighlights, start+m[0], start+m[1], p.group)
		}
	}
}

func (h *Highlighter) highlightRegions(fullHighlights []Group, start int, lineNum int, line []byte, curRegion *region, regions []*region, nestedRegion bool) {
	lineLen := util.CharacterCount(line)
	if lineLen == 0 {
		return
	}

	if nestedRegion {
		h.highlightPatterns(fullHighlights, start, lineNum, line, curRegion)
	} else {
		h.highlightPatterns(fullHighlights, start, lineNum, line, nil)
	}

	for _, r := range regions {
		if !nestedRegion && curRegion != nil && curRegion != r {
			continue
		}
		startMatches := findAllIndex(r.start, r.skip, line)
		endMatches := findAllIndex(r.end, r.skip, line)
		samePattern := false
	startLoop:
		for startIdx := 0; startIdx < len(startMatches); startIdx++ {
			startMatch := startMatches[startIdx]
			for endIdx := 0; endIdx < len(endMatches); endIdx++ {
				endMatch := endMatches[endIdx]
				if startMatch[0] == endMatch[0] {
					// start and end are the same (pattern)
					samePattern = true
					if len(startMatches) == len(endMatches) {
						// special case in the moment both are the same
						if curRegion == r {
							if len(startMatches) > 1 {
								// end < start
								continue startLoop
							} else if len(startMatches) > 0 {
								// ... end
								startIdx = len(startMatches)
								continue startLoop
							}
						} else {
							// start ... or start < end
						}
					}
				} else if startMatch[1] <= endMatch[0] {
					if !nestedRegion && h.lastStart < start+startMatch[0] && start+startMatch[0] < h.lastEnd {
						continue
					}
					// start and end at the current line
					update := false
					if h.lastStart == 0 || h.lastStart < start+endMatch[1] {
						h.lastStart = start + startMatch[0]
						h.lastEnd = start + endMatch[1]
						update = true
					}
					h.highlightRange(fullHighlights, start+startMatch[0], start+endMatch[1], r.limitGroup)
					h.highlightRegions(fullHighlights, start+startMatch[1], lineNum, util.SliceStartEnd(line, startMatch[1], endMatch[0]), r, r.rules.regions, true)
					if samePattern {
						startIdx += 1
					}
					if update {
						if curRegion != nil {
							h.lastRegion = r.parent
						} else {
							h.lastRegion = nil
						}
						curRegion = h.lastRegion
					}
					continue startLoop
				} else if endMatch[1] <= startMatch[0] {
					if start+endMatch[0] < h.lastEnd || curRegion == nil {
						continue
					}
					// start and end at the current line, but switched
					h.lastStart = start
					h.lastEnd = start + endMatch[1]
					h.highlightRange(fullHighlights, start, start+endMatch[1], r.limitGroup)
					h.highlightRegions(fullHighlights, start, lineNum, util.SliceStart(line, endMatch[0]), r, r.rules.regions, true)
					h.highlightPatterns(fullHighlights, start+endMatch[1], lineNum, util.SliceStartEnd(line, endMatch[1], startMatch[0]), nil)
					if curRegion != nil {
						h.lastRegion = r.parent
					} else {
						h.lastRegion = nil
					}
					curRegion = h.lastRegion
				}
			}
			if nestedRegion || start+startMatch[0] <= h.lastStart || h.lastEnd <= start+startMatch[0] {
				// start at the current line
				h.lastStart = start + startMatch[0]
				h.lastEnd = start + lineLen
				h.highlightRange(fullHighlights, start+startMatch[0], start+lineLen, r.limitGroup)
				h.highlightRegions(fullHighlights, start+startMatch[1], lineNum, util.SliceEnd(line, startMatch[1]), r, r.rules.regions, true)
				if h.lastStart == 0 || h.lastStart <= start+startMatch[0] {
					h.lastRegion = r
				}
			}
		}
		if curRegion == r {
			if (len(startMatches) == 0 && len(endMatches) > 0) || (samePattern && (len(startMatches) == len(endMatches))) {
				for _, endMatch := range endMatches {
					// end at the current line
					h.lastStart = start
					h.lastEnd = start + endMatch[1]
					h.highlightRange(fullHighlights, start, start+endMatch[1], r.limitGroup)
					h.highlightRegions(fullHighlights, start, lineNum, util.SliceStart(line, endMatch[0]), r, r.rules.regions, true)
					if curRegion != nil {
						h.lastRegion = r.parent
					} else {
						h.lastRegion = nil
					}
					curRegion = h.lastRegion
					h.highlightRegions(fullHighlights, start+endMatch[1], lineNum, util.SliceEnd(line, endMatch[1]), curRegion, h.Def.rules.regions, false)
					break
				}
			} else if len(startMatches) == 0 && len(endMatches) == 0 {
				// no start and end found in this region
				h.highlightRange(fullHighlights, start, start+lineLen, curRegion.group)
			}
		}
	}

	if curRegion != nil && !nestedRegion {
		// current region still open
		if curRegion.rules != nil {
			h.highlightRegions(fullHighlights, start, lineNum, line, curRegion, curRegion.rules.regions, true)
		}
		if curRegion == h.lastRegion && curRegion.parent != nil {
			var regions []*region
			regions = append(regions, curRegion)
			h.highlightRegions(fullHighlights, start, lineNum, line, curRegion, regions, true)
		}
	}
}

func (h *Highlighter) highlight(highlights LineMatch, start int, lineNum int, line []byte, curRegion *region) LineMatch {
	lineLen := util.CharacterCount(line)
	if lineLen == 0 {
		return highlights
	}

	h.lastStart = 0
	h.lastEnd = 0

	fullHighlights := make([]Group, lineLen)
	h.highlightRegions(fullHighlights, start, lineNum, line, curRegion, h.Def.rules.regions, false)

	for i, h := range fullHighlights {
		if i == 0 || h != fullHighlights[i-1] {
			highlights[i] = h
		}
	}

	return highlights
}

// HighlightString syntax highlights a string
// Use this function for simple syntax highlighting and use the other functions for
// more advanced syntax highlighting. They are optimized for quick rehighlighting of the same
// text with minor changes made
func (h *Highlighter) HighlightString(input string) []LineMatch {
	h.lastRegion = nil
	lines := strings.Split(input, "\n")
	var lineMatches []LineMatch

	for i := 0; i < len(lines); i++ {
		line := []byte(lines[i])
		highlights := make(LineMatch)
		lineMatches = append(lineMatches, h.highlight(highlights, 0, i, line, h.lastRegion))
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

		match := h.highlight(highlights, 0, i, line, h.lastRegion)

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

	match := h.highlight(highlights, 0, lineN, line, h.lastRegion)

	input.SetState(lineN, h.lastRegion)
	input.SetMatch(lineN, match)
}
