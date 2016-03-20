package main

import (
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"strconv"
)

// The View struct stores information about a view into a buffer.
// It has a value for the cursor, and the window that the user sees
// the buffer from.
type View struct {
	cursor        Cursor
	topline       int
	height        int
	width         int
	lineNumOffset int

	eh *EventHandler

	buf *Buffer
	sl  Statusline

	mouseReleased bool

	s tcell.Screen
}

// NewView returns a new view with fullscreen width and height
func NewView(buf *Buffer, s tcell.Screen) *View {
	w, h := s.Size()
	return NewViewWidthHeight(buf, s, w, h-1)
}

// NewViewWidthHeight returns a new view with the specified width and height
func NewViewWidthHeight(buf *Buffer, s tcell.Screen, w, h int) *View {
	v := new(View)

	v.buf = buf
	v.s = s

	v.topline = 0
	v.height = h - 1
	v.width = w
	v.cursor = Cursor{
		x:   0,
		y:   0,
		loc: 0,
		v:   v,
	}

	v.eh = NewEventHandler(v)

	v.sl = Statusline{
		v: v,
	}

	return v
}

// ScrollUp scrolls the view up n lines (if possible)
func (v *View) ScrollUp(n int) {
	// Try to scroll by n but if it would overflow, scroll by 1
	if v.topline-n >= 0 {
		v.topline -= n
	} else if v.topline > 0 {
		v.topline--
	}
}

// ScrollDown scrolls the view down n lines (if possible)
func (v *View) ScrollDown(n int) {
	// Try to scroll by n but if it would overflow, scroll by 1
	if v.topline+n <= len(v.buf.lines)-v.height {
		v.topline += n
	} else if v.topline < len(v.buf.lines)-v.height {
		v.topline++
	}
}

// PageUp scrolls the view up a page
func (v *View) PageUp() {
	if v.topline > v.height {
		v.ScrollUp(v.height)
	} else {
		v.topline = 0
	}
}

// PageDown scrolls the view down a page
func (v *View) PageDown() {
	if len(v.buf.lines)-(v.topline+v.height) > v.height {
		v.ScrollDown(v.height)
	} else {
		v.topline = len(v.buf.lines) - v.height
	}
}

// HalfPageUp scrolls the view up half a page
func (v *View) HalfPageUp() {
	if v.topline > v.height/2 {
		v.ScrollUp(v.height / 2)
	} else {
		v.topline = 0
	}
}

// HalfPageDown scrolls the view down half a page
func (v *View) HalfPageDown() {
	if len(v.buf.lines)-(v.topline+v.height) > v.height/2 {
		v.ScrollDown(v.height / 2)
	} else {
		v.topline = len(v.buf.lines) - v.height
	}
}

