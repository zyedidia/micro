package main

import (
	"strings"
)

// FromCharPos converts from a character position to an x, y position
func FromCharPos(loc int, buf *Buffer) (int, int) {
	return FromCharPosStart(0, 0, 0, loc, buf)
}

// FromCharPosStart converts from a character position to an x, y position, starting at the specified character location
func FromCharPosStart(startLoc, startX, startY, loc int, buf *Buffer) (int, int) {
	charNum := startLoc
	x, y := startX, startY

	lineLen := Count(buf.lines[y]) + 1
	for charNum+lineLen <= loc {
		charNum += lineLen
		y++
		lineLen = Count(buf.lines[y]) + 1
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

	// Last cursor x position
	lastVisualX int

	// The current selection as a range of character numbers (inclusive)
	curSelection [2]int
	// The original selection as a range of character numbers
	// This is used for line and word selection where it is necessary
	// to know what the original selection was
	origSelection [2]int
}

// SetLoc sets the location of the cursor in terms of character number
// and not x, y location
// It's just a simple wrapper of FromCharPos
func (c *Cursor) SetLoc(loc int) {
	c.x, c.y = FromCharPos(loc, c.v.buf)
	c.lastVisualX = c.GetVisualX()
}

// Loc gets the cursor location in terms of character number instead
// of x, y location
// It's just a simple wrapper of ToCharPos
func (c *Cursor) Loc() int {
	return ToCharPos(c.x, c.y, c.v.buf)
}

// ResetSelection resets the user's selection
func (c *Cursor) ResetSelection() {
	c.curSelection[0] = 0
	c.curSelection[1] = 0
}

// HasSelection returns whether or not the user has selected anything
func (c *Cursor) HasSelection() bool {
	return c.curSelection[0] != c.curSelection[1]
}

// DeleteSelection deletes the currently selected text
func (c *Cursor) DeleteSelection() {
	if c.curSelection[0] > c.curSelection[1] {
		c.v.eh.Remove(c.curSelection[1], c.curSelection[0])
		c.SetLoc(c.curSelection[1])
	} else if c.GetSelection() == "" {
		return
	} else {
		c.v.eh.Remove(c.curSelection[0], c.curSelection[1])
		c.SetLoc(c.curSelection[0])
	}
}

// GetSelection returns the cursor's selection
func (c *Cursor) GetSelection() string {
	if c.curSelection[0] > c.curSelection[1] {
		return string([]rune(c.v.buf.text)[c.curSelection[1]:c.curSelection[0]])
	}
	return string([]rune(c.v.buf.text)[c.curSelection[0]:c.curSelection[1]])
}

// SelectLine selects the current line
func (c *Cursor) SelectLine() {
	c.Start()
	c.curSelection[0] = c.Loc()
	c.End()
	if len(c.v.buf.lines)-1 > c.y {
		c.curSelection[1] = c.Loc() + 1
	} else {
		c.curSelection[1] = c.Loc()
	}

	c.origSelection = c.curSelection
}

// AddLineToSelection adds the current line to the selection
func (c *Cursor) AddLineToSelection() {
	loc := c.Loc()

	if loc < c.origSelection[0] {
		c.Start()
		c.curSelection[0] = c.Loc()
		c.curSelection[1] = c.origSelection[1]
	}
	if loc > c.origSelection[1] {
		c.End()
		c.curSelection[1] = c.Loc()
		c.curSelection[0] = c.origSelection[0]
	}

	if loc < c.origSelection[1] && loc > c.origSelection[0] {
		c.curSelection = c.origSelection
	}
}

// SelectWord selects the word the cursor is currently on
func (c *Cursor) SelectWord() {
	if len(c.v.buf.lines[c.y]) == 0 {
		return
	}

	if !IsWordChar(string(c.RuneUnder(c.x))) {
		loc := c.Loc()
		c.curSelection[0] = loc
		c.curSelection[1] = loc + 1
		c.origSelection = c.curSelection
		return
	}

	forward, backward := c.x, c.x

	for backward > 0 && IsWordChar(string(c.RuneUnder(backward-1))) {
		backward--
	}

	c.curSelection[0] = ToCharPos(backward, c.y, c.v.buf)
	c.origSelection[0] = c.curSelection[0]

	for forward < Count(c.v.buf.lines[c.y])-1 && IsWordChar(string(c.RuneUnder(forward+1))) {
		forward++
	}

	c.curSelection[1] = ToCharPos(forward, c.y, c.v.buf) + 1
	c.origSelection[1] = c.curSelection[1]
}

// AddWordToSelection adds the word the cursor is currently on to the selection
func (c *Cursor) AddWordToSelection() {
	loc := c.Loc()

	if loc > c.origSelection[0] && loc < c.origSelection[1] {
		c.curSelection = c.origSelection
		return
	}

	if loc < c.origSelection[0] {
		backward := c.x

		for backward > 0 && IsWordChar(string(c.RuneUnder(backward-1))) {
			backward--
		}

		c.curSelection[0] = ToCharPos(backward, c.y, c.v.buf)
		c.curSelection[1] = c.origSelection[1]
	}

	if loc > c.origSelection[1] {
		forward := c.x

		for forward < Count(c.v.buf.lines[c.y])-1 && IsWordChar(string(c.RuneUnder(forward+1))) {
			forward++
		}

		c.curSelection[1] = ToCharPos(forward, c.y, c.v.buf) + 1
		c.curSelection[0] = c.origSelection[0]
	}
}

// SelectTo selects from the current cursor location to the given location
func (c *Cursor) SelectTo(loc int) {
	if loc > c.origSelection[0] {
		c.curSelection[0] = c.origSelection[0]
		c.curSelection[1] = loc
	} else {
		c.curSelection[0] = loc
		c.curSelection[1] = c.origSelection[0] + 1
	}
}

// WordRight moves the cursor one word to the right
func (c *Cursor) WordRight() {
	c.Right()
	for IsWhitespace(c.RuneUnder(c.x)) {
		c.Right()
	}
	for !IsWhitespace(c.RuneUnder(c.x)) {
		c.Right()
	}
}

// WordLeft moves the cursor one word to the left
func (c *Cursor) WordLeft() {
	c.Left()
	for IsWhitespace(c.RuneUnder(c.x)) {
		c.Left()
	}
	for !IsWhitespace(c.RuneUnder(c.x)) {
		c.Left()
	}
	c.Right()
}

// RuneUnder returns the rune under the given x position
func (c *Cursor) RuneUnder(x int) rune {
	line := []rune(c.v.buf.lines[c.y])
	if len(line) == 0 {
		return '\n'
	}
	if x >= len(line) {
		return '\n'
	} else if x < 0 {
		x = 0
	}
	return line[x]
}

// Up moves the cursor up one line (if possible)
func (c *Cursor) Up() {
	if c.y > 0 {
		c.y--

		runes := []rune(c.v.buf.lines[c.y])
		c.x = c.GetCharPosInLine(c.y, c.lastVisualX)
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
		c.x = c.GetCharPosInLine(c.y, c.lastVisualX)
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
	c.lastVisualX = c.GetVisualX()
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
	c.lastVisualX = c.GetVisualX()
}

// End moves the cursor to the end of the line it is on
func (c *Cursor) End() {
	c.x = Count(c.v.buf.lines[c.y])
	c.lastVisualX = c.GetVisualX()
}

// Start moves the cursor to the start of the line it is on
func (c *Cursor) Start() {
	c.x = 0
	c.lastVisualX = c.GetVisualX()
}

// GetCharPosInLine gets the char position of a visual x y coordinate (this is necessary because tabs are 1 char but 4 visual spaces)
func (c *Cursor) GetCharPosInLine(lineNum, visualPos int) int {
	// Get the tab size
	tabSize := int(settings["tabsize"].(float64))
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
	tabSize := int(settings["tabsize"].(float64))
	return c.x + NumOccurences(string(runes[:c.x]), '\t')*(tabSize-1)
}

// Relocate makes sure that the cursor is inside the bounds of the buffer
// If it isn't, it moves it to be within the buffer's lines
func (c *Cursor) Relocate() {
	if c.y < 0 {
		c.y = 0
	} else if c.y >= len(c.v.buf.lines) {
		c.y = len(c.v.buf.lines) - 1
	}

	if c.x < 0 {
		c.x = 0
	} else if c.x > Count(c.v.buf.lines[c.y]) {
		c.x = Count(c.v.buf.lines[c.y])
	}
}

// Display draws the cursor to the screen at the correct position
func (c *Cursor) Display() {
	// Don't draw the cursor if it is out of the viewport or if it has a selection
	if (c.y-c.v.topline < 0 || c.y-c.v.topline > c.v.height-1) || c.HasSelection() {
		screen.HideCursor()
	} else {
		screen.ShowCursor(c.GetVisualX()+c.v.lineNumOffset-c.v.leftCol, c.y-c.v.topline)
	}
}
