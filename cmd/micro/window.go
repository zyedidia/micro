package main

import (
	"strconv"
	"unicode/utf8"

	runewidth "github.com/mattn/go-runewidth"
	"github.com/zyedidia/tcell"
)

type Window struct {
	// X and Y coordinates for the top left of the window
	X int
	Y int

	// Width and Height for the window
	Width  int
	Height int

	// Which line in the buffer to start displaying at (vertical scroll)
	StartLine int
	// Which visual column in the to start displaying at (horizontal scroll)
	StartCol int

	// Buffer being shown in this window
	Buf *Buffer

	sline *StatusLine
}

func NewWindow(x, y, width, height int, buf *Buffer) *Window {
	w := new(Window)
	w.X, w.Y, w.Width, w.Height, w.Buf = x, y, width, height, buf

	w.sline = NewStatusLine(w)

	return w
}

func (w *Window) DrawLineNum(lineNumStyle tcell.Style, softwrapped bool, maxLineNumLength int, vloc *Loc, bloc *Loc) {
	lineNum := strconv.Itoa(bloc.Y + 1)

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

// GetStyle returns the highlight style for the given character position
// If there is no change to the current highlight style it just returns that
func (w *Window) GetStyle(style tcell.Style, bloc Loc, r rune) tcell.Style {
	if group, ok := w.Buf.Match(bloc.Y)[bloc.X]; ok {
		s := GetColor(group.String())
		return s
	}
	return style
}

// DisplayBuffer draws the buffer being shown in this window on the screen
func (w *Window) DisplayBuffer() {
	b := w.Buf

	bufHeight := w.Height
	if b.Settings["statusline"].(bool) {
		bufHeight--
	}

	// TODO: Rehighlighting
	// start := w.StartLine
	if b.Settings["syntax"].(bool) && b.syntaxDef != nil {
		// 	if start > 0 && b.lines[start-1].rehighlight {
		// 		b.highlighter.ReHighlightLine(b, start-1)
		// 		b.lines[start-1].rehighlight = false
		// 	}
		//
		// 	b.highlighter.ReHighlightStates(b, start)
		//
		b.highlighter.HighlightMatches(b, w.StartLine, w.StartLine+bufHeight)
	}

	lineNumStyle := defStyle
	if style, ok := colorscheme["line-number"]; ok {
		lineNumStyle = style
	}

	// We need to know the string length of the largest line number
	// so we can pad appropriately when displaying line numbers
	maxLineNumLength := len(strconv.Itoa(len(b.lines)))

	tabsize := int(b.Settings["tabsize"].(float64))
	softwrap := b.Settings["softwrap"].(bool)

	// this represents the current draw position
	// within the current window
	vloc := Loc{0, 0}

	// this represents the current draw position in the buffer (char positions)
	bloc := Loc{w.StartCol, w.StartLine}

	curStyle := defStyle
	for vloc.Y = 0; vloc.Y < bufHeight; vloc.Y++ {
		vloc.X = 0
		if b.Settings["ruler"].(bool) {
			w.DrawLineNum(lineNumStyle, false, maxLineNumLength, &vloc, &bloc)
		}

		line := b.LineBytes(bloc.Y)
		line, nColsBeforeStart := SliceVisualEnd(line, bloc.X, tabsize)
		totalwidth := bloc.X - nColsBeforeStart
		for len(line) > 0 {
			r, size := utf8.DecodeRune(line)

			curStyle = w.GetStyle(curStyle, bloc, r)

			if nColsBeforeStart <= 0 {
				screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, r, nil, curStyle)
				vloc.X++
			}
			nColsBeforeStart--

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

			bloc.X++
			line = line[size:]

			// Draw any extra characters either spaces for tabs or @ for incomplete wide runes
			if width > 1 {
				for i := 1; i < width; i++ {
					if nColsBeforeStart <= 0 {
						screen.SetContent(w.X+vloc.X, w.Y+vloc.Y, char, nil, curStyle)
						vloc.X++
					}
					nColsBeforeStart--
				}
			}
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
					// This will draw an empty line number because the current line is wrapped
					w.DrawLineNum(lineNumStyle, true, maxLineNumLength, &vloc, &bloc)
				}
			}
		}
		bloc.X = w.StartCol
		bloc.Y++
		if bloc.Y >= len(b.lines) {
			break
		}
	}
}

func (w *Window) DisplayStatusLine() {
	w.sline.Display()
}
