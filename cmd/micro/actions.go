package main

import (
	"os"
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

func (v *View) deselect(index int) bool {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[index]
		v.Cursor.ResetSelection()
		return true
	}
	return false
}

// Center centers the view on the cursor
func (v *View) Center(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Center", v) {
		return false
	}

	v.Topline = v.Cursor.Y - v.height/2
	if v.Topline+v.height > v.Buf.NumLines {
		v.Topline = v.Buf.NumLines - v.height
	}
	if v.Topline < 0 {
		v.Topline = 0
	}

	if usePlugin {
		return PostActionCall("Center", v)
	}
	return true
}

// CursorUp moves the cursor up
func (v *View) CursorUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorUp", v) {
		return false
	}

	v.deselect(0)
	v.Cursor.Up()

	if usePlugin {
		return PostActionCall("CursorUp", v)
	}
	return true
}

// CursorDown moves the cursor down
func (v *View) CursorDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorDown", v) {
		return false
	}

	v.deselect(1)
	v.Cursor.Down()

	if usePlugin {
		return PostActionCall("CursorDown", v)
	}
	return true
}

// CursorLeft moves the cursor left
func (v *View) CursorLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorLeft", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Left()
	}

	if usePlugin {
		return PostActionCall("CursorLeft", v)
	}
	return true
}

// CursorRight moves the cursor right
func (v *View) CursorRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorRight", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1].Move(-1, v.Buf)
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Right()
	}

	if usePlugin {
		return PostActionCall("CursorRight", v)
	}
	return true
}

// WordRight moves the cursor one word to the right
func (v *View) WordRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("WordRight", v) {
		return false
	}

	v.Cursor.WordRight()

	if usePlugin {
		return PostActionCall("WordRight", v)
	}
	return true
}

// WordLeft moves the cursor one word to the left
func (v *View) WordLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("WordLeft", v) {
		return false
	}

	v.Cursor.WordLeft()

	if usePlugin {
		return PostActionCall("WordLeft", v)
	}
	return true
}

// SelectUp selects up one line
func (v *View) SelectUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectUp", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Up()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectUp", v)
	}
	return true
}

// SelectDown selects down one line
func (v *View) SelectDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectDown", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Down()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectDown", v)
	}
	return true
}

// SelectLeft selects the character to the left of the cursor
func (v *View) SelectLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectLeft", v) {
		return false
	}

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

	if usePlugin {
		return PostActionCall("SelectLeft", v)
	}
	return true
}

// SelectRight selects the character to the right of the cursor
func (v *View) SelectRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectRight", v) {
		return false
	}

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

	if usePlugin {
		return PostActionCall("SelectRight", v)
	}
	return true
}

// SelectWordRight selects the word to the right of the cursor
func (v *View) SelectWordRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectWordRight", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordRight()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectWordRight", v)
	}
	return true
}

// SelectWordLeft selects the word to the left of the cursor
func (v *View) SelectWordLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectWordLeft", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordLeft()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectWordLeft", v)
	}
	return true
}

// StartOfLine moves the cursor to the start of the line
func (v *View) StartOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("StartOfLine", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.Start()

	if usePlugin {
		return PostActionCall("StartOfLine", v)
	}
	return true
}

// EndOfLine moves the cursor to the end of the line
func (v *View) EndOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("EndOfLine", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.End()

	if usePlugin {
		return PostActionCall("EndOfLine", v)
	}
	return true
}

// SelectToStartOfLine selects to the start of the current line
func (v *View) SelectToStartOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToStartOfLine", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Start()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectToStartOfLine", v)
	}
	return true
}

