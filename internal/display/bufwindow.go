package display

import (
	"strconv"

	runewidth "github.com/mattn/go-runewidth"
	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/micro/v2/internal/util"
	"github.com/zyedidia/tcell"
)

// The BufWindow provides a way of displaying a certain section
// of a buffer
type BufWindow struct {
	*View

	// Buffer being shown in this window
	Buf *buffer.Buffer

	active bool

	sline *StatusLine

	gutterOffset int
	drawStatus   bool
}

// NewBufWindow creates a new window at a location in the screen with a width and height
func NewBufWindow(x, y, width, height int, buf *buffer.Buffer) *BufWindow {
	w := new(BufWindow)
	w.View = new(View)
	w.X, w.Y, w.Width, w.Height, w.Buf = x, y, width, height, buf
	w.active = true

	w.sline = NewStatusLine(w)

	return w
}

func (w *BufWindow) SetBuffer(b *buffer.Buffer) {
	w.Buf = b
}

func (w *BufWindow) GetView() *View {
	return w.View
}

func (w *BufWindow) SetView(view *View) {
	w.View = view
}

func (w *BufWindow) Resize(width, height int) {
	w.Width, w.Height = width, height
	w.Relocate()
}

func (w *BufWindow) SetActive(b bool) {
	w.active = b
}

func (w *BufWindow) IsActive() bool {
	return w.active
}

func (w *BufWindow) getStartInfo(n, lineN int) ([]byte, int, int, *tcell.Style) {
	tabsize := util.IntOpt(w.Buf.Settings["tabsize"])
	width := 0
	bloc := buffer.Loc{0, lineN}
	b := w.Buf.LineBytes(lineN)
	curStyle := config.DefStyle
	var s *tcell.Style
	for len(b) > 0 {
		r, _, size := util.DecodeCharacter(b)

		curStyle, found := w.getStyle(curStyle, bloc)
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
			screen.SetContent(w.X+x, w.Y+y, ' ', nil, config.DefStyle)
		}
	}
}

// Bottomline returns the line number of the lowest line in the view
// You might think that this is obviously just v.StartLine + v.Height
// but if softwrap is enabled things get complicated since one buffer
// line can take up multiple lines in the view
func (w *BufWindow) Bottomline() int {
	if !w.Buf.Settings["softwrap"].(bool) {
		h := w.StartLine + w.Height - 1
		if w.drawStatus {
			h--
		}
		return h
	}

	l := w.LocFromVisual(buffer.Loc{0, w.Y + w.Height})

	return l.Y
}

// Relocate moves the view window so that the cursor is in view
// This is useful if the user has scrolled far away, and then starts typing
// Returns true if the window location is moved
func (w *BufWindow) Relocate() bool {
	b := w.Buf
	// how many buffer lines are in the view
	height := w.Bottomline() + 1 - w.StartLine
	h := w.Height
	if w.drawStatus {
		h--
	}
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

// LocFromVisual takes a visual location (x and y position) and returns the
// position in the buffer corresponding to the visual location
// Computing the buffer location requires essentially drawing the entire screen
// to account for complications like softwrap, wide characters, and horizontal scrolling
// If the requested position does not correspond to a buffer location it returns
// the nearest position
func (w *BufWindow) LocFromVisual(svloc buffer.Loc) buffer.Loc {
	b := w.Buf

	hasMessage := len(b.Messages) > 0
	bufHeight := w.Height
	if w.drawStatus {
		bufHeight--
	}

	bufWidth := w.Width
	if w.Buf.Settings["scrollbar"].(bool) && w.Buf.LinesNum() > w.Height {
		bufWidth--
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
		if hasMessage {
			vloc.X += 2
		}
		if b.Settings["diffgutter"].(bool) {
			vloc.X++
		}
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

		totalwidth := w.StartCol - nColsBeforeStart

		if svloc.X <= vloc.X+w.X && vloc.Y+w.Y == svloc.Y {
			return bloc
		}
		for len(line) > 0 {
			if vloc.X+w.X == svloc.X && vloc.Y+w.Y == svloc.Y {
				return bloc
			}

			r, _, size := util.DecodeCharacter(line)
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
					if vloc.X+w.X == svloc.X && vloc.Y+w.Y == svloc.Y {
						return bloc
					}
					draw()
				}
			}
			bloc.X++
			line = line[size:]

			totalwidth += width

			// If we reach the end of the window then we either stop or we wrap for softwrap
			if vloc.X >= bufWidth {
				if !softwrap {
					break
				} else {
					vloc.Y++
					if vloc.Y >= bufHeight {
						break
					}
					vloc.X = w.gutterOffset
				}
			}
		}
		if vloc.Y+w.Y == svloc.Y {
			return bloc
		}

		if bloc.Y+1 >= b.LinesNum() || vloc.Y+1 >= bufHeight {
			return bloc
		}

		bloc.X = w.StartCol
		bloc.Y++
	}

	return buffer.Loc{}
}

