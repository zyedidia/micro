package main

import (
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/zyedidia/tcell"
)

// The View struct stores information about a view into a buffer.
// It has a stores information about the cursor, and the viewport
// that the user sees the buffer from.
type View struct {
	cursor []Cursor

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

	messages []GutterMessage

	// The buffer
	buf *Buffer
	// The statusline
	sline Statusline

	// Since tcell doesn't differentiate between a mouse release event
	// and a mouse move event with no keys pressed, we need to keep
	// track of whether or not the mouse was pressed (or not released) last event to determine
	// mouse release events
	mouseReleased bool

	// This stores when the last click was
	// This is useful for detecting double and triple clicks
	lastClickTime time.Time

	// lastCutTime stores when the last ctrl+k was issued.
	// It is used for clearing the clipboard to replace it with fresh cut lines.
	lastCutTime time.Time

	// freshClip returns true if the clipboard has never been pasted.
	freshClip bool

	// Was the last mouse event actually a double click?
	// Useful for detecting triple clicks -- if a double click is detected
	// but the last mouse event was actually a double click, it's a triple click
	doubleClick bool
	// Same here, just to keep track for mouse move events
	tripleClick bool

	// Syntax highlighting matches
	matches SyntaxMatches
	// The matches from the last frame
	lastMatches SyntaxMatches
}

// NewView returns a new fullscreen view
func NewView(buf *Buffer) *View {
	return NewViewWidthHeight(buf, 100, 100)
}

