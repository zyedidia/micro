package main

import (
	"os"
	"strconv"
	"strings"
	"time"
	"reflect"

	"github.com/yuin/gopher-lua"
	"github.com/zyedidia/clipboard"
)

// PreActionCall executes the lua pre callback if possible
func PreActionCall(funcName string, view *View) bool {
	executeAction := true
	for _, pl := range loadedPlugins {
		ret, err := Call(pl+".pre"+funcName, view)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
		if ret == lua.LFalse {
			executeAction = false
		}
	}
	return executeAction
}

// PostActionCall executes the lua plugin callback if possible
func PostActionCall(funcName string, view *View) bool {
	relocate := true
	for _, pl := range loadedPlugins {
		ret, err := Call(pl+".on"+funcName, view)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
		if ret == lua.LFalse {
			relocate = false
		}
	}
	return relocate
}

// DoActions Performs view actions (e.g. "IndentSelection,InsertTab")
// This handles pre and post actions for plugins
func (v *View) DoActions(actions string) bool {
	relocate := false

	for _, action := range strings.Split(actions, ",") {
		_, ok := reflect.TypeOf(v).MethodByName(action)
		if ok {
			if PreActionCall(action, v) {
				fn := reflect.ValueOf(v).MethodByName(action)
				relocate = fn.Call([]reflect.Value{})[0].Bool() || relocate
				relocate = PostActionCall(action, v) || relocate
			}
			if action != "ToggleMacro" && action != "PlayMacro" {
				if recordingMacro {
					curMacro = append(curMacro, action)
				}
			}
		} else {
			relocate = LuaAction(action) || relocate
		}
	}

	return relocate
}

func (v *View) deselect(index int) bool {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[index]
		v.Cursor.ResetSelection()
		return true
	}
	return false
}

// Center centers the view on the cursor
func (v *View) Center() bool {
	v.Topline = v.Cursor.Y - v.height/2
	if v.Topline+v.height > v.Buf.NumLines {
		v.Topline = v.Buf.NumLines - v.height
	}
	if v.Topline < 0 {
		v.Topline = 0
	}

	return true
}

// CursorUp moves the cursor up
func (v *View) CursorUp() bool {
	v.deselect(0)
	v.Cursor.Up()
	return true
}

// CursorDown moves the cursor down
func (v *View) CursorDown() bool {
	v.deselect(1)
	v.Cursor.Down()
	return true
}

// CursorLeft moves the cursor left
func (v *View) CursorLeft() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Left()
	}

	return true
}

// CursorRight moves the cursor right
func (v *View) CursorRight() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1].Move(-1, v.Buf)
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Right()
	}
	return true
}

// WordRight moves the cursor one word to the right
func (v *View) WordRight() bool {
	v.Cursor.WordRight()
	return true
}

// WordLeft moves the cursor one word to the left
func (v *View) WordLeft() bool {
	v.Cursor.WordLeft()
	return true
}

// SelectUp selects up one line
func (v *View) SelectUp() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Up()
	v.Cursor.SelectTo(v.Cursor.Loc)
	return true
}

// SelectDown selects down one line
func (v *View) SelectDown() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Down()
	v.Cursor.SelectTo(v.Cursor.Loc)
	return true
}

// SelectLeft selects the character to the left of the cursor
func (v *View) SelectLeft() bool {
	loc := v.Cursor.Loc
	count := v.Buf.End().Move(-1, v.Buf)
	if loc.GreaterThan(count) {
		loc = count
	}
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = loc
	}
	v.Cursor.Left()
	v.Cursor.SelectTo(v.Cursor.Loc)
	return true
}

// SelectRight selects the character to the right of the cursor
func (v *View) SelectRight() bool {
	loc := v.Cursor.Loc
	count := v.Buf.End().Move(-1, v.Buf)
	if loc.GreaterThan(count) {
		loc = count
	}
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = loc
	}
	v.Cursor.Right()
	v.Cursor.SelectTo(v.Cursor.Loc)
	return true
}

