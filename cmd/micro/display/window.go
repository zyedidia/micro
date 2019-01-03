package display

import (
	"strconv"
	"unicode/utf8"

	runewidth "github.com/mattn/go-runewidth"
	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/micro/cmd/micro/util"
	"github.com/zyedidia/tcell"
)

type View struct {
	X, Y          int // X,Y location of the view
	Width, Height int // Width and height of the view

	// Start line and start column of the view (vertical/horizontal scroll)
	// note that since the starting column of every line is different if the view
	// is scrolled, StartCol is a visual index (will be the same for every line)
	StartLine, StartCol int
}

type Window interface {
	Display()
	Clear()
	Relocate() bool
	GetView() *View
	SetView(v *View)
	GetMouseLoc(vloc buffer.Loc) buffer.Loc
}

// The BufWindow provides a way of displaying a certain section
// of a buffer
type BufWindow struct {
	*View

	// Buffer being shown in this window
	Buf *buffer.Buffer

	sline *StatusLine

	lineHeight   []int
	gutterOffset int
}

// NewBufWindow creates a new window at a location in the screen with a width and height
func NewBufWindow(x, y, width, height int, buf *buffer.Buffer) *BufWindow {
	w := new(BufWindow)
	w.View = new(View)
	w.X, w.Y, w.Width, w.Height, w.Buf = x, y, width, height, buf
	w.lineHeight = make([]int, height)

	w.sline = NewStatusLine(w)

	return w
}

func (v *View) GetView() *View {
	return v
}

func (v *View) SetView(view *View) {
	v = view
}

func (w *BufWindow) getStartInfo(n, lineN int) ([]byte, int, int, *tcell.Style) {
	tabsize := util.IntOpt(w.Buf.Settings["tabsize"])
	width := 0
	bloc := buffer.Loc{0, lineN}
	b := w.Buf.LineBytes(lineN)
	curStyle := config.DefStyle
	var s *tcell.Style
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)

		curStyle, found := w.getStyle(curStyle, bloc, r)
		if found {
			s = &curStyle
		}

		w := 0
		switch r {
		case '\t':
			ts := tabsize - (width % tabsize)
			w = ts
		default:
			w = runewidth.RuneWidth(r)
		}
		if width+w > n {
			return b, n - width, bloc.X, s
		}
		width += w
		b = b[size:]
		bloc.X++
	}
	return b, n - width, bloc.X, s
}

// Clear resets all cells in this window to the default style
func (w *BufWindow) Clear() {
	for y := 0; y < w.Height; y++ {
		for x := 0; x < w.Width; x++ {
			screen.Screen.SetContent(w.X+x, w.Y+y, ' ', nil, config.DefStyle)
		}
	}
}

// Bottomline returns the line number of the lowest line in the view
// You might think that this is obviously just v.StartLine + v.Height
// but if softwrap is enabled things get complicated since one buffer
// line can take up multiple lines in the view
func (w *BufWindow) Bottomline() int {
	// b := w.Buf

	// TODO: possible non-softwrap optimization
	// if !b.Settings["softwrap"].(bool) {
	// 	return w.StartLine + w.Height
	// }

	prev := 0
	for _, l := range w.lineHeight {
		if l >= prev {
			prev = l
		} else {
			break
		}
	}
	return prev
}

// Relocate moves the view window so that the cursor is in view
// This is useful if the user has scrolled far away, and then starts typing
// Returns true if the window location is moved
func (w *BufWindow) Relocate() bool {
	b := w.Buf
	height := w.Bottomline() + 1 - w.StartLine
	ret := false
	activeC := w.Buf.GetActiveCursor()
	cy := activeC.Y
	scrollmargin := int(b.Settings["scrollmargin"].(float64))
	if cy < w.StartLine+scrollmargin && cy > scrollmargin-1 {
		w.StartLine = cy - scrollmargin
		ret = true
	} else if cy < w.StartLine {
		w.StartLine = cy
		ret = true
	}
	if cy > w.StartLine+height-1-scrollmargin && cy < b.LinesNum()-scrollmargin {
		w.StartLine = cy - height + 1 + scrollmargin
		ret = true
	} else if cy >= b.LinesNum()-scrollmargin && cy >= height {
		w.StartLine = b.LinesNum() - height
		ret = true
	}

	// horizontal relocation (scrolling)
	if !b.Settings["softwrap"].(bool) {
		cx := activeC.GetVisualX()
		if cx < w.StartCol {
			w.StartCol = cx
			ret = true
		}
		if cx+w.gutterOffset+1 > w.StartCol+w.Width {
			w.StartCol = cx - w.Width + w.gutterOffset + 1
			ret = true
		}
	}
	return ret
}

