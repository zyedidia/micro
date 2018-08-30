package action

import (
	"os"
	"unicode/utf8"

	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/micro/cmd/micro/util"
	"github.com/zyedidia/tcell"
)

// MousePress is the event that should happen when a normal click happens
// This is almost always bound to left click
func (h *BufHandler) MousePress(e *tcell.EventMouse) bool {
	return false
}

// ScrollUpAction scrolls the view up
func (h *BufHandler) ScrollUpAction() bool {
	return false
}

// ScrollDownAction scrolls the view up
func (h *BufHandler) ScrollDownAction() bool {
	return false
}

// Center centers the view on the cursor
func (h *BufHandler) Center() bool {
	return true
}

// CursorUp moves the cursor up
func (h *BufHandler) CursorUp() bool {
	h.Cursor.Deselect(true)
	h.Cursor.Up()
	return true
}

// CursorDown moves the cursor down
func (h *BufHandler) CursorDown() bool {
	h.Cursor.Deselect(true)
	h.Cursor.Down()
	return true
}

// CursorLeft moves the cursor left
func (h *BufHandler) CursorLeft() bool {
	h.Cursor.Deselect(true)
	h.Cursor.Left()
	return true
}

// CursorRight moves the cursor right
func (h *BufHandler) CursorRight() bool {
	h.Cursor.Deselect(true)
	h.Cursor.Right()
	return true
}

// WordRight moves the cursor one word to the right
func (h *BufHandler) WordRight() bool {
	h.Cursor.Deselect(true)
	h.Cursor.WordRight()
	return true
}

// WordLeft moves the cursor one word to the left
func (h *BufHandler) WordLeft() bool {
	h.Cursor.Deselect(true)
	h.Cursor.WordLeft()
	return true
}

// SelectUp selects up one line
func (h *BufHandler) SelectUp() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.Up()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectDown selects down one line
func (h *BufHandler) SelectDown() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.Down()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectLeft selects the character to the left of the cursor
func (h *BufHandler) SelectLeft() bool {
	loc := h.Cursor.Loc
	count := h.Buf.End()
	if loc.GreaterThan(count) {
		loc = count
	}
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = loc
	}
	h.Cursor.Left()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectRight selects the character to the right of the cursor
func (h *BufHandler) SelectRight() bool {
	loc := h.Cursor.Loc
	count := h.Buf.End()
	if loc.GreaterThan(count) {
		loc = count
	}
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = loc
	}
	h.Cursor.Right()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectWordRight selects the word to the right of the cursor
func (h *BufHandler) SelectWordRight() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.WordRight()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectWordLeft selects the word to the left of the cursor
func (h *BufHandler) SelectWordLeft() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.WordLeft()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// StartOfLine moves the cursor to the start of the line
func (h *BufHandler) StartOfLine() bool {
	h.Cursor.Deselect(true)
	if h.Cursor.X != 0 {
		h.Cursor.Start()
	} else {
		h.Cursor.StartOfText()
	}
	return true
}

// EndOfLine moves the cursor to the end of the line
func (h *BufHandler) EndOfLine() bool {
	h.Cursor.Deselect(true)
	h.Cursor.End()
	return true
}

// SelectLine selects the entire current line
func (h *BufHandler) SelectLine() bool {
	h.Cursor.SelectLine()
	return true
}

// SelectToStartOfLine selects to the start of the current line
func (h *BufHandler) SelectToStartOfLine() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.Start()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectToEndOfLine selects to the end of the current line
func (h *BufHandler) SelectToEndOfLine() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.End()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// ParagraphPrevious moves the cursor to the previous empty line, or beginning of the buffer if there's none
func (h *BufHandler) ParagraphPrevious() bool {
	var line int
	for line = h.Cursor.Y; line > 0; line-- {
		if len(h.Buf.LineBytes(line)) == 0 && line != h.Cursor.Y {
			h.Cursor.X = 0
			h.Cursor.Y = line
			break
		}
	}
	// If no empty line found. move cursor to end of buffer
	if line == 0 {
		h.Cursor.Loc = h.Buf.Start()
	}
	return true
}

// ParagraphNext moves the cursor to the next empty line, or end of the buffer if there's none
func (h *BufHandler) ParagraphNext() bool {
	var line int
	for line = h.Cursor.Y; line < h.Buf.LinesNum(); line++ {
		if len(h.Buf.LineBytes(line)) == 0 && line != h.Cursor.Y {
			h.Cursor.X = 0
			h.Cursor.Y = line
			break
		}
	}
	// If no empty line found. move cursor to end of buffer
	if line == h.Buf.LinesNum() {
		h.Cursor.Loc = h.Buf.End()
	}
	return true
}

