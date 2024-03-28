package highlight

import (
	// "log"
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

// highlightStorage is used to store the found ranges
type highlightStorage struct {
	start   int
	end     int
	group   Group
	region  *region
	childs  []*highlightStorage
	pattern bool
}

// A Highlighter contains the information needed to highlight a string
type Highlighter struct {
	lastRegion *region
	lastStart  int
	lastEnd    int
	Def        *Def
	storage    []highlightStorage
	removed    []highlightStorage
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

func (h *Highlighter) removeRange(start int, end int, removeStart int) {
	var childs []highlightStorage
	removeEnd := removeStart
	for i := removeStart; i < len(h.storage); i++ {
		e := h.storage[i]
		if start < e.start && e.start < end {
			// log.Println("remove: start:", e.start, "end:", e.end, "group:", e.group)
			removeEnd++
			h.removed = append(h.removed, e)
			for childIdx, _ := range h.storage[i].childs {
				// log.Println("attached child: start:", h.storage[i].childs[childIdx].start, "end:", h.storage[i].childs[childIdx].end, "group:", h.storage[i].childs[childIdx].group)
				childs = append(childs, *(h.storage[i].childs[childIdx]))
			}
		}
	}
	if removeStart < removeEnd {
		h.storage = append(h.storage[:removeStart], h.storage[removeEnd:]...)
	}

	// remove possible childs too
childLoop:
	for childIdx, _ := range childs {
		for storageIdx, _ := range h.storage {
			if childs[childIdx].start == h.storage[storageIdx].start && childs[childIdx].end == h.storage[storageIdx].end && childs[childIdx].group == h.storage[storageIdx].group && childs[childIdx].region == h.storage[storageIdx].region {
				// log.Println("remove child: start:", h.storage[storageIdx].start, "end:", h.storage[storageIdx].end, "group:", h.storage[storageIdx].group)
				h.storage = append(h.storage[:storageIdx], h.storage[storageIdx+1:]...)
				continue childLoop
			}
		}
	}
}

func (h *Highlighter) storeRange(start int, end int, group Group, r *region, statesOnly bool, isPattern bool) {
	if statesOnly {
		return
	}

	// log.Println("storeRange: start:", start, "end:", end, "group:", group)

	var parent *region
	if isPattern {
		parent = r
	} else if r != nil {
		parent = r.parent
	}

	updated := false
	for k, e := range h.storage {
		if r == e.region && group == e.group && start == e.end {
			// same region, update ...
			h.storage[k].end = end
			// log.Println("exchanged to: start:", h.storage[k].start, "end:", h.storage[k].end, "group:", h.storage[k].group)
			updated = true
			start = h.storage[k].start
		}
	}

	for k, e := range h.storage {
		if e.region != nil && r != nil {
			if e.region.parent == parent {
				if r != e.region {
					// sibling regions, search for overlaps ...
					if start < e.start && end > e.start {
						// overlap from left
					} else if start == e.start && end == e.end {
						// same match
						continue
					} else if start <= e.start && end >= e.end {
						// larger match
					} else if start >= e.start && end <= e.end {
						// smaller match
						return
					} else if start > e.start && start < e.end && end > e.end {
						// overlap from right
						return
					} else {
						continue
					}

					if !updated {
						// log.Println("exchanged from: start:", e.start, "end:", e.end, "group:", e.group)
						h.storage[k] = highlightStorage{start, end, group, r, nil, isPattern}

						// check and remove follow-ups matching the same
						h.removeRange(start, end, k+1)
					} else {
						h.removeRange(start, end, k)
					}
					return
				}
			} else {
				if parent != e.region && start >= e.start && end <= e.end {
					return
				}
			}
		}
	}

	if !updated {
		h.storage = append(h.storage, highlightStorage{start, end, group, r, nil, isPattern})
	}

	// add possible child entry
	if parent != nil {
	storageLoop:
		for k, e := range h.storage {
			if e.region == parent && e.start < start && end < e.end {
				for _, child := range h.storage[k].childs {
					if child == &(h.storage[len(h.storage)-1]) {
						continue storageLoop
					}
				}

				// log.Println("add child: start:", h.storage[k].start, "end:", h.storage[k].end, "group:", h.storage[k].group)
				h.storage[k].childs = append(h.storage[k].childs, &(h.storage[len(h.storage)-1]))
			}
		}
	}
}

func (h *Highlighter) storePatterns(start int, lineNum int, line []byte, curRegion *region, statesOnly bool) {
	lineLen := util.CharacterCount(line)
	// log.Println("storePatterns: lineNum:", lineNum, "start:", start, "line:", string(line))
	if lineLen == 0 || statesOnly {
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
				h.storeRange(start+m[0], start+m[1], p.group, curRegion, statesOnly, true)
			}
		}
	}
}

