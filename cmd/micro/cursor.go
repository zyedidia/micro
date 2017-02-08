package main

import "github.com/zyedidia/clipboard"

// The Cursor struct stores the location of the cursor in the view
// The complicated part about the cursor is storing its location.
// The cursor must be displayed at an x, y location, but since the buffer
// uses a rope to store text, to insert text we must have an index. It
// is also simpler to use character indicies for other tasks such as
// selection.
type Cursor struct {
	buf *Buffer
	Loc

	// Last cursor x position
	LastVisualX int

	// The current selection as a range of character numbers (inclusive)
	CurSelection [2]Loc
	// The original selection as a range of character numbers
	// This is used for line and word selection where it is necessary
	// to know what the original selection was
	OrigSelection [2]Loc
}

// Goto puts the cursor at the given cursor's location and gives the current cursor its selection too
func (c *Cursor) Goto(b Cursor) {
	c.X, c.Y, c.LastVisualX = b.X, b.Y, b.LastVisualX
	c.OrigSelection, c.CurSelection = b.OrigSelection, b.CurSelection
}

// CopySelection copies the user's selection to either "primary" or "clipboard"
func (c *Cursor) CopySelection(target string) {
	if c.HasSelection() {
		if target != "primary" || c.buf.Settings["useprimary"].(bool) {
			clipboard.WriteAll(c.GetSelection(), target)
		}
	}
}

// ResetSelection resets the user's selection
func (c *Cursor) ResetSelection() {
	c.CurSelection[0] = c.buf.Start()
	c.CurSelection[1] = c.buf.Start()
}

// SetSelectionStart sets the start of the selection
func (c *Cursor) SetSelectionStart(pos Loc) {
	c.CurSelection[0] = pos
}

// SetSelectionEnd sets the end of the selection
func (c *Cursor) SetSelectionEnd(pos Loc) {
	c.CurSelection[1] = pos
}

// HasSelection returns whether or not the user has selected anything
func (c *Cursor) HasSelection() bool {
	return c.CurSelection[0] != c.CurSelection[1]
}

// DeleteSelection deletes the currently selected text
func (c *Cursor) DeleteSelection() {
	if c.CurSelection[0].GreaterThan(c.CurSelection[1]) {
		c.buf.Remove(c.CurSelection[1], c.CurSelection[0])
		c.Loc = c.CurSelection[1]
	} else if !c.HasSelection() {
		return
	} else {
		c.buf.Remove(c.CurSelection[0], c.CurSelection[1])
		c.Loc = c.CurSelection[0]
	}
}

// GetSelection returns the cursor's selection
func (c *Cursor) GetSelection() string {
	if InBounds(c.CurSelection[0], c.buf) && InBounds(c.CurSelection[1], c.buf) {
		if c.CurSelection[0].GreaterThan(c.CurSelection[1]) {
			return c.buf.Substr(c.CurSelection[1], c.CurSelection[0])
		}
		return c.buf.Substr(c.CurSelection[0], c.CurSelection[1])
	}
	return ""
}

// SelectLine selects the current line
func (c *Cursor) SelectLine() {
	c.Start()
	c.SetSelectionStart(c.Loc)
	c.End()
	if c.buf.NumLines-1 > c.Y {
		c.SetSelectionEnd(c.Loc.Move(1, c.buf))
	} else {
		c.SetSelectionEnd(c.Loc)
	}

	c.OrigSelection = c.CurSelection
}

// AddLineToSelection adds the current line to the selection
func (c *Cursor) AddLineToSelection() {
	if c.Loc.LessThan(c.OrigSelection[0]) {
		c.Start()
		c.SetSelectionStart(c.Loc)
		c.SetSelectionEnd(c.OrigSelection[1])
	}
	if c.Loc.GreaterThan(c.OrigSelection[1]) {
		c.End()
		c.SetSelectionEnd(c.Loc.Move(1, c.buf))
		c.SetSelectionStart(c.OrigSelection[0])
	}

	if c.Loc.LessThan(c.OrigSelection[1]) && c.Loc.GreaterThan(c.OrigSelection[0]) {
		c.CurSelection = c.OrigSelection
	}
}