func (w *BufWindow) GetMouseLoc(svloc buffer.Loc) buffer.Loc {
	b := w.Buf

	// TODO: possible non-softwrap optimization
	// if !b.Settings["softwrap"].(bool) {
	// 	l := b.LineBytes(svloc.Y)
	// 	return buffer.Loc{b.GetActiveCursor().GetCharPosInLine(l, svloc.X), svloc.Y}
	// }

	bufHeight := w.Height
	if b.Settings["statusline"].(bool) {
		bufHeight--
	}

	// We need to know the string length of the largest line number
	// so we can pad appropriately when displaying line numbers
	maxLineNumLength := len(strconv.Itoa(b.LinesNum()))

	tabsize := int(b.Settings["tabsize"].(float64))
	softwrap := b.Settings["softwrap"].(bool)

	// this represents the current draw position
	// within the current window
	vloc := buffer.Loc{X: 0, Y: 0}

	// this represents the current draw position in the buffer (char positions)
	bloc := buffer.Loc{X: -1, Y: w.StartLine}

	for vloc.Y = 0; vloc.Y < bufHeight; vloc.Y++ {
		vloc.X = 0
		if b.Settings["ruler"].(bool) {
			vloc.X += maxLineNumLength + 1
		}

		line := b.LineBytes(bloc.Y)
		line, nColsBeforeStart, bslice := util.SliceVisualEnd(line, w.StartCol, tabsize)
		bloc.X = bslice

		draw := func() {
			if nColsBeforeStart <= 0 {
				vloc.X++
			}
			nColsBeforeStart--
		}

		w.lineHeight[vloc.Y] = bloc.Y

		totalwidth := w.StartCol - nColsBeforeStart

		if svloc.X <= vloc.X && vloc.Y == svloc.Y {
			return bloc
		}
		for len(line) > 0 {
			if vloc.X == svloc.X && vloc.Y == svloc.Y {
				return bloc
			}

			r, size := utf8.DecodeRune(line)
			draw()
			width := 0

			switch r {
			case '\t':
				ts := tabsize - (totalwidth % tabsize)
				width = ts
			default:
				width = runewidth.RuneWidth(r)
			}

			// Draw any extra characters either spaces for tabs or @ for incomplete wide runes
			if width > 1 {
				for i := 1; i < width; i++ {
					if vloc.X == svloc.X && vloc.Y == svloc.Y {
						return bloc
					}
					draw()
				}
			}
			bloc.X++
			line = line[size:]

			totalwidth += width

			// If we reach the end of the window then we either stop or we wrap for softwrap
			if vloc.X >= w.Width {
				if !softwrap {
					break
				} else {
					vloc.Y++
					if vloc.Y >= bufHeight {
						break
					}
					vloc.X = 0
					w.lineHeight[vloc.Y] = bloc.Y
					// This will draw an empty line number because the current line is wrapped
					vloc.X += maxLineNumLength + 1
				}
			}
		}
		if vloc.Y == svloc.Y {
			return bloc
		}

		bloc.X = w.StartCol
		bloc.Y++
		if bloc.Y >= b.LinesNum() {
			break
		}
	}

	return buffer.Loc{X: -1, Y: -1}
}

func (w *BufWindow) drawLineNum(lineNumStyle tcell.Style, softwrapped bool, maxLineNumLength int, vloc *buffer.Loc, bloc *buffer.Loc) {
	lineNum := strconv.Itoa(bloc.Y + 1)

	// Write the spaces before the line number if necessary
	for i := 0; i < maxLineNumLength-len(lineNum); i++ {
		screen.Screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, ' ', nil, lineNumStyle)
		vloc.X++
	}
	// Write the actual line number
	for _, ch := range lineNum {
		if softwrapped {
			screen.Screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, ' ', nil, lineNumStyle)
		} else {
			screen.Screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, ch, nil, lineNumStyle)
		}
		vloc.X++
	}

	// Write the extra space
	screen.Screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, ' ', nil, lineNumStyle)
	vloc.X++
}

// getStyle returns the highlight style for the given character position
// If there is no change to the current highlight style it just returns that
func (w *BufWindow) getStyle(style tcell.Style, bloc buffer.Loc, r rune) (tcell.Style, bool) {
	if group, ok := w.Buf.Match(bloc.Y)[bloc.X]; ok {
		s := config.GetColor(group.String())
		return s, true
	}
	return style, false
}

func (w *BufWindow) showCursor(x, y int, main bool) {
	if main {
		screen.Screen.ShowCursor(x, y)
	} else {
		r, _, _, _ := screen.Screen.GetContent(x, y)
		screen.Screen.SetContent(x, y, r, nil, config.DefStyle.Reverse(true))
	}
}

