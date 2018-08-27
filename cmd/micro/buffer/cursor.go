package buffer

import (
	"unicode/utf8"

	runewidth "github.com/mattn/go-runewidth"
	"github.com/zyedidia/clipboard"
	"github.com/zyedidia/micro/cmd/micro/util"
)

// InBounds returns whether the given location is a valid character position in the given buffer
func InBounds(pos Loc, buf *Buffer) bool {
	if pos.Y < 0 || pos.Y >= len(buf.lines) || pos.X < 0 || pos.X > utf8.RuneCount(buf.LineBytes(pos.Y)) {
		return false
	}

	return true
}

// The Cursor struct stores the location of the cursor in the buffer
// as well as the selection
type Cursor struct {
	Buf *Buffer
	Loc

	// Last cursor x position
	LastVisualX int

	// The current selection as a range of character numbers (inclusive)
	CurSelection [2]Loc
	// The original selection as a range of character numbers
	// This is used for line and word selection where it is necessary
	// to know what the original selection was
	OrigSelection [2]Loc

	// Which cursor index is this (for multiple cursors)
	Num int
}

// Goto puts the cursor at the given cursor's location and gives
// the current cursor its selection too
func (c *Cursor) Goto(b Cursor) {
	c.X, c.Y, c.LastVisualX = b.X, b.Y, b.LastVisualX
	c.OrigSelection, c.CurSelection = b.OrigSelection, b.CurSelection
}

// GotoLoc puts the cursor at the given cursor's location and gives
// the current cursor its selection too
func (c *Cursor) GotoLoc(l Loc) {
	c.X, c.Y = l.X, l.Y
	c.StoreVisualX()
}

// GetVisualX returns the x value of the cursor in visual spaces
func (c *Cursor) GetVisualX() int {
	if c.X <= 0 {
		c.X = 0
		return 0
	}

	bytes := c.Buf.LineBytes(c.Y)
	tabsize := int(c.Buf.Settings["tabsize"].(float64))
	if c.X > utf8.RuneCount(bytes) {
		c.X = utf8.RuneCount(bytes) - 1
	}

	return util.StringWidth(bytes, c.X, tabsize)
}

// GetCharPosInLine gets the char position of a visual x y
// coordinate (this is necessary because tabs are 1 char but
// 4 visual spaces)
func (c *Cursor) GetCharPosInLine(b []byte, visualPos int) int {
	tabsize := int(c.Buf.Settings["tabsize"].(float64))

	// Scan rune by rune until we exceed the visual width that we are
	// looking for. Then we can return the character position we have found
	i := 0     // char pos
	width := 0 // string visual width
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		b = b[size:]

		switch r {
		case '\t':
			ts := tabsize - (width % tabsize)
			width += ts
		default:
			width += runewidth.RuneWidth(r)
		}

		if width >= visualPos {
			if width == visualPos {
				i++
			}
			break
		}
		i++
	}

	return i
}

// Start moves the cursor to the start of the line it is on
func (c *Cursor) Start() {
	c.X = 0
	c.LastVisualX = c.GetVisualX()
}

// End moves the cursor to the end of the line it is on
func (c *Cursor) End() {
	c.X = utf8.RuneCount(c.Buf.LineBytes(c.Y))
	c.LastVisualX = c.GetVisualX()
}

// CopySelection copies the user's selection to either "primary"
// or "clipboard"
func (c *Cursor) CopySelection(target string) {
	if c.HasSelection() {
		if target != "primary" || c.Buf.Settings["useprimary"].(bool) {
			clipboard.WriteAll(string(c.GetSelection()), target)
		}
	}
}

// ResetSelection resets the user's selection
func (c *Cursor) ResetSelection() {
	c.CurSelection[0] = c.Buf.Start()
	c.CurSelection[1] = c.Buf.Start()
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
		c.Buf.Remove(c.CurSelection[1], c.CurSelection[0])
		c.Loc = c.CurSelection[1]
	} else if !c.HasSelection() {
		return
	} else {
		c.Buf.Remove(c.CurSelection[0], c.CurSelection[1])
		c.Loc = c.CurSelection[0]
	}
}

// Deselect closes the cursor's current selection
// Start indicates whether the cursor should be placed
// at the start or end of the selection
func (c *Cursor) Deselect(start bool) {
	if c.HasSelection() {
		if start {
			c.Loc = c.CurSelection[0]
		} else {
			c.Loc = c.CurSelection[1]
		}
		c.ResetSelection()
		c.StoreVisualX()
	}
}

// GetSelection returns the cursor's selection
func (c *Cursor) GetSelection() []byte {
	if InBounds(c.CurSelection[0], c.Buf) && InBounds(c.CurSelection[1], c.Buf) {
		if c.CurSelection[0].GreaterThan(c.CurSelection[1]) {
			return c.Buf.Substr(c.CurSelection[1], c.CurSelection[0])
		}
		return c.Buf.Substr(c.CurSelection[0], c.CurSelection[1])
	}
	return []byte{}
}

// SelectLine selects the current line
func (c *Cursor) SelectLine() {
	c.Start()
	c.SetSelectionStart(c.Loc)
	c.End()
	if len(c.Buf.lines)-1 > c.Y {
		c.SetSelectionEnd(c.Loc.Move(1, c.Buf))
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
		c.SetSelectionEnd(c.Loc.Move(1, c.Buf))
		c.SetSelectionStart(c.OrigSelection[0])
	}

	if c.Loc.LessThan(c.OrigSelection[1]) && c.Loc.GreaterThan(c.OrigSelection[0]) {
		c.CurSelection = c.OrigSelection
	}
}

// UpN moves the cursor up N lines (if possible)
func (c *Cursor) UpN(amount int) {
	proposedY := c.Y - amount
	if proposedY < 0 {
		proposedY = 0
	} else if proposedY >= len(c.Buf.lines) {
		proposedY = len(c.Buf.lines) - 1
	}

	bytes := c.Buf.LineBytes(proposedY)
	c.X = c.GetCharPosInLine(bytes, c.LastVisualX)

	if c.X > utf8.RuneCount(bytes) || (amount < 0 && proposedY == c.Y) {
		c.X = utf8.RuneCount(bytes)
	}

	c.Y = proposedY
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

// Left moves the cursor left one cell (if possible) or to
// the previous line if it is at the beginning
func (c *Cursor) Left() {
	if c.Loc == c.Buf.Start() {
		return
	}
	if c.X > 0 {
		c.X--
	} else {
		c.Up()
		c.End()
	}
	c.StoreVisualX()
}

// Right moves the cursor right one cell (if possible) or
// to the next line if it is at the end
func (c *Cursor) Right() {
	if c.Loc == c.Buf.End() {
		return
	}
	if c.X < utf8.RuneCount(c.Buf.LineBytes(c.Y)) {
		c.X++
	} else {
		c.Down()
		c.Start()
	}
	c.StoreVisualX()
}

// Relocate makes sure that the cursor is inside the bounds
// of the buffer If it isn't, it moves it to be within the
// buffer's lines
func (c *Cursor) Relocate() {
	if c.Y < 0 {
		c.Y = 0
	} else if c.Y >= len(c.Buf.lines) {
		c.Y = len(c.Buf.lines) - 1
	}

	if c.X < 0 {
		c.X = 0
	} else if c.X > utf8.RuneCount(c.Buf.LineBytes(c.Y)) {
		c.X = utf8.RuneCount(c.Buf.LineBytes(c.Y))
	}
}

func (c *Cursor) StoreVisualX() {
	c.LastVisualX = c.GetVisualX()
}
