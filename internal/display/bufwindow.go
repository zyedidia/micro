package display

import (
	"strconv"

	runewidth "github.com/mattn/go-runewidth"
	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/micro/v2/internal/util"
	"github.com/zyedidia/tcell/v2"
)

// The BufWindow provides a way of displaying a certain section of a buffer.
type BufWindow struct {
	*View

	// Buffer being shown in this window
	Buf *buffer.Buffer

	active bool

	sline *StatusLine

	bufWidth         int
	bufHeight        int
	gutterOffset     int
	hasMessage       bool
	maxLineNumLength int
	drawDivider      bool
}

// NewBufWindow creates a new window at a location in the screen with a width and height
func NewBufWindow(x, y, width, height int, buf *buffer.Buffer) *BufWindow {
	w := new(BufWindow)
	w.View = new(View)
	w.X, w.Y, w.Width, w.Height = x, y, width, height
	w.SetBuffer(buf)
	w.active = true

	w.sline = NewStatusLine(w)

	return w
}

// SetBuffer sets this window's buffer.
func (w *BufWindow) SetBuffer(b *buffer.Buffer) {
	w.Buf = b
	b.OptionCallback = func(option string, nativeValue interface{}) {
		if option == "softwrap" {
			if nativeValue.(bool) {
				w.StartCol = 0
			} else {
				w.StartLine.Row = 0
			}
		}

		if option == "softwrap" || option == "wordwrap" {
			w.Relocate()
			for _, c := range w.Buf.GetCursors() {
				c.LastVisualX = c.GetVisualX()
			}
		}
	}
	b.GetVisualX = func(loc buffer.Loc) int {
		return w.VLocFromLoc(loc).VisualX
	}
}

// GetView gets the view.
func (w *BufWindow) GetView() *View {
	return w.View
}

// GetView sets the view.
func (w *BufWindow) SetView(view *View) {
	w.View = view
}

// Resize resizes this window.
func (w *BufWindow) Resize(width, height int) {
	w.Width, w.Height = width, height
	w.updateDisplayInfo()

	w.Relocate()
}

// SetActive marks the window as active.
func (w *BufWindow) SetActive(b bool) {
	w.active = b
}

// IsActive returns true if this window is active.
func (w *BufWindow) IsActive() bool {
	return w.active
}

// BufView returns the width, height and x,y location of the actual buffer.
// It is not exactly the same as the whole window which also contains gutter,
// ruler, scrollbar and statusline.
func (w *BufWindow) BufView() View {
	return View{
		X:         w.X + w.gutterOffset,
		Y:         w.Y,
		Width:     w.bufWidth,
		Height:    w.bufHeight,
		StartLine: w.StartLine,
		StartCol:  w.StartCol,
	}
}

