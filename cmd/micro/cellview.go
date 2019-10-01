package main

import (
	"github.com/mattn/go-runewidth"
	"github.com/zyedidia/tcell"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func visualToCharPos(visualIndex, lineN int, str string, buf *Buffer, tabsize int) (int, int, *tcell.Style) {
	var (
		charPos   int
		lineIdx   int
		lastWidth int
		style     *tcell.Style
		width     int
		rw        int
	)
	for i, c := range str {
		// width := StringWidth(str[:i], tabsize)

		if group, ok := buf.Match(lineN)[charPos]; ok {
			s := GetColor(group.String())
			style = &s
		}

		if width >= visualIndex {
			return charPos, visualIndex - lastWidth, style
		}

		if i != 0 {
			charPos++
			lineIdx += rw
		}
		lastWidth = width
		if c == '\t' {
			rw = tabsize - (lineIdx % tabsize)
			width += rw
		} else {
			rw = runewidth.RuneWidth(c)
			width += rw
		}
	}

	return -1, -1, style
}

type Char struct {
	visualLoc Loc
	realLoc   Loc
	char      rune
	// The actual character that is drawn
	// This is only different from char if it's for example hidden character
	drawChar rune
	style    tcell.Style
	width    int
}

type CellView struct {
	lines [][]*Char
}

func (c *CellView) Draw(buf *Buffer, top, height, left, width int) {
	if width <= 0 {
		return
	}

	matchingBrace := Loc{-1, -1}
	// bracePairs is defined in buffer.go
	if buf.Settings["matchbrace"].(bool) {
		for _, bp := range bracePairs {
			curX := buf.Cursor.X
			curLoc := buf.Cursor.Loc
			if buf.Settings["matchbraceleft"].(bool) {
				if curX > 0 {
					curX--
					curLoc = curLoc.Move(-1, buf)
				}
			}

			r := buf.Cursor.RuneUnder(curX)
			if r == bp[0] || r == bp[1] {
				matchingBrace = buf.FindMatchingBrace(bp, curLoc)
			}
		}
	}

	tabsize := int(buf.Settings["tabsize"].(float64))
	softwrap := buf.Settings["softwrap"].(bool)
	indentrunes := []rune(buf.Settings["indentchar"].(string))
	// if empty indentchar settings, use space
	if indentrunes == nil || len(indentrunes) == 0 {
		indentrunes = []rune{' '}
	}
	indentchar := indentrunes[0]

	start := buf.Cursor.Y
	if buf.Settings["syntax"].(bool) && buf.syntaxDef != nil {
		if start > 0 && buf.lines[start-1].rehighlight {
			buf.highlighter.ReHighlightLine(buf, start-1)
			buf.lines[start-1].rehighlight = false
		}

		buf.highlighter.ReHighlightStates(buf, start)

		buf.highlighter.HighlightMatches(buf, top, top+height)
	}

	c.lines = make([][]*Char, 0)

	viewLine := 0
	lineN := top

	curStyle := defStyle
	for viewLine < height {
		if lineN >= len(buf.lines) {
			break
		}

		lineStr := buf.Line(lineN)
		line := []rune(lineStr)

		colN, startOffset, startStyle := visualToCharPos(left, lineN, lineStr, buf, tabsize)
		if colN < 0 {
			colN = len(line)
		}
		viewCol := -startOffset
		if startStyle != nil {
			curStyle = *startStyle
		}

		// We'll either draw the length of the line, or the width of the screen
		// whichever is smaller
		lineLength := min(StringWidth(lineStr, tabsize), width)
		c.lines = append(c.lines, make([]*Char, lineLength))

		wrap := false
		// We only need to wrap if the length of the line is greater than the width of the terminal screen
		if softwrap && StringWidth(lineStr, tabsize) > width {
			wrap = true
			// We're going to draw the entire line now
			lineLength = StringWidth(lineStr, tabsize)
		}

		for viewCol < lineLength {
			if colN >= len(line) {
				break
			}
			if group, ok := buf.Match(lineN)[colN]; ok {
				curStyle = GetColor(group.String())
			}

			char := line[colN]

			if viewCol >= 0 {
				st := curStyle
				if colN == matchingBrace.X && lineN == matchingBrace.Y && !buf.Cursor.HasSelection() {
					st = curStyle.Reverse(true)
				}
				if viewCol < len(c.lines[viewLine]) {
					c.lines[viewLine][viewCol] = &Char{Loc{viewCol, viewLine}, Loc{colN, lineN}, char, char, st, 1}
				}
			}
			if char == '\t' {
				charWidth := tabsize - (viewCol+left)%tabsize
				if viewCol >= 0 {
					c.lines[viewLine][viewCol].drawChar = indentchar
					c.lines[viewLine][viewCol].width = charWidth

					indentStyle := curStyle
					ch := buf.Settings["indentchar"].(string)
					if group, ok := colorscheme["indent-char"]; ok && !IsStrWhitespace(ch) && ch != "" {
						indentStyle = group
					}

					c.lines[viewLine][viewCol].style = indentStyle
				}

				for i := 1; i < charWidth; i++ {
					viewCol++
					if viewCol >= 0 && viewCol < lineLength && viewCol < len(c.lines[viewLine]) {
						c.lines[viewLine][viewCol] = &Char{Loc{viewCol, viewLine}, Loc{colN, lineN}, char, ' ', curStyle, 1}
					}
				}
				viewCol++
			} else if runewidth.RuneWidth(char) > 1 {
				charWidth := runewidth.RuneWidth(char)
				if viewCol >= 0 {
					c.lines[viewLine][viewCol].width = charWidth
				}
				for i := 1; i < charWidth; i++ {
					viewCol++
					if viewCol >= 0 && viewCol < lineLength && viewCol < len(c.lines[viewLine]) {
						c.lines[viewLine][viewCol] = &Char{Loc{viewCol, viewLine}, Loc{colN, lineN}, char, ' ', curStyle, 1}
					}
				}
				viewCol++
			} else {
				viewCol++
			}
			colN++

			if wrap && viewCol >= width {
				viewLine++

				// If we go too far soft wrapping we have to cut off
				if viewLine >= height {
					break
				}

				nextLine := line[colN:]
				lineLength := min(StringWidth(string(nextLine), tabsize), width)
				c.lines = append(c.lines, make([]*Char, lineLength))

				viewCol = 0
			}

		}
		if group, ok := buf.Match(lineN)[len(line)]; ok {
			curStyle = GetColor(group.String())
		}

		// newline
		viewLine++
		lineN++
	}

	for i := top; i < top+height; i++ {
		if i >= buf.NumLines {
			break
		}
		buf.SetMatch(i, nil)
	}
}
