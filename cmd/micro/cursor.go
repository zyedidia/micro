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
	X int
	Y int

	// Last cursor x position
	LastVisualX int

	// The current selection as a range of character numbers (inclusive)
	CurSelection [2]int
	// The original selection as a range of character numbers
	// This is used for line and word selection where it is necessary
	// to know what the original selection was
	OrigSelection [2]int
}

// SetLoc sets the location of the cursor in terms of character number
// and not x, y location
// It's just a simple wrapper of FromCharPos
func (c *Cursor) SetLoc(loc int) {
	c.X, c.Y = FromCharPos(loc, c.v.buf)
	c.LastVisualX = c.GetVisualX()
}

// Loc gets the cursor location in terms of character number instead
// of x, y location
// It's just a simple wrapper of ToCharPos
func (c *Cursor) Loc() int {
	return ToCharPos(c.X, c.Y, c.v.buf)
}

// ResetSelection resets the user's selection
func (c *Cursor) ResetSelection() {
	c.CurSelection[0] = 0
	c.CurSelection[1] = 0
}

// HasSelection returns whether or not the user has selected anything
func (c *Cursor) HasSelection() bool {
	return c.CurSelection[0] != c.CurSelection[1]
}

// DeleteSelection deletes the currently selected text
func (c *Cursor) DeleteSelection() {
	if c.CurSelection[0] > c.CurSelection[1] {
		c.v.buf.eh.Remove(c.CurSelection[1], c.CurSelection[0])
		c.SetLoc(c.CurSelection[1])
	} else {
		c.v.buf.eh.Remove(c.CurSelection[0], c.CurSelection[1])
		c.SetLoc(c.CurSelection[0])
	}
}

// GetSelection returns the cursor's selection
func (c *Cursor) GetSelection() string {
	if c.CurSelection[0] > c.CurSelection[1] {
		return string([]rune(c.v.buf.text)[c.CurSelection[1]:c.CurSelection[0]])
	}
	return string([]rune(c.v.buf.text)[c.CurSelection[0]:c.CurSelection[1]])
}

// SelectLine selects the current line
func (c *Cursor) SelectLine() {
	c.Start()
	c.CurSelection[0] = c.Loc()
	c.End()
	c.CurSelection[1] = c.Loc() + 1

	c.OrigSelection = c.CurSelection
}

// AddLineToSelection adds the current line to the selection
func (c *Cursor) AddLineToSelection() {
	loc := c.Loc()

	if loc < c.OrigSelection[0] {
		c.Start()
		c.CurSelection[0] = c.Loc()
		c.CurSelection[1] = c.OrigSelection[1]
	}
	if loc > c.OrigSelection[1] {
		c.End()
		c.CurSelection[1] = c.Loc()
		c.CurSelection[0] = c.OrigSelection[0]
	}

	if loc < c.OrigSelection[1] && loc > c.OrigSelection[0] {
		c.CurSelection = c.OrigSelection
	}
}

// SelectWord selects the word the cursor is currently on
func (c *Cursor) SelectWord() {
	if len(c.v.buf.lines[c.Y]) == 0 {
		return
	}

	if !IsWordChar(string(c.RuneUnder(c.X))) {
		loc := c.Loc()
		c.CurSelection[0] = loc
		c.CurSelection[1] = loc + 1
		c.OrigSelection = c.CurSelection
		return
	}

	forward, backward := c.X, c.X

	for backward > 0 && IsWordChar(string(c.RuneUnder(backward-1))) {
		backward--
	}

	c.CurSelection[0] = ToCharPos(backward, c.Y, c.v.buf)
	c.OrigSelection[0] = c.CurSelection[0]

	for forward < Count(c.v.buf.lines[c.Y])-1 && IsWordChar(string(c.RuneUnder(forward+1))) {
		forward++
	}

	c.CurSelection[1] = ToCharPos(forward, c.Y, c.v.buf) + 1
	c.OrigSelection[1] = c.CurSelection[1]
}

