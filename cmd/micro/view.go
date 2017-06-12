package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/zyedidia/tcell"
)

type ViewType struct {
	kind     int
	readonly bool // The file cannot be edited
	scratch  bool // The file cannot be saved
}

var (
	vtDefault = ViewType{0, false, false}
	vtHelp    = ViewType{1, true, true}
	vtLog     = ViewType{2, true, true}
	vtScratch = ViewType{3, false, true}
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

	// Specifies whether or not this view holds a help buffer
	Type ViewType

	// Actual width and height
	Width  int
	Height int

	LockWidth  bool
	LockHeight bool

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

	cellview *CellView

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

	v.Width = w
	v.Height = h
	v.cellview = new(CellView)

	v.ToggleTabbar()

	v.OpenBuffer(buf)

	v.messages = make(map[string][]GutterMessage)

	v.sline = Statusline{
		view: v,
	}

	if v.Buf.Settings["statusline"].(bool) {
		v.Height--
	}

	for pl := range loadedPlugins {
		_, err := Call(pl+".onViewOpen", v)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
	}

	return v
}

// ToggleStatusLine creates an extra row for the statusline if necessary
func (v *View) ToggleStatusLine() {
	if v.Buf.Settings["statusline"].(bool) {
		v.Height--
	} else {
		v.Height++
	}
}