// HandleEvent handles an event passed by the main loop
// It returns an int describing how the screen needs to be redrawn
// 0: Screen does not need to be redrawn
// 1: Only the cursor/statusline needs to be redrawn
// 2: Everything needs to be redrawn
func (v *View) HandleEvent(event tcell.Event) int {
	var ret int
	switch e := event.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyUp:
			v.cursor.Up()
			ret = 1
		case tcell.KeyDown:
			v.cursor.Down()
			ret = 1
		case tcell.KeyLeft:
			v.cursor.Left()
			ret = 1
		case tcell.KeyRight:
			v.cursor.Right()
			ret = 1
		case tcell.KeyEnter:
			v.eh.Insert(v.cursor.loc, "\n")
			v.cursor.Right()
			ret = 2
		case tcell.KeySpace:
			v.eh.Insert(v.cursor.loc, " ")
			v.cursor.Right()
			ret = 2
		case tcell.KeyBackspace2:
			if v.cursor.HasSelection() {
				v.cursor.DeleteSelection()
				v.cursor.ResetSelection()
				ret = 2
			} else if v.cursor.loc > 0 {
				// We have to do something a bit hacky here because we want to
				// delete the line by first moving left and then deleting backwards
				// but the undo redo would place the cursor in the wrong place
				// So instead we move left, save the position, move back, delete
				// and restore the position
				v.cursor.Left()
				cx, cy, cloc := v.cursor.x, v.cursor.y, v.cursor.loc
				v.cursor.Right()
				v.eh.Remove(v.cursor.loc-1, v.cursor.loc)
				v.cursor.x, v.cursor.y, v.cursor.loc = cx, cy, cloc
				ret = 2
			}
		case tcell.KeyTab:
			v.eh.Insert(v.cursor.loc, "\t")
			v.cursor.Right()
			ret = 2
		case tcell.KeyCtrlS:
			err := v.buf.Save()
			if err != nil {
				// Error!
			}
			// Need to redraw the status line
			ret = 1
		case tcell.KeyCtrlZ:
			v.eh.Undo()
			ret = 2
		case tcell.KeyCtrlY:
			v.eh.Redo()
			ret = 2
		case tcell.KeyCtrlC:
			if v.cursor.HasSelection() {
				if !clipboard.Unsupported {
					clipboard.WriteAll(v.cursor.GetSelection())
					ret = 2
				}
			}
		case tcell.KeyCtrlX:
			if v.cursor.HasSelection() {
				if !clipboard.Unsupported {
					clipboard.WriteAll(v.cursor.GetSelection())
					v.cursor.DeleteSelection()
					v.cursor.ResetSelection()
					ret = 2
				}
			}
		case tcell.KeyCtrlV:
			if !clipboard.Unsupported {
				if v.cursor.HasSelection() {
					v.cursor.DeleteSelection()
					v.cursor.ResetSelection()
				}
				clip, _ := clipboard.ReadAll()
				v.eh.Insert(v.cursor.loc, clip)
				// This is a bit weird... Not sure if there's a better way
				for i := 0; i < Count(clip); i++ {
					v.cursor.Right()
				}
				ret = 2
			}
		case tcell.KeyPgUp:
			v.PageUp()
			return 2
		case tcell.KeyPgDn:
			v.PageDown()
			return 2
		case tcell.KeyCtrlU:
			v.HalfPageUp()
			return 2
		case tcell.KeyCtrlD:
			v.HalfPageDown()
			return 2
		case tcell.KeyRune:
			if v.cursor.HasSelection() {
				v.cursor.DeleteSelection()
				v.cursor.ResetSelection()
			}
			v.eh.Insert(v.cursor.loc, string(e.Rune()))
			v.cursor.Right()
			ret = 2
		}
	case *tcell.EventMouse:
		x, y := e.Position()
		x -= v.lineNumOffset
		y += v.topline
		// Position always seems to be off by one
		x--
		y--

		button := e.Buttons()

		switch button {
		case tcell.Button1:
			if y-v.topline > v.height-1 {
				v.ScrollDown(1)
				y = v.height + v.topline - 1
			}
			if y > len(v.buf.lines) {
				y = len(v.buf.lines) - 2
			}
			if x < 0 {
				x = 0
			}

			x = v.cursor.GetCharPosInLine(y, x)
			if x > Count(v.buf.lines[y]) {
				x = Count(v.buf.lines[y])
			}
			d := v.cursor.Distance(x, y)
			v.cursor.loc += d
			v.cursor.x = x
			v.cursor.y = y

			if v.mouseReleased {
				v.cursor.selectionStart = v.cursor.loc
				v.cursor.selectionStartX = v.cursor.x
				v.cursor.selectionStartY = v.cursor.y
			}
			v.cursor.selectionEnd = v.cursor.loc
			v.mouseReleased = false
			return 2
		case tcell.ButtonNone:
			v.mouseReleased = true
			return 0
		case tcell.WheelUp:
			v.ScrollUp(2)
			return 2
		case tcell.WheelDown:
			v.ScrollDown(2)
			return 2
		}
	}

	cy := v.cursor.y
	if cy < v.topline {
		v.topline = cy
		ret = 2
	}
	if cy > v.topline+v.height-1 {
		v.topline = cy - v.height + 1
		ret = 2
	}

	return ret
}

// Display renders the view to the screen
func (v *View) Display() {
	var x int

	charNum := v.cursor.loc + v.cursor.Distance(0, v.topline)

	matches := Match(v.buf.text)

	// Convert the length of buffer to a string, and get the length of the string
	// We are going to have to offset by that amount
	maxLineLength := len(strconv.Itoa(len(v.buf.lines)))
	// + 1 for the little space after the line number
	v.lineNumOffset = maxLineLength + 1

	var lineStyle tcell.Style

	for lineN := 0; lineN < v.height; lineN++ {
		if lineN+v.topline >= len(v.buf.lines) {
			break
		}
		line := v.buf.lines[lineN+v.topline]

		// Write the line number
		lineNumStyle := tcell.StyleDefault
		// Write the spaces before the line number if necessary
		lineNum := strconv.Itoa(lineN + v.topline + 1)
		for i := 0; i < maxLineLength-len(lineNum); i++ {
			v.s.SetContent(x, lineN, ' ', nil, lineNumStyle)
			x++
		}
		// Write the actual line number
		for _, ch := range lineNum {
			v.s.SetContent(x, lineN, ch, nil, lineNumStyle)
			x++
		}
		// Write the extra space
		v.s.SetContent(x, lineN, ' ', nil, lineNumStyle)
		x++

		// Write the line
		tabchars := 0
		for _, ch := range line {
			st, ok := matches[charNum]
			if ok {
				lineStyle = st
			}

			if v.cursor.HasSelection() &&
				(charNum >= v.cursor.selectionStart && charNum <= v.cursor.selectionEnd ||
					charNum <= v.cursor.selectionStart && charNum >= v.cursor.selectionEnd) {
				lineStyle = lineStyle.Reverse(true)
			}

			if ch == '\t' {
				v.s.SetContent(x+tabchars, lineN, ' ', nil, lineStyle)
				for i := 0; i < tabSize-1; i++ {
					tabchars++
					v.s.SetContent(x+tabchars, lineN, ' ', nil, lineStyle)
				}
			} else {
				v.s.SetContent(x+tabchars, lineN, ch, nil, lineStyle)
			}
			charNum++
			x++
		}
		x = 0
		st, ok := matches[charNum]
		if ok {
			lineStyle = st
		}
		charNum++
	}
}
