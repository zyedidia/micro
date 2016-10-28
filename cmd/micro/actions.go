package main

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

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
func PostActionCall(funcName string, view *View) {
	for _, pl := range loadedPlugins {
		_, err := Call(pl+".on"+funcName, view)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
	}
}

// DoActions Performs view actions (e.g. "IndentSelection,InsertTab")
// This handles pre and post actions for plugins
func (v *View) DoActions(actions string) {
	for _, action := range strings.Split(actions, ",") {
		_, ok := reflect.TypeOf(v).MethodByName(action)
		if ok {
			if PreActionCall(action, v) {
				fn := reflect.ValueOf(v).MethodByName(action)
				fn.Call([]reflect.Value{})
				PostActionCall(action, v)
			}
			if action != "ToggleMacro" && action != "PlayMacro" {
				if recordingMacro {
					curMacro = append(curMacro, action)
				}
			}
		} else {
			LuaAction(action)
		}
	}
}

func (v *View) deselect(index int) {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[index]
		v.Cursor.ResetSelection()
	}
}

// Center centers the view on the cursor
func (v *View) Center(deprecatedUsePlugin ...bool) {
	v.Topline = v.Cursor.Y - v.height/2
	if v.Topline+v.height > v.Buf.NumLines {
		v.Topline = v.Buf.NumLines - v.height
	}
	if v.Topline < 0 {
		v.Topline = 0
	}
}

// CursorUp moves the cursor up
func (v *View) CursorUp(deprecatedUsePlugin ...bool) {
	v.deselect(0)
	v.Cursor.Up()
}

// CursorDown moves the cursor down
func (v *View) CursorDown(deprecatedUsePlugin ...bool) {
	v.deselect(1)
	v.Cursor.Down()
}

// CursorLeft moves the cursor left
func (v *View) CursorLeft(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Left()
	}
}

// CursorRight moves the cursor right
func (v *View) CursorRight(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1].Move(-1, v.Buf)
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Right()
	}
}

// WordRight moves the cursor one word to the right
func (v *View) WordRight(deprecatedUsePlugin ...bool) {
	v.Cursor.WordRight()
}

// WordLeft moves the cursor one word to the left
func (v *View) WordLeft(deprecatedUsePlugin ...bool) {
	v.Cursor.WordLeft()
}

// SelectUp selects up one line
func (v *View) SelectUp(deprecatedUsePlugin ...bool) {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Up()
	v.Cursor.SelectTo(v.Cursor.Loc)
}

// SelectDown selects down one line
func (v *View) SelectDown(deprecatedUsePlugin ...bool) {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Down()
	v.Cursor.SelectTo(v.Cursor.Loc)
}

// SelectLeft selects the character to the left of the cursor
func (v *View) SelectLeft(deprecatedUsePlugin ...bool) {
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
}

// SelectRight selects the character to the right of the cursor
func (v *View) SelectRight(deprecatedUsePlugin ...bool) {
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
}

// SelectWordRight selects the word to the right of the cursor
func (v *View) SelectWordRight(deprecatedUsePlugin ...bool) {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordRight()
	v.Cursor.SelectTo(v.Cursor.Loc)
}

// SelectWordLeft selects the word to the left of the cursor
func (v *View) SelectWordLeft(deprecatedUsePlugin ...bool) {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordLeft()
	v.Cursor.SelectTo(v.Cursor.Loc)
}

// StartOfLine moves the cursor to the start of the line
func (v *View) StartOfLine(deprecatedUsePlugin ...bool) {
	v.deselect(0)
	v.Cursor.Start()
}

// EndOfLine moves the cursor to the end of the line
func (v *View) EndOfLine(deprecatedUsePlugin ...bool) {
	v.deselect(0)
	v.Cursor.End()
}

// SelectToStartOfLine selects to the start of the current line
func (v *View) SelectToStartOfLine(deprecatedUsePlugin ...bool) {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Start()
	v.Cursor.SelectTo(v.Cursor.Loc)
}