func (h *Highlighter) storeRegions(start int, lineNum int, line []byte, curRegion *region, regions []*region, nestedRegion bool, statesOnly bool) {
	lineLen := util.CharacterCount(line)
	// log.Println("storeRegions: lineNum:", lineNum, "start:", start, "line:", string(line))
	if lineLen == 0 {
		return
	}

	if nestedRegion {
		h.storePatterns(start, lineNum, line, curRegion, statesOnly)
	} else {
		h.storePatterns(start, lineNum, line, nil, statesOnly)
	}

regionLoop:
	for _, r := range regions {
		// log.Println("r.start:", r.start.String(), "r.end", r.end.String())
		if !nestedRegion && curRegion != nil && curRegion != r {
			continue
		}
		startMatches := findAllIndex(r.start, r.skip, line)
		endMatches := findAllIndex(r.end, r.skip, line)
		samePattern := false
	startLoop:
		for startIdx := 0; startIdx < len(startMatches); startIdx++ {
			// log.Println("startIdx:", startIdx, "of", len(startMatches))
			if startMatches[startIdx][0] < lineLen {
				for endIdx := 0; endIdx < len(endMatches); endIdx++ {
					// log.Println("startIdx:", startIdx, "of", len(startMatches), "/ endIdx:", endIdx, "of", len(endMatches), "/ h.lastStart:", h.lastStart, "/ h.lastEnd:", h.lastEnd)
					if startMatches[startIdx][0] == endMatches[endIdx][0] {
						// start and end are the same (pattern)
						// log.Println("start == end")
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
					} else if startMatches[startIdx][1] <= endMatches[endIdx][0] {
						if !nestedRegion && h.lastStart < start+startMatches[startIdx][0] && start+startMatches[startIdx][0] < h.lastEnd {
							continue
						}
						// start and end at the current line
						// log.Println("start < end")
						update := false
						if h.lastStart == -1 || h.lastStart < start+endMatches[endIdx][1] {
							h.lastStart = start + startMatches[startIdx][0]
							h.lastEnd = start + endMatches[endIdx][1]
							update = true
						}
						h.storeRange(start+startMatches[startIdx][0], start+startMatches[startIdx][1], r.limitGroup, r, statesOnly, false)
						h.storeRange(start+startMatches[startIdx][1], start+endMatches[endIdx][0], r.group, r, statesOnly, false)
						h.storeRange(start+endMatches[endIdx][0], start+endMatches[endIdx][1], r.limitGroup, r, statesOnly, false)
						h.storeRegions(start+startMatches[startIdx][1], lineNum, util.SliceStartEnd(line, startMatches[startIdx][1], endMatches[endIdx][0]), r, r.rules.regions, true, statesOnly)
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
					} else if endMatches[endIdx][1] <= startMatches[startIdx][0] {
						if start+endMatches[endIdx][0] < h.lastEnd || curRegion == nil {
							continue
						}
						// start and end at the current line, but switched
						// log.Println("end < start")
						h.lastStart = start
						h.lastEnd = start + endMatches[endIdx][1]
						h.storeRange(start, start+endMatches[endIdx][0], r.group, r, statesOnly, false)
						h.storeRange(start+endMatches[endIdx][0], start+endMatches[endIdx][1], r.limitGroup, r, statesOnly, false)
						h.storeRegions(start, lineNum, util.SliceStart(line, endMatches[endIdx][0]), r, r.rules.regions, true, statesOnly)
						h.storePatterns(start+endMatches[endIdx][1], lineNum, util.SliceStartEnd(line, endMatches[endIdx][1], startMatches[startIdx][0]), nil, statesOnly)
						if curRegion != nil {
							h.lastRegion = r.parent
						} else {
							h.lastRegion = nil
						}
						curRegion = h.lastRegion
					}
				}
				if nestedRegion || start+startMatches[startIdx][0] < h.lastStart || h.lastEnd < start+startMatches[startIdx][0] {
					// start at the current, but end at the next line
					// log.Println("start ...")
					if h.lastStart == -1 || start+startMatches[startIdx][0] < h.lastStart || h.lastEnd < start+startMatches[startIdx][0] {
						h.lastStart = start + startMatches[startIdx][0]
						h.lastEnd = start + lineLen - 1
						h.lastRegion = r
					}
					h.storeRange(start+startMatches[startIdx][0], start+startMatches[startIdx][1], r.limitGroup, r, statesOnly, false)
					h.storeRange(start+startMatches[startIdx][1], start+lineLen, r.group, r, statesOnly, false)
					h.storeRegions(start+startMatches[startIdx][1], lineNum, util.SliceEnd(line, startMatches[startIdx][1]), r, r.rules.regions, true, statesOnly)
					continue regionLoop
				}
			}
		}
		if curRegion == r {
			if (len(startMatches) == 0 && len(endMatches) > 0) || (samePattern && (len(startMatches) == len(endMatches))) {
				for _, endMatch := range endMatches {
					// end at the current, but start at the previous line
					// log.Println("... end")
					h.lastStart = start
					h.lastEnd = start + endMatch[1]
					h.storeRange(start, start+endMatch[0], r.group, r, statesOnly, false)
					h.storeRange(start+endMatch[0], start+endMatch[1], r.limitGroup, r, statesOnly, false)
					h.storeRegions(start, lineNum, util.SliceStart(line, endMatch[0]), r, r.rules.regions, true, statesOnly)
					if curRegion != nil {
						h.lastRegion = r.parent
					} else {
						h.lastRegion = nil
					}
					curRegion = h.lastRegion
					h.storeRegions(start+endMatch[1], lineNum, util.SliceEnd(line, endMatch[1]), curRegion, h.Def.rules.regions, false, statesOnly)
					break
				}
			} else if len(startMatches) == 0 && len(endMatches) == 0 {
				// no start and end found in this region
				h.storeRange(start, start+lineLen, curRegion.group, r, statesOnly, false)
			}
		}
	}

	if curRegion != nil && !nestedRegion {
		// current region still open
		// log.Println("...")
		if curRegion.rules != nil {
			h.storeRegions(start, lineNum, line, curRegion, curRegion.rules.regions, true, statesOnly)
		}
		if curRegion == h.lastRegion && curRegion.parent != nil {
			var regions []*region
			regions = append(regions, curRegion)
			h.storeRegions(start, lineNum, line, curRegion, regions, true, statesOnly)
		}
	}
}

