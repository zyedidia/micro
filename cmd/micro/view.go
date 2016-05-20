package main

import (
	"io/ioutil"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/zyedidia/tcell"
)

// The View struct stores information about a view into a buffer.
// It stores information about the cursor, and the viewport
// that the user sees the buffer from.
type View struct {
	Cursor Cursor

	// The topmost line, used for vertical scrolling
	Topline int
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

	// Holds the list of gutter messages
	messages map[string][]GutterMessage

	// The buffer
	Buf *Buffer
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

	v.widthPercent = w
	v.heightPercent = h
	v.Resize(screen.Size())

	v.OpenBuffer(buf)

	v.eh = NewEventHandler(v)

	v.messages = make(map[string][]GutterMessage)

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
	v.height = int(float32(h) * float32(v.heightPercent) / 100)
	if settings["statusline"].(bool) {
		// Make room for the status line if it is enabled
		v.height--
	}
}

// ScrollUp scrolls the view up n lines (if possible)
func (v *View) ScrollUp(n int) {
	// Try to scroll by n but if it would overflow, scroll by 1
	if v.Topline-n >= 0 {
		v.Topline -= n
	} else if v.Topline > 0 {
		v.Topline--
	}
}

// ScrollDown scrolls the view down n lines (if possible)
func (v *View) ScrollDown(n int) {
	// Try to scroll by n but if it would overflow, scroll by 1
	if v.Topline+n <= v.Buf.NumLines-v.height {
		v.Topline += n
	} else if v.Topline < v.Buf.NumLines-v.height {
		v.Topline++
	}
}