// SelectToEndOfLine selects to the end of the current line
func (v *View) SelectToEndOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToEndOfLine", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.End()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectToEndOfLine", v)
	}
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (v *View) CursorStart(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorStart", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.X = 0
	v.Cursor.Y = 0

	if usePlugin {
		return PostActionCall("CursorStart", v)
	}
	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (v *View) CursorEnd(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorEnd", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.Loc = v.Buf.End()

	if usePlugin {
		return PostActionCall("CursorEnd", v)
	}
	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (v *View) SelectToStart(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToStart", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorStart(false)
	v.Cursor.SelectTo(v.Buf.Start())

	if usePlugin {
		return PostActionCall("SelectToStart", v)
	}
	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (v *View) SelectToEnd(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToEnd", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorEnd(false)
	v.Cursor.SelectTo(v.Buf.End())

	if usePlugin {
		return PostActionCall("SelectToEnd", v)
	}
	return true
}

// InsertSpace inserts a space
func (v *View) InsertSpace(usePlugin bool) bool {
	if usePlugin && !PreActionCall("InsertSpace", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.Buf.Insert(v.Cursor.Loc, " ")
	v.Cursor.Right()

	if usePlugin {
		return PostActionCall("InsertSpace", v)
	}
	return true
}

// InsertNewline inserts a newline plus possible some whitespace if autoindent is on
func (v *View) InsertNewline(usePlugin bool) bool {
	if usePlugin && !PreActionCall("InsertNewline", v) {
		return false
	}

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

	if usePlugin {
		return PostActionCall("InsertNewline", v)
	}
	return true
}

// Backspace deletes the previous character
func (v *View) Backspace(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Backspace", v) {
		return false
	}

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
		// whitespace at the start of the line, we should delete as if its a
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

	if usePlugin {
		return PostActionCall("Backspace", v)
	}
	return true
}

// DeleteWordRight deletes the word to the right of the cursor
func (v *View) DeleteWordRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteWordRight", v) {
		return false
	}

	v.SelectWordRight(false)
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	if usePlugin {
		return PostActionCall("DeleteWordRight", v)
	}
	return true
}

// DeleteWordLeft deletes the word to the left of the cursor
func (v *View) DeleteWordLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteWordLeft", v) {
		return false
	}

	v.SelectWordLeft(false)
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	if usePlugin {
		return PostActionCall("DeleteWordLeft", v)
	}
	return true
}

// Delete deletes the next character
func (v *View) Delete(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Delete", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else {
		loc := v.Cursor.Loc
		if loc.LessThan(v.Buf.End()) {
			v.Buf.Remove(loc, loc.Move(1, v.Buf))
		}
	}

	if usePlugin {
		return PostActionCall("Delete", v)
	}
	return true
}

// IndentSelection indents the current selection
func (v *View) IndentSelection(usePlugin bool) bool {
	if usePlugin && !PreActionCall("IndentSelection", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		start := v.Cursor.CurSelection[0].Y
		end := v.Cursor.CurSelection[1].Move(-1, v.Buf).Y
		endX := v.Cursor.CurSelection[1].Move(-1, v.Buf).X
		for i := start; i <= end; i++ {
			if v.Buf.Settings["tabstospaces"].(bool) {
				tabsize := int(v.Buf.Settings["tabsize"].(float64))
				v.Buf.Insert(Loc{0, i}, Spaces(tabsize))
				if i == start {
					if v.Cursor.CurSelection[0].X > 0 {
						v.Cursor.SetSelectionStart(v.Cursor.CurSelection[0].Move(tabsize, v.Buf))
					}
				}
				if i == end {
					v.Cursor.SetSelectionEnd(Loc{endX + tabsize + 1, end})
				}
			} else {
				v.Buf.Insert(Loc{0, i}, "\t")
				if i == start {
					if v.Cursor.CurSelection[0].X > 0 {
						v.Cursor.SetSelectionStart(v.Cursor.CurSelection[0].Move(1, v.Buf))
					}
				}
				if i == end {
					v.Cursor.SetSelectionEnd(Loc{endX + 2, end})
				}
			}
		}
		v.Cursor.Relocate()

		if usePlugin {
			return PostActionCall("IndentSelection", v)
		}
		return true
	}
	return false
}

