package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/yuin/gopher-lua"
	"github.com/zyedidia/clipboard"
)

// PreActionCall executes the lua pre callback if possible
func PreActionCall(funcName string) bool {
	executeAction := true
	for _, pl := range loadedPlugins {
		ret, err := Call(pl+".pre"+funcName, nil)
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
func PostActionCall(funcName string) {
	for _, pl := range loadedPlugins {
		_, err := Call(pl+".on"+funcName, nil)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
	}
}

// CursorUp moves the cursor up
func (v *View) CursorUp() bool {
	if !PreActionCall("CursorUp") {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	}
	v.Cursor.Up()

	PostActionCall("CursorUp")
	return true
}

// CursorDown moves the cursor down
func (v *View) CursorDown() bool {
	if !PreActionCall("CursorDown") {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1]
		v.Cursor.ResetSelection()
	}
	v.Cursor.Down()

	PostActionCall("CursorDown")
	return true
}

// CursorLeft moves the cursor left
func (v *View) CursorLeft() bool {
	if !PreActionCall("CursorLeft") {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Left()
	}

	PostActionCall("CursorLeft")
	return true
}

// CursorRight moves the cursor right
func (v *View) CursorRight() bool {
	if !PreActionCall("CursorRight") {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1].Move(-1, v.Buf)
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Right()
	}

	PostActionCall("CursorRight")
	return true
}

// WordRight moves the cursor one word to the right
func (v *View) WordRight() bool {
	if !PreActionCall("WordRight") {
		return false
	}

	v.Cursor.WordRight()

	PostActionCall("WordRight")
	return true
}

// WordLeft moves the cursor one word to the left
func (v *View) WordLeft() bool {
	if !PreActionCall("WordLeft") {
		return false
	}

	v.Cursor.WordLeft()

	PostActionCall("WordLeft")
	return true
}

