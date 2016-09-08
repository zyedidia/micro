package main

import (
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/mitchellh/go-homedir"
	"github.com/zyedidia/tcell"
)

// The View struct stores information about a view into a buffer.
// It stores information about the cursor, and the viewport
// that the user sees the buffer from.
type View struct {
	// A pointer to the buffer's cursor for ease of access
	Cursor *Cursor

	// The topmost line, used for vertical scrolling
	Topline int
	// The leftmost column, used for horizontal scrolling
	leftCol int

	// Percentage of the terminal window that this view takes up (from 0 to 100)
	widthPercent  int
	heightPercent int

	// Specifies whether or not this view holds a help buffer
	Help bool

	// Actual with and height
	width  int
	height int

	// Where this view is located
	x, y int

	// How much to offset because of line numbers
	lineNumOffset int

	// Holds the list of gutter messages
	messages map[string][]GutterMessage

	// This is the index of this view in the views array
	Num int
	// What tab is this view stored in
	TabNum int

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

	splitNode *LeafNode
}

// NewView returns a new fullscreen view
func NewView(buf *Buffer) *View {
	screenW, screenH := screen.Size()
	return NewViewWidthHeight(buf, screenW, screenH)
}

// NewViewWidthHeight returns a new view with the specified width and height
// Note that w and h are raw column and row values
func NewViewWidthHeight(buf *Buffer, w, h int) *View {
	v := new(View)

	v.x, v.y = 0, 0

	v.width = w
	v.height = h

	v.ToggleTabbar()

	v.OpenBuffer(buf)

	v.messages = make(map[string][]GutterMessage)

	v.sline = Statusline{
		view: v,
	}

	if v.Buf.Settings["statusline"].(bool) {
		v.height--
	}

	for _, pl := range loadedPlugins {
		_, err := Call(pl+".onViewOpen", v)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
	}

	return v
}

func (v *View) ToggleStatusLine() {
	if v.Buf.Settings["statusline"].(bool) {
		v.height--
	} else {
		v.height++
	}
}

func (v *View) ToggleTabbar() {
	if len(tabs) > 1 {
		if v.y == 0 {
			// Include one line for the tab bar at the top
			v.height--
			v.y = 1
		}
	} else {
		if v.y == 1 {
			v.y = 0
			v.height++
		}
	}
}

