package display

import (
	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/util"
)

// SLoc represents a vertical scrolling location, i.e. a location of a visual line
// in the buffer. When softwrap is enabled, a buffer line may be displayed as
// multiple visual lines (rows). So SLoc stores a number of a line in the buffer
// and a number of a row within this line.
type SLoc struct {
	Line, Row int
}

// LessThan returns true if s is less b
func (s SLoc) LessThan(b SLoc) bool {
	if s.Line < b.Line {
		return true
	}
	return s.Line == b.Line && s.Row < b.Row
}

// GreaterThan returns true if s is bigger than b
func (s SLoc) GreaterThan(b SLoc) bool {
	if s.Line > b.Line {
		return true
	}
	return s.Line == b.Line && s.Row > b.Row
}

type SoftWrap interface {
	Scroll(s SLoc, n int) SLoc
	Diff(s1, s2 SLoc) int
	SLocFromLoc(loc buffer.Loc) SLoc
}

func (w *BufWindow) getRow(loc buffer.Loc) int {
	if w.bufWidth <= 0 {
		return 0
	}
	// TODO: this doesn't work quite correctly if there is an incomplete tab
	// or wide character at the end of a row. See also issue #1979
	x := util.StringWidth(w.Buf.LineBytes(loc.Y), loc.X, util.IntOpt(w.Buf.Settings["tabsize"]))
	return x / w.bufWidth
}

func (w *BufWindow) getRowCount(line int) int {
	return w.getRow(buffer.Loc{X: util.CharacterCount(w.Buf.LineBytes(line)), Y: line}) + 1
}

func (w *BufWindow) scrollUp(s SLoc, n int) SLoc {
	for n > 0 {
		if n <= s.Row {
			s.Row -= n
			n = 0
		} else if s.Line > 0 {
			s.Line--
			n -= s.Row + 1
			s.Row = w.getRowCount(s.Line) - 1
		} else {
			s.Row = 0
			break
		}
	}
	return s
}

func (w *BufWindow) scrollDown(s SLoc, n int) SLoc {
	for n > 0 {
		rc := w.getRowCount(s.Line)
		if n < rc-s.Row {
			s.Row += n
			n = 0
		} else if s.Line < w.Buf.LinesNum()-1 {
			s.Line++
			n -= rc - s.Row
			s.Row = 0
		} else {
			s.Row = rc - 1
			break
		}
	}
	return s
}

func (w *BufWindow) scroll(s SLoc, n int) SLoc {
	if n < 0 {
		return w.scrollUp(s, -n)
	}
	return w.scrollDown(s, n)
}

func (w *BufWindow) diff(s1, s2 SLoc) int {
	n := 0
	for s1.LessThan(s2) {
		if s1.Line < s2.Line {
			n += w.getRowCount(s1.Line) - s1.Row
			s1.Line++
			s1.Row = 0
		} else {
			n += s2.Row - s1.Row
			s1.Row = s2.Row
		}
	}
	return n
}

// Scroll returns the location which is n visual lines below the location s
// i.e. the result of scrolling n lines down. n can be negative,
// which means scrolling up. The returned location is guaranteed to be
// within the buffer boundaries.
func (w *BufWindow) Scroll(s SLoc, n int) SLoc {
	if !w.Buf.Settings["softwrap"].(bool) {
		s.Line += n
		if s.Line < 0 {
			s.Line = 0
		}
		if s.Line > w.Buf.LinesNum()-1 {
			s.Line = w.Buf.LinesNum() - 1
		}
		return s
	}
	return w.scroll(s, n)
}

// Diff returns the difference (the vertical distance) between two SLocs.
func (w *BufWindow) Diff(s1, s2 SLoc) int {
	if !w.Buf.Settings["softwrap"].(bool) {
		return s2.Line - s1.Line
	}
	if s1.GreaterThan(s2) {
		return -w.diff(s2, s1)
	}
	return w.diff(s1, s2)
}

// SLocFromLoc takes a position in the buffer and returns the location
// of the visual line containing this position.
func (w *BufWindow) SLocFromLoc(loc buffer.Loc) SLoc {
	if !w.Buf.Settings["softwrap"].(bool) {
		return SLoc{loc.Y, 0}
	}
	return SLoc{loc.Y, w.getRow(loc)}
}