func (h *Highlighter) highlight(highlights LineMatch, start int, lineNum int, line []byte, curRegion *region, statesOnly bool) LineMatch {
	lineLen := util.CharacterCount(line)
	// log.Println("highlight: lineNum:", lineNum, "start:", start, "line:", string(line))
	if lineLen == 0 {
		return highlights
	}

	h.lastStart = -1
	h.lastEnd = -1
	h.storage = h.storage[:0]
	h.removed = h.removed[:0]

	h.storeRegions(start, lineNum, line, curRegion, h.Def.rules.regions, false, statesOnly)

	if !statesOnly {
		// check if entries have been removed by invalid region
		for _, e := range h.removed {
			h.storeRange(e.start, e.end, e.group, e.region, statesOnly, e.pattern)
		}

		fullHighlights := make([]Group, lineLen)

		for _, e := range h.storage {
			if e.start <= e.end && e.end <= len(fullHighlights) {
				for i := e.start; i < e.end; i++ {
					fullHighlights[i] = e.group
					// log.Println("fullHighlights[", i, "]:", e.group)
				}
			}
		}

		for i, h := range fullHighlights {
			if i == 0 || h != fullHighlights[i-1] {
				highlights[i] = h
			}
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
		lineMatches = append(lineMatches, h.highlight(highlights, 0, i, line, nil, false))
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
		match = h.highlight(highlights, 0, i, line, h.lastRegion, false)

		input.SetState(i, h.lastRegion)
		input.SetMatch(i, match)
	}
}

// ReHighlightStates will scan down from `startline` and set the appropriate end of line state
// for each line until it comes across a line whose state does not change
// returns the number of the final line
func (h *Highlighter) ReHighlightStates(input LineStates, startline int) int {
	h.lastRegion = nil
	if startline > 0 {
		h.lastRegion = input.State(startline - 1)
	}
	for i := startline; i < input.LinesNum(); i++ {
		line := input.LineBytes(i)

		h.highlight(nil, 0, i, line, h.lastRegion, true)

		curState := h.lastRegion
		lastState := input.State(i)

		input.SetState(i, curState)

		if curState == lastState {
			return i
		}
	}

	return input.LinesNum() - 1
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
	match = h.highlight(highlights, 0, lineN, line, h.lastRegion, false)

	input.SetState(lineN, h.lastRegion)
	input.SetMatch(lineN, match)
}