// ToggleTabbar creates an extra row for the tabbar if necessary
func (v *View) ToggleTabbar() {
	if len(tabs) > 1 {
		if v.y == 0 {
			// Include one line for the tab bar at the top
			v.Height--
			v.y = 1
		}
	} else {
		if v.y == 1 {
			v.y = 0
			v.Height++
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
	if v.Topline+n <= v.Buf.NumLines {
		v.Topline += n
	} else if v.Topline < v.Buf.NumLines-1 {
		v.Topline++
	}
}

// CanClose returns whether or not the view can be closed
// If there are unsaved changes, the user will be asked if the view can be closed
// causing them to lose the unsaved changes
func (v *View) CanClose() bool {
	if v.Type == vtDefault && v.Buf.IsModified {
		var choice bool
		var canceled bool
		if v.Buf.Settings["autosave"].(bool) {
			choice = true
		} else {
			choice, canceled = messenger.YesNoPrompt("Save changes to " + v.Buf.GetName() + " before closing? (y,n,esc) ")
		}
		if !canceled {
			//if char == 'y' {
			if choice {
				v.Save(true)
				return true
			} else {
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

	// Set mouseReleased to true because we assume the mouse is not being pressed when
	// the editor is opened
	v.mouseReleased = true
	v.lastClickTime = time.Time{}
}

// Open opens the given file in the view
func (v *View) Open(filename string) {
	home, _ := homedir.Dir()
	filename = strings.Replace(filename, "~", home, 1)
	file, err := os.Open(filename)
	fileInfo, _ := os.Stat(filename)

	if err == nil && fileInfo.IsDir() {
		messenger.Error(filename, " is a directory")
		return
	}

	defer file.Close()

	var buf *Buffer
	if err != nil {
		messenger.Message(err.Error())
		// File does not exist -- create an empty buffer with that name
		buf = NewBufferFromString("", filename)
	} else {
		buf = NewBuffer(file, FSize(file), filename)
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
	}
}

// HSplit opens a horizontal split with the given buffer
func (v *View) HSplit(buf *Buffer) {
	i := 0
	if v.Buf.Settings["splitBottom"].(bool) {
		i = 1
	}
	v.splitNode.HSplit(buf, v.Num+i)
}

// VSplit opens a vertical split with the given buffer
func (v *View) VSplit(buf *Buffer) {
	i := 0
	if v.Buf.Settings["splitRight"].(bool) {
		i = 1
	}
	v.splitNode.VSplit(buf, v.Num+i)
}

// HSplitIndex opens a horizontal split with the given buffer at the given index
func (v *View) HSplitIndex(buf *Buffer, splitIndex int) {
	v.splitNode.HSplit(buf, splitIndex)
}

// VSplitIndex opens a vertical split with the given buffer at the given index
func (v *View) VSplitIndex(buf *Buffer, splitIndex int) {
	v.splitNode.VSplit(buf, splitIndex)
}

// GetSoftWrapLocation gets the location of a visual click on the screen and converts it to col,line
func (v *View) GetSoftWrapLocation(vx, vy int) (int, int) {
	if !v.Buf.Settings["softwrap"].(bool) {
		if vy >= v.Buf.NumLines {
			vy = v.Buf.NumLines - 1
		}
		vx = v.Cursor.GetCharPosInLine(vy, vx)
		return vx, vy
	}

	screenX, screenY := 0, v.Topline
	for lineN := v.Topline; lineN < v.Bottomline(); lineN++ {
		line := v.Buf.Line(lineN)
		if lineN >= v.Buf.NumLines {
			return 0, v.Buf.NumLines - 1
		}

		colN := 0
		for _, ch := range line {
			if screenX >= v.Width-v.lineNumOffset {
				screenX = 0
				screenY++
			}

			if screenX == vx && screenY == vy {
				return colN, lineN
			}

			if ch == '\t' {
				screenX += int(v.Buf.Settings["tabsize"].(float64)) - 1
			}

			screenX++
			colN++
		}
		if screenY == vy {
			return colN, lineN
		}
		screenX = 0
		screenY++
	}

	return 0, 0
}

func (v *View) Bottomline() int {
	if !v.Buf.Settings["softwrap"].(bool) {
		return v.Topline + v.Height
	}

	screenX, screenY := 0, 0
	numLines := 0
	for lineN := v.Topline; lineN < v.Topline+v.Height; lineN++ {
		line := v.Buf.Line(lineN)

		colN := 0
		for _, ch := range line {
			if screenX >= v.Width-v.lineNumOffset {
				screenX = 0
				screenY++
			}

			if ch == '\t' {
				screenX += int(v.Buf.Settings["tabsize"].(float64)) - 1
			}

			screenX++
			colN++
		}
		screenX = 0
		screenY++
		numLines++

		if screenY >= v.Height {
			break
		}
	}
	return numLines + v.Topline
}

// Relocate moves the view window so that the cursor is in view
// This is useful if the user has scrolled far away, and then starts typing
func (v *View) Relocate() bool {
	height := v.Bottomline() - v.Topline
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
	if cy > v.Topline+height-1-scrollmargin && cy < v.Buf.NumLines-scrollmargin {
		v.Topline = cy - height + 1 + scrollmargin
		ret = true
	} else if cy >= v.Buf.NumLines-scrollmargin && cy > height {
		v.Topline = v.Buf.NumLines - height
		ret = true
	}

	if !v.Buf.Settings["softwrap"].(bool) {
		cx := v.Cursor.GetVisualX()
		if cx < v.leftCol {
			v.leftCol = cx
			ret = true
		}
		if cx+v.lineNumOffset+1 > v.leftCol+v.Width {
			v.leftCol = cx - v.Width + v.lineNumOffset + 1
			ret = true
		}
	}
	return ret
}

// MoveToMouseClick moves the cursor to location x, y assuming x, y were given
// by a mouse click
func (v *View) MoveToMouseClick(x, y int) {
	if y-v.Topline > v.Height-1 {
		v.ScrollDown(1)
		y = v.Height + v.Topline - 1
	}
	if y < 0 {
		y = 0
	}
	if x < 0 {
		x = 0
	}

	x, y = v.GetSoftWrapLocation(x, y)
	// x = v.Cursor.GetCharPosInLine(y, x)
	if x > Count(v.Buf.Line(y)) {
		x = Count(v.Buf.Line(y))
	}
	v.Cursor.X = x
	v.Cursor.Y = y
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()
}

func (v *View) ExecuteActions(actions []func(*View, bool) bool) bool {
	relocate := false
	readonlyBindingsList := []string{"Delete", "Insert", "Backspace", "Cut", "Play", "Paste", "Move", "Add", "DuplicateLine", "Macro"}
	for _, action := range actions {
		readonlyBindingsResult := false
		funcName := ShortFuncName(action)
		if v.Type.readonly == true {
			// check for readonly and if true only let key bindings get called if they do not change the contents.
			for _, readonlyBindings := range readonlyBindingsList {
				if strings.Contains(funcName, readonlyBindings) {
					readonlyBindingsResult = true
				}
			}
		}
		if !readonlyBindingsResult {
			// call the key binding
			relocate = action(v, true) || relocate
			// Macro
			if funcName != "ToggleMacro" && funcName != "PlayMacro" {
				if recordingMacro {
					curMacro = append(curMacro, action)
				}
			}
		}
	}

	return relocate
}

// HandleEvent handles an event passed by the main loop
func (v *View) HandleEvent(event tcell.Event) {
	// This bool determines whether the view is relocated at the end of the function
	// By default it's true because most events should cause a relocate
	relocate := true

	v.Buf.CheckModTime()

	switch e := event.(type) {
	case *tcell.EventKey:
		// Check first if input is a key binding, if it is we 'eat' the input and don't insert a rune
		isBinding := false
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
					relocate = v.ExecuteActions(actions)
					break
				}
			}
		}
		if !isBinding && e.Key() == tcell.KeyRune {
			// Check viewtype if readonly don't insert a rune (readonly help and log view etc.)
			if v.Type.readonly == false {
				// Insert a character
				if v.Cursor.HasSelection() {
					v.Cursor.DeleteSelection()
					v.Cursor.ResetSelection()
				}
				v.Buf.Insert(v.Cursor.Loc, string(e.Rune()))
				v.Cursor.Right()

				for pl := range loadedPlugins {
					_, err := Call(pl+".onRune", string(e.Rune()), v)
					if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
						TermMessage(err)
					}
				}

				if recordingMacro {
					curMacro = append(curMacro, e.Rune())
				}
			}
		}
	case *tcell.EventPaste:
		// Check viewtype if readonly don't paste (readonly help and log view etc.)
		if v.Type.readonly == false {
			if !PreActionCall("Paste", v) {
				break
			}

			v.paste(e.Text())

			PostActionCall("Paste", v)
		}
	case *tcell.EventMouse:
		x, y := e.Position()
		x -= v.lineNumOffset - v.leftCol + v.x
		y += v.Topline - v.y
		// Don't relocate for mouse events
		relocate = false

		button := e.Buttons()

		for key, actions := range bindings {
			if button == key.buttons {
				relocate = v.ExecuteActions(actions)
			}
		}

		for key, actions := range mouseBindings {
			if button == key.buttons {
				for _, action := range actions {
					action(v, true, e)
				}
			}
		}

		switch button {
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
					v.Cursor.CopySelection("primary")
				}
				v.mouseReleased = true
			}
		}
	}

	if relocate {
		v.Relocate()
		// We run relocate again because there's a bug with relocating with softwrap
		// when for example you jump to the bottom of the buffer and it tries to
		// calculate where to put the topline so that the bottom line is at the bottom
		// of the terminal and it runs into problems with visual lines vs real lines.
		// This is (hopefully) a temporary solution
		v.Relocate()
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
	if data, err := FindRuntimeFile(RTHelp, helpPage).Data(); err != nil {
		TermMessage("Unable to load help text", helpPage, "\n", err)
	} else {
		helpBuffer := NewBufferFromString(string(data), helpPage+".md")
		helpBuffer.name = "Help"

		if v.Type == vtHelp {
			v.OpenBuffer(helpBuffer)
		} else {
			v.HSplit(helpBuffer)
			CurView().Type = vtHelp
		}
	}
}