// NewViewWidthHeight returns a new view with the specified width and height percentages
// Note that w and h are percentages not actual values
func NewViewWidthHeight(buf *Buffer, w, h int) *View {
	v := new(View)

	// Initialize first item in cursor array. This is our default cursor
	v.cursor = []Cursor{
		Cursor{
			x: 0,
			y: 0,
			v: v,
		},
	}

	v.widthPercent = w
	v.heightPercent = h
	v.Resize(screen.Size())

	v.OpenBuffer(buf)

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

// CanClose returns whether or not the view can be closed
// If there are unsaved changes, the user will be asked if the view can be closed
// causing them to lose the unsaved changes
// The message is what to print after saying "You have unsaved changes. "
func (v *View) CanClose(msg string) bool {
	if v.buf.IsDirty() {
		quit, canceled := messenger.Prompt("You have unsaved changes. " + msg)
		if !canceled {
			if strings.ToLower(quit) == "yes" || strings.ToLower(quit) == "y" {
				return true
			} else if strings.ToLower(quit) == "save" || strings.ToLower(quit) == "s" {
				v.Save()
				return true
			}
		}
	} else {
		return true
	}
	return false
}

// OpenBuffer opens a new buffer in this view.
// This resets the topline, event handler and cursor.
func (v *View) OpenBuffer(buf *Buffer) {
	v.buf = buf
	v.topline = 0
	v.leftCol = 0

	// Put the default cursor at the first spot
	v.cursor[0] = Cursor{
		x: 0,
		y: 0,
		v: v,
	}
	v.cursor[0].ResetSelection()

	v.eh = NewEventHandler(v)
	v.matches = Match(v)

	// Set mouseReleased to true because we assume the mouse is not being pressed when
	// the editor is opened
	v.mouseReleased = true
	v.lastClickTime = time.Time{}
}

// Close and Re-open the current file.
func (v *View) reOpen() {
	if v.CanClose("Continue? (yes, no, save) ") {
		file, err := ioutil.ReadFile(v.buf.path)
		filename := v.buf.name

		if err != nil {
			messenger.Error(err.Error())
			return
		}
		buf := NewBuffer(string(file), filename)
		v.buf = buf
		v.matches = Match(v)
		v.cursor[0].Relocate()
		v.Relocate()
	}
}

// Relocate moves the view window so that the cursor is in view
// This is useful if the user has scrolled far away, and then starts typing.
// In multi-cursor mode, we always use cursor[0], since it is the "leader".
func (v *View) Relocate() bool {
	ret := false
	cy := v.cursor[0].y
	if cy < v.topline {
		v.topline = cy
		ret = true
	}
	if cy > v.topline+v.height-1 {
		v.topline = cy - v.height + 1
		ret = true
	}

	cx := v.cursor[0].GetVisualX()
	if cx < v.leftCol {
		v.leftCol = cx
		ret = true
	}
	if cx+v.lineNumOffset+1 > v.leftCol+v.width {
		v.leftCol = cx - v.width + v.lineNumOffset + 1
		ret = true
	}
	return ret
}

// MoveToMouseClick moves the cursor to location x, y assuming x, y were given
// by a mouse click. It will also cancel
func (v *View) MoveToMouseClick(x, y int) {
	if y-v.topline > v.height-1 {
		v.ScrollDown(1)
		y = v.height + v.topline - 1
	}
	if y >= len(v.buf.lines) {
		y = len(v.buf.lines) - 1
	}
	if y < 0 {
		y = 0
	}
	if x < 0 {
		x = 0
	}

	v.exitMultiCursorMode()
	x = v.cursor[0].GetCharPosInLine(y, x)
	if x > Count(v.buf.lines[y]) {
		x = Count(v.buf.lines[y])
	}
	v.cursor[0].x = x
	v.cursor[0].y = y
	v.cursor[0].lastVisualX = v.cursor[0].GetVisualX()
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
		if e.Key() == tcell.KeyRune {
			for i := range v.cursor {
				if v.cursor[i].HasSelection() {
					v.cursor[i].DeleteSelection()
					v.cursor[i].ResetSelection()
				}
				v.eh.Insert(v.cursor[i].Loc(), string(e.Rune()))
				v.cursor[i].Right()
			}
		} else {
			for key, action := range bindings {
				if e.Key() == key {
					relocate = action(v)
				}
			}
		}
	case *tcell.EventMouse:
		x, y := e.Position()
		x -= v.lineNumOffset - v.leftCol
		y += v.topline

		button := e.Buttons()

		switch button {
		case tcell.Button1:
			// Clicking will exit multi-cursor mode
			v.exitMultiCursorMode()

			// Left click
			origX, origY := v.cursor[0].x, v.cursor[0].y
			if v.mouseReleased && !e.HasMotion() {
				v.MoveToMouseClick(x, y)
				if (time.Since(v.lastClickTime)/time.Millisecond < doubleClickThreshold) &&
					(origX == v.cursor[0].x && origY == v.cursor[0].y) {
					if v.doubleClick {
						// Triple click
						v.lastClickTime = time.Now()

						v.tripleClick = true
						v.doubleClick = false

						v.cursor[0].SelectLine()
					} else {
						// Double click
						v.lastClickTime = time.Now()

						v.doubleClick = true
						v.tripleClick = false

						v.cursor[0].SelectWord()
					}
				} else {
					v.doubleClick = false
					v.tripleClick = false
					v.lastClickTime = time.Now()

					loc := v.cursor[0].Loc()
					v.cursor[0].curSelection[0] = loc
					v.cursor[0].curSelection[1] = loc
				}
				v.mouseReleased = false
			} else if !v.mouseReleased {
				v.MoveToMouseClick(x, y)
				if v.tripleClick {
					v.cursor[0].AddLineToSelection()
				} else if v.doubleClick {
					v.cursor[0].AddWordToSelection()
				} else {
					v.cursor[0].curSelection[1] = v.cursor[0].Loc()
				}
			}
		case tcell.ButtonNone:
			// Mouse event with no click
			if !v.mouseReleased {
				// Mouse was just released

				// Relocating here isn't really necessary because the cursor will
				// be in the right place from the last mouse event
				// However, if we are running in a terminal that doesn't support mouse motion
				// events, this still allows the user to make selections, except only after they
				// release the mouse

				if !v.doubleClick && !v.tripleClick {
					v.exitMultiCursorMode()
					v.MoveToMouseClick(x, y)
					v.cursor[0].curSelection[1] = v.cursor[0].Loc()
				}
				v.mouseReleased = true
			}
			// We don't want to relocate because otherwise the view will be relocated
			// every time the user moves the cursor
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
	if settings.Syntax {
		v.matches = Match(v)
	}
}

// GutterMessage creates a message in this view's gutter
func (v *View) GutterMessage(lineN int, msg string, kind int) {
	gutterMsg := GutterMessage{
		lineNum: lineN,
		msg:     msg,
		kind:    kind,
	}
	for _, gmsg := range v.messages {
		if gmsg.lineNum == lineN {
			return
		}
	}
	v.messages = append(v.messages, gutterMsg)
}

// DisplayView renders the view to the screen
func (v *View) DisplayView() {
	// The character number of the character in the top left of the screen
	charNum := ToCharPos(0, v.topline, v.buf)

	// Convert the length of buffer to a string, and get the length of the string
	// We are going to have to offset by that amount
	maxLineLength := len(strconv.Itoa(len(v.buf.lines)))
	// + 1 for the little space after the line number
	if settings.Ruler == true {
		v.lineNumOffset = maxLineLength + 1
	} else {
		v.lineNumOffset = 0
	}
	var highlightStyle tcell.Style

	if len(v.messages) > 0 {
		v.lineNumOffset += 2
	}

	for lineN := 0; lineN < v.height; lineN++ {
		var x int
		// If the buffer is smaller than the view height
		// and we went too far, break
		if lineN+v.topline >= len(v.buf.lines) {
			break
		}
		line := v.buf.lines[lineN+v.topline]

		if len(v.messages) > 0 {
			msgOnLine := false
			for _, msg := range v.messages {
				if msg.lineNum == lineN+v.topline {
					msgOnLine = true
					gutterStyle := tcell.StyleDefault
					switch msg.kind {
					case GutterInfo:
						if style, ok := colorscheme["gutter-info"]; ok {
							gutterStyle = style
						}
					case GutterWarning:
						if style, ok := colorscheme["gutter-warning"]; ok {
							gutterStyle = style
						}
					case GutterError:
						if style, ok := colorscheme["gutter-error"]; ok {
							gutterStyle = style
						}
					}
					screen.SetContent(x, lineN, '>', nil, gutterStyle)
					x++
					screen.SetContent(x, lineN, '>', nil, gutterStyle)
					x++
					if v.cursor[0].y == lineN {
						messenger.Message(msg.msg)
						messenger.gutterMessage = true
					}
				}
			}
			if !msgOnLine {
				screen.SetContent(x, lineN, ' ', nil, tcell.StyleDefault)
				x++
				screen.SetContent(x, lineN, ' ', nil, tcell.StyleDefault)
				x++
				if v.cursor[0].y == lineN && messenger.gutterMessage {
					messenger.Reset()
					messenger.gutterMessage = false
				}
			}
		}

		// Write the line number
		lineNumStyle := defStyle
		if style, ok := colorscheme["line-number"]; ok {
			lineNumStyle = style
		}
		// Write the spaces before the line number if necessary
		var lineNum string
		if settings.Ruler == true {
			lineNum = strconv.Itoa(lineN + v.topline + 1)
			for i := 0; i < maxLineLength-len(lineNum); i++ {
				screen.SetContent(x, lineN, ' ', nil, lineNumStyle)
				x++
			}
			// Write the actual line number
			for _, ch := range lineNum {
				screen.SetContent(x, lineN, ch, nil, lineNumStyle)
				x++
			}

			if settings.Ruler == true {
				// Write the extra space
				screen.SetContent(x, lineN, ' ', nil, lineNumStyle)
				x++
			}
		}
		// Write the line
		tabchars := 0
		for colN, ch := range line {
			var lineStyle tcell.Style

			if settings.Syntax {
				// Syntax highlighting is enabled
				highlightStyle = v.matches[lineN][colN]
			}

			// I think I need to do something like, "if anything in the array is true" here
			if selectionExistsInLine(v, charNum) {

				lineStyle = tcell.StyleDefault.Reverse(true)

				if style, ok := colorscheme["selection"]; ok {
					lineStyle = style
				}
			} else {
				lineStyle = highlightStyle
			}

			if ch == '\t' {
				screen.SetContent(x+tabchars, lineN, ' ', nil, lineStyle)
				tabSize := settings.TabSize
				for i := 0; i < tabSize-1; i++ {
					tabchars++
					if x-v.leftCol+tabchars >= v.lineNumOffset {
						screen.SetContent(x-v.leftCol+tabchars, lineN, ' ', nil, lineStyle)
					}
				}
			} else {
				if x-v.leftCol+tabchars >= v.lineNumOffset {
					screen.SetContent(x-v.leftCol+tabchars, lineN, ch, nil, lineStyle)
				}
			}
			charNum++
			x++
		}
		// Here we are at a newline

		// The newline may be selected, in which case we should draw the selection style
		// with a space to represent it
		if selectionExistsInLine(v, charNum) {

			selectStyle := defStyle.Reverse(true)

			if style, ok := colorscheme["selection"]; ok {
				selectStyle = style
			}
			screen.SetContent(x-v.leftCol+tabchars, lineN, ' ', nil, selectStyle)
		}

		charNum++
	}
}

// Display renders the view, the cursor, and statusline
func (v *View) Display() {
	v.DisplayView()
	v.cursor[0].Display()
	v.sline.Display()
}

// exitMultiCursorMode removes all additonal cursors.
func (v *View) exitMultiCursorMode() {
	if len(v.cursor) > 1 {
		v.cursor = v.cursor[:1]
	}
}

func selectionExistsInLine(v *View, charNum int) bool {
	for i := range v.cursor {
		if v.cursor[i].HasSelection() &&
			(charNum >= v.cursor[i].curSelection[0] && charNum < v.cursor[i].curSelection[1] ||
				charNum < v.cursor[i].curSelection[0] && charNum >= v.cursor[i].curSelection[1]) {
			return true
		}
	}
	return false
}