func (v *View) paste(clip string) {
	leadingWS := GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	clip = strings.Replace(clip, "\n", "\n"+leadingWS, -1)
	v.Buf.Insert(v.Cursor.Loc, clip)
	v.Cursor.Loc = v.Cursor.Loc.Move(Count(clip), v.Buf)
	v.freshClip = false
	messenger.Message("Pasted clipboard")
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
func (v *View) CanClose() bool {
	if v.Buf.IsModified {
		char, canceled := messenger.LetterPrompt("Save changes to "+v.Buf.Name+" before closing? (y,n,esc) ", 'y', 'n')
		if !canceled {
			if char == 'y' {
				v.Save(true)
				return true
			} else if char == 'n' {
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
	screen.Clear()
	v.CloseBuffer()
	v.Buf = buf
	v.Cursor = &buf.Cursor
	v.Topline = 0
	v.leftCol = 0
	v.Cursor.ResetSelection()
	v.Relocate()
	v.Center(false)
	v.messages = make(map[string][]GutterMessage)

	v.matches = Match(v)

	// Set mouseReleased to true because we assume the mouse is not being pressed when
	// the editor is opened
	v.mouseReleased = true
	v.lastClickTime = time.Time{}
}

func (v *View) Open(filename string) {
	home, _ := homedir.Dir()
	filename = strings.Replace(filename, "~", home, 1)
	file, err := ioutil.ReadFile(filename)

	var buf *Buffer
	if err != nil {
		messenger.Message(err.Error())
		// File does not exist -- create an empty buffer with that name
		buf = NewBuffer([]byte{}, filename)
	} else {
		buf = NewBuffer(file, filename)
	}
	v.OpenBuffer(buf)
}

// CloseBuffer performs any closing functions on the buffer
func (v *View) CloseBuffer() {
	if v.Buf != nil {
		v.Buf.Serialize()
	}
}

// ReOpen reloads the current buffer
func (v *View) ReOpen() {
	if v.CanClose() {
		screen.Clear()
		v.Buf.ReOpen()
		v.Relocate()
		v.matches = Match(v)
	}
}

// HSplit opens a horizontal split with the given buffer
func (v *View) HSplit(buf *Buffer) bool {
	v.splitNode.HSplit(buf)
	tabs[v.TabNum].Resize()
	return false
}

// VSplit opens a vertical split with the given buffer
func (v *View) VSplit(buf *Buffer) bool {
	v.splitNode.VSplit(buf)
	tabs[v.TabNum].Resize()
	return false
}

// Relocate moves the view window so that the cursor is in view
// This is useful if the user has scrolled far away, and then starts typing
func (v *View) Relocate() bool {
	ret := false
	cy := v.Cursor.Y
	scrollmargin := int(v.Buf.Settings["scrollmargin"].(float64))
	if cy < v.Topline+scrollmargin && cy > scrollmargin-1 {
		v.Topline = cy - scrollmargin
		ret = true
	} else if cy < v.Topline {
		v.Topline = cy
		ret = true
	}
	if cy > v.Topline+v.height-1-scrollmargin && cy < v.Buf.NumLines-scrollmargin {
		v.Topline = cy - v.height + 1 + scrollmargin
		ret = true
	} else if cy >= v.Buf.NumLines-scrollmargin && cy > v.height {
		v.Topline = v.Buf.NumLines - v.height
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
	if x > Count(v.Buf.Line(y)) {
		x = Count(v.Buf.Line(y))
	}
	v.Cursor.X = x
	v.Cursor.Y = y
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()
}

// HandleEvent handles an event passed by the main loop
func (v *View) HandleEvent(event tcell.Event) {
	// This bool determines whether the view is relocated at the end of the function
	// By default it's true because most events should cause a relocate
	relocate := true

	v.Buf.CheckModTime()

	switch e := event.(type) {
	case *tcell.EventResize:
		// Window resized
		tabs[v.TabNum].Resize()
	case *tcell.EventKey:
		// Check first if input is a key binding, if it is we 'eat' the input and don't insert a rune
		isBinding := false
		if e.Key() != tcell.KeyRune || e.Modifiers() != 0 {
			for key, actions := range bindings {
				if e.Key() == key.keyCode {
					if e.Key() == tcell.KeyRune {
						if e.Rune() != key.r {
							continue
						}
					}
					if e.Modifiers() == key.modifiers {
						relocate = false
						isBinding = true
						for _, action := range actions {
							relocate = action(v, true) || relocate
							funcName := FuncName(action)
							if funcName != "main.(*View).ToggleMacro" && funcName != "main.(*View).PlayMacro" {
								if recordingMacro {
									curMacro = append(curMacro, action)
								}
							}
						}
					}
				}
			}
		}
		if !isBinding && e.Key() == tcell.KeyRune {
			// Insert a character
			if v.Cursor.HasSelection() {
				v.Cursor.DeleteSelection()
				v.Cursor.ResetSelection()
			}
			v.Buf.Insert(v.Cursor.Loc, string(e.Rune()))
			v.Cursor.Right()

			for _, pl := range loadedPlugins {
				_, err := Call(pl+".onRune", string(e.Rune()), v)
				if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
					TermMessage(err)
				}
			}

			if recordingMacro {
				curMacro = append(curMacro, e.Rune())
			}
		}
	case *tcell.EventPaste:
		if !PreActionCall("Paste", v) {
			break
		}

		leadingWS := GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))

		if v.Cursor.HasSelection() {
			v.Cursor.DeleteSelection()
			v.Cursor.ResetSelection()
		}
		clip := e.Text()
		clip = strings.Replace(clip, "\n", "\n"+leadingWS, -1)
		v.Buf.Insert(v.Cursor.Loc, clip)
		v.Cursor.Loc = v.Cursor.Loc.Move(Count(clip), v.Buf)
		v.freshClip = false
		messenger.Message("Pasted clipboard")

		PostActionCall("Paste", v)
	case *tcell.EventMouse:
		x, y := e.Position()
		x -= v.lineNumOffset - v.leftCol + v.x
		y += v.Topline - v.y
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

					v.Cursor.OrigSelection[0] = v.Cursor.Loc
					v.Cursor.CurSelection[0] = v.Cursor.Loc
					v.Cursor.CurSelection[1] = v.Cursor.Loc
				}
				v.mouseReleased = false
			} else if !v.mouseReleased {
				v.MoveToMouseClick(x, y)
				if v.tripleClick {
					v.Cursor.AddLineToSelection()
				} else if v.doubleClick {
					v.Cursor.AddWordToSelection()
				} else {
					v.Cursor.SetSelectionEnd(v.Cursor.Loc)
				}
			}
		case tcell.Button2:
			// Middle mouse button was clicked,
			// We should paste primary
			v.PastePrimary(true)
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
					v.Cursor.SetSelectionEnd(v.Cursor.Loc)
				}
				v.mouseReleased = true
			}
		case tcell.WheelUp:
			// Scroll up
			scrollspeed := int(v.Buf.Settings["scrollspeed"].(float64))
			v.ScrollUp(scrollspeed)
		case tcell.WheelDown:
			// Scroll down
			scrollspeed := int(v.Buf.Settings["scrollspeed"].(float64))
			v.ScrollDown(scrollspeed)
		}
	}

	if relocate {
		v.Relocate()
	}
	if v.Buf.Settings["syntax"].(bool) {
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

// Opens the given help page in a new horizontal split
func (v *View) openHelp(helpPage string) {
	if v.Help {
		helpBuffer := NewBuffer([]byte(helpPages[helpPage]), helpPage+".md")
		helpBuffer.Name = "Help"
		v.OpenBuffer(helpBuffer)
	} else {
		helpBuffer := NewBuffer([]byte(helpPages[helpPage]), helpPage+".md")
		helpBuffer.Name = "Help"
		v.HSplit(helpBuffer)
		CurView().Help = true
	}
}

func (v *View) drawCell(x, y int, ch rune, combc []rune, style tcell.Style) {
	if x >= v.x && x < v.x+v.width && y >= v.y && y < v.y+v.height {
		screen.SetContent(x, y, ch, combc, style)
	}
}

// DisplayView renders the view to the screen
func (v *View) DisplayView() {
	// The charNum we are currently displaying
	// starts at the start of the viewport
	charNum := Loc{0, v.Topline}

	// Convert the length of buffer to a string, and get the length of the string
	// We are going to have to offset by that amount
	maxLineLength := len(strconv.Itoa(v.Buf.NumLines))

	if v.Buf.Settings["ruler"] == true {
		// + 1 for the little space after the line number
		v.lineNumOffset = maxLineLength + 1
	} else {
		v.lineNumOffset = 0
	}

	// We need to add to the line offset if there are gutter messages
	var hasGutterMessages bool
	for _, v := range v.messages {
		if len(v) > 0 {
			hasGutterMessages = true
		}
	}
	if hasGutterMessages {
		v.lineNumOffset += 2
	}

	if v.x != 0 {
		// One space for the extra split divider
		v.lineNumOffset++
	}

	// These represent the current screen coordinates
	screenX, screenY := 0, 0

	highlightStyle := defStyle

	// ViewLine is the current line from the top of the viewport
	for viewLine := 0; viewLine < v.height; viewLine++ {
		screenY = v.y + viewLine
		screenX = v.x

		// This is the current line number of the buffer that we are drawing
		curLineN := viewLine + v.Topline

		if v.x != 0 {
			// Draw the split divider
			v.drawCell(screenX, screenY, '|', nil, defStyle.Reverse(true))
			screenX++
		}

		// If the buffer is smaller than the view height we have to clear all this space
		if curLineN >= v.Buf.NumLines {
			for i := screenX; i < v.x+v.width; i++ {
				v.drawCell(i, screenY, ' ', nil, defStyle)
			}

			continue
		}
		line := v.Buf.Line(curLineN)

		// If there are gutter messages we need to display the '>>' symbol here
		if hasGutterMessages {
			// msgOnLine stores whether or not there is a gutter message on this line in particular
			msgOnLine := false
			for k := range v.messages {
				for _, msg := range v.messages[k] {
					if msg.lineNum == curLineN {
						msgOnLine = true
						gutterStyle := defStyle
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
						v.drawCell(screenX, screenY, '>', nil, gutterStyle)
						screenX++
						v.drawCell(screenX, screenY, '>', nil, gutterStyle)
						screenX++
						if v.Cursor.Y == curLineN && !messenger.hasPrompt {
							messenger.Message(msg.msg)
							messenger.gutterMessage = true
						}
					}
				}
			}
			// If there is no message on this line we just display an empty offset
			if !msgOnLine {
				v.drawCell(screenX, screenY, ' ', nil, defStyle)
				screenX++
				v.drawCell(screenX, screenY, ' ', nil, defStyle)
				screenX++
				if v.Cursor.Y == curLineN && messenger.gutterMessage {
					messenger.Reset()
					messenger.gutterMessage = false
				}
			}
		}

		if v.Buf.Settings["ruler"] == true {
			// Write the line number
			lineNumStyle := defStyle
			if style, ok := colorscheme["line-number"]; ok {
				lineNumStyle = style
			}
			if style, ok := colorscheme["current-line-number"]; ok {
				if curLineN == v.Cursor.Y && tabs[curTab].curView == v.Num && !v.Cursor.HasSelection() {
					lineNumStyle = style
				}
			}

			lineNum := strconv.Itoa(curLineN + 1)

			// Write the spaces before the line number if necessary
			for i := 0; i < maxLineLength-len(lineNum); i++ {
				v.drawCell(screenX, screenY, ' ', nil, lineNumStyle)
				screenX++
			}
			// Write the actual line number
			for _, ch := range lineNum {
				v.drawCell(screenX, screenY, ch, nil, lineNumStyle)
				screenX++
			}

			// Write the extra space
			v.drawCell(screenX, screenY, ' ', nil, lineNumStyle)
			screenX++
		}

		// Now we actually draw the line
		colN := 0
		for _, ch := range line {
			lineStyle := defStyle

			if v.Buf.Settings["syntax"].(bool) {
				// Syntax highlighting is enabled
				highlightStyle = v.matches[viewLine][colN]
			}

			if v.Cursor.HasSelection() &&
				(charNum.GreaterEqual(v.Cursor.CurSelection[0]) && charNum.LessThan(v.Cursor.CurSelection[1]) ||
					charNum.LessThan(v.Cursor.CurSelection[0]) && charNum.GreaterEqual(v.Cursor.CurSelection[1])) {
				// The current character is selected
				lineStyle = defStyle.Reverse(true)

				if style, ok := colorscheme["selection"]; ok {
					lineStyle = style
				}
			} else {
				lineStyle = highlightStyle
			}

			// We need to display the background of the linestyle with the correct color if cursorline is enabled
			// and this is the current view and there is no selection on this line and the cursor is on this line
			if v.Buf.Settings["cursorline"].(bool) && tabs[curTab].curView == v.Num && !v.Cursor.HasSelection() && v.Cursor.Y == curLineN {
				if style, ok := colorscheme["cursor-line"]; ok {
					fg, _, _ := style.Decompose()
					lineStyle = lineStyle.Background(fg)
				}
			}

			if ch == '\t' {
				// If the character we are displaying is a tab, we need to do a bunch of special things

				// First the user may have configured an `indent-char` to be displayed to show that this
				// is a tab character
				lineIndentStyle := defStyle
				if style, ok := colorscheme["indent-char"]; ok {
					lineIndentStyle = style
				}
				if v.Cursor.HasSelection() &&
					(charNum.GreaterEqual(v.Cursor.CurSelection[0]) && charNum.LessThan(v.Cursor.CurSelection[1]) ||
						charNum.LessThan(v.Cursor.CurSelection[0]) && charNum.GreaterEqual(v.Cursor.CurSelection[1])) {

					lineIndentStyle = defStyle.Reverse(true)

					if style, ok := colorscheme["selection"]; ok {
						lineIndentStyle = style
					}
				}
				if v.Buf.Settings["cursorline"].(bool) && tabs[curTab].curView == v.Num && !v.Cursor.HasSelection() && v.Cursor.Y == curLineN {
					if style, ok := colorscheme["cursor-line"]; ok {
						fg, _, _ := style.Decompose()
						lineIndentStyle = lineIndentStyle.Background(fg)
					}
				}
				// Here we get the indent char
				indentChar := []rune(v.Buf.Settings["indentchar"].(string))
				if screenX-v.x-v.leftCol >= v.lineNumOffset {
					v.drawCell(screenX-v.leftCol, screenY, indentChar[0], nil, lineIndentStyle)
				}
				// Now the tab has to be displayed as a bunch of spaces
				tabSize := int(v.Buf.Settings["tabsize"].(float64))
				for i := 0; i < tabSize-1; i++ {
					screenX++
					if screenX-v.x-v.leftCol >= v.lineNumOffset {
						v.drawCell(screenX-v.leftCol, screenY, ' ', nil, lineStyle)
					}
				}
			} else if runewidth.RuneWidth(ch) > 1 {
				if screenX-v.x-v.leftCol >= v.lineNumOffset {
					v.drawCell(screenX, screenY, ch, nil, lineStyle)
				}
				for i := 0; i < runewidth.RuneWidth(ch)-1; i++ {
					screenX++
					if screenX-v.x-v.leftCol >= v.lineNumOffset {
						v.drawCell(screenX-v.leftCol, screenY, '<', nil, lineStyle)
					}
				}
			} else {
				if screenX-v.x-v.leftCol >= v.lineNumOffset {
					v.drawCell(screenX-v.leftCol, screenY, ch, nil, lineStyle)
				}
			}
			charNum = charNum.Move(1, v.Buf)
			screenX++
			colN++
		}
		// Here we are at a newline

		// The newline may be selected, in which case we should draw the selection style
		// with a space to represent it
		if v.Cursor.HasSelection() &&
			(charNum.GreaterEqual(v.Cursor.CurSelection[0]) && charNum.LessThan(v.Cursor.CurSelection[1]) ||
				charNum.LessThan(v.Cursor.CurSelection[0]) && charNum.GreaterEqual(v.Cursor.CurSelection[1])) {

			selectStyle := defStyle.Reverse(true)

			if style, ok := colorscheme["selection"]; ok {
				selectStyle = style
			}
			v.drawCell(screenX, screenY, ' ', nil, selectStyle)
			screenX++
		}

		charNum = charNum.Move(1, v.Buf)

		for i := 0; i < v.width; i++ {
			lineStyle := defStyle
			if v.Buf.Settings["cursorline"].(bool) && tabs[curTab].curView == v.Num && !v.Cursor.HasSelection() && v.Cursor.Y == curLineN {
				if style, ok := colorscheme["cursor-line"]; ok {
					fg, _, _ := style.Decompose()
					lineStyle = lineStyle.Background(fg)
				}
			}
			if screenX-v.x-v.leftCol+i >= v.lineNumOffset {
				colorcolumn := int(v.Buf.Settings["colorcolumn"].(float64))
				if colorcolumn != 0 && screenX-v.leftCol+i == colorcolumn-1 {
					if style, ok := colorscheme["color-column"]; ok {
						fg, _, _ := style.Decompose()
						lineStyle = lineStyle.Background(fg)
					}
					v.drawCell(screenX-v.leftCol+i, screenY, ' ', nil, lineStyle)
				} else {
					v.drawCell(screenX-v.leftCol+i, screenY, ' ', nil, lineStyle)
				}
			}
		}
	}
}

// DisplayCursor draws the current buffer's cursor to the screen
func (v *View) DisplayCursor() {
	// Don't draw the cursor if it is out of the viewport or if it has a selection
	if (v.Cursor.Y-v.Topline < 0 || v.Cursor.Y-v.Topline > v.height-1) || v.Cursor.HasSelection() {
		screen.HideCursor()
	} else {
		screen.ShowCursor(v.x+v.Cursor.GetVisualX()+v.lineNumOffset-v.leftCol, v.Cursor.Y-v.Topline+v.y)
	}
}

// Display renders the view, the cursor, and statusline
func (v *View) Display() {
	v.DisplayView()
	if v.Num == tabs[curTab].curView {
		v.DisplayCursor()
	}
	_, screenH := screen.Size()
	if v.Buf.Settings["statusline"].(bool) {
		v.sline.Display()
	} else if (v.y + v.height) != screenH-1 {
		for x := 0; x < v.width; x++ {
			screen.SetContent(v.x+x, v.y+v.height, '-', nil, defStyle.Reverse(true))
		}
	}
}