func (w *BufWindow) updateDisplayInfo() {
	b := w.Buf

	w.drawDivider = false
	if !b.Settings["statusline"].(bool) {
		_, h := screen.Screen.Size()
		infoY := h
		if config.GetGlobalOption("infobar").(bool) {
			infoY--
		}
		if w.Y+w.Height != infoY {
			w.drawDivider = true
		}
	}

	w.bufHeight = w.Height
	if b.Settings["statusline"].(bool) || w.drawDivider {
		w.bufHeight--
	}

	w.hasMessage = len(b.Messages) > 0

	// We need to know the string length of the largest line number
	// so we can pad appropriately when displaying line numbers
	w.maxLineNumLength = len(strconv.Itoa(b.LinesNum()))

	w.gutterOffset = 0
	if w.hasMessage {
		w.gutterOffset += 2
	}
	if b.Settings["diffgutter"].(bool) {
		w.gutterOffset++
	}
	if b.Settings["ruler"].(bool) {
		w.gutterOffset += w.maxLineNumLength + 1
	}

	prevBufWidth := w.bufWidth

	w.bufWidth = w.Width - w.gutterOffset
	if w.Buf.Settings["scrollbar"].(bool) && w.Buf.LinesNum() > w.Height {
		w.bufWidth--
	}

	if w.bufWidth != prevBufWidth && w.Buf.Settings["softwrap"].(bool) {
		for _, c := range w.Buf.GetCursors() {
			c.LastVisualX = c.GetVisualX()
		}
	}
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

// Relocate moves the view window so that the cursor is in view
// This is useful if the user has scrolled far away, and then starts typing
// Returns true if the window location is moved
func (w *BufWindow) Relocate() bool {
	b := w.Buf
	height := w.bufHeight
	ret := false
	activeC := w.Buf.GetActiveCursor()
	scrollmargin := int(b.Settings["scrollmargin"].(float64))

	c := w.SLocFromLoc(activeC.Loc)
	bStart := SLoc{0, 0}
	bEnd := w.SLocFromLoc(b.End())

	if c.LessThan(w.Scroll(w.StartLine, scrollmargin)) && c.GreaterThan(w.Scroll(bStart, scrollmargin-1)) {
		w.StartLine = w.Scroll(c, -scrollmargin)
		ret = true
	} else if c.LessThan(w.StartLine) {
		w.StartLine = c
		ret = true
	}
	if c.GreaterThan(w.Scroll(w.StartLine, height-1-scrollmargin)) && c.LessEqual(w.Scroll(bEnd, -scrollmargin)) {
		w.StartLine = w.Scroll(c, -height+1+scrollmargin)
		ret = true
	} else if c.GreaterThan(w.Scroll(bEnd, -scrollmargin)) && c.GreaterThan(w.Scroll(w.StartLine, height-1)) {
		w.StartLine = w.Scroll(bEnd, -height+1)
		ret = true
	}

	// horizontal relocation (scrolling)
	if !b.Settings["softwrap"].(bool) {
		cx := activeC.GetVisualX()
		rw := runewidth.RuneWidth(activeC.RuneUnder(activeC.X))
		if rw == 0 {
			rw = 1 // tab or newline
		}

		if cx < w.StartCol {
			w.StartCol = cx
			ret = true
		}
		if cx+w.gutterOffset+rw > w.StartCol+w.Width {
			w.StartCol = cx - w.Width + w.gutterOffset + rw
			ret = true
		}
	}
	return ret
}

// LocFromVisual takes a visual location (x and y position) and returns the
// position in the buffer corresponding to the visual location
// If the requested position does not correspond to a buffer location it returns
// the nearest position
func (w *BufWindow) LocFromVisual(svloc buffer.Loc) buffer.Loc {
	vx := svloc.X - w.X - w.gutterOffset
	if vx < 0 {
		vx = 0
	}
	vloc := VLoc{
		SLoc:    w.Scroll(w.StartLine, svloc.Y-w.Y),
		VisualX: vx + w.StartCol,
	}
	return w.LocFromVLoc(vloc)
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

func (w *BufWindow) drawLineNum(lineNumStyle tcell.Style, softwrapped bool, vloc *buffer.Loc, bloc *buffer.Loc) {
	cursorLine := w.Buf.GetActiveCursor().Loc.Y
	var lineInt int
	if w.Buf.Settings["relativeruler"] == false || cursorLine == bloc.Y {
		lineInt = bloc.Y + 1
	} else {
		lineInt = bloc.Y - cursorLine
	}
	lineNum := strconv.Itoa(util.Abs(lineInt))

	// Write the spaces before the line number if necessary
	for i := 0; i < w.maxLineNumLength-len(lineNum); i++ {
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

	maxWidth := w.gutterOffset + w.bufWidth

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

	softwrap := b.Settings["softwrap"].(bool)
	wordwrap := softwrap && b.Settings["wordwrap"].(bool)

	tabsize := util.IntOpt(b.Settings["tabsize"])
	colorcolumn := util.IntOpt(b.Settings["colorcolumn"])

	// this represents the current draw position
	// within the current window
	vloc := buffer.Loc{X: 0, Y: 0}
	if softwrap {
		// the start line may be partially out of the current window
		vloc.Y = -w.StartLine.Row
	}

	// this represents the current draw position in the buffer (char positions)
	bloc := buffer.Loc{X: -1, Y: w.StartLine.Line}

	cursors := b.GetCursors()

	curStyle := config.DefStyle
	for ; vloc.Y < w.bufHeight; vloc.Y++ {
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

		if vloc.Y >= 0 {
			if w.hasMessage {
				w.drawGutter(&vloc, &bloc)
			}

			if b.Settings["diffgutter"].(bool) {
				w.drawDiffGutter(s, false, &vloc, &bloc)
			}

			if b.Settings["ruler"].(bool) {
				w.drawLineNum(s, false, &vloc, &bloc)
			}
		} else {
			vloc.X = w.gutterOffset
		}

		line, nColsBeforeStart, bslice, startStyle := w.getStartInfo(w.StartCol, bloc.Y)
		if startStyle != nil {
			curStyle = *startStyle
		}
		bloc.X = bslice

		draw := func(r rune, combc []rune, style tcell.Style, highlight bool, showcursor bool) {
			if nColsBeforeStart <= 0 && vloc.Y >= 0 {
				if highlight {
					if w.Buf.HighlightSearch && w.Buf.SearchMatch(bloc) {
						style = config.DefStyle.Reverse(true)
						if s, ok := config.Colorscheme["hlsearch"]; ok {
							style = s
						}
					}

					_, origBg, _ := style.Decompose()
					_, defBg, _ := config.DefStyle.Decompose()

					// syntax or hlsearch highlighting with non-default background takes precedence
					// over cursor-line and color-column
					dontOverrideBackground := origBg != defBg

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

						if b.Settings["cursorline"].(bool) && w.active && !dontOverrideBackground &&
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
						if len(indentrunes) == 0 {
							indentrunes = []rune{' '}
						}

						r = indentrunes[0]
						if s, ok := config.Colorscheme["indent-char"]; ok && r != ' ' {
							fg, _, _ := s.Decompose()
							style = style.Foreground(fg)
						}
					}

					if s, ok := config.Colorscheme["color-column"]; ok {
						if colorcolumn != 0 && vloc.X-w.gutterOffset+w.StartCol == colorcolumn && !dontOverrideBackground {
							fg, _, _ := s.Decompose()
							style = style.Background(fg)
						}
					}

					for _, mb := range matchingBraces {
						if mb.X == bloc.X && mb.Y == bloc.Y {
							style = style.Underline(true)
						}
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
			}
			if nColsBeforeStart <= 0 {
				vloc.X++
			}
			nColsBeforeStart--
		}

		wrap := func() {
			vloc.X = 0
			if w.hasMessage {
				w.drawGutter(&vloc, &bloc)
			}
			if b.Settings["diffgutter"].(bool) {
				w.drawDiffGutter(lineNumStyle, true, &vloc, &bloc)
			}

			// This will draw an empty line number because the current line is wrapped
			if b.Settings["ruler"].(bool) {
				w.drawLineNum(lineNumStyle, true, &vloc, &bloc)
			}
		}

		type glyph struct {
			r     rune
			combc []rune
			style tcell.Style
			width int
		}

		var word []glyph
		if wordwrap {
			word = make([]glyph, 0, w.bufWidth)
		} else {
			word = make([]glyph, 0, 1)
		}
		wordwidth := 0

		totalwidth := w.StartCol - nColsBeforeStart
		for len(line) > 0 {
			r, combc, size := util.DecodeCharacter(line)
			line = line[size:]

			loc := buffer.Loc{X: bloc.X + len(word), Y: bloc.Y}
			curStyle, _ = w.getStyle(curStyle, loc)

			width := 0

			switch r {
			case '\t':
				ts := tabsize - (totalwidth % tabsize)
				width = util.Min(ts, maxWidth-vloc.X)
				totalwidth += ts
			default:
				width = runewidth.RuneWidth(r)
				totalwidth += width
			}

			word = append(word, glyph{r, combc, curStyle, width})
			wordwidth += width

			// Collect a complete word to know its width.
			// If wordwrap is off, every single character is a complete "word".
			if wordwrap {
				if !util.IsWhitespace(r) && len(line) > 0 && wordwidth < w.bufWidth {
					continue
				}
			}

			// If a word (or just a wide rune) does not fit in the window
			if vloc.X+wordwidth > maxWidth && vloc.X > w.gutterOffset {
				for vloc.X < maxWidth {
					draw(' ', nil, config.DefStyle, false, false)
				}

				// We either stop or we wrap to draw the word in the next line
				if !softwrap {
					break
				} else {
					vloc.Y++
					if vloc.Y >= w.bufHeight {
						break
					}
					wrap()
				}
			}

			for _, r := range word {
				draw(r.r, r.combc, r.style, true, true)

				// Draw any extra characters either spaces for tabs or @ for incomplete wide runes
				if r.width > 1 {
					char := ' '
					if r.r != '\t' {
						char = '@'
					}

					for i := 1; i < r.width; i++ {
						draw(char, nil, r.style, true, false)
					}
				}
				bloc.X++
			}

			word = word[:0]
			wordwidth = 0

			// If we reach the end of the window then we either stop or we wrap for softwrap
			if vloc.X >= maxWidth {
				if !softwrap {
					break
				} else {
					vloc.Y++
					if vloc.Y >= w.bufHeight {
						break
					}
					wrap()
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
		for i := vloc.X; i < maxWidth; i++ {
			curStyle := style
			if s, ok := config.Colorscheme["color-column"]; ok {
				if colorcolumn != 0 && i-w.gutterOffset+w.StartCol == colorcolumn {
					fg, _, _ := s.Decompose()
					curStyle = style.Background(fg)
				}
			}
			screen.SetContent(i+w.X, vloc.Y+w.Y, ' ', nil, curStyle)
		}

		if vloc.X != maxWidth {
			// Display newline within a selection
			draw(' ', nil, config.DefStyle, true, true)
		}

		bloc.X = w.StartCol
		bloc.Y++
		if bloc.Y >= b.LinesNum() {
			break
		}
	}
}

func (w *BufWindow) displayStatusLine() {
	if w.Buf.Settings["statusline"].(bool) {
		w.sline.Display()
	} else if w.drawDivider {
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
	}
}

func (w *BufWindow) displayScrollBar() {
	if w.Buf.Settings["scrollbar"].(bool) && w.Buf.LinesNum() > w.Height {
		scrollX := w.X + w.Width - 1
		barsize := int(float64(w.Height) / float64(w.Buf.LinesNum()) * float64(w.Height))
		if barsize < 1 {
			barsize = 1
		}
		barstart := w.Y + int(float64(w.StartLine.Line)/float64(w.Buf.LinesNum())*float64(w.Height))

		scrollBarStyle := config.DefStyle.Reverse(true)
		if style, ok := config.Colorscheme["scrollbar"]; ok {
			scrollBarStyle = style
		}

		for y := barstart; y < util.Min(barstart+barsize, w.Y+w.bufHeight); y++ {
			screen.SetContent(scrollX, y, '|', nil, scrollBarStyle)
		}
	}
}

// Display displays the buffer and the statusline
func (w *BufWindow) Display() {
	w.updateDisplayInfo()

	w.displayStatusLine()
	w.displayScrollBar()
	w.displayBuffer()
}
