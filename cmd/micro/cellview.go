package main

import (
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/zyedidia/tcell"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func VisualToCharPos(visualIndex int, str string, tabsize int) int {
	visualPos := 0
	charPos := 0
	for _, c := range str {
		width := StringWidth(string(c), tabsize)

		if visualPos+width > visualIndex {
			return charPos
		}

		visualPos += width
		charPos++
	}

	return 0
}

type Char struct {
	visualLoc Loc
	realLoc   Loc
	char      rune
	style     tcell.Style
}

type CellView struct {
	lines [][]*Char
}

func (c *CellView) Draw(buf *Buffer, top, height, left, width int) {
	tabsize := int(buf.Settings["tabsize"].(float64))
	softwrap := buf.Settings["softwrap"].(bool)
	indentchar := []rune(buf.Settings["indentchar"].(string))[0]

	start := buf.Cursor.Y
	startTime := time.Now()
	matches := buf.highlighter.ReHighlight(buf, start)
	elapsed := time.Since(startTime)
	for i, m := range matches {
		buf.matches[start+i] = m
	}
	messenger.Message("Rehighlighted ", len(matches), " lines in ", elapsed)

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

		colN := VisualToCharPos(left, lineStr, tabsize)
		viewCol := 0

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
			if group, ok := buf.matches[lineN][colN]; ok {
				curStyle = GetColor(group)
			}

			char := line[colN]

			if char == '\t' {
				c.lines[viewLine][viewCol] = &Char{Loc{viewCol, viewLine}, Loc{colN, lineN}, indentchar, curStyle}
				viewCol += tabsize - viewCol%tabsize
			} else if runewidth.RuneWidth(char) > 1 {
				c.lines[viewLine][viewCol] = &Char{Loc{viewCol, viewLine}, Loc{colN, lineN}, char, curStyle}
				viewCol += runewidth.RuneWidth(char)
			} else {
				c.lines[viewLine][viewCol] = &Char{Loc{viewCol, viewLine}, Loc{colN, lineN}, char, curStyle}
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
		if group, ok := buf.matches[lineN][len(line)]; ok {
			curStyle = GetColor(group)
		}

		// newline
		viewLine++
		lineN++
	}
}
