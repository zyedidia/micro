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

	lineLen := Count(buf.Lines[y]) + 1
	for charNum+lineLen <= loc {
		charNum += lineLen
		y++
		if y >= buf.NumLines {
			return 0, 0
		}
		lineLen = Count(buf.Lines[y]) + 1
	}
	x = loc - charNum

	return x, y
}

// ToCharPos converts from an x, y position to a character position
func ToCharPos(x, y int, buf *Buffer) int {
	loc := 0
	for i := 0; i < y; i++ {
		// + 1 for the newline
		loc += Count(buf.Lines[i]) + 1
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
	buf *Buffer

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

// Goto puts the cursor at the given cursor's location and gives the current cursor its selection too
func (c *Cursor) Goto(b Cursor) {
	c.X, c.Y, c.LastVisualX = b.X, b.Y, b.LastVisualX
	c.OrigSelection, c.CurSelection = b.OrigSelection, b.CurSelection
}

// SetLoc sets the location of the cursor in terms of character number
// and not x, y location
// It's just a simple wrapper of FromCharPos
func (c *Cursor) SetLoc(loc int) {
	c.X, c.Y = FromCharPos(loc, c.buf)
	c.LastVisualX = c.GetVisualX()
}

// Loc gets the cursor location in terms of character number instead
// of x, y location
// It's just a simple wrapper of ToCharPos
func (c *Cursor) Loc() int {
	return ToCharPos(c.X, c.Y, c.buf)
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
		c.buf.Remove(c.CurSelection[1], c.CurSelection[0])
		c.SetLoc(c.CurSelection[1])
	} else if c.GetSelection() == "" {
		return
	} else {
		c.buf.Remove(c.CurSelection[0], c.CurSelection[1])
		c.SetLoc(c.CurSelection[0])
	}
}

// GetSelection returns the cursor's selection
func (c *Cursor) GetSelection() string {
	if c.CurSelection[0] > c.CurSelection[1] {
		return c.buf.Substr(c.CurSelection[1], c.CurSelection[0])
	}
	return c.buf.Substr(c.CurSelection[0], c.CurSelection[1])
}

// SelectLine selects the current line
func (c *Cursor) SelectLine() {
	c.Start()
	c.CurSelection[0] = c.Loc()
	c.End()
	if c.buf.NumLines-1 > c.Y {
		c.CurSelection[1] = c.Loc() + 1
	} else {
		c.CurSelection[1] = c.Loc()
	}

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
	if len(c.buf.Lines[c.Y]) == 0 {
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

	c.CurSelection[0] = ToCharPos(backward, c.Y, c.buf)
	c.OrigSelection[0] = c.CurSelection[0]

	for forward < Count(c.buf.Lines[c.Y])-1 && IsWordChar(string(c.RuneUnder(forward+1))) {
		forward++
	}

	c.CurSelection[1] = ToCharPos(forward, c.Y, c.buf) + 1
	c.OrigSelection[1] = c.CurSelection[1]
	c.SetLoc(c.CurSelection[1])
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

		c.CurSelection[0] = ToCharPos(backward, c.Y, c.buf)
		c.CurSelection[1] = c.OrigSelection[1]
	}

	if loc > c.OrigSelection[1] {
		forward := c.X

		for forward < Count(c.buf.Lines[c.Y])-1 && IsWordChar(string(c.RuneUnder(forward+1))) {
			forward++
		}

		c.CurSelection[1] = ToCharPos(forward, c.Y, c.buf) + 1
		c.CurSelection[0] = c.OrigSelection[0]
	}

	c.SetLoc(c.CurSelection[1])
}

// SelectTo selects from the current cursor location to the given location
func (c *Cursor) SelectTo(loc int) {
	if loc > c.OrigSelection[0] {
		c.CurSelection[0] = c.OrigSelection[0]
		c.CurSelection[1] = loc
	} else {
		c.CurSelection[0] = loc
		c.CurSelection[1] = c.OrigSelection[0]
	}
}

// WordRight moves the cursor one word to the right
func (c *Cursor) WordRight() {
	c.Right()
	for IsWhitespace(c.RuneUnder(c.X)) {
		if c.X == Count(c.buf.Lines[c.Y]) {
			return
		}
		c.Right()
	}
	for !IsWhitespace(c.RuneUnder(c.X)) {
		if c.X == Count(c.buf.Lines[c.Y]) {
			return
		}
		c.Right()
	}
}

// WordLeft moves the cursor one word to the left
func (c *Cursor) WordLeft() {
	c.Left()
	for IsWhitespace(c.RuneUnder(c.X)) {
		if c.X == 0 {
			return
		}
		c.Left()
	}
	for !IsWhitespace(c.RuneUnder(c.X)) {
		if c.X == 0 {
			return
		}
		c.Left()
	}
	c.Right()
}

// RuneUnder returns the rune under the given x position
func (c *Cursor) RuneUnder(x int) rune {
	line := []rune(c.buf.Lines[c.Y])
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

// UpN moves the cursor up N lines (if possible)
func (c *Cursor) UpN(amount int) {
	proposedY := c.Y - amount
	if proposedY < 0 {
		proposedY = 0
	} else if proposedY >= c.buf.NumLines {
		proposedY = c.buf.NumLines - 1
	}
	if proposedY == c.Y {
		return
	}

	c.Y = proposedY
	runes := []rune(c.buf.Lines[c.Y])
	c.X = c.GetCharPosInLine(c.Y, c.LastVisualX)
	if c.X > len(runes) {
		c.X = len(runes)
	}
}

// DownN moves the cursor down N lines (if possible)
func (c *Cursor) DownN(amount int) {
	c.UpN(-amount)
}

// Up moves the cursor up one line (if possible)
func (c *Cursor) Up() {
	c.UpN(1)
}

// Down moves the cursor down one line (if possible)
func (c *Cursor) Down() {
	c.DownN(1)
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
	if c.Loc() == c.buf.Len() {
		return
	}
	// TermMessage(Count(c.buf.Lines[c.Y]))
	if c.X < Count(c.buf.Lines[c.Y]) {
		c.X++
	} else {
		c.Down()
		c.Start()
	}
	c.LastVisualX = c.GetVisualX()
}

// End moves the cursor to the end of the line it is on
func (c *Cursor) End() {
	c.X = Count(c.buf.Lines[c.Y])
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
	tabSize := int(settings["tabsize"].(float64))
	// This is the visual line -- every \t replaced with the correct number of spaces
	visualLine := strings.Replace(c.buf.Lines[lineNum], "\t", "\t"+Spaces(tabSize-1), -1)
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
	runes := []rune(c.buf.Lines[c.Y])
	tabSize := int(settings["tabsize"].(float64))
	return c.X + NumOccurences(string(runes[:c.X]), '\t')*(tabSize-1)
}

// Relocate makes sure that the cursor is inside the bounds of the buffer
// If it isn't, it moves it to be within the buffer's lines
func (c *Cursor) Relocate() {
	if c.Y < 0 {
		c.Y = 0
	} else if c.Y >= c.buf.NumLines {
		c.Y = c.buf.NumLines - 1
	}

	if c.X < 0 {
		c.X = 0
	} else if c.X > Count(c.buf.Lines[c.Y]) {
		c.X = Count(c.buf.Lines[c.Y])
	}
}