// SelectWordRight selects the word to the right of the cursor
func (v *View) SelectWordRight() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordRight()
	v.Cursor.SelectTo(v.Cursor.Loc)
	return true
}

// SelectWordLeft selects the word to the left of the cursor
func (v *View) SelectWordLeft() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordLeft()
	v.Cursor.SelectTo(v.Cursor.Loc)
	return true
}

// StartOfLine moves the cursor to the start of the line
func (v *View) StartOfLine() bool {
	v.deselect(0)
	v.Cursor.Start()
	return true
}

// EndOfLine moves the cursor to the end of the line
func (v *View) EndOfLine() bool {
	v.deselect(0)
	v.Cursor.End()
	return true
}

// SelectToStartOfLine selects to the start of the current line
func (v *View) SelectToStartOfLine() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Start()
	v.Cursor.SelectTo(v.Cursor.Loc)
	return true
}

// SelectToEndOfLine selects to the end of the current line
func (v *View) SelectToEndOfLine() bool {
	v.Cursor.End()
	v.Cursor.SelectTo(v.Cursor.Loc)
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (v *View) CursorStart() bool {
	v.deselect(0)

	v.Cursor.X = 0
	v.Cursor.Y = 0

	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (v *View) CursorEnd() bool {
	v.deselect(0)

	v.Cursor.Loc = v.Buf.End()

	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (v *View) SelectToStart() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorStart()
	v.Cursor.SelectTo(v.Buf.Start())

	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (v *View) SelectToEnd() bool {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorEnd()
	v.Cursor.SelectTo(v.Buf.End())

	return true
}

// InsertSpace inserts a space
func (v *View) InsertSpace() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.Buf.Insert(v.Cursor.Loc, " ")
	v.Cursor.Right()

	return true
}

// InsertNewline inserts a newline plus possible some whitespace if autoindent is on
func (v *View) InsertNewline() bool {
	// Insert a newline
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	v.Buf.Insert(v.Cursor.Loc, "\n")
	ws := GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))
	v.Cursor.Right()

	if v.Buf.Settings["autoindent"].(bool) {
		v.Buf.Insert(v.Cursor.Loc, ws)
		for i := 0; i < len(ws); i++ {
			v.Cursor.Right()
		}

		if IsSpacesOrTabs(v.Buf.Line(v.Cursor.Y - 1)) {
			line := v.Buf.Line(v.Cursor.Y - 1)
			v.Buf.Remove(Loc{0, v.Cursor.Y - 1}, Loc{Count(line), v.Cursor.Y - 1})
		}
	}
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()

	return true
}

// InsertEnter calls InsertNewline for backwards compatability
func (v *View) InsertEnter() bool {
	return v.InsertNewline()
}

// Backspace deletes the previous character
func (v *View) Backspace() bool {
	// Delete a character
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else if v.Cursor.Loc.GreaterThan(v.Buf.Start()) {
		// We have to do something a bit hacky here because we want to
		// delete the line by first moving left and then deleting backwards
		// but the undo redo would place the cursor in the wrong place
		// So instead we move left, save the position, move back, delete
		// and restore the position

		// If the user is using spaces instead of tabs and they are deleting
		// whitespace at the start of the line, we should delete as if it's a
		// tab (tabSize number of spaces)
		lineStart := v.Buf.Line(v.Cursor.Y)[:v.Cursor.X]
		tabSize := int(v.Buf.Settings["tabsize"].(float64))
		if v.Buf.Settings["tabstospaces"].(bool) && IsSpaces(lineStart) && len(lineStart) != 0 && len(lineStart)%tabSize == 0 {
			loc := v.Cursor.Loc
			v.Cursor.Loc = loc.Move(-tabSize, v.Buf)
			cx, cy := v.Cursor.X, v.Cursor.Y
			v.Cursor.Loc = loc
			v.Buf.Remove(loc.Move(-tabSize, v.Buf), loc)
			v.Cursor.X, v.Cursor.Y = cx, cy
		} else {
			v.Cursor.Left()
			cx, cy := v.Cursor.X, v.Cursor.Y
			v.Cursor.Right()
			loc := v.Cursor.Loc
			v.Buf.Remove(loc.Move(-1, v.Buf), loc)
			v.Cursor.X, v.Cursor.Y = cx, cy
		}
	}
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()

	return true
}