// SelectWord selects the word the cursor is currently on
func (c *Cursor) SelectWord() {
	if len(c.buf.Line(c.Y)) == 0 {
		return
	}

	if !IsWordChar(string(c.RuneUnder(c.X))) {
		c.SetSelectionStart(c.Loc)
		c.SetSelectionEnd(c.Loc.Move(1, c.buf))
		c.OrigSelection = c.CurSelection
		return
	}

	forward, backward := c.X, c.X

	for backward > 0 && IsWordChar(string(c.RuneUnder(backward-1))) {
		backward--
	}

	c.SetSelectionStart(Loc{backward, c.Y})
	c.OrigSelection[0] = c.CurSelection[0]

	for forward < Count(c.buf.Line(c.Y))-1 && IsWordChar(string(c.RuneUnder(forward+1))) {
		forward++
	}

	c.SetSelectionEnd(Loc{forward, c.Y}.Move(1, c.buf))
	c.OrigSelection[1] = c.CurSelection[1]
	c.Loc = c.CurSelection[1]
}

// AddWordToSelection adds the word the cursor is currently on to the selection
func (c *Cursor) AddWordToSelection() {
	if c.Loc.GreaterThan(c.OrigSelection[0]) && c.Loc.LessThan(c.OrigSelection[1]) {
		c.CurSelection = c.OrigSelection
		return
	}

	if c.Loc.LessThan(c.OrigSelection[0]) {
		backward := c.X

		for backward > 0 && IsWordChar(string(c.RuneUnder(backward-1))) {
			backward--
		}

		c.SetSelectionStart(Loc{backward, c.Y})
		c.SetSelectionEnd(c.OrigSelection[1])
	}

	if c.Loc.GreaterThan(c.OrigSelection[1]) {
		forward := c.X

		for forward < Count(c.buf.Line(c.Y))-1 && IsWordChar(string(c.RuneUnder(forward+1))) {
			forward++
		}

		c.SetSelectionEnd(Loc{forward, c.Y}.Move(1, c.buf))
		c.SetSelectionStart(c.OrigSelection[0])
	}

	c.Loc = c.CurSelection[1]
}

// SelectTo selects from the current cursor location to the given location
func (c *Cursor) SelectTo(loc Loc) {
	if loc.GreaterThan(c.OrigSelection[0]) {
		c.SetSelectionStart(c.OrigSelection[0])
		c.SetSelectionEnd(loc)
	} else {
		c.SetSelectionStart(loc)
		c.SetSelectionEnd(c.OrigSelection[0])
	}
}

// WordRight moves the cursor one word to the right
func (c *Cursor) WordRight() {
	for IsWhitespace(c.RuneUnder(c.X)) {
		if c.X == Count(c.buf.Line(c.Y)) {
			c.Right()
			return
		}
		c.Right()
	}
	c.Right()
	for IsWordChar(string(c.RuneUnder(c.X))) {
		if c.X == Count(c.buf.Line(c.Y)) {
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
	c.Left()
	for IsWordChar(string(c.RuneUnder(c.X))) {
		if c.X == 0 {
			return
		}
		c.Left()
	}
	c.Right()
}

// RuneUnder returns the rune under the given x position
func (c *Cursor) RuneUnder(x int) rune {
	line := []rune(c.buf.Line(c.Y))
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
	runes := []rune(c.buf.Line(c.Y))
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
	if c.Loc == c.buf.Start() {
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
	if c.Loc == c.buf.End() {
		return
	}
	if c.X < Count(c.buf.Line(c.Y)) {
		c.X++
	} else {
		c.Down()
		c.Start()
	}
	c.LastVisualX = c.GetVisualX()
}

// End moves the cursor to the end of the line it is on
func (c *Cursor) End() {
	c.X = Count(c.buf.Line(c.Y))
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
	tabSize := int(c.buf.Settings["tabsize"].(float64))
	visualLineLen := StringWidth(c.buf.Line(lineNum), tabSize)
	if visualPos > visualLineLen {
		visualPos = visualLineLen
	}
	width := WidthOfLargeRunes(c.buf.Line(lineNum), tabSize)
	if visualPos >= width {
		return visualPos - width
	}
	return visualPos / tabSize
}

// GetVisualX returns the x value of the cursor in visual spaces
func (c *Cursor) GetVisualX() int {
	runes := []rune(c.buf.Line(c.Y))
	tabSize := int(c.buf.Settings["tabsize"].(float64))
	return StringWidth(string(runes[:c.X]), tabSize)
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
	} else if c.X > Count(c.buf.Line(c.Y)) {
		c.X = Count(c.buf.Line(c.Y))
	}
}