// SelectToEndOfLine selects to the end of the current line
func (v *View) SelectToEndOfLine(deprecatedUsePlugin ...bool) {
	v.Cursor.End()
	v.Cursor.SelectTo(v.Cursor.Loc)
}

// CursorStart moves the cursor to the start of the buffer
func (v *View) CursorStart(deprecatedUsePlugin ...bool) {
	v.deselect(0)

	v.Cursor.X = 0
	v.Cursor.Y = 0
}

// CursorEnd moves the cursor to the end of the buffer
func (v *View) CursorEnd(deprecatedUsePlugin ...bool) {
	v.deselect(0)

	v.Cursor.Loc = v.Buf.End()
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (v *View) SelectToStart(deprecatedUsePlugin ...bool) {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorStart()
	v.Cursor.SelectTo(v.Buf.Start())
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (v *View) SelectToEnd(deprecatedUsePlugin ...bool) {
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorEnd()
	v.Cursor.SelectTo(v.Buf.End())
}

// InsertSpace inserts a space
func (v *View) InsertSpace(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.Buf.Insert(v.Cursor.Loc, " ")
	v.Cursor.Right()
}

// InsertNewline inserts a newline plus possible some whitespace if autoindent is on
func (v *View) InsertNewline(deprecatedUsePlugin ...bool) {
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
}

// InsertEnter calls InsertNewline for backwards compatability
func (v *View) InsertEnter(deprecatedUsePlugin ...bool) {
	v.InsertNewline()
}

// Backspace deletes the previous character
func (v *View) Backspace(deprecatedUsePlugin ...bool) {
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
}

// DeleteWordRight deletes the word to the right of the cursor
func (v *View) DeleteWordRight(deprecatedUsePlugin ...bool) {
	v.SelectWordRight()
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
}

// DeleteWordLeft deletes the word to the left of the cursor
func (v *View) DeleteWordLeft(deprecatedUsePlugin ...bool) {
	v.SelectWordLeft()
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
}

// Delete deletes the next character
func (v *View) Delete(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else {
		loc := v.Cursor.Loc
		if loc.LessThan(v.Buf.End()) {
			v.Buf.Remove(loc, loc.Move(1, v.Buf))
		}
	}
}

// IndentSelection indents the current selection
func (v *View) IndentSelection(deprecatedUsePlugin ...bool) {
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
	}
}

// OutdentLine moves the current line back one indentation
func (v *View) OutdentLine(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		return
	}

	for x := 0; x < len(v.Buf.IndentString()); x++ {
		if len(GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))) == 0 {
			break
		}
		v.Buf.Remove(Loc{0, v.Cursor.Y}, Loc{1, v.Cursor.Y})
		v.Cursor.X -= 1
	}
	v.Cursor.Relocate()
}

// OutdentSelection takes the current selection and moves it back one indent level
func (v *View) OutdentSelection(deprecatedUsePlugin ...bool) {
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
	}
}

// InsertTab inserts a tab or spaces
func (v *View) InsertTab(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		return
	}

	tabBytes := len(v.Buf.IndentString())
	bytesUntilIndent := tabBytes - (v.Cursor.GetVisualX() % tabBytes)
	v.Buf.Insert(v.Cursor.Loc, v.Buf.IndentString()[:bytesUntilIndent])
	for i := 0; i < bytesUntilIndent; i++ {
		v.Cursor.Right()
	}
}

