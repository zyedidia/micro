package main

import (
	"io/ioutil"
	"strconv"
	"strings"
)

// FromCharPos converts from a character position to an x, y position
func FromCharPos(loc int, buf *Buffer) (int, int) {
	charNum := 0
	x, y := 0, 0

	for charNum+Count(buf.lines[y])+1 <= loc {
		charNum += Count(buf.lines[y]) + 1
		y++
	}
	x = loc - charNum

	return x, y
}

// ToCharPos converts from an x, y position to a character position
func ToCharPos(x, y int, buf *Buffer) int {
	loc := 0
	for i := 0; i < y; i++ {
		// + 1 for the newline
		loc += Count(buf.lines[i]) + 1
	}
	loc += x
	return loc
}

// The Cursor struct stores the location of the cursor in the view
// The complicated part about the cursor is storing its location.
// The cursor must be displayed at an x, y location, but since the buffer
// uses a rope to store text, to insert text we must have an index. It
// is also simpler to use character indicies for other tasks such as
// selection.
type Cursor struct {
	v *View

	// The cursor display location
	x int
	y int

	selectionStart int
	selectionEnd   int
}

// SetLoc sets the location of the cursor in terms of character number
// and not x, y location
// It's just a simple wrapper of FromCharPos
func (c *Cursor) SetLoc(loc int) {
	c.x, c.y = FromCharPos(loc, c.v.buf)
}

// Loc gets the cursor location in terms of character number instead
// of x, y location
// It's just a simple wrapper of ToCharPos
func (c *Cursor) Loc() int {
	return ToCharPos(c.x, c.y, c.v.buf)
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
		c.SetLoc(c.selectionEnd)
	} else {
		c.v.eh.Remove(c.selectionStart, c.selectionEnd+1)
		c.SetLoc(c.selectionStart)
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
		c.y--

		runes := []rune(c.v.buf.lines[c.y])
		if c.x > len(runes) {
			c.x = len(runes)
		}
	}
}

// Down moves the cursor down one line (if possible)
func (c *Cursor) Down() {
	if c.y < len(c.v.buf.lines)-1 {
		c.y++

		runes := []rune(c.v.buf.lines[c.y])
		if c.x > len(runes) {
			c.x = len(runes)
		}
	}
}

// Left moves the cursor left one cell (if possible) or to the last line if it is at the beginning
func (c *Cursor) Left() {
	if c.Loc() == 0 {
		return
	}
	if c.x > 0 {
		c.x--
	} else {
		c.Up()
		c.End()
	}
}

// Right moves the cursor right one cell (if possible) or to the next line if it is at the end
func (c *Cursor) Right() {
	if c.Loc() == c.v.buf.Len() {
		return
	}
	if c.x < Count(c.v.buf.lines[c.y]) {
		c.x++
	} else {
		c.Down()
		c.Start()
	}
}

// End moves the cursor to the end of the line it is on
func (c *Cursor) End() {
	c.x = Count(c.v.buf.lines[c.y])
}

// Start moves the cursor to the start of the line it is on
func (c *Cursor) Start() {
	c.x = 0
}

// GetCharPosInLine gets the char position of a visual x y coordinate (this is necessary because tabs are 1 char but 4 visual spaces)
func (c *Cursor) GetCharPosInLine(lineNum, visualPos int) int {
	// Get the tab size
	tabSize := options["tabsize"].(int)
	// This is the visual line -- every \t replaced with the correct number of spaces
	visualLine := strings.Replace(c.v.buf.lines[lineNum], "\t", "\t"+Spaces(tabSize-1), -1)
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
	runes := []rune(c.v.buf.lines[c.y])
	tabSize := options["tabsize"].(int)
	return c.x + NumOccurences(string(runes[:c.x]), '\t')*(tabSize-1)
}

// Display draws the cursor to the screen at the correct position
func (c *Cursor) Display() {
	if c.y-c.v.topline < 0 || c.y-c.v.topline > c.v.height-1 {
		screen.HideCursor()
	} else {
		screen.ShowCursor(c.GetVisualX()+c.v.lineNumOffset, c.y-c.v.topline)
	}
}