// OutdentSelection takes the current selection and moves it back one indent level
func (v *View) OutdentSelection(usePlugin bool) bool {
	if usePlugin && !PreActionCall("OutdentSelection", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		start := v.Cursor.CurSelection[0].Y
		end := v.Cursor.CurSelection[1].Move(-1, v.Buf).Y
		endX := v.Cursor.CurSelection[1].Move(-1, v.Buf).X
		for i := start; i <= end; i++ {
			if len(GetLeadingWhitespace(v.Buf.Line(i))) > 0 {
				if v.Buf.Settings["tabstospaces"].(bool) {
					tabsize := int(v.Buf.Settings["tabsize"].(float64))
					for j := 0; j < tabsize; j++ {
						if len(GetLeadingWhitespace(v.Buf.Line(i))) == 0 {
							break
						}
						v.Buf.Remove(Loc{0, i}, Loc{1, i})
						if i == start {
							if v.Cursor.CurSelection[0].X > 0 {
								v.Cursor.SetSelectionStart(v.Cursor.CurSelection[0].Move(-1, v.Buf))
							}
						}
						if i == end {
							v.Cursor.SetSelectionEnd(Loc{endX - j, end})
						}
					}
				} else {
					v.Buf.Remove(Loc{0, i}, Loc{1, i})
					if i == start {
						if v.Cursor.CurSelection[0].X > 0 {
							v.Cursor.SetSelectionStart(v.Cursor.CurSelection[0].Move(-1, v.Buf))
						}
					}
					if i == end {
						v.Cursor.SetSelectionEnd(Loc{endX, end})
					}
				}
			}
		}
		v.Cursor.Relocate()

		if usePlugin {
			return PostActionCall("OutdentSelection", v)
		}
		return true
	}
	return false
}

// InsertTab inserts a tab or spaces
func (v *View) InsertTab(usePlugin bool) bool {
	if usePlugin && !PreActionCall("InsertTab", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		return false
	}
	// Insert a tab
	if v.Buf.Settings["tabstospaces"].(bool) {
		tabSize := int(v.Buf.Settings["tabsize"].(float64))
		if remainder := v.Cursor.Loc.X % tabSize; remainder != 0 {
			tabSize = tabSize - remainder
		}
		v.Buf.Insert(v.Cursor.Loc, Spaces(tabSize))
		for i := 0; i < tabSize; i++ {
			v.Cursor.Right()
		}
	} else {
		v.Buf.Insert(v.Cursor.Loc, "\t")
		v.Cursor.Right()
	}

	if usePlugin {
		return PostActionCall("InsertTab", v)
	}
	return true
}

// Save the buffer to disk
func (v *View) Save(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Save", v) {
		return false
	}

	if v.Help {
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

	if usePlugin {
		return PostActionCall("Save", v)
	}
	return false
}

// Find opens a prompt and searches forward for the input
func (v *View) Find(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Find", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	BeginSearch()

	if usePlugin {
		return PostActionCall("Find", v)
	}
	return true
}

// FindNext searches forwards for the last used search term
func (v *View) FindNext(usePlugin bool) bool {
	if usePlugin && !PreActionCall("FindNext", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, true)

	if usePlugin {
		return PostActionCall("FindNext", v)
	}
	return true
}

// FindPrevious searches backwards for the last used search term
func (v *View) FindPrevious(usePlugin bool) bool {
	if usePlugin && !PreActionCall("FindPrevious", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[0], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, false)

	if usePlugin {
		return PostActionCall("FindPrevious", v)
	}
	return true
}

// Undo undoes the last action
func (v *View) Undo(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Undo", v) {
		return false
	}

	v.Buf.Undo()
	messenger.Message("Undid action")

	if usePlugin {
		return PostActionCall("Undo", v)
	}
	return true
}