func (v *View) DisplayView() {
	if v.Buf.Settings["softwrap"].(bool) && v.leftCol != 0 {
		v.leftCol = 0
	}

	if v.Type == vtLog {
		// Log views should always follow the cursor...
		v.Relocate()
	}

	// We need to know the string length of the largest line number
	// so we can pad appropriately when displaying line numbers
	maxLineNumLength := len(strconv.Itoa(v.Buf.NumLines))

	if v.Buf.Settings["ruler"] == true {
		// + 1 for the little space after the line number
		v.lineNumOffset = maxLineNumLength + 1
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

	divider := 0
	if v.x != 0 {
		// One space for the extra split divider
		v.lineNumOffset++
		divider = 1
	}

	xOffset := v.x + v.lineNumOffset
	yOffset := v.y

	height := v.Height
	width := v.Width
	left := v.leftCol
	top := v.Topline

	v.cellview.Draw(v.Buf, top, height, left, width-v.lineNumOffset)

	screenX := v.x
	realLineN := top - 1
	visualLineN := 0
	var line []*Char
	for visualLineN, line = range v.cellview.lines {
		var firstChar *Char
		if len(line) > 0 {
			firstChar = line[0]
		}

		var softwrapped bool
		if firstChar != nil {
			if firstChar.realLoc.Y == realLineN {
				softwrapped = true
			}
			realLineN = firstChar.realLoc.Y
		} else {
			realLineN++
		}

		colorcolumn := int(v.Buf.Settings["colorcolumn"].(float64))
		if colorcolumn != 0 {
			style := GetColor("color-column")
			fg, _, _ := style.Decompose()
			st := defStyle.Background(fg)
			screen.SetContent(xOffset+colorcolumn, yOffset+visualLineN, ' ', nil, st)
		}

		screenX = v.x

		// If there are gutter messages we need to display the '>>' symbol here
		if hasGutterMessages {
			// msgOnLine stores whether or not there is a gutter message on this line in particular
			msgOnLine := false
			for k := range v.messages {
				for _, msg := range v.messages[k] {
					if msg.lineNum == realLineN {
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
						screen.SetContent(screenX, yOffset+visualLineN, '>', nil, gutterStyle)
						screenX++
						screen.SetContent(screenX, yOffset+visualLineN, '>', nil, gutterStyle)
						screenX++
						if v.Cursor.Y == realLineN && !messenger.hasPrompt {
							messenger.Message(msg.msg)
							messenger.gutterMessage = true
						}
					}
				}
			}
			// If there is no message on this line we just display an empty offset
			if !msgOnLine {
				screen.SetContent(screenX, yOffset+visualLineN, ' ', nil, defStyle)
				screenX++
				screen.SetContent(screenX, yOffset+visualLineN, ' ', nil, defStyle)
				screenX++
				if v.Cursor.Y == realLineN && messenger.gutterMessage {
					messenger.Reset()
					messenger.gutterMessage = false
				}
			}
		}

		lineNumStyle := defStyle
		if v.Buf.Settings["ruler"] == true {
			// Write the line number
			if style, ok := colorscheme["line-number"]; ok {
				lineNumStyle = style
			}
			if style, ok := colorscheme["current-line-number"]; ok {
				if realLineN == v.Cursor.Y && tabs[curTab].CurView == v.Num && !v.Cursor.HasSelection() {
					lineNumStyle = style
				}
			}

			lineNum := strconv.Itoa(realLineN + 1)

			// Write the spaces before the line number if necessary
			for i := 0; i < maxLineNumLength-len(lineNum); i++ {
				screen.SetContent(screenX+divider, yOffset+visualLineN, ' ', nil, lineNumStyle)
				screenX++
			}
			if softwrapped && visualLineN != 0 {
				// Pad without the line number because it was written on the visual line before
				for range lineNum {
					screen.SetContent(screenX+divider, yOffset+visualLineN, ' ', nil, lineNumStyle)
					screenX++
				}
			} else {
				// Write the actual line number
				for _, ch := range lineNum {
					screen.SetContent(screenX+divider, yOffset+visualLineN, ch, nil, lineNumStyle)
					screenX++
				}
			}

			// Write the extra space
			screen.SetContent(screenX+divider, yOffset+visualLineN, ' ', nil, lineNumStyle)
			screenX++
		}

		var lastChar *Char
		cursorSet := false
		for _, char := range line {
			if char != nil {
				lineStyle := char.style

				colorcolumn := int(v.Buf.Settings["colorcolumn"].(float64))
				if colorcolumn != 0 && char.visualLoc.X == colorcolumn {
					style := GetColor("color-column")
					fg, _, _ := style.Decompose()
					lineStyle = lineStyle.Background(fg)
				}

				charLoc := char.realLoc
				if v.Cursor.HasSelection() &&
					(charLoc.GreaterEqual(v.Cursor.CurSelection[0]) && charLoc.LessThan(v.Cursor.CurSelection[1]) ||
						charLoc.LessThan(v.Cursor.CurSelection[0]) && charLoc.GreaterEqual(v.Cursor.CurSelection[1])) {
					// The current character is selected
					lineStyle = defStyle.Reverse(true)

					if style, ok := colorscheme["selection"]; ok {
						lineStyle = style
					}
				}

				if tabs[curTab].CurView == v.Num && !v.Cursor.HasSelection() &&
					v.Cursor.Y == char.realLoc.Y && v.Cursor.X == char.realLoc.X && !cursorSet {
					screen.ShowCursor(xOffset+char.visualLoc.X, yOffset+char.visualLoc.Y)
					cursorSet = true
				}

				if v.Buf.Settings["cursorline"].(bool) && tabs[curTab].CurView == v.Num &&
					!v.Cursor.HasSelection() && v.Cursor.Y == realLineN {
					style := GetColor("cursor-line")
					fg, _, _ := style.Decompose()
					lineStyle = lineStyle.Background(fg)
				}

				screen.SetContent(xOffset+char.visualLoc.X, yOffset+char.visualLoc.Y, char.drawChar, nil, lineStyle)

				lastChar = char
			}
		}

		lastX := 0
		var realLoc Loc
		var visualLoc Loc
		var cx, cy int
		if lastChar != nil {
			lastX = xOffset + lastChar.visualLoc.X + lastChar.width
			if tabs[curTab].CurView == v.Num && !v.Cursor.HasSelection() &&
				v.Cursor.Y == lastChar.realLoc.Y && v.Cursor.X == lastChar.realLoc.X+1 {
				screen.ShowCursor(lastX, yOffset+lastChar.visualLoc.Y)
				cx, cy = lastX, yOffset+lastChar.visualLoc.Y
			}
			realLoc = Loc{lastChar.realLoc.X, realLineN}
			visualLoc = Loc{lastX - xOffset, lastChar.visualLoc.Y}
		} else if len(line) == 0 {
			if tabs[curTab].CurView == v.Num && !v.Cursor.HasSelection() &&
				v.Cursor.Y == realLineN {
				screen.ShowCursor(xOffset, yOffset+visualLineN)
				cx, cy = xOffset, yOffset+visualLineN
			}
			lastX = xOffset
			realLoc = Loc{0, realLineN}
			visualLoc = Loc{0, visualLineN}
		}

		if v.Cursor.HasSelection() &&
			(realLoc.GreaterEqual(v.Cursor.CurSelection[0]) && realLoc.LessThan(v.Cursor.CurSelection[1]) ||
				realLoc.LessThan(v.Cursor.CurSelection[0]) && realLoc.GreaterEqual(v.Cursor.CurSelection[1])) {
			// The current character is selected
			selectStyle := defStyle.Reverse(true)

			if style, ok := colorscheme["selection"]; ok {
				selectStyle = style
			}
			screen.SetContent(xOffset+visualLoc.X, yOffset+visualLoc.Y, ' ', nil, selectStyle)
		}

		if v.Buf.Settings["cursorline"].(bool) && tabs[curTab].CurView == v.Num &&
			!v.Cursor.HasSelection() && v.Cursor.Y == realLineN {
			for i := lastX; i < xOffset+v.Width; i++ {
				style := GetColor("cursor-line")
				fg, _, _ := style.Decompose()
				style = style.Background(fg)
				if !(tabs[curTab].CurView == v.Num && !v.Cursor.HasSelection() && i == cx && yOffset+visualLineN == cy) {
					screen.SetContent(i, yOffset+visualLineN, ' ', nil, style)
				}
			}
		}
	}

	if divider != 0 {
		dividerStyle := defStyle
		if style, ok := colorscheme["divider"]; ok {
			dividerStyle = style
		}
		for i := 0; i < v.Height; i++ {
			screen.SetContent(v.x, yOffset+i, '|', nil, dividerStyle.Reverse(true))
		}
	}
}

// Display renders the view, the cursor, and statusline
func (v *View) Display() {
	if globalSettings["termtitle"].(bool) {
		screen.SetTitle("micro: " + v.Buf.GetName())
	}
	v.DisplayView()
	// Don't draw the cursor if it is out of the viewport or if it has a selection
	if v.Num == tabs[curTab].CurView && (v.Cursor.Y-v.Topline < 0 || v.Cursor.Y-v.Topline > v.Height-1 || v.Cursor.HasSelection()) {
		screen.HideCursor()
	}
	_, screenH := screen.Size()
	if v.Buf.Settings["statusline"].(bool) {
		v.sline.Display()
	} else if (v.y + v.Height) != screenH-1 {
		for x := 0; x < v.Width; x++ {
			screen.SetContent(v.x+x, v.y+v.Height, '-', nil, defStyle.Reverse(true))
		}
	}
}
