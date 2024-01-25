package highlight

import (
	"regexp"
	"strings"

	"github.com/zyedidia/micro/v2/internal/util"
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
type State *region

// EmptyDef is an empty definition.
var EmptyDef = Def{nil, &rules{}}

// LineStates is an interface for a buffer-like object which can also store the states and matches for every line
type LineStates interface {
	LineBytes(n int) []byte
	LinesNum() int
	State(lineN int) State
	SetState(lineN int, s State)
	SetMatch(lineN int, m LineMatch)
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
		if curRegion == nil || curRegion.group == curRegion.limitGroup || p.group == curRegion.limitGroup {
			matches := findAllIndex(p.regex, nil, line)
			for _, m := range matches {
				h.highlightRange(fullHighlights, start+m[0], start+m[1], p.group)
			}
		}
	}
}

func (h *Highlighter) highlightRegions(fullHighlights []Group, start int, lineNum int, line []byte, curRegion *region, regions []*region, nestedRegion bool) {
	lineLen := util.CharacterCount(line)
	if lineLen == 0 {
		return
	}

	h.highlightPatterns(fullHighlights, start, lineNum, line, curRegion)

	lastStart := 0
	regionStart := 0
	regionEnd := 0
	for _, r := range regions {
		if !nestedRegion && curRegion != nil && (curRegion.group != r.group || curRegion.start != r.start) {
			continue
		}
		startMatches := findAllIndex(r.start, r.skip, line)
		endMatches := findAllIndex(r.end, r.skip, line)
		samePattern := false
	startLoop:
		for startIdx := 0; startIdx < len(startMatches); startIdx++ {
			if startMatches[startIdx][0] < lineLen {
				for endIdx := 0; endIdx < len(endMatches); endIdx++ {
					if startMatches[startIdx][0] == endMatches[endIdx][0] {
						// start and end are the same (pattern)
						samePattern = true
						if len(startMatches) == len(endMatches) {
							// special case in the moment both are the same
							if curRegion != nil && curRegion.group == r.group && curRegion.start == r.start {
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
					} else if startMatches[startIdx][1] <= endMatches[endIdx][0] {
						if regionStart < startMatches[startIdx][0] && startMatches[startIdx][0] < regionEnd {
							continue
						}
						// start and end at the current line
						regionStart = startMatches[startIdx][0]
						regionEnd = endMatches[endIdx][1]
						h.highlightRange(fullHighlights, start+startMatches[startIdx][0], start+endMatches[endIdx][1], r.limitGroup)
						h.highlightRegions(fullHighlights, start+startMatches[startIdx][1], lineNum, util.SliceStartEnd(line, startMatches[startIdx][1], endMatches[endIdx][0]), r, r.rules.regions, true)
						if samePattern {
							startIdx += 1
						}
						if lastStart == 0 || lastStart < endMatches[endIdx][1] {
							if curRegion != nil {
								h.lastRegion = curRegion.parent
							} else {
								h.lastRegion = nil
							}
							curRegion = h.lastRegion
						}
						continue startLoop
					} else if endMatches[endIdx][1] <= startMatches[startIdx][0] {
						if endMatches[endIdx][0] < regionEnd || curRegion == nil {
							continue
						}
						// start and end at the current line, but switched
						h.highlightRange(fullHighlights, start, start+endMatches[endIdx][1], r.limitGroup)
						h.highlightRegions(fullHighlights, start, lineNum, util.SliceStart(line, endMatches[endIdx][0]), r, r.rules.regions, true)
						h.highlightPatterns(fullHighlights, start+endMatches[endIdx][1], lineNum, util.SliceStartEnd(line, endMatches[endIdx][1], startMatches[startIdx][0]), nil)
						if curRegion != nil {
							h.lastRegion = curRegion.parent
						} else {
							h.lastRegion = nil
						}
						curRegion = h.lastRegion
					}
				}
				if regionEnd <= startMatches[startIdx][0] {
					// start at the current, but end at the next line
					regionStart = startMatches[startIdx][0]
					regionEnd = 0
					h.highlightRange(fullHighlights, start+startMatches[startIdx][0], lineLen, r.limitGroup)
					h.highlightRegions(fullHighlights, start+startMatches[startIdx][1], lineNum, util.SliceEnd(line, startMatches[startIdx][1]), r, r.rules.regions, true)
					if lastStart <= startMatches[startIdx][0] {
						lastStart = startMatches[startIdx][0]
						h.lastRegion = r
					}
				}
			}
		}
		if curRegion != nil && curRegion.group == r.group && curRegion.start == r.start {
			if (len(startMatches) == 0 && len(endMatches) > 0) || (samePattern && (len(startMatches) == len(endMatches))) {
				for _, endMatch := range endMatches {
					// end at the current, but start at the previous line
					h.highlightRange(fullHighlights, start, start+endMatch[1], r.limitGroup)
					h.highlightRegions(fullHighlights, start, lineNum, util.SliceStart(line, endMatch[0]), r, r.rules.regions, true)
					if curRegion != nil {
						h.lastRegion = curRegion.parent
					} else {
						h.lastRegion = nil
					}
					curRegion = h.lastRegion
					break
				}
			} else if len(startMatches) == 0 && len(endMatches) == 0 {
				// no start and end found in this region
				h.highlightRange(fullHighlights, start, lineLen, curRegion.group)
			}
		}
	}
}

func (h *Highlighter) highlight(highlights LineMatch, start int, lineNum int, line []byte, curRegion *region) LineMatch {
	lineLen := util.CharacterCount(line)
	if lineLen == 0 {
		return highlights
	}

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
	lines := strings.Split(input, "\n")
	var lineMatches []LineMatch

	for i := 0; i < len(lines); i++ {
		line := []byte(lines[i])
		highlights := make(LineMatch)
		lineMatches = append(lineMatches, h.highlight(highlights, 0, i, line, nil))
	}

	return lineMatches
}

// Highlight sets the state and matches for each line from startline to endline
// It sets all other matches in the buffer to nil to conserve memory
func (h *Highlighter) Highlight(input LineStates, startline, endline int) {
	h.lastRegion = nil
	if startline > 0 {
		h.lastRegion = input.State(startline - 1)
	}

	for i := startline; i <= endline; i++ {
		if i >= input.LinesNum() {
			break
		}

		line := input.LineBytes(i)
		highlights := make(LineMatch)

		var match LineMatch
		match = h.highlight(highlights, 0, i, line, h.lastRegion)

		input.SetState(i, h.lastRegion)
		input.SetMatch(i, match)
	}
}

// ReHighlightLine will rehighlight the state and match for a single line
func (h *Highlighter) ReHighlightLine(input LineStates, lineN int) {
	line := input.LineBytes(lineN)
	highlights := make(LineMatch)

	h.lastRegion = nil
	if lineN > 0 {
		h.lastRegion = input.State(lineN - 1)
	}

	var match LineMatch
	match = h.highlight(highlights, 0, lineN, line, h.lastRegion)

	input.SetState(lineN, h.lastRegion)
	input.SetMatch(lineN, match)
}
