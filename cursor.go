package main

import (
	"strings"
)

// The Cursor struct stores the location of the cursor in the view
type Cursor struct {
	v *View

	// We need three variables here because we insert text at loc but
	// display the cursor at x, y
	x   int
	y   int
	loc int

	// Start of the selection in charNum
	selectionStart int

	// We store the x, y of the start because when the user deletes the selection
	// the cursor needs to go back to the start, and this is the simplest way
	selectionStartX int
	selectionStartY int

	// End of the selection in charNum
	// We don't need to store the x, y here because when if the user is selecting backwards
	// and they delete the selection, the cursor is already in the right place
	selectionEnd int
}

// ResetSelection resets the user's selection
func (c *Cursor) ResetSelection() {
	c.selectionStart = 0
	c.selectionEnd = 0
}

// HasSelection returns whether or not the user has selected anything
func (c *Cursor) HasSelection() bool {
	return c.selectionEnd != c.selectionStart
}

// DeleteSelection deletes the currently selected text
func (c *Cursor) DeleteSelection() {
	if c.selectionStart > c.selectionEnd {
		c.v.eh.Remove(c.selectionEnd, c.selectionStart+1)
		// Since the cursor is already at the selection start we don't need to move
	} else {
		c.v.eh.Remove(c.selectionStart, c.selectionEnd+1)
		c.loc -= c.selectionEnd - c.selectionStart
		c.x = c.selectionStartX
		c.y = c.selectionStartY
	}
}

// GetSelection returns the cursor's selection
func (c *Cursor) GetSelection() string {
	if c.selectionStart > c.selectionEnd {
		return string([]rune(c.v.buf.text)[c.selectionEnd : c.selectionStart+1])
	}
	return string([]rune(c.v.buf.text)[c.selectionStart : c.selectionEnd+1])
}

// RuneUnder returns the rune under the cursor
func (c *Cursor) RuneUnder() rune {
	line := c.v.buf.lines[c.y]
	if c.x >= Count(line) {
		return ' '
	}
	return []rune(line)[c.x]
}

// Up moves the cursor up one line (if possible)
func (c *Cursor) Up() {
	if c.y > 0 {
		c.loc -= Count(c.v.buf.lines[c.y][:c.x])
		// Count the newline
		c.loc--
		c.y--

		if c.x > Count(c.v.buf.lines[c.y]) {
			c.x = Count(c.v.buf.lines[c.y])
		}

		c.loc -= Count(c.v.buf.lines[c.y][c.x:])
	}
}

// Down moves the cursor down one line (if possible)
func (c *Cursor) Down() {
	if c.y < len(c.v.buf.lines)-1 {
		c.loc += Count(c.v.buf.lines[c.y][c.x:])
		// Count the newline
		c.loc++
		c.y++

		if c.x > Count(c.v.buf.lines[c.y]) {
			c.x = Count(c.v.buf.lines[c.y])
		}

		c.loc += Count(c.v.buf.lines[c.y][:c.x])
	}
}

// Left moves the cursor left one cell (if possible) or to the last line if it is at the beginning
func (c *Cursor) Left() {
	if c.loc == 0 {
		return
	}
	if c.x > 0 {
		c.loc--
		c.x--
	} else {
		c.Up()
		c.End()
	}
}

// Right moves the cursor right one cell (if possible) or to the next line if it is at the end
func (c *Cursor) Right() {
	if c.loc == c.v.buf.Len() {
		return
	}
	if c.x < Count(c.v.buf.lines[c.y]) {
		c.loc++
		c.x++
	} else {
		c.Down()
		c.Start()
	}
}

// End moves the cursor to the end of the line it is on
func (c *Cursor) End() {
	c.loc += Count(c.v.buf.lines[c.y][c.x:])
	c.x = Count(c.v.buf.lines[c.y])
}

// Start moves the cursor to the start of the line it is on
func (c *Cursor) Start() {
	c.loc -= Count(c.v.buf.lines[c.y][:c.x])
	c.x = 0
}

// GetCharPosInLine gets the char position of a visual x y coordinate (this is necessary because tabs are 1 char but 4 visual spaces)
func (c *Cursor) GetCharPosInLine(lineNum, visualPos int) int {
	visualLine := strings.Replace(c.v.buf.lines[lineNum], "\t", "\t"+EmptyString(tabSize-1), -1)
	if visualPos > Count(visualLine) {
		visualPos = Count(visualLine)
	}
	numTabs := NumOccurences(visualLine[:visualPos], '\t')
	if visualPos >= (tabSize-1)*numTabs {
		return visualPos - (tabSize-1)*numTabs
	}
	return visualPos / tabSize
}

// GetVisualX returns the x value of the cursor in visual spaces
func (c *Cursor) GetVisualX() int {
	return c.x + NumOccurences(c.v.buf.lines[c.y][:c.x], '\t')*(tabSize-1)
}

// Distance returns the distance between the cursor and x, y in runes
func (c *Cursor) Distance(x, y int) int {
	// Same line
	if y == c.y {
		return x - c.x
	}

	var distance int
	if y > c.y {
		distance += Count(c.v.buf.lines[c.y][c.x:])
		// Newline
		distance++
		i := 1
		for y != c.y+i {
			distance += Count(c.v.buf.lines[c.y+i])
			// Newline
			distance++
			i++
		}
		if x < Count(c.v.buf.lines[y]) {
			distance += Count(c.v.buf.lines[y][:x])
		} else {
			distance += Count(c.v.buf.lines[y])
		}
		return distance
	}

	distance -= Count(c.v.buf.lines[c.y][:c.x])
	// Newline
	distance--
	i := 1
	for y != c.y-i {
		distance -= Count(c.v.buf.lines[c.y-i])
		// Newline
		distance--
		i++
	}
	if x >= 0 {
		distance -= Count(c.v.buf.lines[y][x:])
	}
	return distance
}

// Display draws the cursor to the screen at the correct position
func (c *Cursor) Display() {
	if c.y-c.v.topline < 0 || c.y-c.v.topline > c.v.height-1 {
		c.v.s.HideCursor()
	} else {
		voffset := NumOccurences(c.v.buf.lines[c.y][:c.x], '\t') * (tabSize - 1)
		c.v.s.ShowCursor(c.x+voffset+c.v.lineNumOffset, c.y-c.v.topline)
		// cursorStyle := tcell.StyleDefault.Reverse(true)
		// c.v.s.SetContent(c.x+voffset, c.y-c.v.topline, c.runeUnder(), nil, cursorStyle)
	}
}