// DeleteWordRight deletes the word to the right of the cursor
func (v *View) DeleteWordRight() bool {
	v.SelectWordRight()
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	return true
}

// DeleteWordLeft deletes the word to the left of the cursor
func (v *View) DeleteWordLeft() bool {
	v.SelectWordLeft()
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	return true
}

// Delete deletes the next character
func (v *View) Delete() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else {
		loc := v.Cursor.Loc
		if loc.LessThan(v.Buf.End()) {
			v.Buf.Remove(loc, loc.Move(1, v.Buf))
		}
	}

	return true
}

// IndentSelection indents the current selection
func (v *View) IndentSelection() bool {
	if v.Cursor.HasSelection() {
		startY := v.Cursor.CurSelection[0].Y
		endY := v.Cursor.CurSelection[1].Move(-1, v.Buf).Y
		endX := v.Cursor.CurSelection[1].Move(-1, v.Buf).X
		for y := startY; y <= endY; y++ {
			tabsize := len(v.Buf.IndentString())
			v.Buf.Insert(Loc{0, y}, v.Buf.IndentString())
			if y == startY && v.Cursor.CurSelection[0].X > 0 {
				v.Cursor.SetSelectionStart(v.Cursor.CurSelection[0].Move(tabsize, v.Buf))
			}
			if y == endY {
				v.Cursor.SetSelectionEnd(Loc{endX + tabsize + 1, endY})
			}
		}
		v.Cursor.Relocate()
		return true
	}
	return false
}

// OutdentLine moves the current line back one indentation
func (v *View) OutdentLine() bool {
	if v.Cursor.HasSelection() {
		return false
	}
	
	for x := 0; x < len(v.Buf.IndentString()); x++ {
		if len(GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))) == 0 {
			break
		}
		v.Buf.Remove(Loc{0, v.Cursor.Y}, Loc{1, v.Cursor.Y})
		v.Cursor.X -= 1
	}
	v.Cursor.Relocate()

	return true
}

// OutdentSelection takes the current selection and moves it back one indent level
func (v *View) OutdentSelection() bool {
	if v.Cursor.HasSelection() {
		startY := v.Cursor.CurSelection[0].Y
		endY := v.Cursor.CurSelection[1].Move(-1, v.Buf).Y
		endX := v.Cursor.CurSelection[1].Move(-1, v.Buf).X
		for y := startY; y <= endY; y++ {
			for x := 0; x < len(v.Buf.IndentString()); x++ {
				if len(GetLeadingWhitespace(v.Buf.Line(y))) == 0 {
					break
				}
				v.Buf.Remove(Loc{0, y}, Loc{1, y})
				if y == startY && v.Cursor.CurSelection[0].X > 0 {
					v.Cursor.SetSelectionStart(v.Cursor.CurSelection[0].Move(-1, v.Buf))
				}
				if y == endY {
					v.Cursor.SetSelectionEnd(Loc{endX - x, endY})
				}
			}
		}
		v.Cursor.Relocate()

		return true
	}
	return false
}

// InsertTab inserts a tab or spaces
func (v *View) InsertTab() bool {
	if v.Cursor.HasSelection() {
		return false
	}
	
	tabBytes := len(v.Buf.IndentString())
	bytesUntilIndent := tabBytes - (v.Cursor.GetVisualX() % tabBytes)
	v.Buf.Insert(v.Cursor.Loc, v.Buf.IndentString()[:bytesUntilIndent])
	for i := 0; i < bytesUntilIndent; i++ {
		v.Cursor.Right()
	}

	return true
}