// Save the buffer to disk
func (v *View) Save(deprecatedUsePlugin ...bool) {
	if v.Type == vtHelp {
		// We can't save the help text
		return
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
			return
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
					return
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
}

// SaveAs saves the buffer to disk with the given name
func (v *View) SaveAs(deprecatedUsePlugin ...bool) {
	filename, canceled := messenger.Prompt("Filename: ", "Save", NoCompletion)
	if !canceled {
		// the filename might or might not be quoted, so unquote first then join the strings.
		filename = strings.Join(SplitCommandArgs(filename), " ")
		v.Buf.Path = filename
		v.Buf.Name = filename

		v.DoActions("Save")
	}
}

// Find opens a prompt and searches forward for the input
func (v *View) Find(deprecatedUsePlugin ...bool) {
	searchStr := ""
	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
		searchStr = v.Cursor.GetSelection()
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	BeginSearch(searchStr)
}

// FindNext searches forwards for the last used search term
func (v *View) FindNext(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
		lastSearch = v.Cursor.GetSelection()
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	if lastSearch != "" {
		messenger.Message("Finding: " + lastSearch)
		Search(lastSearch, v, true)
	}
}

// FindPrevious searches backwards for the last used search term
func (v *View) FindPrevious(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[0], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, false)
}

// Undo undoes the last action
func (v *View) Undo(deprecatedUsePlugin ...bool) {
	v.Buf.Undo()
	messenger.Message("Undid action")
}

// Redo redoes the last action
func (v *View) Redo(deprecatedUsePlugin ...bool) {
	v.Buf.Redo()
	messenger.Message("Redid action")
}

// Copy the selection to the system clipboard
func (v *View) Copy(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection(), "clipboard")
		v.freshClip = true
		messenger.Message("Copied selection")
	}
}

// CutLine cuts the current line to the clipboard
func (v *View) CutLine(deprecatedUsePlugin ...bool) {
	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return
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
}

// Cut the selection to the system clipboard
func (v *View) Cut(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection(), "clipboard")
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		v.freshClip = true
		messenger.Message("Cut selection")
	}
}

// DuplicateLine duplicates the current line or selection
func (v *View) DuplicateLine(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		v.Buf.Insert(v.Cursor.CurSelection[1], v.Cursor.GetSelection())
	} else {
		v.Cursor.End()
		v.Buf.Insert(v.Cursor.Loc, "\n"+v.Buf.Line(v.Cursor.Y))
		v.Cursor.Right()
	}

	messenger.Message("Duplicated line")
}

// DeleteLine deletes the current line
func (v *View) DeleteLine(deprecatedUsePlugin ...bool) {
	v.Cursor.SelectLine()
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		messenger.Message("Deleted line")
	}
}

// MoveLinesUp moves up the current line or selected lines if any
func (v *View) MoveLinesUp(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		if v.Cursor.CurSelection[0].Y == 0 {
			messenger.Message("Can not move further up")
			return
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
			return
		}
		v.Buf.MoveLinesUp(
			v.Cursor.Loc.Y,
			v.Cursor.Loc.Y+1,
		)
		v.Cursor.UpN(1)
		messenger.Message("Moved up current line")
	}
	v.Buf.IsModified = true
}