// AddWordToSelection adds the word the cursor is currently on to the selection
func (c *Cursor) AddWordToSelection() {
	loc := c.Loc()

	if loc > c.OrigSelection[0] && loc < c.OrigSelection[1] {
		c.CurSelection = c.OrigSelection
		return
	}

	if loc < c.OrigSelection[0] {
		backward := c.X

		for backward > 0 && IsWordChar(string(c.RuneUnder(backward-1))) {
			backward--
		}

		c.CurSelection[0] = ToCharPos(backward, c.Y, c.v.buf)
		c.CurSelection[1] = c.OrigSelection[1]
	}

	if loc > c.OrigSelection[1] {
		forward := c.X

		for forward < Count(c.v.buf.lines[c.Y])-1 && IsWordChar(string(c.RuneUnder(forward+1))) {
			forward++
		}

		c.CurSelection[1] = ToCharPos(forward, c.Y, c.v.buf) + 1
		c.CurSelection[0] = c.OrigSelection[0]
	}
}

// RuneUnder returns the rune under the given x position
func (c *Cursor) RuneUnder(x int) rune {
	line := []rune(c.v.buf.lines[c.Y])
	if x >= len(line) {
		x = len(line) - 1
	} else if x < 0 {
		x = 0
	}
	return line[x]
}

// Up moves the cursor up one line (if possible)
func (c *Cursor) Up() {
	if c.Y > 0 {
		c.Y--

		runes := []rune(c.v.buf.lines[c.Y])
		c.X = c.GetCharPosInLine(c.Y, c.LastVisualX)
		if c.X > len(runes) {
			c.X = len(runes)
		}
	}
}

// Down moves the cursor down one line (if possible)
func (c *Cursor) Down() {
	if c.Y < len(c.v.buf.lines)-1 {
		c.Y++

		runes := []rune(c.v.buf.lines[c.Y])
		c.X = c.GetCharPosInLine(c.Y, c.LastVisualX)
		if c.X > len(runes) {
			c.X = len(runes)
		}
	}
}

// Left moves the cursor left one cell (if possible) or to the last line if it is at the beginning
func (c *Cursor) Left() {
	if c.Loc() == 0 {
		return
	}
	if c.X > 0 {
		c.X--
	} else {
		c.Up()
		c.End()
	}
	c.LastVisualX = c.GetVisualX()
}

// Right moves the cursor right one cell (if possible) or to the next line if it is at the end
func (c *Cursor) Right() {
	if c.Loc() == c.v.buf.Len() {
		return
	}
	if c.X < Count(c.v.buf.lines[c.Y]) {
		c.X++
	} else {
		c.Down()
		c.Start()
	}
	c.LastVisualX = c.GetVisualX()
}

// End moves the cursor to the end of the line it is on
func (c *Cursor) End() {
	c.X = Count(c.v.buf.lines[c.Y])
	c.LastVisualX = c.GetVisualX()
}

// Start moves the cursor to the start of the line it is on
func (c *Cursor) Start() {
	c.X = 0
	c.LastVisualX = c.GetVisualX()
}

// GetCharPosInLine gets the char position of a visual x y coordinate (this is necessary because tabs are 1 char but 4 visual spaces)
func (c *Cursor) GetCharPosInLine(lineNum, visualPos int) int {
	// Get the tab size
	tabSize := settings.TabSize
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
	runes := []rune(c.v.buf.lines[c.Y])
	tabSize := settings.TabSize
	return c.X + NumOccurences(string(runes[:c.X]), '\t')*(tabSize-1)
}

// Display draws the cursor to the screen at the correct position
func (c *Cursor) Display() {
	// Don't draw the cursor if it is out of the viewport or if it has a selection
	if (c.Y-c.v.topline < 0 || c.Y-c.v.topline > c.v.height-1) || c.HasSelection() {
		screen.HideCursor()
	} else {
		screen.ShowCursor(c.GetVisualX()+c.v.lineNumOffset-c.v.leftCol, c.Y-c.v.topline)
	}
}