// displayBuffer draws the buffer being shown in this window on the screen.Screen
func (w *BufWindow) displayBuffer() {
	b := w.Buf

	bufHeight := w.Height
	if b.Settings["statusline"].(bool) {
		bufHeight--
	}

	// TODO: Rehighlighting
	start := w.StartLine
	if b.Settings["syntax"].(bool) && b.SyntaxDef != nil {
		if start > 0 && b.Rehighlight(start-1) {
			b.Highlighter.ReHighlightLine(b, start-1)
			b.SetRehighlight(start-1, false)
		}

		b.Highlighter.ReHighlightStates(b, start)

		b.Highlighter.HighlightMatches(b, w.StartLine, w.StartLine+bufHeight)
	}

	lineNumStyle := config.DefStyle
	if style, ok := config.Colorscheme["line-number"]; ok {
		lineNumStyle = style
	}
	curNumStyle := config.DefStyle
	if style, ok := config.Colorscheme["current-line-number"]; ok {
		curNumStyle = style
	}

	// We need to know the string length of the largest line number
	// so we can pad appropriately when displaying line numbers
	maxLineNumLength := len(strconv.Itoa(b.LinesNum()))

	tabsize := int(b.Settings["tabsize"].(float64))
	softwrap := b.Settings["softwrap"].(bool)

	// this represents the current draw position
	// within the current window
	vloc := buffer.Loc{X: 0, Y: 0}

	// this represents the current draw position in the buffer (char positions)
	bloc := buffer.Loc{X: -1, Y: w.StartLine}

	cursors := b.GetCursors()

	curStyle := config.DefStyle
	for vloc.Y = 0; vloc.Y < bufHeight; vloc.Y++ {
		vloc.X = 0
		if b.Settings["ruler"].(bool) {
			s := lineNumStyle
			for _, c := range cursors {
				if bloc.Y == c.Y {
					s = curNumStyle
					break
				}
			}
			w.drawLineNum(s, false, maxLineNumLength, &vloc, &bloc)
		}

		w.gutterOffset = vloc.X

		line, nColsBeforeStart, bslice, startStyle := w.getStartInfo(w.StartCol, bloc.Y)
		if startStyle != nil {
			curStyle = *startStyle
		}
		bloc.X = bslice

		draw := func(r rune, style tcell.Style, showcursor bool) {
			if nColsBeforeStart <= 0 {
				for _, c := range cursors {
					if c.HasSelection() &&
						(bloc.GreaterEqual(c.CurSelection[0]) && bloc.LessThan(c.CurSelection[1]) ||
							bloc.LessThan(c.CurSelection[0]) && bloc.GreaterEqual(c.CurSelection[1])) {
						// The current character is selected
						style = config.DefStyle.Reverse(true)

						if s, ok := config.Colorscheme["selection"]; ok {
							style = s
						}
					}

					if b.Settings["cursorline"].(bool) &&
						!c.HasSelection() && c.Y == bloc.Y {
						if s, ok := config.Colorscheme["cursor-line"]; ok {
							fg, _, _ := s.Decompose()
							style = style.Background(fg)
						}
					}
				}

				screen.Screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, r, nil, style)

				if showcursor {
					for _, c := range cursors {
						if c.X == bloc.X && c.Y == bloc.Y && !c.HasSelection() {
							w.showCursor(w.X+vloc.X, w.Y+vloc.Y, c.Num == 0)
						}
					}
				}
				vloc.X++
			}
			nColsBeforeStart--
		}

		w.lineHeight[vloc.Y] = bloc.Y

		totalwidth := w.StartCol - nColsBeforeStart
		for len(line) > 0 {
			r, size := utf8.DecodeRune(line)
			curStyle, _ = w.getStyle(curStyle, bloc, r)

			draw(r, curStyle, true)

			width := 0

			char := ' '
			switch r {
			case '\t':
				ts := tabsize - (totalwidth % tabsize)
				width = ts
			default:
				width = runewidth.RuneWidth(r)
				char = '@'
			}

			// Draw any extra characters either spaces for tabs or @ for incomplete wide runes
			if width > 1 {
				for i := 1; i < width; i++ {
					draw(char, curStyle, false)
				}
			}
			bloc.X++
			line = line[size:]

			totalwidth += width

			// If we reach the end of the window then we either stop or we wrap for softwrap
			if vloc.X >= w.Width {
				if !softwrap {
					break
				} else {
					vloc.Y++
					if vloc.Y >= bufHeight {
						break
					}
					vloc.X = 0
					w.lineHeight[vloc.Y] = bloc.Y
					// This will draw an empty line number because the current line is wrapped
					w.drawLineNum(lineNumStyle, true, maxLineNumLength, &vloc, &bloc)
				}
			}
		}

		for _, c := range cursors {
			if b.Settings["cursorline"].(bool) &&
				!c.HasSelection() && c.Y == bloc.Y {
				style := config.DefStyle
				if s, ok := config.Colorscheme["cursor-line"]; ok {
					fg, _, _ := s.Decompose()
					style = style.Background(fg)
				}
				for i := vloc.X; i < w.Width; i++ {
					screen.Screen.SetContent(i, vloc.Y, ' ', nil, style)
				}
			}
		}

		for _, c := range cursors {
			if c.X == bloc.X && c.Y == bloc.Y && !c.HasSelection() {
				w.showCursor(w.X+vloc.X, w.Y+vloc.Y, c.Num == 0)
			}
		}

		bloc.X = w.StartCol
		bloc.Y++
		if bloc.Y >= b.LinesNum() {
			break
		}
	}
}

func (w *BufWindow) displayStatusLine() {
	w.sline.Display()
}

// Display displays the buffer and the statusline
func (w *BufWindow) Display() {
	w.displayBuffer()
	w.displayStatusLine()
}