// Redo redoes the last action
func (v *View) Redo(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Redo", v) {
		return false
	}

	v.Buf.Redo()
	messenger.Message("Redid action")

	if usePlugin {
		return PostActionCall("Redo", v)
	}
	return true
}

// Copy the selection to the system clipboard
func (v *View) Copy(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Copy", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection(), "clipboard")
		v.freshClip = true
		messenger.Message("Copied selection")
	}

	if usePlugin {
		return PostActionCall("Copy", v)
	}
	return true
}

// CutLine cuts the current line to the clipboard
func (v *View) CutLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CutLine", v) {
		return false
	}

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
		v.Copy(true)
	}
	v.freshClip = true
	v.lastCutTime = time.Now()
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()
	messenger.Message("Cut line")

	if usePlugin {
		return PostActionCall("CutLine", v)
	}
	return true
}

// Cut the selection to the system clipboard
func (v *View) Cut(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Cut", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection(), "clipboard")
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		v.freshClip = true
		messenger.Message("Cut selection")

		if usePlugin {
			return PostActionCall("Cut", v)
		}
		return true
	}

	return false
}

// DuplicateLine duplicates the current line
func (v *View) DuplicateLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DuplicateLine", v) {
		return false
	}

	v.Cursor.End()
	v.Buf.Insert(v.Cursor.Loc, "\n"+v.Buf.Line(v.Cursor.Y))
	v.Cursor.Right()
	messenger.Message("Duplicated line")

	if usePlugin {
		return PostActionCall("DuplicateLine", v)
	}
	return true
}

// DeleteLine deletes the current line
func (v *View) DeleteLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteLine", v) {
		return false
	}

	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()
	messenger.Message("Deleted line")

	if usePlugin {
		return PostActionCall("DeleteLine", v)
	}
	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (v *View) Paste(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Paste", v) {
		return false
	}

	clip, _ := clipboard.ReadAll("clipboard")
	v.paste(clip)

	if usePlugin {
		return PostActionCall("Paste", v)
	}
	return true
}

// PastePrimary pastes from the primary clipboard (only use on linux)
func (v *View) PastePrimary(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Paste", v) {
		return false
	}

	clip, _ := clipboard.ReadAll("primary")
	v.paste(clip)

	if usePlugin {
		return PostActionCall("Paste", v)
	}
	return true
}

// SelectAll selects the entire buffer
func (v *View) SelectAll(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectAll", v) {
		return false
	}

	v.Cursor.SetSelectionStart(v.Buf.Start())
	v.Cursor.SetSelectionEnd(v.Buf.End())
	// Put the cursor at the beginning
	v.Cursor.X = 0
	v.Cursor.Y = 0

	if usePlugin {
		return PostActionCall("SelectAll", v)
	}
	return true
}

// OpenFile opens a new file in the buffer
func (v *View) OpenFile(usePlugin bool) bool {
	if usePlugin && !PreActionCall("OpenFile", v) {
		return false
	}

	if v.CanClose() {
		filename, canceled := messenger.Prompt("File to open: ", "Open", FileCompletion)
		if canceled {
			return false
		}
		// the filename might or might not be quoted, so unquote first then join the strings.
		filename = strings.Join(SplitCommandArgs(filename), " ")

		v.Open(filename)

		if usePlugin {
			return PostActionCall("OpenFile", v)
		}
		return true
	}
	return false
}

// Start moves the viewport to the start of the buffer
func (v *View) Start(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Start", v) {
		return false
	}

	v.Topline = 0

	if usePlugin {
		return PostActionCall("Start", v)
	}
	return false
}

// End moves the viewport to the end of the buffer
func (v *View) End(usePlugin bool) bool {
	if usePlugin && !PreActionCall("End", v) {
		return false
	}

	if v.height > v.Buf.NumLines {
		v.Topline = 0
	} else {
		v.Topline = v.Buf.NumLines - v.height
	}

	if usePlugin {
		return PostActionCall("End", v)
	}
	return false
}

