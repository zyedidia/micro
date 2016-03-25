package main

import (
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"io/ioutil"
	"strconv"
	"strings"
)

// The View struct stores information about a view into a buffer.
// It has a value for the cursor, and the window that the user sees
// the buffer from.
type View struct {
	cursor Cursor

	// The topmost line, used for vertical scrolling
	topline int
	// The leftmost column, used for horizontal scrolling
	leftCol int

	// Percentage of the terminal window that this view takes up (from 0 to 100)
	widthPercent  int
	heightPercent int

	// Actual with and height
	width  int
	height int

	// How much to offset because of line numbers
	lineNumOffset int

	// The eventhandler for undo/redo
	eh *EventHandler

	// The buffer
	buf *Buffer
	// The statusline
	sline Statusline

	// Since tcell doesn't differentiate between a mouse release event
	// and a mouse move event with no keys pressed, we need to keep
	// track of whether or not the mouse was pressed (or not released) last event to determine
	// mouse release events
	mouseReleased bool

	// Syntax higlighting matches
	matches SyntaxMatches

	// The messenger so we can send messages to the user and get input from them
	m *Messenger
}

// NewView returns a new fullscreen view
func NewView(buf *Buffer, m *Messenger) *View {
	return NewViewWidthHeight(buf, m, 100, 100)
}

// NewViewWidthHeight returns a new view with the specified width and height percentages
// Note that w and h are percentages not actual values
func NewViewWidthHeight(buf *Buffer, m *Messenger, w, h int) *View {
	v := new(View)

	v.buf = buf
	// Messenger
	v.m = m

	v.widthPercent = w
	v.heightPercent = h
	v.Resize(screen.Size())

	v.topline = 0
	// Put the cursor at the first spot
	v.cursor = Cursor{
		x:   0,
		y:   0,
		loc: 0,
		v:   v,
	}

	v.eh = NewEventHandler(v)

	v.sline = Statusline{
		view: v,
	}

	return v
}

