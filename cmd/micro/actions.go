package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/mitchellh/go-homedir"
)

// CursorUp moves the cursor up
func (v *View) CursorUp() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	}
	v.Cursor.Up()
	return true
}

// CursorDown moves the cursor down
func (v *View) CursorDown() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1]
		v.Cursor.ResetSelection()
	}
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
	v.Cursor.Start()
	return true
}

// EndOfLine moves the cursor to the end of the line
func (v *View) EndOfLine() bool {
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
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.End()
	v.Cursor.SelectTo(v.Cursor.Loc)
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (v *View) CursorStart() bool {
	v.Cursor.X = 0
	v.Cursor.Y = 0
	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (v *View) CursorEnd() bool {
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

// InsertEnter inserts a newline plus possible some whitespace if autoindent is on
func (v *View) InsertEnter() bool {
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
	return true
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

// InsertTab inserts a tab or spaces
func (v *View) InsertTab() bool {
	// Insert a tab
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
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
	return true
}

// Save the buffer to disk
func (v *View) Save() bool {
	if v.helpOpen {
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
	return false
}

// Find opens a prompt and searches forward for the input
func (v *View) Find() bool {
	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	BeginSearch()
	return true
}

// FindNext searches forwards for the last used search term
func (v *View) FindNext() bool {
	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
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
		clipboard.WriteAll(v.Cursor.GetSelection())
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
	return true
}

// Cut the selection to the system clipboard
func (v *View) Cut() bool {
	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection())
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		v.freshClip = true
		messenger.Message("Cut selection")
	}
	return true
}

// DuplicateLine duplicates the current line
func (v *View) DuplicateLine() bool {
	v.Cursor.End()
	v.Buf.Insert(v.Cursor.Loc, "\n"+v.Buf.Line(v.Cursor.Y))
	v.Cursor.Right()
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

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (v *View) Paste() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	clip, _ := clipboard.ReadAll()
	v.Buf.Insert(v.Cursor.Loc, clip)
	v.Cursor.Loc = v.Cursor.Loc.Move(Count(clip), v.Buf)
	v.freshClip = false
	messenger.Message("Pasted clipboard")
	return true
}

// SelectAll selects the entire buffer
func (v *View) SelectAll() bool {
	v.Cursor.CurSelection[0] = v.Buf.Start()
	v.Cursor.CurSelection[1] = v.Buf.End()
	// Put the cursor at the beginning
	v.Cursor.X = 0
	v.Cursor.Y = 0
	return true
}

// OpenFile opens a new file in the buffer
func (v *View) OpenFile() bool {
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
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	}
	v.Cursor.UpN(v.height)
	return true
}

// CursorPageDown places the cursor a page up
func (v *View) CursorPageDown() bool {
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
	if settings["ruler"] == false {
		settings["ruler"] = true
		messenger.Message("Enabled ruler")
	} else {
		settings["ruler"] = false
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
	if !v.helpOpen {
		v.lastBuffer = v.Buf
		helpBuffer := NewBuffer([]byte(helpTxt), "help.md")
		helpBuffer.Name = "Help"
		v.helpOpen = true
		v.OpenBuffer(helpBuffer)
	} else {
		v.OpenBuffer(v.lastBuffer)
		v.helpOpen = false
	}
	return true
}

// ShellMode opens a terminal to run a shell command
func (v *View) ShellMode() bool {
	input, canceled := messenger.Prompt("$ ", "Shell", NoCompletion)
	if !canceled {
		// The true here is for openTerm to make the command interactive
		HandleShellCommand(input, true)
	}
	return false
}

// CommandMode lets the user enter a command
func (v *View) CommandMode() bool {
	input, canceled := messenger.Prompt("> ", "Command", NoCompletion)
	if !canceled {
		HandleCommand(input)
	}
	return false
}

// Quit quits the editor
// This behavior needs to be changed and should really only quit the editor if this
// is the last view
// However, since micro only supports one view for now, it doesn't really matter
func (v *View) Quit() bool {
	if v.helpOpen {
		return v.ToggleHelp()
	}
	// Make sure not to quit if there are unsaved changes
	if v.CanClose("Quit anyway? (yes, no, save) ") {
		v.CloseBuffer()
		if len(tabs[curTab].views) > 1 {
			var view *View
			if v.splitChild != nil {
				view = v.splitChild
				view.splitParent = v.splitParent
			} else if v.splitParent != nil {
				view = v.splitParent
				v.splitParent.splitChild = nil
			}
			view.x, view.y = view.splitOrigPos[0], view.splitOrigPos[1]
			view.widthPercent, view.heightPercent = view.splitOrigDimensions[0], view.splitOrigDimensions[1]
			view.Resize(screen.Size())
			if settings["syntax"].(bool) {
				view.matches = Match(view)
			}
			tabs[curTab].views = tabs[curTab].views[:v.Num+copy(tabs[curTab].views[v.Num:], tabs[curTab].views[v.Num+1:])]
			for i, v := range tabs[curTab].views {
				v.Num = i
			}
			tabs[curTab].curView = view.Num
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
					CurView().Resize(screen.Size())
					CurView().matches = Match(CurView())
				}
			}
		} else {
			screen.Fini()
			os.Exit(0)
		}
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
				v.Resize(screen.Size())
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

// Changes the view to the next split
func (v *View) NextSplit() bool {
	tab := tabs[curTab]
	if tab.curView < len(tab.views)-1 {
		tab.curView++
	} else {
		tab.curView = 0
	}
	return false
}

// Changes the view to the previous split
func (v *View) PreviousSplit() bool {
	tab := tabs[curTab]
	if tab.curView > 0 {
		tab.curView--
	} else {
		tab.curView = len(tab.views) - 1
	}
	return false
}

// None is no action
func None() bool {
	return false
}