func (w *BufWindow) drawGutter(vloc *buffer.Loc, bloc *buffer.Loc) {
	char := ' '
	s := config.DefStyle
	for _, m := range w.Buf.Messages {
		if m.Start.Y == bloc.Y || m.End.Y == bloc.Y {
			s = m.Style()
			char = '>'
			break
		}
	}
	screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, char, nil, s)
	vloc.X++
	screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, char, nil, s)
	vloc.X++
}

func (w *BufWindow) drawDiffGutter(backgroundStyle tcell.Style, softwrapped bool, vloc *buffer.Loc, bloc *buffer.Loc) {
	symbol := ' '
	styleName := ""

	switch w.Buf.DiffStatus(bloc.Y) {
	case buffer.DSAdded:
		symbol = '\u258C' // Left half block
		styleName = "diff-added"
	case buffer.DSModified:
		symbol = '\u258C' // Left half block
		styleName = "diff-modified"
	case buffer.DSDeletedAbove:
		if !softwrapped {
			symbol = '\u2594' // Upper one eighth block
			styleName = "diff-deleted"
		}
	}

	style := backgroundStyle
	if s, ok := config.Colorscheme[styleName]; ok {
		foreground, _, _ := s.Decompose()
		style = style.Foreground(foreground)
	}

	screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, symbol, nil, style)
	vloc.X++
}

func (w *BufWindow) drawLineNum(lineNumStyle tcell.Style, softwrapped bool, maxLineNumLength int, vloc *buffer.Loc, bloc *buffer.Loc) {
	cursorLine := w.Buf.GetActiveCursor().Loc.Y
	var lineInt int
	if w.Buf.Settings["relativeruler"] == false || cursorLine == bloc.Y {
		lineInt = bloc.Y + 1
	} else {
		lineInt = bloc.Y - cursorLine
	}
	lineNum := strconv.Itoa(util.Abs(lineInt))

	// Write the spaces before the line number if necessary
	for i := 0; i < maxLineNumLength-len(lineNum); i++ {
		screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, ' ', nil, lineNumStyle)
		vloc.X++
	}
	// Write the actual line number
	for _, ch := range lineNum {
		if softwrapped {
			screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, ' ', nil, lineNumStyle)
		} else {
			screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, ch, nil, lineNumStyle)
		}
		vloc.X++
	}

	// Write the extra space
	screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, ' ', nil, lineNumStyle)
	vloc.X++
}

// getStyle returns the highlight style for the given character position
// If there is no change to the current highlight style it just returns that
func (w *BufWindow) getStyle(style tcell.Style, bloc buffer.Loc) (tcell.Style, bool) {
	if group, ok := w.Buf.Match(bloc.Y)[bloc.X]; ok {
		s := config.GetColor(group.String())
		return s, true
	}
	return style, false
}

func (w *BufWindow) showCursor(x, y int, main bool) {
	if w.active {
		if main {
			screen.ShowCursor(x, y)
		} else {
			screen.ShowFakeCursorMulti(x, y)
		}
	}
}