// PageUp scrolls the view up a page
func (v *View) PageUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("PageUp", v) {
		return false
	}

	if v.Topline > v.height {
		v.ScrollUp(v.height)
	} else {
		v.Topline = 0
	}

	if usePlugin {
		return PostActionCall("PageUp", v)
	}
	return false
}

// PageDown scrolls the view down a page
func (v *View) PageDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("PageDown", v) {
		return false
	}

	if v.Buf.NumLines-(v.Topline+v.height) > v.height {
		v.ScrollDown(v.height)
	} else if v.Buf.NumLines >= v.height {
		v.Topline = v.Buf.NumLines - v.height
	}

	if usePlugin {
		return PostActionCall("PageDown", v)
	}
	return false
}

// CursorPageUp places the cursor a page up
func (v *View) CursorPageUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorPageUp", v) {
		return false
	}

	v.deselect(0)

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	}
	v.Cursor.UpN(v.height)

	if usePlugin {
		return PostActionCall("CursorPageUp", v)
	}
	return true
}

// CursorPageDown places the cursor a page up
func (v *View) CursorPageDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorPageDown", v) {
		return false
	}

	v.deselect(0)

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1]
		v.Cursor.ResetSelection()
	}
	v.Cursor.DownN(v.height)

	if usePlugin {
		return PostActionCall("CursorPageDown", v)
	}
	return true
}

// HalfPageUp scrolls the view up half a page
func (v *View) HalfPageUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("HalfPageUp", v) {
		return false
	}

	if v.Topline > v.height/2 {
		v.ScrollUp(v.height / 2)
	} else {
		v.Topline = 0
	}

	if usePlugin {
		return PostActionCall("HalfPageUp", v)
	}
	return false
}

// HalfPageDown scrolls the view down half a page
func (v *View) HalfPageDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("HalfPageDown", v) {
		return false
	}

	if v.Buf.NumLines-(v.Topline+v.height) > v.height/2 {
		v.ScrollDown(v.height / 2)
	} else {
		if v.Buf.NumLines >= v.height {
			v.Topline = v.Buf.NumLines - v.height
		}
	}

	if usePlugin {
		return PostActionCall("HalfPageDown", v)
	}
	return false
}

// ToggleRuler turns line numbers off and on
func (v *View) ToggleRuler(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ToggleRuler", v) {
		return false
	}

	if v.Buf.Settings["ruler"] == false {
		v.Buf.Settings["ruler"] = true
		messenger.Message("Enabled ruler")
	} else {
		v.Buf.Settings["ruler"] = false
		messenger.Message("Disabled ruler")
	}

	if usePlugin {
		return PostActionCall("ToggleRuler", v)
	}
	return false
}

// JumpLine jumps to a line and moves the view accordingly.
func (v *View) JumpLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("JumpLine", v) {
		return false
	}

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

		if usePlugin {
			return PostActionCall("JumpLine", v)
		}
		return true
	}
	messenger.Error("Only ", v.Buf.NumLines, " lines to jump")
	return false
}

// ClearStatus clears the messenger bar
func (v *View) ClearStatus(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ClearStatus", v) {
		return false
	}

	messenger.Message("")

	if usePlugin {
		return PostActionCall("ClearStatus", v)
	}
	return false
}

// ToggleHelp toggles the help screen
func (v *View) ToggleHelp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ToggleHelp", v) {
		return false
	}

	if !v.Help {
		// Open the default help
		v.openHelp("help")
	} else {
		v.Quit(true)
	}

	if usePlugin {
		return PostActionCall("ToggleHelp", v)
	}
	return true
}

// ShellMode opens a terminal to run a shell command
func (v *View) ShellMode(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ShellMode", v) {
		return false
	}

	input, canceled := messenger.Prompt("$ ", "Shell", NoCompletion)
	if !canceled {
		// The true here is for openTerm to make the command interactive
		HandleShellCommand(input, true, true)
		if usePlugin {
			return PostActionCall("ShellMode", v)
		}
	}
	return false
}