// Resize recalculates the actual width and height of the view from the width and height
// percentages
// This is usually called when the window is resized, or when a split has been added and
// the percentages have changed
func (v *View) Resize(w, h int) {
	// Always include 1 line for the command line at the bottom
	h--
	v.width = int(float32(w) * float32(v.widthPercent) / 100)
	// We subtract 1 for the statusline
	v.height = int(float32(h)*float32(v.heightPercent)/100) - 1
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

// CanClose returns whether or not the view can be closed
// If there are unsaved changes, the user will be asked if the view can be closed
// causing them to lose the unsaved changes
// The message is what to print after saying "You have unsaved changes. "
func (v *View) CanClose(msg string) bool {
	if v.buf.IsDirty() {
		quit, canceled := v.m.Prompt("You have unsaved changes. " + msg)
		if !canceled {
			if strings.ToLower(quit) == "yes" || strings.ToLower(quit) == "y" {
				return true
			}
		}
	} else {
		return true
	}
	return false
}

// Save the buffer to disk
func (v *View) Save() {
	// If this is an empty buffer, ask for a filename
	if v.buf.path == "" {
		filename, canceled := v.m.Prompt("Filename: ")
		if !canceled {
			v.buf.path = filename
			v.buf.name = filename
		} else {
			return
		}
	}
	err := v.buf.Save()
	if err != nil {
		v.m.Error(err.Error())
	}
}

// Copy the selection to the system clipboard
func (v *View) Copy() {
	if v.cursor.HasSelection() {
		if !clipboard.Unsupported {
			clipboard.WriteAll(v.cursor.GetSelection())
		} else {
			v.m.Error("Clipboard is not supported on your system")
		}
	}
}

// Cut the selection to the system clipboard
func (v *View) Cut() {
	if v.cursor.HasSelection() {
		if !clipboard.Unsupported {
			clipboard.WriteAll(v.cursor.GetSelection())
			v.cursor.DeleteSelection()
			v.cursor.ResetSelection()
		} else {
			v.m.Error("Clipboard is not supported on your system")
		}
	}
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (v *View) Paste() {
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
	} else {
		v.m.Error("Clipboard is not supported on your system")
	}
}

// SelectAll selects the entire buffer
func (v *View) SelectAll() {
	v.cursor.selectionEnd = 0
	v.cursor.selectionStart = v.buf.Len()
	// Put the cursor at the beginning
	v.cursor.x = 0
	v.cursor.y = 0
	v.cursor.loc = 0
}

// OpenFile opens a new file in the current view
// It makes sure that the current buffer can be closed first (unsaved changes)
func (v *View) OpenFile() {
	if v.CanClose("Continue? ") {
		filename, canceled := v.m.Prompt("File to open: ")
		if canceled {
			return
		}
		file, err := ioutil.ReadFile(filename)

		if err != nil {
			v.m.Error(err.Error())
			return
		}
		v.buf = NewBuffer(string(file), filename)
	}
}

// Relocate moves the view window so that the cursor is in view
// This is useful if the user has scrolled far away, and then starts typing
func (v *View) Relocate() {
	cy := v.cursor.y
	if cy < v.topline {
		v.topline = cy
	}
	if cy > v.topline+v.height-1 {
		v.topline = cy - v.height + 1
	}
}

// MoveToMouseClick moves the cursor to location x, y assuming x, y were given
// by a mouse click
func (v *View) MoveToMouseClick(x, y int) {
	if y-v.topline > v.height-1 {
		v.ScrollDown(1)
		y = v.height + v.topline - 1
	}
	if y >= len(v.buf.lines) {
		y = len(v.buf.lines) - 1
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
}

// HandleEvent handles an event passed by the main loop
func (v *View) HandleEvent(event tcell.Event) {
	// This bool determines whether the view is relocated at the end of the function
	// By default it's true because most events should cause a relocate
	relocate := true
	switch e := event.(type) {
	case *tcell.EventResize:
		// Window resized
		v.Resize(e.Size())
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyUp:
			// Cursor up
			v.cursor.Up()
		case tcell.KeyDown:
			// Cursor down
			v.cursor.Down()
		case tcell.KeyLeft:
			// Cursor left
			v.cursor.Left()
		case tcell.KeyRight:
			// Cursor right
			v.cursor.Right()
		case tcell.KeyEnter:
			// Insert a newline
			v.eh.Insert(v.cursor.loc, "\n")
			v.cursor.Right()
		case tcell.KeySpace:
			// Insert a space
			v.eh.Insert(v.cursor.loc, " ")
			v.cursor.Right()
		case tcell.KeyBackspace2:
			// Delete a character
			if v.cursor.HasSelection() {
				v.cursor.DeleteSelection()
				v.cursor.ResetSelection()
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
			}
		case tcell.KeyTab:
			// Insert a tab
			v.eh.Insert(v.cursor.loc, "\t")
			v.cursor.Right()
		case tcell.KeyCtrlS:
			v.Save()
		case tcell.KeyCtrlZ:
			v.eh.Undo()
		case tcell.KeyCtrlY:
			v.eh.Redo()
		case tcell.KeyCtrlC:
			v.Copy()
		case tcell.KeyCtrlX:
			v.Cut()
		case tcell.KeyCtrlV:
			v.Paste()
		case tcell.KeyCtrlA:
			v.SelectAll()
		case tcell.KeyCtrlO:
			v.OpenFile()
		case tcell.KeyPgUp:
			v.PageUp()
		case tcell.KeyPgDn:
			v.PageDown()
		case tcell.KeyCtrlU:
			v.HalfPageUp()
		case tcell.KeyCtrlD:
			v.HalfPageDown()
		case tcell.KeyRune:
			// Insert a character
			if v.cursor.HasSelection() {
				v.cursor.DeleteSelection()
				v.cursor.ResetSelection()
			}
			v.eh.Insert(v.cursor.loc, string(e.Rune()))
			v.cursor.Right()
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
			// Left click
			v.MoveToMouseClick(x, y)

			if v.mouseReleased {
				v.cursor.selectionStart = v.cursor.loc
				v.cursor.selectionStartX = v.cursor.x
				v.cursor.selectionStartY = v.cursor.y
			}
			v.cursor.selectionEnd = v.cursor.loc
			v.mouseReleased = false
		case tcell.ButtonNone:
			// Mouse event with no click
			if !v.mouseReleased {
				// Mouse was just released

				// Relocating here isn't really necessary because the cursor will
				// be in the right place from the last mouse event
				// However, if we are running in a terminal that doesn't support mouse motion
				// events, this still allows the user to make selections, except only after they
				// release the mouse
				v.MoveToMouseClick(x, y)
				v.cursor.selectionEnd = v.cursor.loc
				v.mouseReleased = true
			}
			// We don't want to relocate because otherwise the view will be relocated
			// everytime the user moves the cursor
			relocate = false
		case tcell.WheelUp:
			// Scroll up two lines
			v.ScrollUp(2)
			// We don't want to relocate if the user is scrolling
			relocate = false
		case tcell.WheelDown:
			// Scroll down two lines
			v.ScrollDown(2)
			// We don't want to relocate if the user is scrolling
			relocate = false
		}
	}

	if relocate {
		v.Relocate()
	}
}

// DisplayView renders the view to the screen
func (v *View) DisplayView() {
	// The character number of the character in the top left of the screen
	charNum := v.cursor.loc + v.cursor.Distance(0, v.topline)

	// Convert the length of buffer to a string, and get the length of the string
	// We are going to have to offset by that amount
	maxLineLength := len(strconv.Itoa(len(v.buf.lines)))
	// + 1 for the little space after the line number
	v.lineNumOffset = maxLineLength + 1

	var highlightStyle tcell.Style

	for lineN := 0; lineN < v.height; lineN++ {
		var x int
		// If the buffer is smaller than the view height
		// and we went too far, break
		if lineN+v.topline >= len(v.buf.lines) {
			break
		}
		line := v.buf.lines[lineN+v.topline]

		// Write the line number
		lineNumStyle := tcell.StyleDefault
		if _, ok := colorscheme["line-number"]; ok {
			lineNumStyle = colorscheme["line-number"]
		}
		// Write the spaces before the line number if necessary
		lineNum := strconv.Itoa(lineN + v.topline + 1)
		for i := 0; i < maxLineLength-len(lineNum); i++ {
			screen.SetContent(x, lineN, ' ', nil, lineNumStyle)
			x++
		}
		// Write the actual line number
		for _, ch := range lineNum {
			screen.SetContent(x, lineN, ch, nil, lineNumStyle)
			x++
		}
		// Write the extra space
		screen.SetContent(x, lineN, ' ', nil, lineNumStyle)
		x++

		// Write the line
		tabchars := 0
		for _, ch := range line {
			var lineStyle tcell.Style
			// Does the current character need to be syntax highlighted?
			st, ok := v.matches[charNum]
			if ok {
				highlightStyle = st
			} else {
				highlightStyle = tcell.StyleDefault
			}

			if v.cursor.HasSelection() &&
				(charNum >= v.cursor.selectionStart && charNum <= v.cursor.selectionEnd ||
					charNum <= v.cursor.selectionStart && charNum >= v.cursor.selectionEnd) {

				lineStyle = tcell.StyleDefault.Reverse(true)

				if _, ok := colorscheme["selection"]; ok {
					lineStyle = colorscheme["selection"]
				}
			} else {
				lineStyle = highlightStyle
			}

			if ch == '\t' {
				screen.SetContent(x+tabchars, lineN, ' ', nil, lineStyle)
				for i := 0; i < tabSize-1; i++ {
					tabchars++
					screen.SetContent(x+tabchars, lineN, ' ', nil, lineStyle)
				}
			} else {
				screen.SetContent(x+tabchars, lineN, ch, nil, lineStyle)
			}
			charNum++
			x++
		}
		// Here we are at a newline

		// The newline may be selected, in which case we should draw the selection style
		// with a space to represent it
		if v.cursor.HasSelection() &&
			(charNum >= v.cursor.selectionStart && charNum <= v.cursor.selectionEnd ||
				charNum <= v.cursor.selectionStart && charNum >= v.cursor.selectionEnd) {

			selectStyle := tcell.StyleDefault.Reverse(true)

			if _, ok := colorscheme["selection"]; ok {
				selectStyle = colorscheme["selection"]
			}
			screen.SetContent(x+tabchars, lineN, ' ', nil, selectStyle)
		}

		x = 0
		charNum++
	}
}

// Display renders the view, the cursor, and statusline
func (v *View) Display() {
	v.DisplayView()
	v.cursor.Display()
	v.sline.Display()
}