// SelectUp selects up one line
func (v *View) SelectUp() bool {
	if !PreActionCall("SelectUp") {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Up()
	v.Cursor.SelectTo(v.Cursor.Loc)

	PostActionCall("SelectUp")
	return true
}

// SelectDown selects down one line
func (v *View) SelectDown() bool {
	if !PreActionCall("SelectDown") {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Down()
	v.Cursor.SelectTo(v.Cursor.Loc)

	PostActionCall("SelectDown")
	return true
}

// SelectLeft selects the character to the left of the cursor
func (v *View) SelectLeft() bool {
	if !PreActionCall("SelectLeft") {
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

	PostActionCall("SelectLeft")
	return true
}

// SelectRight selects the character to the right of the cursor
func (v *View) SelectRight() bool {
	if !PreActionCall("SelectRight") {
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

	PostActionCall("SelectRight")
	return true
}

// SelectWordRight selects the word to the right of the cursor
func (v *View) SelectWordRight() bool {
	if !PreActionCall("SelectWordRight") {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordRight()
	v.Cursor.SelectTo(v.Cursor.Loc)

	PostActionCall("SelectWordRight")
	return true
}

// SelectWordLeft selects the word to the left of the cursor
func (v *View) SelectWordLeft() bool {
	if !PreActionCall("SelectWordLeft") {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordLeft()
	v.Cursor.SelectTo(v.Cursor.Loc)

	PostActionCall("SelectWordLeft")
	return true
}

// StartOfLine moves the cursor to the start of the line
func (v *View) StartOfLine() bool {
	if !PreActionCall("StartOfLine") {
		return false
	}

	v.Cursor.Start()

	PostActionCall("StartOfLine")
	return true
}

// EndOfLine moves the cursor to the end of the line
func (v *View) EndOfLine() bool {
	if !PreActionCall("EndOfLine") {
		return false
	}

	v.Cursor.End()

	PostActionCall("EndOfLine")
	return true
}

// SelectToStartOfLine selects to the start of the current line
func (v *View) SelectToStartOfLine() bool {
	if !PreActionCall("SelectToStartOfLine") {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Start()
	v.Cursor.SelectTo(v.Cursor.Loc)

	PostActionCall("SelectToStartOfLine")
	return true
}

// SelectToEndOfLine selects to the end of the current line
func (v *View) SelectToEndOfLine() bool {
	if !PreActionCall("SelectToEndOfLine") {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.End()
	v.Cursor.SelectTo(v.Cursor.Loc)

	PostActionCall("SelectToEndOfLine")
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (v *View) CursorStart() bool {
	if !PreActionCall("CursorStart") {
		return false
	}

	v.Cursor.X = 0
	v.Cursor.Y = 0

	PostActionCall("CursorStart")
	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (v *View) CursorEnd() bool {
	if !PreActionCall("CursorEnd") {
		return false
	}

	v.Cursor.Loc = v.Buf.End()

	PostActionCall("CursorEnd")
	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (v *View) SelectToStart() bool {
	if !PreActionCall("SelectToStart") {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorStart()
	v.Cursor.SelectTo(v.Buf.Start())

	PostActionCall("SelectToStart")
	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (v *View) SelectToEnd() bool {
	if !PreActionCall("SelectToEnd") {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorEnd()
	v.Cursor.SelectTo(v.Buf.End())

	PostActionCall("SelectToEnd")
	return true
}

// InsertSpace inserts a space
func (v *View) InsertSpace() bool {
	if !PreActionCall("InsertSpace") {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.Buf.Insert(v.Cursor.Loc, " ")
	v.Cursor.Right()

	PostActionCall("InsertSpace")
	return true
}

// InsertNewline inserts a newline plus possible some whitespace if autoindent is on
func (v *View) InsertNewline() bool {
	if !PreActionCall("InsertNewline") {
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

	if settings["autoindent"].(bool) {
		v.Buf.Insert(v.Cursor.Loc, ws)
		for i := 0; i < len(ws); i++ {
			v.Cursor.Right()
		}
	}
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()

	PostActionCall("InsertNewline")
	return true
}

// Backspace deletes the previous character
func (v *View) Backspace() bool {
	if !PreActionCall("Backspace") {
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
		tabSize := int(settings["tabsize"].(float64))
		if settings["tabstospaces"].(bool) && IsSpaces(lineStart) && len(lineStart) != 0 && len(lineStart)%tabSize == 0 {
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

	PostActionCall("Backspace")
	return true
}

// DeleteWordRight deletes the word to the right of the cursor
func (v *View) DeleteWordRight() bool {
	if !PreActionCall("DeleteWordRight") {
		return false
	}

	v.SelectWordRight()
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	PostActionCall("DeleteWordRight")
	return true
}

// DeleteWordLeft deletes the word to the left of the cursor
func (v *View) DeleteWordLeft() bool {
	if !PreActionCall("DeleteWordLeft") {
		return false
	}

	v.SelectWordLeft()
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	PostActionCall("DeleteWordLeft")
	return true
}

// Delete deletes the next character
func (v *View) Delete() bool {
	if !PreActionCall("Delete") {
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

	PostActionCall("Delete")
	return true
}

// IndentSelection indents the current selection
func (v *View) IndentSelection() bool {
	if !PreActionCall("IndentSelection") {
		return false
	}

	if v.Cursor.HasSelection() {
		start := v.Cursor.CurSelection[0].Y
		end := v.Cursor.CurSelection[1].Move(-1, v.Buf).Y
		endX := v.Cursor.CurSelection[1].Move(-1, v.Buf).X
		for i := start; i <= end; i++ {
			if settings["tabstospaces"].(bool) {
				tabsize := int(settings["tabsize"].(float64))
				v.Buf.Insert(Loc{0, i}, Spaces(tabsize))
				if i == start {
					if v.Cursor.CurSelection[0].X > 0 {
						v.Cursor.CurSelection[0] = v.Cursor.CurSelection[0].Move(tabsize, v.Buf)
					}
				}
				if i == end {
					v.Cursor.CurSelection[1] = Loc{endX + tabsize + 1, end}
				}
			} else {
				v.Buf.Insert(Loc{0, i}, "\t")
				if i == start {
					if v.Cursor.CurSelection[0].X > 0 {
						v.Cursor.CurSelection[0] = v.Cursor.CurSelection[0].Move(1, v.Buf)
					}
				}
				if i == end {
					v.Cursor.CurSelection[1] = Loc{endX + 2, end}
				}
			}
		}
		v.Cursor.Relocate()

		PostActionCall("IndentSelection")
		return true
	}
	return false
}

// OutdentSelection takes the current selection and moves it back one indent level
func (v *View) OutdentSelection() bool {
	if !PreActionCall("OutdentSelection") {
		return false
	}

	if v.Cursor.HasSelection() {
		start := v.Cursor.CurSelection[0].Y
		end := v.Cursor.CurSelection[1].Move(-1, v.Buf).Y
		endX := v.Cursor.CurSelection[1].Move(-1, v.Buf).X
		for i := start; i <= end; i++ {
			if len(GetLeadingWhitespace(v.Buf.Line(i))) > 0 {
				if settings["tabstospaces"].(bool) {
					tabsize := int(settings["tabsize"].(float64))
					for j := 0; j < tabsize; j++ {
						if len(GetLeadingWhitespace(v.Buf.Line(i))) == 0 {
							break
						}
						v.Buf.Remove(Loc{0, i}, Loc{1, i})
						if i == start {
							if v.Cursor.CurSelection[0].X > 0 {
								v.Cursor.CurSelection[0] = v.Cursor.CurSelection[0].Move(-1, v.Buf)
							}
						}
						if i == end {
							v.Cursor.CurSelection[1] = Loc{endX - j, end}
						}
					}
				} else {
					v.Buf.Remove(Loc{0, i}, Loc{1, i})
					if i == start {
						if v.Cursor.CurSelection[0].X > 0 {
							v.Cursor.CurSelection[0] = v.Cursor.CurSelection[0].Move(-1, v.Buf)
						}
					}
					if i == end {
						v.Cursor.CurSelection[1] = Loc{endX, end}
					}
				}
			}
		}
		v.Cursor.Relocate()

		PostActionCall("OutdentSelection")
		return true
	}
	return false
}

// InsertTab inserts a tab or spaces
func (v *View) InsertTab() bool {
	if !PreActionCall("InsertTab") {
		return false
	}

	if v.Cursor.HasSelection() {
		return false
	}
	// Insert a tab
	if settings["tabstospaces"].(bool) {
		tabSize := int(settings["tabsize"].(float64))
		v.Buf.Insert(v.Cursor.Loc, Spaces(tabSize))
		for i := 0; i < tabSize; i++ {
			v.Cursor.Right()
		}
	} else {
		v.Buf.Insert(v.Cursor.Loc, "\t")
		v.Cursor.Right()
	}

	PostActionCall("InsertTab")
	return true
}

// Save the buffer to disk
func (v *View) Save() bool {
	if !PreActionCall("Save") {
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

	PostActionCall("Save")
	return false
}

// Find opens a prompt and searches forward for the input
func (v *View) Find() bool {
	if !PreActionCall("Find") {
		return false
	}

	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	BeginSearch()

	PostActionCall("Find")
	return true
}

// FindNext searches forwards for the last used search term
func (v *View) FindNext() bool {
	if !PreActionCall("FindNext") {
		return false
	}

	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, true)

	PostActionCall("FindNext")
	return true
}

// FindPrevious searches backwards for the last used search term
func (v *View) FindPrevious() bool {
	if !PreActionCall("FindPrevious") {
		return false
	}

	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[0], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, false)

	PostActionCall("FindPrevious")
	return true
}

// Undo undoes the last action
func (v *View) Undo() bool {
	if !PreActionCall("Undo") {
		return false
	}

	v.Buf.Undo()
	messenger.Message("Undid action")

	PostActionCall("Undo")
	return true
}

// Redo redoes the last action
func (v *View) Redo() bool {
	if !PreActionCall("Redo") {
		return false
	}

	v.Buf.Redo()
	messenger.Message("Redid action")

	PostActionCall("Redo")
	return true
}

// Copy the selection to the system clipboard
func (v *View) Copy() bool {
	if !PreActionCall("Copy") {
		return false
	}

	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection())
		v.freshClip = true
		messenger.Message("Copied selection")
	}

	PostActionCall("Copy")
	return true
}

// CutLine cuts the current line to the clipboard
func (v *View) CutLine() bool {
	if !PreActionCall("CutLine") {
		return false
	}

	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	if v.freshClip == true {
		if v.Cursor.HasSelection() {
			if clip, err := clipboard.ReadAll(); err != nil {
				messenger.Error(err)
			} else {
				clipboard.WriteAll(clip + v.Cursor.GetSelection())
			}
		}
	} else if time.Since(v.lastCutTime)/time.Second > 10*time.Second || v.freshClip == false {
		v.Copy()
	}
	v.freshClip = true
	v.lastCutTime = time.Now()
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()
	messenger.Message("Cut line")

	PostActionCall("CutLine")
	return true
}

// Cut the selection to the system clipboard
func (v *View) Cut() bool {
	if !PreActionCall("Cut") {
		return false
	}

	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection())
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		v.freshClip = true
		messenger.Message("Cut selection")

		PostActionCall("Cut")
		return true
	}

	return false
}

// DuplicateLine duplicates the current line
func (v *View) DuplicateLine() bool {
	if !PreActionCall("DuplicateLine") {
		return false
	}

	v.Cursor.End()
	v.Buf.Insert(v.Cursor.Loc, "\n"+v.Buf.Line(v.Cursor.Y))
	v.Cursor.Right()
	messenger.Message("Duplicated line")

	PostActionCall("DuplicateLine")
	return true
}

// DeleteLine deletes the current line
func (v *View) DeleteLine() bool {
	if !PreActionCall("DeleteLine") {
		return false
	}

	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()
	messenger.Message("Deleted line")

	PostActionCall("DeleteLine")
	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (v *View) Paste() bool {
	if !PreActionCall("Paste") {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	clip, _ := clipboard.ReadAll()
	v.Buf.Insert(v.Cursor.Loc, clip)
	v.Cursor.Loc = v.Cursor.Loc.Move(Count(clip), v.Buf)
	v.freshClip = false
	messenger.Message("Pasted clipboard")

	PostActionCall("Paste")
	return true
}

// SelectAll selects the entire buffer
func (v *View) SelectAll() bool {
	if !PreActionCall("SelectAll") {
		return false
	}

	v.Cursor.CurSelection[0] = v.Buf.Start()
	v.Cursor.CurSelection[1] = v.Buf.End()
	// Put the cursor at the beginning
	v.Cursor.X = 0
	v.Cursor.Y = 0

	PostActionCall("SelectAll")
	return true
}

// OpenFile opens a new file in the buffer
func (v *View) OpenFile() bool {
	if !PreActionCall("OpenFile") {
		return false
	}

	if v.CanClose("Continue? (yes, no, save) ") {
		filename, canceled := messenger.Prompt("File to open: ", "Open", FileCompletion)
		if canceled {
			return false
		}
		home, _ := homedir.Dir()
		filename = strings.Replace(filename, "~", home, 1)
		file, err := ioutil.ReadFile(filename)

		var buf *Buffer
		if err != nil {
			// File does not exist -- create an empty buffer with that name
			buf = NewBuffer([]byte{}, filename)
		} else {
			buf = NewBuffer(file, filename)
		}
		v.OpenBuffer(buf)

		PostActionCall("OpenFile")
		return true
	}
	return false
}

// Start moves the viewport to the start of the buffer
func (v *View) Start() bool {
	if !PreActionCall("Start") {
		return false
	}

	v.Topline = 0

	PostActionCall("Start")
	return false
}

// End moves the viewport to the end of the buffer
func (v *View) End() bool {
	if !PreActionCall("End") {
		return false
	}

	if v.height > v.Buf.NumLines {
		v.Topline = 0
	} else {
		v.Topline = v.Buf.NumLines - v.height
	}

	PostActionCall("End")
	return false
}

// PageUp scrolls the view up a page
func (v *View) PageUp() bool {
	if !PreActionCall("PageUp") {
		return false
	}

	if v.Topline > v.height {
		v.ScrollUp(v.height)
	} else {
		v.Topline = 0
	}

	PostActionCall("PageUp")
	return false
}

// PageDown scrolls the view down a page
func (v *View) PageDown() bool {
	if !PreActionCall("PageDown") {
		return false
	}

	if v.Buf.NumLines-(v.Topline+v.height) > v.height {
		v.ScrollDown(v.height)
	} else if v.Buf.NumLines >= v.height {
		v.Topline = v.Buf.NumLines - v.height
	}

	PostActionCall("PageDown")
	return false
}

// CursorPageUp places the cursor a page up
func (v *View) CursorPageUp() bool {
	if !PreActionCall("CursorPageUp") {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	}
	v.Cursor.UpN(v.height)

	PostActionCall("CursorPageUp")
	return true
}

// CursorPageDown places the cursor a page up
func (v *View) CursorPageDown() bool {
	if !PreActionCall("CursorPageDown") {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1]
		v.Cursor.ResetSelection()
	}
	v.Cursor.DownN(v.height)

	PostActionCall("CursorPageDown")
	return true
}

// HalfPageUp scrolls the view up half a page
func (v *View) HalfPageUp() bool {
	if !PreActionCall("HalfPageUp") {
		return false
	}

	if v.Topline > v.height/2 {
		v.ScrollUp(v.height / 2)
	} else {
		v.Topline = 0
	}

	PostActionCall("HalfPageUp")
	return false
}

// HalfPageDown scrolls the view down half a page
func (v *View) HalfPageDown() bool {
	if !PreActionCall("HalfPageDown") {
		return false
	}

	if v.Buf.NumLines-(v.Topline+v.height) > v.height/2 {
		v.ScrollDown(v.height / 2)
	} else {
		if v.Buf.NumLines >= v.height {
			v.Topline = v.Buf.NumLines - v.height
		}
	}

	PostActionCall("HalfPageDown")
	return false
}

// ToggleRuler turns line numbers off and on
func (v *View) ToggleRuler() bool {
	if !PreActionCall("ToggleRuler") {
		return false
	}

	if settings["ruler"] == false {
		settings["ruler"] = true
		messenger.Message("Enabled ruler")
	} else {
		settings["ruler"] = false
		messenger.Message("Disabled ruler")
	}

	PostActionCall("ToggleRuler")
	return false
}

// JumpLine jumps to a line and moves the view accordingly.
func (v *View) JumpLine() bool {
	if !PreActionCall("JumpLine") {
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

		PostActionCall("JumpLine")
		return true
	}
	messenger.Error("Only ", v.Buf.NumLines, " lines to jump")
	return false
}

// ClearStatus clears the messenger bar
func (v *View) ClearStatus() bool {
	if !PreActionCall("ClearStatus") {
		return false
	}

	messenger.Message("")

	PostActionCall("ClearStatus")
	return false
}

// ToggleHelp toggles the help screen
func (v *View) ToggleHelp() bool {
	if !PreActionCall("ToggleHelp") {
		return false
	}

	if !v.Help {
		// Open the default help
		v.openHelp("help")
	} else {
		v.Quit()
	}

	PostActionCall("ToggleHelp")
	return true
}

// ShellMode opens a terminal to run a shell command
func (v *View) ShellMode() bool {
	if !PreActionCall("ShellMode") {
		return false
	}

	input, canceled := messenger.Prompt("$ ", "Shell", NoCompletion)
	if !canceled {
		// The true here is for openTerm to make the command interactive
		HandleShellCommand(input, true)
		PostActionCall("ShellMode")
	}
	return false
}

// CommandMode lets the user enter a command
func (v *View) CommandMode() bool {
	if !PreActionCall("CommandMode") {
		return false
	}

	input, canceled := messenger.Prompt("> ", "Command", FileCompletion)
	if !canceled {
		HandleCommand(input)
		PostActionCall("CommandMode")
	}

	return false
}

// Quit quits the editor
// This behavior needs to be changed and should really only quit the editor if this
// is the last view
// However, since micro only supports one view for now, it doesn't really matter
func (v *View) Quit() bool {
	if !PreActionCall("Quit") {
		return false
	}

	// Make sure not to quit if there are unsaved changes
	if v.CanClose("Quit anyway? (yes, no, save) ") {
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
			PostActionCall("Quit")

			screen.Fini()
			os.Exit(0)
		}
	}

	PostActionCall("Quit")
	return false
}

// AddTab adds a new tab with an empty buffer
func (v *View) AddTab() bool {
	if !PreActionCall("AddTab") {
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

	PostActionCall("AddTab")
	return true
}

// PreviousTab switches to the previous tab in the tab list
func (v *View) PreviousTab() bool {
	if !PreActionCall("PreviousTab") {
		return false
	}

	if curTab > 0 {
		curTab--
	} else if curTab == 0 {
		curTab = len(tabs) - 1
	}

	PostActionCall("PreviousTab")
	return false
}

// NextTab switches to the next tab in the tab list
func (v *View) NextTab() bool {
	if !PreActionCall("NextTab") {
		return false
	}

	if curTab < len(tabs)-1 {
		curTab++
	} else if curTab == len(tabs)-1 {
		curTab = 0
	}

	PostActionCall("NextTab")
	return false
}

// NextSplit changes the view to the next split
func (v *View) NextSplit() bool {
	if !PreActionCall("NextSplit") {
		return false
	}

	tab := tabs[curTab]
	if tab.curView < len(tab.views)-1 {
		tab.curView++
	} else {
		tab.curView = 0
	}

	PostActionCall("NextSplit")
	return false
}

// PreviousSplit changes the view to the previous split
func (v *View) PreviousSplit() bool {
	if !PreActionCall("PreviousSplit") {
		return false
	}

	tab := tabs[curTab]
	if tab.curView > 0 {
		tab.curView--
	} else {
		tab.curView = len(tab.views) - 1
	}

	PostActionCall("PreviousSplit")
	return false
}

// None is no action
func None() bool {
	if !PreActionCall("None") {
		return false
	}

	PostActionCall("None")
	return false
}