// displayBuffer draws the buffer being shown in this window on the screen.Screen
func (w *BufWindow) displayBuffer() {
	b := w.Buf

	if w.Height <= 0 || w.Width <= 0 {
		return
	}

	hasMessage := len(b.Messages) > 0
	bufHeight := w.Height
	if w.drawStatus {
		bufHeight--
	}

	bufWidth := w.Width
	if w.Buf.Settings["scrollbar"].(bool) && w.Buf.LinesNum() > w.Height {
		bufWidth--
	}

	if b.ModifiedThisFrame {
		if b.Settings["diffgutter"].(bool) {
			b.UpdateDiff(func(synchronous bool) {
				// If the diff was updated asynchronously, the outer call to
				// displayBuffer might already be completed and we need to
				// schedule a redraw in order to display the new diff.
				// Note that this cannot lead to an infinite recursion
				// because the modifications were cleared above so there won't
				// be another call to UpdateDiff when displayBuffer is called
				// during the redraw.
				if !synchronous {
					screen.Redraw()
				}
			})
		}
		b.ModifiedThisFrame = false
	}

	var matchingBraces []buffer.Loc
	// bracePairs is defined in buffer.go
	if b.Settings["matchbrace"].(bool) {
		for _, bp := range buffer.BracePairs {
			for _, c := range b.GetCursors() {
				if c.HasSelection() {
					continue
				}
				curX := c.X
				curLoc := c.Loc

				r := c.RuneUnder(curX)
				rl := c.RuneUnder(curX - 1)
				if r == bp[0] || r == bp[1] || rl == bp[0] || rl == bp[1] {
					mb, left, found := b.FindMatchingBrace(bp, curLoc)
					if found {
						matchingBraces = append(matchingBraces, mb)
						if !left {
							matchingBraces = append(matchingBraces, curLoc)
						} else {
							matchingBraces = append(matchingBraces, curLoc.Move(-1, b))
						}
					}
				}
			}
		}
	}

	lineNumStyle := config.DefStyle
	if style, ok := config.Colorscheme["line-number"]; ok {
		lineNumStyle = style
	}
	curNumStyle := config.DefStyle
	if style, ok := config.Colorscheme["current-line-number"]; ok {
		if !b.Settings["cursorline"].(bool) {
			curNumStyle = lineNumStyle
		} else {
			curNumStyle = style
		}
	}

	// We need to know the string length of the largest line number
	// so we can pad appropriately when displaying line numbers
	maxLineNumLength := len(strconv.Itoa(b.LinesNum()))

	softwrap := b.Settings["softwrap"].(bool)
	tabsize := util.IntOpt(b.Settings["tabsize"])
	colorcolumn := util.IntOpt(b.Settings["colorcolumn"])

	// this represents the current draw position
	// within the current window
	vloc := buffer.Loc{X: 0, Y: 0}

	// this represents the current draw position in the buffer (char positions)
	bloc := buffer.Loc{X: -1, Y: w.StartLine}

	cursors := b.GetCursors()

	curStyle := config.DefStyle
	for vloc.Y = 0; vloc.Y < bufHeight; vloc.Y++ {
		vloc.X = 0

		currentLine := false
		for _, c := range cursors {
			if bloc.Y == c.Y && w.active {
				currentLine = true
				break
			}
		}

		s := lineNumStyle
		if currentLine {
			s = curNumStyle
		}

		if hasMessage {
			w.drawGutter(&vloc, &bloc)
		}

		if b.Settings["diffgutter"].(bool) {
			w.drawDiffGutter(s, false, &vloc, &bloc)
		}

		if b.Settings["ruler"].(bool) {
			w.drawLineNum(s, false, maxLineNumLength, &vloc, &bloc)
		}

		w.gutterOffset = vloc.X

		line, nColsBeforeStart, bslice, startStyle := w.getStartInfo(w.StartCol, bloc.Y)
		if startStyle != nil {
			curStyle = *startStyle
		}
		bloc.X = bslice

		draw := func(r rune, combc []rune, style tcell.Style, showcursor bool) {
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

					if b.Settings["cursorline"].(bool) && w.active &&
						!c.HasSelection() && c.Y == bloc.Y {
						if s, ok := config.Colorscheme["cursor-line"]; ok {
							fg, _, _ := s.Decompose()
							style = style.Background(fg)
						}
					}
				}

				for _, m := range b.Messages {
					if bloc.GreaterEqual(m.Start) && bloc.LessThan(m.End) ||
						bloc.LessThan(m.End) && bloc.GreaterEqual(m.Start) {
						style = style.Underline(true)
						break
					}
				}

				if r == '\t' {
					indentrunes := []rune(b.Settings["indentchar"].(string))
					// if empty indentchar settings, use space
					if indentrunes == nil || len(indentrunes) == 0 {
						indentrunes = []rune{' '}
					}

					r = indentrunes[0]
					if s, ok := config.Colorscheme["indent-char"]; ok && r != ' ' {
						fg, _, _ := s.Decompose()
						style = style.Foreground(fg)
					}
				}

				if s, ok := config.Colorscheme["color-column"]; ok {
					if colorcolumn != 0 && vloc.X-w.gutterOffset+w.StartCol == colorcolumn {
						fg, _, _ := s.Decompose()
						style = style.Background(fg)
					}
				}

				for _, mb := range matchingBraces {
					if mb.X == bloc.X && mb.Y == bloc.Y {
						style = style.Underline(true)
					}
				}

				screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, r, combc, style)

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

		totalwidth := w.StartCol - nColsBeforeStart
		for len(line) > 0 {
			r, combc, size := util.DecodeCharacter(line)

			curStyle, _ = w.getStyle(curStyle, bloc)

			draw(r, combc, curStyle, true)

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
					draw(char, nil, curStyle, false)
				}
			}
			bloc.X++
			line = line[size:]

			totalwidth += width

			// If we reach the end of the window then we either stop or we wrap for softwrap
			if vloc.X >= bufWidth {
				if !softwrap {
					break
				} else {
					vloc.Y++
					if vloc.Y >= bufHeight {
						break
					}
					vloc.X = 0
					if hasMessage {
						w.drawGutter(&vloc, &bloc)
					}
					if b.Settings["diffgutter"].(bool) {
						w.drawDiffGutter(lineNumStyle, true, &vloc, &bloc)
					}

					// This will draw an empty line number because the current line is wrapped
					if b.Settings["ruler"].(bool) {
						w.drawLineNum(lineNumStyle, true, maxLineNumLength, &vloc, &bloc)
					}
				}
			}
		}

		style := config.DefStyle
		for _, c := range cursors {
			if b.Settings["cursorline"].(bool) && w.active &&
				!c.HasSelection() && c.Y == bloc.Y {
				if s, ok := config.Colorscheme["cursor-line"]; ok {
					fg, _, _ := s.Decompose()
					style = style.Background(fg)
				}
			}
		}
		for i := vloc.X; i < bufWidth; i++ {
			curStyle := style
			if s, ok := config.Colorscheme["color-column"]; ok {
				if colorcolumn != 0 && i-w.gutterOffset+w.StartCol == colorcolumn {
					fg, _, _ := s.Decompose()
					curStyle = style.Background(fg)
				}
			}
			screen.SetContent(i+w.X, vloc.Y+w.Y, ' ', nil, curStyle)
		}

		if vloc.X != bufWidth {
			draw(' ', nil, curStyle, true)
		}

		bloc.X = w.StartCol
		bloc.Y++
		if bloc.Y >= b.LinesNum() {
			break
		}
	}
}