// MoveLinesDown moves down the current line or selected lines if any
func (v *View) MoveLinesDown(deprecatedUsePlugin ...bool) {
	if v.Cursor.HasSelection() {
		if v.Cursor.CurSelection[1].Y >= len(v.Buf.lines) {
			messenger.Message("Can not move further down")
			return
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
			return
		}
		v.Buf.MoveLinesDown(
			v.Cursor.Loc.Y,
			v.Cursor.Loc.Y+1,
		)
		v.Cursor.DownN(1)
		messenger.Message("Moved down current line")
	}
	v.Buf.IsModified = true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (v *View) Paste(deprecatedUsePlugin ...bool) {
	clip, _ := clipboard.ReadAll("clipboard")
	v.paste(clip)
}

// PastePrimary pastes from the primary clipboard (only use on linux)
func (v *View) PastePrimary(deprecatedUsePlugin ...bool) {
	clip, _ := clipboard.ReadAll("primary")
	v.paste(clip)
}

// SelectAll selects the entire buffer
func (v *View) SelectAll(deprecatedUsePlugin ...bool) {
	v.Cursor.SetSelectionStart(v.Buf.Start())
	v.Cursor.SetSelectionEnd(v.Buf.End())
	// Put the cursor at the beginning
	v.Cursor.X = 0
	v.Cursor.Y = 0
}

// OpenFile opens a new file in the buffer
func (v *View) OpenFile(deprecatedUsePlugin ...bool) {
	if v.CanClose() {
		filename, canceled := messenger.Prompt("File to open: ", "Open", FileCompletion)
		if !canceled {
			// the filename might or might not be quoted, so unquote first then join the strings.
			filename = strings.Join(SplitCommandArgs(filename), " ")

			v.Open(filename)
		}
	}
}

// Start moves the viewport to the start of the buffer
func (v *View) Start(deprecatedUsePlugin ...bool) {
	v.Topline = 0
}

// End moves the viewport to the end of the buffer
func (v *View) End(deprecatedUsePlugin ...bool) {
	if v.height > v.Buf.NumLines {
		v.Topline = 0
	} else {
		v.Topline = v.Buf.NumLines - v.height
	}
}

// PageUp scrolls the view up a page
func (v *View) PageUp(deprecatedUsePlugin ...bool) {
	if v.Topline > v.height {
		v.ScrollUp(v.height)
	} else {
		v.Topline = 0
	}
}

// PageDown scrolls the view down a page
func (v *View) PageDown(deprecatedUsePlugin ...bool) {
	if v.Buf.NumLines-(v.Topline+v.height) > v.height {
		v.ScrollDown(v.height)
	} else if v.Buf.NumLines >= v.height {
		v.Topline = v.Buf.NumLines - v.height
	}
}

// CursorPageUp places the cursor a page up
func (v *View) CursorPageUp(deprecatedUsePlugin ...bool) {
	v.deselect(0)

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	}
	v.Cursor.UpN(v.height)
}

// CursorPageDown places the cursor a page up
func (v *View) CursorPageDown(deprecatedUsePlugin ...bool) {
	v.deselect(0)

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1]
		v.Cursor.ResetSelection()
	}
	v.Cursor.DownN(v.height)
}

// HalfPageUp scrolls the view up half a page
func (v *View) HalfPageUp(deprecatedUsePlugin ...bool) {
	if v.Topline > v.height/2 {
		v.ScrollUp(v.height / 2)
	} else {
		v.Topline = 0
	}
}

// HalfPageDown scrolls the view down half a page
func (v *View) HalfPageDown(deprecatedUsePlugin ...bool) {
	if v.Buf.NumLines-(v.Topline+v.height) > v.height/2 {
		v.ScrollDown(v.height / 2)
	} else {
		if v.Buf.NumLines >= v.height {
			v.Topline = v.Buf.NumLines - v.height
		}
	}
}

// ToggleRuler turns line numbers off and on
func (v *View) ToggleRuler(deprecatedUsePlugin ...bool) {
	if v.Buf.Settings["ruler"] == false {
		v.Buf.Settings["ruler"] = true
		messenger.Message("Enabled ruler")
	} else {
		v.Buf.Settings["ruler"] = false
		messenger.Message("Disabled ruler")
	}
}

// JumpLine jumps to a line and moves the view accordingly.
func (v *View) JumpLine(deprecatedUsePlugin ...bool) {
	// Prompt for line number
	linestring, canceled := messenger.Prompt("Jump to line # ", "LineNumber", NoCompletion)
	if canceled {
		return
	}
	lineint, err := strconv.Atoi(linestring)
	lineint = lineint - 1 // fix offset
	if err != nil {
		messenger.Error(err) // return errors
		return
	}
	// Move cursor and view if possible.
	if lineint < v.Buf.NumLines && lineint >= 0 {
		v.Cursor.X = 0
		v.Cursor.Y = lineint
	} else {
		messenger.Error("Only ", v.Buf.NumLines, " lines to jump")
	}
}

// ClearStatus clears the messenger bar
func (v *View) ClearStatus(deprecatedUsePlugin ...bool) {
	messenger.Message("")
}