// Save the buffer to disk
func (v *View) Save() bool {
	if v.Type == vtHelp {
		// We can't save the help text
		return false
	}
	// If this is an empty buffer, ask for a filename
	if v.Buf.Path == "" {
		filename, canceled := messenger.Prompt("Filename: ", "Save", NoCompletion)
		if !canceled {
			// the filename might or might not be quoted, so unquote first then join the strings.
			filename = strings.Join(SplitCommandArgs(filename), " ")
			v.Buf.Path = filename
			v.Buf.Name = filename
		} else {
			return false
		}
	}
	err := v.Buf.Save()
	if err != nil {
		if strings.HasSuffix(err.Error(), "permission denied") {
			choice, _ := messenger.YesNoPrompt("Permission denied. Do you want to save this file using sudo? (y,n)")
			if choice {
				err = v.Buf.SaveWithSudo()
				if err != nil {
					messenger.Error(err.Error())
					return false
				}
				messenger.Message("Saved " + v.Buf.Path)
			}
			messenger.Reset()
			messenger.Clear()
		} else {
			messenger.Error(err.Error())
		}
	} else {
		messenger.Message("Saved " + v.Buf.Path)
	}

	return false
}

// SaveAs saves the buffer to disk with the given name
func (v *View) SaveAs() bool {
	filename, canceled := messenger.Prompt("Filename: ", "Save", NoCompletion)
	if !canceled {
		// the filename might or might not be quoted, so unquote first then join the strings.
		filename = strings.Join(SplitCommandArgs(filename), " ")
		v.Buf.Path = filename
		v.Buf.Name = filename

		v.DoActions("Save")
	}

	return false
}

// Find opens a prompt and searches forward for the input
func (v *View) Find() bool {
	searchStr := ""
	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
		searchStr = v.Cursor.GetSelection()
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	BeginSearch(searchStr)

	return true
}

// FindNext searches forwards for the last used search term
func (v *View) FindNext() bool {
	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
		lastSearch = v.Cursor.GetSelection()
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	if lastSearch == "" {
		return true
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, true)

	return true
}

// FindPrevious searches backwards for the last used search term
func (v *View) FindPrevious() bool {
	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[0], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, false)

	return true
}

// Undo undoes the last action
func (v *View) Undo() bool {
	v.Buf.Undo()
	messenger.Message("Undid action")

	return true
}

// Redo redoes the last action
func (v *View) Redo() bool {
	v.Buf.Redo()
	messenger.Message("Redid action")

	return true
}

// Copy the selection to the system clipboard
func (v *View) Copy() bool {
	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection(), "clipboard")
		v.freshClip = true
		messenger.Message("Copied selection")
	}
	return true
}

// CutLine cuts the current line to the clipboard
func (v *View) CutLine() bool {
	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	if v.freshClip == true {
		if v.Cursor.HasSelection() {
			if clip, err := clipboard.ReadAll("clipboard"); err != nil {
				messenger.Error(err)
			} else {
				clipboard.WriteAll(clip+v.Cursor.GetSelection(), "clipboard")
			}
		}
	} else if time.Since(v.lastCutTime)/time.Second > 10*time.Second || v.freshClip == false {
		v.DoActions("Copy")
	}
	v.freshClip = true
	v.lastCutTime = time.Now()
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()
	messenger.Message("Cut line")

	return true
}

// Cut the selection to the system clipboard
func (v *View) Cut() bool {
	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection(), "clipboard")
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		v.freshClip = true
		messenger.Message("Cut selection")

		return true
	}

	return false
}

// DuplicateLine duplicates the current line or selection
func (v *View) DuplicateLine() bool {
	if v.Cursor.HasSelection() {
		v.Buf.Insert(v.Cursor.CurSelection[1], v.Cursor.GetSelection())
	} else {
		v.Cursor.End()
		v.Buf.Insert(v.Cursor.Loc, "\n"+v.Buf.Line(v.Cursor.Y))
		v.Cursor.Right()
	}

	messenger.Message("Duplicated line")

	return true
}

// DeleteLine deletes the current line
func (v *View) DeleteLine() bool {
	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()
	messenger.Message("Deleted line")

	return true
}