// CanClose returns whether or not the view can be closed
// If there are unsaved changes, the user will be asked if the view can be closed
// causing them to lose the unsaved changes
// The message is what to print after saying "You have unsaved changes. "
func (v *View) CanClose(msg string) bool {
	if v.Buf.IsModified {
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
	v.Buf = buf
	v.Topline = 0
	v.leftCol = 0
	// Put the cursor at the first spot
	v.Cursor = Cursor{
		x: 0,
		y: 0,
		v: v,
	}
	v.Cursor.ResetSelection()
	v.messages = make(map[string][]GutterMessage)

	v.eh = NewEventHandler(v)
	v.matches = Match(v)

	// Set mouseReleased to true because we assume the mouse is not being pressed when
	// the editor is opened
	v.mouseReleased = true
	v.lastClickTime = time.Time{}
}

// ReOpen reloads the current buffer
func (v *View) ReOpen() {
	if v.CanClose("Continue? (yes, no, save) ") {
		file, err := ioutil.ReadFile(v.Buf.Path)
		filename := v.Buf.Name

		if err != nil {
			messenger.Error(err.Error())
			return
		}
		buf := NewBuffer(string(file), filename)
		v.Buf = buf
		v.matches = Match(v)
		v.Cursor.Relocate()
		v.Relocate(0)
	}
}

// Relocate moves the view window so that the cursor is in view
// This is useful if the user has scrolled far away, and then starts typing
func (v *View) Relocate(x int) bool {
	ret := false
	cy := v.Cursor.y
	if cy < v.Topline+x && cy > x-1 {
		v.Topline = cy - x
		ret = true
	} else if cy < v.Topline {
		v.Topline = cy
		ret = true
	}
	if cy > v.Topline+v.height-1-x {
		v.Topline = cy - v.height + 1 + x
		ret = true
	}

	cx := v.Cursor.GetVisualX()
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
// by a mouse click
func (v *View) MoveToMouseClick(x, y int) {
	if y-v.Topline > v.height-1 {
		v.ScrollDown(1)
		y = v.height + v.Topline - 1
	}
	if y >= v.Buf.NumLines {
		y = v.Buf.NumLines - 1
	}
	if y < 0 {
		y = 0
	}
	if x < 0 {
		x = 0
	}

	x = v.Cursor.GetCharPosInLine(y, x)
	if x > Count(v.Buf.Lines[y]) {
		x = Count(v.Buf.Lines[y])
	}
	v.Cursor.x = x
	v.Cursor.y = y
	v.Cursor.lastVisualX = v.Cursor.GetVisualX()
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
			// Insert a character
			if v.Cursor.HasSelection() {
				v.Cursor.DeleteSelection()
				v.Cursor.ResetSelection()
			}
			v.eh.Insert(v.Cursor.Loc(), string(e.Rune()))
			v.Cursor.Right()
		} else {
			for key, action := range bindings {
				if e.Key() == key.keyCode {
					if e.Modifiers() == key.modifiers {
						relocate = action(v)
						for _, pl := range loadedPlugins {
							funcName := strings.Split(runtime.FuncForPC(reflect.ValueOf(action).Pointer()).Name(), ".")
							err := Call(pl + "_on" + funcName[len(funcName)-1])
							if err != nil {
								TermMessage(err)
							}
						}
					}
				}
			}
		}
	case *tcell.EventPaste:
		if v.Cursor.HasSelection() {
			v.Cursor.DeleteSelection()
			v.Cursor.ResetSelection()
		}
		clip := e.Text()
		v.eh.Insert(v.Cursor.Loc(), clip)
		v.Cursor.SetLoc(v.Cursor.Loc() + Count(clip))
		v.freshClip = false
	case *tcell.EventMouse:
		x, y := e.Position()
		x -= v.lineNumOffset - v.leftCol
		y += v.Topline
		// Don't relocate for mouse events
		relocate = false

		button := e.Buttons()

		switch button {
		case tcell.Button1:
			// Left click
			if v.mouseReleased {
				v.MoveToMouseClick(x, y)
				if time.Since(v.lastClickTime)/time.Millisecond < doubleClickThreshold {
					if v.doubleClick {
						// Triple click
						v.lastClickTime = time.Now()

						v.tripleClick = true
						v.doubleClick = false

						v.Cursor.SelectLine()
					} else {
						// Double click
						v.lastClickTime = time.Now()

						v.doubleClick = true
						v.tripleClick = false

						v.Cursor.SelectWord()
					}
				} else {
					v.doubleClick = false
					v.tripleClick = false
					v.lastClickTime = time.Now()

					loc := v.Cursor.Loc()
					v.Cursor.origSelection[0] = loc
					v.Cursor.curSelection[0] = loc
					v.Cursor.curSelection[1] = loc
				}
				v.mouseReleased = false
			} else if !v.mouseReleased {
				v.MoveToMouseClick(x, y)
				if v.tripleClick {
					v.Cursor.AddLineToSelection()
				} else if v.doubleClick {
					v.Cursor.AddWordToSelection()
				} else {
					v.Cursor.curSelection[1] = v.Cursor.Loc()
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
					v.MoveToMouseClick(x, y)
					v.Cursor.curSelection[1] = v.Cursor.Loc()
				}
				v.mouseReleased = true
			}
		case tcell.WheelUp:
			// Scroll up
			scrollspeed := int(settings["scrollspeed"].(float64))
			v.ScrollUp(scrollspeed)
		case tcell.WheelDown:
			// Scroll down
			scrollspeed := int(settings["scrollspeed"].(float64))
			v.ScrollDown(scrollspeed)
		}
	}

	if relocate {
		scrollmargin := int(settings["scrollmargin"].(float64))
		v.Relocate(scrollmargin)
	}
	if settings["syntax"].(bool) {
		v.matches = Match(v)
	}
}

// GutterMessage creates a message in this view's gutter
func (v *View) GutterMessage(section string, lineN int, msg string, kind int) {
	lineN--
	gutterMsg := GutterMessage{
		lineNum: lineN,
		msg:     msg,
		kind:    kind,
	}
	for _, v := range v.messages {
		for _, gmsg := range v {
			if gmsg.lineNum == lineN {
				return
			}
		}
	}
	messages := v.messages[section]
	v.messages[section] = append(messages, gutterMsg)
}

// ClearGutterMessages clears all gutter messages from a given section
func (v *View) ClearGutterMessages(section string) {
	v.messages[section] = []GutterMessage{}
}

// ClearAllGutterMessages clears all the gutter messages
func (v *View) ClearAllGutterMessages() {
	for k := range v.messages {
		v.messages[k] = []GutterMessage{}
	}
}