// ToggleHelp toggles the help screen
func (v *View) ToggleHelp(deprecatedUsePlugin ...bool) {
	if v.Type != vtHelp {
		// Open the default help
		v.openHelp("help")
	} else {
		v.DoActions("Quit")
	}
}

// ShellMode opens a terminal to run a shell command
func (v *View) ShellMode(deprecatedUsePlugin ...bool) {
	input, canceled := messenger.Prompt("$ ", "Shell", NoCompletion)
	if !canceled {
		// The true here is for openTerm to make the command interactive
		HandleShellCommand(input, true, true)
	}
}

// CommandMode lets the user enter a command
func (v *View) CommandMode(deprecatedUsePlugin ...bool) {
	input, canceled := messenger.Prompt("> ", "Command", CommandCompletion)
	if !canceled {
		HandleCommand(input)
	}
}

// Escape leaves current mode / quits the editor
func (v *View) Escape(deprecatedUsePlugin ...bool) {
	// check if user is searching, or the last search is still active
	if searching || lastSearch != "" {
		ExitSearch(v)
		return
	}
	// check if a prompt is shown, hide it and don't quit
	if messenger.hasPrompt {
		messenger.Reset() // FIXME
		return
	}
	v.Quit()
}

// Quit quits the editor
// This behavior needs to be changed and should really only quit the editor if this
// is the last view
// However, since micro only supports one view for now, it doesn't really matter
func (v *View) Quit(deprecatedUsePlugin ...bool) {
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
}

// QuitAll quits the whole editor; all splits and tabs
func (v *View) QuitAll(deprecatedUsePlugin ...bool) {
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
}

// AddTab adds a new tab with an empty buffer
func (v *View) AddTab(deprecatedUsePlugin ...bool) {
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
}

// PreviousTab switches to the previous tab in the tab list
func (v *View) PreviousTab(deprecatedUsePlugin ...bool) {
	if curTab > 0 {
		curTab--
	} else if curTab == 0 {
		curTab = len(tabs) - 1
	}
}

// NextTab switches to the next tab in the tab list
func (v *View) NextTab(deprecatedUsePlugin ...bool) {
	if curTab < len(tabs)-1 {
		curTab++
	} else if curTab == len(tabs)-1 {
		curTab = 0
	}
}

// VSplitBinding opens an empty vertical split
func (v *View) VSplitBinding(deprecatedUsePlugin ...bool) {
	v.VSplit(NewBuffer([]byte{}, ""))
}

// HSplitBinding opens an empty horizontal split
func (v *View) HSplitBinding(deprecatedUsePlugin ...bool) {
	v.HSplit(NewBuffer([]byte{}, ""))
}

// Unsplit closes all splits in the current tab except the active one
func (v *View) Unsplit(deprecatedUsePlugin ...bool) {
	curView := tabs[curTab].curView
	for i := len(tabs[curTab].views) - 1; i >= 0; i-- {
		view := tabs[curTab].views[i]
		if view != nil && view.Num != curView {
			v.DoActions("Quit")
		}
	}
}

// NextSplit changes the view to the next split
func (v *View) NextSplit(deprecatedUsePlugin ...bool) {
	tab := tabs[curTab]
	if tab.curView < len(tab.views)-1 {
		tab.curView++
	} else {
		tab.curView = 0
	}
}

// PreviousSplit changes the view to the previous split
func (v *View) PreviousSplit(deprecatedUsePlugin ...bool) {
	tab := tabs[curTab]
	if tab.curView > 0 {
		tab.curView--
	} else {
		tab.curView = len(tab.views) - 1
	}
}

var curMacro []interface{}
var recordingMacro bool

func (v *View) ToggleMacro(deprecatedUsePlugin ...bool) {
	recordingMacro = !recordingMacro

	if recordingMacro {
		curMacro = []interface{}{}
		messenger.Message("Recording")
	} else {
		messenger.Message("Stopped recording")
	}
}

func (v *View) PlayMacro(deprecatedUsePlugin ...bool) {
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
}

// None is no action
func None() {
}