// MoveLinesUp moves up the current line or selected lines if any
func (v *View) MoveLinesUp() bool {
	if v.Cursor.HasSelection() {
		if v.Cursor.CurSelection[0].Y == 0 {
			messenger.Message("Can not move further up")
			return true
		}
		v.Buf.MoveLinesUp(
			v.Cursor.CurSelection[0].Y,
			v.Cursor.CurSelection[1].Y,
		)
		v.Cursor.UpN(1)
		v.Cursor.CurSelection[0].Y -= 1
		v.Cursor.CurSelection[1].Y -= 1
		messenger.Message("Moved up selected line(s)")
	} else {
		if v.Cursor.Loc.Y == 0 {
			messenger.Message("Can not move further up")
			return true
		}
		v.Buf.MoveLinesUp(
			v.Cursor.Loc.Y,
			v.Cursor.Loc.Y+1,
		)
		v.Cursor.UpN(1)
		messenger.Message("Moved up current line")
	}
	v.Buf.IsModified = true

	return true
}

// MoveLinesDown moves down the current line or selected lines if any
func (v *View) MoveLinesDown() bool {
	if v.Cursor.HasSelection() {
		if v.Cursor.CurSelection[1].Y >= len(v.Buf.lines) {
			messenger.Message("Can not move further down")
			return true
		}
		v.Buf.MoveLinesDown(
			v.Cursor.CurSelection[0].Y,
			v.Cursor.CurSelection[1].Y,
		)
		v.Cursor.DownN(1)
		v.Cursor.CurSelection[0].Y += 1
		v.Cursor.CurSelection[1].Y += 1
		messenger.Message("Moved down selected line(s)")
	} else {
		if v.Cursor.Loc.Y >= len(v.Buf.lines)-1 {
			messenger.Message("Can not move further down")
			return true
		}
		v.Buf.MoveLinesDown(
			v.Cursor.Loc.Y,
			v.Cursor.Loc.Y+1,
		)
		v.Cursor.DownN(1)
		messenger.Message("Moved down current line")
	}
	v.Buf.IsModified = true

	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (v *View) Paste() bool {
	clip, _ := clipboard.ReadAll("clipboard")
	v.paste(clip)
	return true
}

// PastePrimary pastes from the primary clipboard (only use on linux)
func (v *View) PastePrimary() bool {
	clip, _ := clipboard.ReadAll("primary")
	v.paste(clip)
	return true
}

// SelectAll selects the entire buffer
func (v *View) SelectAll() bool {
	v.Cursor.SetSelectionStart(v.Buf.Start())
	v.Cursor.SetSelectionEnd(v.Buf.End())
	// Put the cursor at the beginning
	v.Cursor.X = 0
	v.Cursor.Y = 0

	return true
}

// OpenFile opens a new file in the buffer
func (v *View) OpenFile() bool {
	if v.CanClose() {
		filename, canceled := messenger.Prompt("File to open: ", "Open", FileCompletion)
		if canceled {
			return false
		}
		// the filename might or might not be quoted, so unquote first then join the strings.
		filename = strings.Join(SplitCommandArgs(filename), " ")

		v.Open(filename)

		return true
	}
	return false
}

// Start moves the viewport to the start of the buffer
func (v *View) Start() bool {
	v.Topline = 0
	return false
}

// End moves the viewport to the end of the buffer
func (v *View) End() bool {
	if v.height > v.Buf.NumLines {
		v.Topline = 0
	} else {
		v.Topline = v.Buf.NumLines - v.height
	}

	return false
}

// PageUp scrolls the view up a page
func (v *View) PageUp() bool {
	if v.Topline > v.height {
		v.ScrollUp(v.height)
	} else {
		v.Topline = 0
	}
	return false
}

// PageDown scrolls the view down a page
func (v *View) PageDown() bool {
	if v.Buf.NumLines-(v.Topline+v.height) > v.height {
		v.ScrollDown(v.height)
	} else if v.Buf.NumLines >= v.height {
		v.Topline = v.Buf.NumLines - v.height
	}

	return false
}

// CursorPageUp places the cursor a page up
func (v *View) CursorPageUp() bool {
	v.deselect(0)

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	}
	v.Cursor.UpN(v.height)

	return true
}