func (w *BufWindow) displayStatusLine() {
	_, h := screen.Screen.Size()
	infoY := h
	if config.GetGlobalOption("infobar").(bool) {
		infoY--
	}

	if w.Buf.Settings["statusline"].(bool) {
		w.drawStatus = true
		w.sline.Display()
	} else if w.Y+w.Height != infoY {
		w.drawStatus = true

		divchars := config.GetGlobalOption("divchars").(string)
		if util.CharacterCountInString(divchars) != 2 {
			divchars = "|-"
		}

		_, _, size := util.DecodeCharacterInString(divchars)
		divchar, combc, _ := util.DecodeCharacterInString(divchars[size:])

		dividerStyle := config.DefStyle
		if style, ok := config.Colorscheme["divider"]; ok {
			dividerStyle = style
		}

		divreverse := config.GetGlobalOption("divreverse").(bool)
		if divreverse {
			dividerStyle = dividerStyle.Reverse(true)
		}

		for x := w.X; x < w.X+w.Width; x++ {
			screen.SetContent(x, w.Y+w.Height-1, divchar, combc, dividerStyle)
		}
	} else {
		w.drawStatus = false
	}
}

func (w *BufWindow) displayScrollBar() {
	if w.Buf.Settings["scrollbar"].(bool) && w.Buf.LinesNum() > w.Height {
		scrollX := w.X + w.Width - 1
		bufHeight := w.Height
		if w.drawStatus {
			bufHeight--
		}
		barsize := int(float64(w.Height) / float64(w.Buf.LinesNum()) * float64(w.Height))
		if barsize < 1 {
			barsize = 1
		}
		barstart := w.Y + int(float64(w.StartLine)/float64(w.Buf.LinesNum())*float64(w.Height))
		for y := barstart; y < util.Min(barstart+barsize, w.Y+bufHeight); y++ {
			screen.SetContent(scrollX, y, '|', nil, config.DefStyle.Reverse(true))
		}
	}
}

// Display displays the buffer and the statusline
func (w *BufWindow) Display() {
	w.displayStatusLine()
	w.displayScrollBar()
	w.displayBuffer()
}
