package main

import (
	// "github.com/gdamore/tcell"
	"strings"
)

// Cursor stores the location of the cursor in the view
type Cursor struct {
	v *View

	x   int
	y   int
	loc int

	selectionStart  int
	selectionStartX int
	selectionStartY int
	selectionEnd    int
}

func (c *Cursor) resetSelection() {
	c.selectionStart = 0
	c.selectionEnd = 0
}

func (c *Cursor) hasSelection() bool {
	return c.selectionEnd != c.selectionStart
}

func (c *Cursor) deleteSelected() {
	if c.selectionStart > c.selectionEnd {
		c.v.buf.remove(c.selectionEnd, c.selectionStart+1)
		// Since the cursor is already at the selection start we don't need to move
	} else {
		c.v.buf.remove(c.selectionStart, c.selectionEnd+1)
		c.loc -= c.selectionEnd - c.selectionStart
		c.x = c.selectionStartX
		c.y = c.selectionStartY
	}
}

func (c *Cursor) runeUnder() rune {
	line := c.v.buf.lines[c.y]
	if c.x >= count(line) {
		return ' '
	}
	return []rune(line)[c.x]
}

func (c *Cursor) up() {
	if c.y > 0 {
		c.loc -= count(c.v.buf.lines[c.y][:c.x])
		// Count the newline
		c.loc--
		c.y--

		if c.x > count(c.v.buf.lines[c.y]) {
			c.x = count(c.v.buf.lines[c.y])
		}

		c.loc -= count(c.v.buf.lines[c.y][c.x:])
	}
}
func (c *Cursor) down() {
	if c.y < len(c.v.buf.lines)-1 {
		c.loc += count(c.v.buf.lines[c.y][c.x:])
		// Count the newline
		c.loc++
		c.y++

		if c.x > count(c.v.buf.lines[c.y]) {
			c.x = count(c.v.buf.lines[c.y])
		}

		c.loc += count(c.v.buf.lines[c.y][:c.x])
	}
}
func (c *Cursor) left() {
	if c.x > 0 {
		c.loc--
		c.x--
	} else {
		c.up()
		c.end()
	}
}
func (c *Cursor) right() {
	if c.x < count(c.v.buf.lines[c.y]) {
		c.loc++
		c.x++
	} else {
		c.down()
		c.start()
	}
}

func (c *Cursor) end() {
	c.loc += count(c.v.buf.lines[c.y][c.x:])
	c.x = count(c.v.buf.lines[c.y])
}

func (c *Cursor) start() {
	c.loc -= count(c.v.buf.lines[c.y][:c.x])
	c.x = 0
}

func (c *Cursor) getCharPosInLine(lineNum, visualPos int) int {
	visualLine := strings.Replace(c.v.buf.lines[lineNum], "\t", "\t"+emptyString(tabSize-1), -1)
	if visualPos > count(visualLine) {
		visualPos = count(visualLine)
	}
	numTabs := numOccurences(visualLine[:visualPos], '\t')
	if visualPos >= (tabSize-1)*numTabs {
		return visualPos - (tabSize-1)*numTabs
	} else {
		return visualPos / tabSize
	}
}

func (c *Cursor) getVisualX() int {
	return c.x + numOccurences(c.v.buf.lines[c.y][:c.x], '\t')*(tabSize-1)
}

func (c *Cursor) distance(x, y int) int {
	// Same line
	if y == c.y {
		return x - c.x
	}

	var distance int
	if y > c.y {
		distance += count(c.v.buf.lines[c.y][c.x:])
		// Newline
		distance++
		i := 1
		for y != c.y+i {
			distance += count(c.v.buf.lines[c.y+i])
			// Newline
			distance++
			i++
		}
		if x < count(c.v.buf.lines[y]) {
			distance += count(c.v.buf.lines[y][:x])
		} else {
			distance += count(c.v.buf.lines[y])
		}
		return distance
	}

	distance -= count(c.v.buf.lines[c.y][:c.x])
	// Newline
	distance--
	i := 1
	for y != c.y-i {
		distance -= count(c.v.buf.lines[c.y-i])
		// Newline
		distance--
		i++
	}
	if x >= 0 {
		distance -= count(c.v.buf.lines[y][x:])
	}
	return distance
}

func (c *Cursor) display() {
	if c.y-c.v.topline < 0 || c.y-c.v.topline > c.v.height-1 {
		c.v.s.HideCursor()
	} else {
		voffset := numOccurences(c.v.buf.lines[c.y][:c.x], '\t') * (tabSize - 1)
		c.v.s.ShowCursor(c.x+voffset, c.y-c.v.topline)
		// cursorStyle := tcell.StyleDefault.Reverse(true)
		// c.v.s.SetContent(c.x+voffset, c.y-c.v.topline, c.runeUnder(), nil, cursorStyle)
	}
}
