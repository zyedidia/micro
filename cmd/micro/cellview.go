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

	if len(c.lines) != height {
		c.lines = make([][]*Char, height)
	}

	viewLine := 0
	lineN := top

	for viewLine < height {
		lineStr := string(buf.lines[lineN])
		line := []rune(lineStr)

		colN := VisualToCharPos(left, lineStr, tabsize)
		viewCol := 0

		// We'll either draw the length of the line, or the width of the screen
		// whichever is smaller
		lineLength := min(StringWidth(lineStr, tabsize), width)
		if len(c.lines[viewLine]) != lineLength {
			c.lines[viewLine] = make([]*Char, lineLength)
		}

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
			char := line[colN]

			if char == '\t' {
				c.lines[viewLine][viewCol] = &Char{Loc{viewCol, viewLine}, Loc{colN, lineN}, indentchar, tcell.StyleDefault}
				// TODO: this always adds 4 spaces but it should really add just the remainder to the next tab location
				viewCol += tabsize
			} else if runewidth.RuneWidth(char) > 1 {
				c.lines[viewLine][viewCol] = &Char{Loc{viewCol, viewLine}, Loc{colN, lineN}, char, tcell.StyleDefault}
				viewCol += runewidth.RuneWidth(char)
			} else {
				c.lines[viewLine][viewCol] = &Char{Loc{viewCol, viewLine}, Loc{colN, lineN}, char, tcell.StyleDefault}
				viewCol++
			}
			colN++

			if wrap && viewCol >= width {
				viewLine++

				nextLine := line[colN:]
				lineLength := min(StringWidth(string(nextLine), tabsize), width)
				if len(c.lines[viewLine]) != lineLength {
					c.lines[viewLine] = make([]*Char, lineLength)
				}

				viewCol = 0

				// If we go too far soft wrapping we have to cut off
				if viewLine >= height {
					break
				}
			}
		}

		// newline
		viewLine++
		lineN++
	}
}