// CursorPageDown places the cursor a page up
func (v *View) CursorPageDown() bool {
	v.deselect(0)

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1]
		v.Cursor.ResetSelection()
	}
	v.Cursor.DownN(v.height)

	return true
}

// HalfPageUp scrolls the view up half a page
func (v *View) HalfPageUp() bool {
	if v.Topline > v.height/2 {
		v.ScrollUp(v.height / 2)
	} else {
		v.Topline = 0
	}

	return false
}

// HalfPageDown scrolls the view down half a page
func (v *View) HalfPageDown() bool {
	if v.Buf.NumLines-(v.Topline+v.height) > v.height/2 {
		v.ScrollDown(v.height / 2)
	} else {
		if v.Buf.NumLines >= v.height {
			v.Topline = v.Buf.NumLines - v.height
		}
	}

	return false
}

// ToggleRuler turns line numbers off and on
func (v *View) ToggleRuler() bool {
	if v.Buf.Settings["ruler"] == false {
		v.Buf.Settings["ruler"] = true
		messenger.Message("Enabled ruler")
	} else {
		v.Buf.Settings["ruler"] = false
		messenger.Message("Disabled ruler")
	}

	return false
}

// JumpLine jumps to a line and moves the view accordingly.
func (v *View) JumpLine() bool {
	// Prompt for line number
	linestring, canceled := messenger.Prompt("Jump to line # ", "LineNumber", NoCompletion)
	if canceled {
		return false
	}
	lineint, err := strconv.Atoi(linestring)
	lineint = lineint - 1 // fix offset
	if err != nil {
		messenger.Error(err) // return errors
		return false
	}
	// Move cursor and view if possible.
	if lineint < v.Buf.NumLines && lineint >= 0 {
		v.Cursor.X = 0
		v.Cursor.Y = lineint

		return true
	}
	messenger.Error("Only ", v.Buf.NumLines, " lines to jump")
	return false
}

// ClearStatus clears the messenger bar
func (v *View) ClearStatus() bool {
	messenger.Message("")
	return false
}

// ToggleHelp toggles the help screen
func (v *View) ToggleHelp() bool {
	if v.Type != vtHelp {
		// Open the default help
		v.openHelp("help")
	} else {
		v.DoActions("Quit")
	}

	return true
}

// ShellMode opens a terminal to run a shell command
func (v *View) ShellMode() bool {
	input, canceled := messenger.Prompt("$ ", "Shell", NoCompletion)
	if !canceled {
		// The true here is for openTerm to make the command interactive
		HandleShellCommand(input, true, true)
	}
	return false
}

// CommandMode lets the user enter a command
func (v *View) CommandMode() bool {
	input, canceled := messenger.Prompt("> ", "Command", CommandCompletion)
	if !canceled {
		HandleCommand(input)
	}

	return false
}

// Escape leaves current mode / quits the editor
func (v *View) Escape() bool {
	// check if user is searching, or the last search is still active
	if searching || lastSearch != "" {
		ExitSearch(v)
		return true
	}
	// check if a prompt is shown, hide it and don't quit
	if messenger.hasPrompt {
		messenger.Reset() // FIXME
		return true
	}
	return v.Quit()
}

// Quit quits the editor
// This behavior needs to be changed and should really only quit the editor if this
// is the last view
// However, since micro only supports one view for now, it doesn't really matter
func (v *View) Quit() bool {
	// Make sure not to quit if there are unsaved changes
	if v.CanClose() {
		v.CloseBuffer()
		if len(tabs[curTab].views) > 1 {
			v.splitNode.Delete()
			tabs[v.TabNum].Cleanup()
			tabs[v.TabNum].Resize()
		} else if len(tabs) > 1 {
			if len(tabs[v.TabNum].views) == 1 {
				tabs = tabs[:v.TabNum+copy(tabs[v.TabNum:], tabs[v.TabNum+1:])]
				for i, t := range tabs {
					t.SetNum(i)
				}
				if curTab >= len(tabs) {
					curTab--
				}
				if curTab == 0 {
					// CurView().Resize(screen.Size())
					CurView().ToggleTabbar()
					CurView().matches = Match(CurView())
				}
			}
		} else {
			PostActionCall("Quit", v)

			screen.Fini()
			os.Exit(0)
		}
	}
	return false
}