// Retab changes all tabs to spaces or all spaces to tabs depending
// on the user's settings
func (h *BufHandler) Retab() bool {
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (h *BufHandler) CursorStart() bool {
	h.Cursor.Deselect(true)
	h.Cursor.X = 0
	h.Cursor.Y = 0
	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (h *BufHandler) CursorEnd() bool {
	h.Cursor.Deselect(true)
	h.Cursor.Loc = h.Buf.End()
	h.Cursor.StoreVisualX()
	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (h *BufHandler) SelectToStart() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.CursorStart()
	h.Cursor.SelectTo(h.Buf.Start())
	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (h *BufHandler) SelectToEnd() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.CursorEnd()
	h.Cursor.SelectTo(h.Buf.End())
	return true
}

// InsertSpace inserts a space
func (h *BufHandler) InsertSpace() bool {
	if h.Cursor.HasSelection() {
		h.Cursor.DeleteSelection()
		h.Cursor.ResetSelection()
	}
	h.Buf.Insert(h.Cursor.Loc, " ")
	return true
}

// InsertNewline inserts a newline plus possible some whitespace if autoindent is on
func (h *BufHandler) InsertNewline() bool {
	return true
}

// Backspace deletes the previous character
func (h *BufHandler) Backspace() bool {
	if h.Cursor.HasSelection() {
		h.Cursor.DeleteSelection()
		h.Cursor.ResetSelection()
	} else if h.Cursor.Loc.GreaterThan(h.Buf.Start()) {
		// We have to do something a bit hacky here because we want to
		// delete the line by first moving left and then deleting backwards
		// but the undo redo would place the cursor in the wrong place
		// So instead we move left, save the position, move back, delete
		// and restore the position

		// If the user is using spaces instead of tabs and they are deleting
		// whitespace at the start of the line, we should delete as if it's a
		// tab (tabSize number of spaces)
		lineStart := util.SliceStart(h.Buf.LineBytes(h.Cursor.Y), h.Cursor.X)
		tabSize := int(h.Buf.Settings["tabsize"].(float64))
		if h.Buf.Settings["tabstospaces"].(bool) && util.IsSpaces(lineStart) && len(lineStart) != 0 && utf8.RuneCount(lineStart)%tabSize == 0 {
			loc := h.Cursor.Loc
			h.Buf.Remove(loc.Move(-tabSize, h.Buf), loc)
		} else {
			loc := h.Cursor.Loc
			h.Buf.Remove(loc.Move(-1, h.Buf), loc)
		}
	}
	h.Cursor.LastVisualX = h.Cursor.GetVisualX()
	return true
}

// DeleteWordRight deletes the word to the right of the cursor
func (h *BufHandler) DeleteWordRight() bool {
	return true
}

// DeleteWordLeft deletes the word to the left of the cursor
func (h *BufHandler) DeleteWordLeft() bool {
	return true
}

// Delete deletes the next character
func (h *BufHandler) Delete() bool {
	return true
}

// IndentSelection indents the current selection
func (h *BufHandler) IndentSelection() bool {
	return false
}

// OutdentLine moves the current line back one indentation
func (h *BufHandler) OutdentLine() bool {
	return true
}

// OutdentSelection takes the current selection and moves it back one indent level
func (h *BufHandler) OutdentSelection() bool {
	return false
}

// InsertTab inserts a tab or spaces
func (h *BufHandler) InsertTab() bool {
	return true
}

// SaveAll saves all open buffers
func (h *BufHandler) SaveAll() bool {
	return false
}

// Save the buffer to disk
func (h *BufHandler) Save() bool {
	h.Buf.Save()
	return false
}

// SaveAs saves the buffer to disk with the given name
func (h *BufHandler) SaveAs() bool {
	return false
}

// Find opens a prompt and searches forward for the input
func (h *BufHandler) Find() bool {
	return true
}

// FindNext searches forwards for the last used search term
func (h *BufHandler) FindNext() bool {
	return true
}

// FindPrevious searches backwards for the last used search term
func (h *BufHandler) FindPrevious() bool {
	return true
}

// Undo undoes the last action
func (h *BufHandler) Undo() bool {
	return true
}

// Redo redoes the last action
func (h *BufHandler) Redo() bool {
	return true
}

// Copy the selection to the system clipboard
func (h *BufHandler) Copy() bool {
	return true
}

// CutLine cuts the current line to the clipboard
func (h *BufHandler) CutLine() bool {
	return true
}

// Cut the selection to the system clipboard
func (h *BufHandler) Cut() bool {
	return true
}

// DuplicateLine duplicates the current line or selection
func (h *BufHandler) DuplicateLine() bool {
	return true
}

// DeleteLine deletes the current line
func (h *BufHandler) DeleteLine() bool {
	return true
}

// MoveLinesUp moves up the current line or selected lines if any
func (h *BufHandler) MoveLinesUp() bool {
	return true
}

// MoveLinesDown moves down the current line or selected lines if any
func (h *BufHandler) MoveLinesDown() bool {
	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (h *BufHandler) Paste() bool {
	return true
}

// PastePrimary pastes from the primary clipboard (only use on linux)
func (h *BufHandler) PastePrimary() bool {
	return true
}

// JumpToMatchingBrace moves the cursor to the matching brace if it is
// currently on a brace
func (h *BufHandler) JumpToMatchingBrace() bool {
	return true
}

// SelectAll selects the entire buffer
func (h *BufHandler) SelectAll() bool {
	return true
}

// OpenFile opens a new file in the buffer
func (h *BufHandler) OpenFile() bool {
	return false
}

// Start moves the viewport to the start of the buffer
func (h *BufHandler) Start() bool {
	return false
}

// End moves the viewport to the end of the buffer
func (h *BufHandler) End() bool {
	return false
}

// PageUp scrolls the view up a page
func (h *BufHandler) PageUp() bool {
	return false
}

// PageDown scrolls the view down a page
func (h *BufHandler) PageDown() bool {
	return false
}

// SelectPageUp selects up one page
func (h *BufHandler) SelectPageUp() bool {
	return true
}

// SelectPageDown selects down one page
func (h *BufHandler) SelectPageDown() bool {
	return true
}

// CursorPageUp places the cursor a page up
func (h *BufHandler) CursorPageUp() bool {
	return true
}

// CursorPageDown places the cursor a page up
func (h *BufHandler) CursorPageDown() bool {
	return true
}

// HalfPageUp scrolls the view up half a page
func (h *BufHandler) HalfPageUp() bool {
	return false
}

// HalfPageDown scrolls the view down half a page
func (h *BufHandler) HalfPageDown() bool {
	return false
}

// ToggleRuler turns line numbers off and on
func (h *BufHandler) ToggleRuler() bool {
	return false
}

// JumpLine jumps to a line and moves the view accordingly.
func (h *BufHandler) JumpLine() bool {
	return false
}

// ClearStatus clears the messenger bar
func (h *BufHandler) ClearStatus() bool {
	return false
}

// ToggleHelp toggles the help screen
func (h *BufHandler) ToggleHelp() bool {
	return true
}

// ToggleKeyMenu toggles the keymenu option and resizes all tabs
func (h *BufHandler) ToggleKeyMenu() bool {
	return true
}

// ShellMode opens a terminal to run a shell command
func (h *BufHandler) ShellMode() bool {
	return false
}

// CommandMode lets the user enter a command
func (h *BufHandler) CommandMode() bool {
	return false
}

// ToggleOverwriteMode lets the user toggle the text overwrite mode
func (h *BufHandler) ToggleOverwriteMode() bool {
	return false
}

// Escape leaves current mode
func (h *BufHandler) Escape() bool {
	return false
}

// Quit this will close the current tab or view that is open
func (h *BufHandler) Quit() bool {
	screen.Screen.Fini()
	os.Exit(0)
	return false
}

// QuitAll quits the whole editor; all splits and tabs
func (h *BufHandler) QuitAll() bool {
	return false
}

// AddTab adds a new tab with an empty buffer
func (h *BufHandler) AddTab() bool {
	return true
}

// PreviousTab switches to the previous tab in the tab list
func (h *BufHandler) PreviousTab() bool {
	return false
}

// NextTab switches to the next tab in the tab list
func (h *BufHandler) NextTab() bool {
	return false
}

// VSplitBinding opens an empty vertical split
func (h *BufHandler) VSplitBinding() bool {
	return false
}

// HSplitBinding opens an empty horizontal split
func (h *BufHandler) HSplitBinding() bool {
	return false
}

// Unsplit closes all splits in the current tab except the active one
func (h *BufHandler) Unsplit() bool {
	return false
}

// NextSplit changes the view to the next split
func (h *BufHandler) NextSplit() bool {
	return false
}

// PreviousSplit changes the view to the previous split
func (h *BufHandler) PreviousSplit() bool {
	return false
}

var curMacro []interface{}
var recordingMacro bool

// ToggleMacro toggles recording of a macro
func (h *BufHandler) ToggleMacro() bool {
	return true
}

// PlayMacro plays back the most recently recorded macro
func (h *BufHandler) PlayMacro() bool {
	return true
}

// SpawnMultiCursor creates a new multiple cursor at the next occurrence of the current selection or current word
func (h *BufHandler) SpawnMultiCursor() bool {
	return false
}

// SpawnMultiCursorSelect adds a cursor at the beginning of each line of a selection
func (h *BufHandler) SpawnMultiCursorSelect() bool {
	return false
}

// MouseMultiCursor is a mouse action which puts a new cursor at the mouse position
func (h *BufHandler) MouseMultiCursor(e *tcell.EventMouse) bool {
	return false
}

// SkipMultiCursor moves the current multiple cursor to the next available position
func (h *BufHandler) SkipMultiCursor() bool {
	return false
}

// RemoveMultiCursor removes the latest multiple cursor
func (h *BufHandler) RemoveMultiCursor() bool {
	return false
}

// RemoveAllMultiCursors removes all cursors except the base cursor
func (h *BufHandler) RemoveAllMultiCursors() bool {
	return false
}