// CommandMode lets the user enter a command
func (v *View) CommandMode(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CommandMode", v) {
		return false
	}

	input, canceled := messenger.Prompt("> ", "Command", CommandCompletion)
	if !canceled {
		HandleCommand(input)
		if usePlugin {
			return PostActionCall("CommandMode", v)
		}
	}

	return false
}

// Quit quits the editor
// This behavior needs to be changed and should really only quit the editor if this
// is the last view
// However, since micro only supports one view for now, it doesn't really matter
func (v *View) Quit(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Quit", v) {
		return false
	}

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
			if usePlugin {
				PostActionCall("Quit", v)
			}

			screen.Fini()
			os.Exit(0)
		}
	}

	if usePlugin {
		return PostActionCall("Quit", v)
	}
	return false
}

// QuitAll quits the whole editor; all splits and tabs
func (v *View) QuitAll(usePlugin bool) bool {
	if usePlugin && !PreActionCall("QuitAll", v) {
		return false
	}

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

		if usePlugin {
			PostActionCall("QuitAll", v)
		}

		screen.Fini()
		os.Exit(0)
	}

	return false
}

// AddTab adds a new tab with an empty buffer
func (v *View) AddTab(usePlugin bool) bool {
	if usePlugin && !PreActionCall("AddTab", v) {
		return false
	}

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

	if usePlugin {
		return PostActionCall("AddTab", v)
	}
	return true
}

// PreviousTab switches to the previous tab in the tab list
func (v *View) PreviousTab(usePlugin bool) bool {
	if usePlugin && !PreActionCall("PreviousTab", v) {
		return false
	}

	if curTab > 0 {
		curTab--
	} else if curTab == 0 {
		curTab = len(tabs) - 1
	}

	if usePlugin {
		return PostActionCall("PreviousTab", v)
	}
	return false
}

// NextTab switches to the next tab in the tab list
func (v *View) NextTab(usePlugin bool) bool {
	if usePlugin && !PreActionCall("NextTab", v) {
		return false
	}

	if curTab < len(tabs)-1 {
		curTab++
	} else if curTab == len(tabs)-1 {
		curTab = 0
	}

	if usePlugin {
		return PostActionCall("NextTab", v)
	}
	return false
}

// NextSplit changes the view to the next split
func (v *View) NextSplit(usePlugin bool) bool {
	if usePlugin && !PreActionCall("NextSplit", v) {
		return false
	}

	tab := tabs[curTab]
	if tab.curView < len(tab.views)-1 {
		tab.curView++
	} else {
		tab.curView = 0
	}

	if usePlugin {
		return PostActionCall("NextSplit", v)
	}
	return false
}

// PreviousSplit changes the view to the previous split
func (v *View) PreviousSplit(usePlugin bool) bool {
	if usePlugin && !PreActionCall("PreviousSplit", v) {
		return false
	}

	tab := tabs[curTab]
	if tab.curView > 0 {
		tab.curView--
	} else {
		tab.curView = len(tab.views) - 1
	}

	if usePlugin {
		return PostActionCall("PreviousSplit", v)
	}
	return false
}

var curMacro []interface{}
var recordingMacro bool

func (v *View) ToggleMacro(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ToggleMacro", v) {
		return false
	}

	recordingMacro = !recordingMacro

	if recordingMacro {
		curMacro = []interface{}{}
		messenger.Message("Recording")
	} else {
		messenger.Message("Stopped recording")
	}

	if usePlugin {
		return PostActionCall("ToggleMacro", v)
	}
	return true
}

func (v *View) PlayMacro(usePlugin bool) bool {
	if usePlugin && !PreActionCall("PlayMacro", v) {
		return false
	}

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
		case func(*View, bool) bool:
			t(v, true)
		}
	}

	if usePlugin {
		return PostActionCall("PlayMacro", v)
	}
	return true
}

// None is no action
func None() bool {
	return false
}