// QuitAll quits the whole editor; all splits and tabs
func (v *View) QuitAll() bool {
	closeAll := true
	for _, tab := range tabs {
		for _, v := range tab.views {
			if !v.CanClose() {
				closeAll = false
			}
		}
	}

	if closeAll {
		for _, tab := range tabs {
			for _, v := range tab.views {
				v.CloseBuffer()
			}
		}

		PostActionCall("QuitAll", v)

		screen.Fini()
		os.Exit(0)
	}

	return false
}

// AddTab adds a new tab with an empty buffer
func (v *View) AddTab() bool {
	tab := NewTabFromView(NewView(NewBuffer([]byte{}, "")))
	tab.SetNum(len(tabs))
	tabs = append(tabs, tab)
	curTab++
	if len(tabs) == 2 {
		for _, t := range tabs {
			for _, v := range t.views {
				v.ToggleTabbar()
			}
		}
	}

	return true
}

// PreviousTab switches to the previous tab in the tab list
func (v *View) PreviousTab() bool {
	if curTab > 0 {
		curTab--
	} else if curTab == 0 {
		curTab = len(tabs) - 1
	}
	return false
}

// NextTab switches to the next tab in the tab list
func (v *View) NextTab() bool {
	if curTab < len(tabs)-1 {
		curTab++
	} else if curTab == len(tabs)-1 {
		curTab = 0
	}

	return false
}

// VSplitBinding opens an empty vertical split
func (v *View) VSplitBinding() bool {
	v.VSplit(NewBuffer([]byte{}, ""))
	return false
}

// HSplitBinding opens an empty horizontal split
func (v *View) HSplitBinding() bool {
	v.HSplit(NewBuffer([]byte{}, ""))
	return false
}

// Unsplit closes all splits in the current tab except the active one
func (v *View) Unsplit() bool {
	curView := tabs[curTab].curView
	for i := len(tabs[curTab].views) - 1; i >= 0; i-- {
		view := tabs[curTab].views[i]
		if view != nil && view.Num != curView {
			v.DoActions("Quit")
		}
	}

	return false
}

// NextSplit changes the view to the next split
func (v *View) NextSplit() bool {
	tab := tabs[curTab]
	if tab.curView < len(tab.views)-1 {
		tab.curView++
	} else {
		tab.curView = 0
	}

	return false
}

// PreviousSplit changes the view to the previous split
func (v *View) PreviousSplit() bool {
	tab := tabs[curTab]
	if tab.curView > 0 {
		tab.curView--
	} else {
		tab.curView = len(tab.views) - 1
	}

	return false
}

var curMacro []interface{}
var recordingMacro bool

func (v *View) ToggleMacro() bool {
	recordingMacro = !recordingMacro

	if recordingMacro {
		curMacro = []interface{}{}
		messenger.Message("Recording")
	} else {
		messenger.Message("Stopped recording")
	}

	return true
}

func (v *View) PlayMacro() bool {
	for _, action := range curMacro {
		switch t := action.(type) {
		case rune:
			// Insert a character
			if v.Cursor.HasSelection() {
				v.Cursor.DeleteSelection()
				v.Cursor.ResetSelection()
			}
			v.Buf.Insert(v.Cursor.Loc, string(t))
			v.Cursor.Right()

			for _, pl := range loadedPlugins {
				_, err := Call(pl+".onRune", string(t), v)
				if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
					TermMessage(err)
				}
			}
		case string:
			v.DoActions(string(t))
		}
	}
	return true
}

// None is no action
func None() bool {
	return false
}