// DisplayView renders the view to the screen
func (v *View) DisplayView() {
	// The character number of the character in the top left of the screen
	charNum := ToCharPos(0, v.Topline, v.Buf)

	// Convert the length of buffer to a string, and get the length of the string
	// We are going to have to offset by that amount
	maxLineLength := len(strconv.Itoa(v.Buf.NumLines))
	// + 1 for the little space after the line number
	if settings["ruler"] == true {
		v.lineNumOffset = maxLineLength + 1
	} else {
		v.lineNumOffset = 0
	}
	var highlightStyle tcell.Style

	var hasGutterMessages bool
	for _, v := range v.messages {
		if len(v) > 0 {
			hasGutterMessages = true
		}
	}
	if hasGutterMessages {
		v.lineNumOffset += 2
	}

	for lineN := 0; lineN < v.height; lineN++ {
		var x int
		// If the buffer is smaller than the view height
		// and we went too far, break
		if lineN+v.Topline >= v.Buf.NumLines {
			break
		}
		line := v.Buf.Lines[lineN+v.Topline]

		if hasGutterMessages {
			msgOnLine := false
			for k := range v.messages {
				for _, msg := range v.messages[k] {
					if msg.lineNum == lineN+v.Topline {
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
						if v.Cursor.y == lineN+v.Topline {
							messenger.Message(msg.msg)
							messenger.gutterMessage = true
						}
					}
				}
			}
			if !msgOnLine {
				screen.SetContent(x, lineN, ' ', nil, tcell.StyleDefault)
				x++
				screen.SetContent(x, lineN, ' ', nil, tcell.StyleDefault)
				x++
				if v.Cursor.y == lineN+v.Topline && messenger.gutterMessage {
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
		if settings["ruler"] == true {
			lineNum = strconv.Itoa(lineN + v.Topline + 1)
			for i := 0; i < maxLineLength-len(lineNum); i++ {
				screen.SetContent(x, lineN, ' ', nil, lineNumStyle)
				x++
			}
			// Write the actual line number
			for _, ch := range lineNum {
				screen.SetContent(x, lineN, ch, nil, lineNumStyle)
				x++
			}

			if settings["ruler"] == true {
				// Write the extra space
				screen.SetContent(x, lineN, ' ', nil, lineNumStyle)
				x++
			}
		}
		// Write the line
		tabchars := 0
		for colN, ch := range line {
			var lineStyle tcell.Style

			if settings["syntax"].(bool) {
				// Syntax highlighting is enabled
				highlightStyle = v.matches[lineN][colN]
			}

			if v.Cursor.HasSelection() &&
				(charNum >= v.Cursor.curSelection[0] && charNum < v.Cursor.curSelection[1] ||
					charNum < v.Cursor.curSelection[0] && charNum >= v.Cursor.curSelection[1]) {

				lineStyle = tcell.StyleDefault.Reverse(true)

				if style, ok := colorscheme["selection"]; ok {
					lineStyle = style
				}
			} else {
				lineStyle = highlightStyle
			}

			if ch == '\t' {
				lineIndentStyle := defStyle
				if style, ok := colorscheme["indent-char"]; ok {
					lineIndentStyle = style
				}
				if v.Cursor.HasSelection() &&
					(charNum >= v.Cursor.curSelection[0] && charNum < v.Cursor.curSelection[1] ||
						charNum < v.Cursor.curSelection[0] && charNum >= v.Cursor.curSelection[1]) {

					lineIndentStyle = tcell.StyleDefault.Reverse(true)

					if style, ok := colorscheme["selection"]; ok {
						lineIndentStyle = style
					}
				}
				indentChar := []rune(settings["indentchar"].(string))
				screen.SetContent(x-v.leftCol+tabchars, lineN, indentChar[0], nil, lineIndentStyle)
				tabSize := int(settings["tabsize"].(float64))
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
		if v.Cursor.HasSelection() &&
			(charNum >= v.Cursor.curSelection[0] && charNum < v.Cursor.curSelection[1] ||
				charNum < v.Cursor.curSelection[0] && charNum >= v.Cursor.curSelection[1]) {

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
	v.Cursor.Display()
	if settings["statusline"].(bool) {
		v.sline.Display()
	}
}
