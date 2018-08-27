package action

import (
	"os"

	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/tcell"
)

// MousePress is the event that should happen when a normal click happens
// This is almost always bound to left click
func (a *BufHandler) MousePress(e *tcell.EventMouse) bool {
	return false
}

// ScrollUpAction scrolls the view up
func (a *BufHandler) ScrollUpAction() bool {
	return false
}

// ScrollDownAction scrolls the view up
func (a *BufHandler) ScrollDownAction() bool {
	return false
}

// Center centers the view on the cursor
func (a *BufHandler) Center() bool {
	return true
}

// CursorUp moves the cursor up
func (a *BufHandler) CursorUp() bool {
	a.Cursor.Deselect(true)
	a.Cursor.Up()
	return true
}

// CursorDown moves the cursor down
func (a *BufHandler) CursorDown() bool {
	a.Cursor.Deselect(true)
	a.Cursor.Down()
	return true
}

// CursorLeft moves the cursor left
func (a *BufHandler) CursorLeft() bool {
	a.Cursor.Deselect(true)
	a.Cursor.Left()
	return true
}

// CursorRight moves the cursor right
func (a *BufHandler) CursorRight() bool {
	a.Cursor.Deselect(true)
	a.Cursor.Right()
	return true
}

// WordRight moves the cursor one word to the right
func (a *BufHandler) WordRight() bool {
	return true
}

// WordLeft moves the cursor one word to the left
func (a *BufHandler) WordLeft() bool {
	return true
}

// SelectUp selects up one line
func (a *BufHandler) SelectUp() bool {
	return true
}

// SelectDown selects down one line
func (a *BufHandler) SelectDown() bool {
	return true
}

// SelectLeft selects the character to the left of the cursor
func (a *BufHandler) SelectLeft() bool {
	return true
}

// SelectRight selects the character to the right of the cursor
func (a *BufHandler) SelectRight() bool {
	return true
}

// SelectWordRight selects the word to the right of the cursor
func (a *BufHandler) SelectWordRight() bool {
	return true
}

// SelectWordLeft selects the word to the left of the cursor
func (a *BufHandler) SelectWordLeft() bool {
	return true
}

// StartOfLine moves the cursor to the start of the line
func (a *BufHandler) StartOfLine() bool {
	return true
}

// EndOfLine moves the cursor to the end of the line
func (a *BufHandler) EndOfLine() bool {
	return true
}

// SelectLine selects the entire current line
func (a *BufHandler) SelectLine() bool {
	return true
}

// SelectToStartOfLine selects to the start of the current line
func (a *BufHandler) SelectToStartOfLine() bool {
	return true
}

// SelectToEndOfLine selects to the end of the current line
func (a *BufHandler) SelectToEndOfLine() bool {
	return true
}

// ParagraphPrevious moves the cursor to the previous empty line, or beginning of the buffer if there's none
func (a *BufHandler) ParagraphPrevious() bool {
	return true
}

// ParagraphNext moves the cursor to the next empty line, or end of the buffer if there's none
func (a *BufHandler) ParagraphNext() bool {
	return true
}

// Retab changes all tabs to spaces or all spaces to tabs depending
// on the user's settings
func (a *BufHandler) Retab() bool {
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (a *BufHandler) CursorStart() bool {
	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (a *BufHandler) CursorEnd() bool {
	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (a *BufHandler) SelectToStart() bool {
	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (a *BufHandler) SelectToEnd() bool {
	return true
}

// InsertSpace inserts a space
func (a *BufHandler) InsertSpace() bool {
	return true
}

// InsertNewline inserts a newline plus possible some whitespace if autoindent is on
func (a *BufHandler) InsertNewline() bool {
	return true
}

// Backspace deletes the previous character
func (a *BufHandler) Backspace() bool {
	return true
}

// DeleteWordRight deletes the word to the right of the cursor
func (a *BufHandler) DeleteWordRight() bool {
	return true
}

// DeleteWordLeft deletes the word to the left of the cursor
func (a *BufHandler) DeleteWordLeft() bool {
	return true
}

// Delete deletes the next character
func (a *BufHandler) Delete() bool {
	return true
}

// IndentSelection indents the current selection
func (a *BufHandler) IndentSelection() bool {
	return false
}

// OutdentLine moves the current line back one indentation
func (a *BufHandler) OutdentLine() bool {
	return true
}

// OutdentSelection takes the current selection and moves it back one indent level
func (a *BufHandler) OutdentSelection() bool {
	return false
}

// InsertTab inserts a tab or spaces
func (a *BufHandler) InsertTab() bool {
	return true
}

// SaveAll saves all open buffers
func (a *BufHandler) SaveAll() bool {
	return false
}

// Save the buffer to disk
func (a *BufHandler) Save() bool {
	return false
}

// SaveAs saves the buffer to disk with the given name
func (a *BufHandler) SaveAs() bool {
	return false
}

// Find opens a prompt and searches forward for the input
func (a *BufHandler) Find() bool {
	return true
}

// FindNext searches forwards for the last used search term
func (a *BufHandler) FindNext() bool {
	return true
}

// FindPrevious searches backwards for the last used search term
func (a *BufHandler) FindPrevious() bool {
	return true
}

// Undo undoes the last action
func (a *BufHandler) Undo() bool {
	return true
}

// Redo redoes the last action
func (a *BufHandler) Redo() bool {
	return true
}

// Copy the selection to the system clipboard
func (a *BufHandler) Copy() bool {
	return true
}

// CutLine cuts the current line to the clipboard
func (a *BufHandler) CutLine() bool {
	return true
}

// Cut the selection to the system clipboard
func (a *BufHandler) Cut() bool {
	return true
}

// DuplicateLine duplicates the current line or selection
func (a *BufHandler) DuplicateLine() bool {
	return true
}

// DeleteLine deletes the current line
func (a *BufHandler) DeleteLine() bool {
	return true
}

// MoveLinesUp moves up the current line or selected lines if any
func (a *BufHandler) MoveLinesUp() bool {
	return true
}

// MoveLinesDown moves down the current line or selected lines if any
func (a *BufHandler) MoveLinesDown() bool {
	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (a *BufHandler) Paste() bool {
	return true
}

// PastePrimary pastes from the primary clipboard (only use on linux)
func (a *BufHandler) PastePrimary() bool {
	return true
}

// JumpToMatchingBrace moves the cursor to the matching brace if it is
// currently on a brace
func (a *BufHandler) JumpToMatchingBrace() bool {
	return true
}

// SelectAll selects the entire buffer
func (a *BufHandler) SelectAll() bool {
	return true
}

// OpenFile opens a new file in the buffer
func (a *BufHandler) OpenFile() bool {
	return false
}

// Start moves the viewport to the start of the buffer
func (a *BufHandler) Start() bool {
	return false
}

// End moves the viewport to the end of the buffer
func (a *BufHandler) End() bool {
	return false
}

// PageUp scrolls the view up a page
func (a *BufHandler) PageUp() bool {
	return false
}

// PageDown scrolls the view down a page
func (a *BufHandler) PageDown() bool {
	return false
}

// SelectPageUp selects up one page
func (a *BufHandler) SelectPageUp() bool {
	return true
}

// SelectPageDown selects down one page
func (a *BufHandler) SelectPageDown() bool {
	return true
}

// CursorPageUp places the cursor a page up
func (a *BufHandler) CursorPageUp() bool {
	return true
}

// CursorPageDown places the cursor a page up
func (a *BufHandler) CursorPageDown() bool {
	return true
}

// HalfPageUp scrolls the view up half a page
func (a *BufHandler) HalfPageUp() bool {
	return false
}

// HalfPageDown scrolls the view down half a page
func (a *BufHandler) HalfPageDown() bool {
	return false
}

// ToggleRuler turns line numbers off and on
func (a *BufHandler) ToggleRuler() bool {
	return false
}

// JumpLine jumps to a line and moves the view accordingly.
func (a *BufHandler) JumpLine() bool {
	return false
}

// ClearStatus clears the messenger bar
func (a *BufHandler) ClearStatus() bool {
	return false
}

// ToggleHelp toggles the help screen
func (a *BufHandler) ToggleHelp() bool {
	return true
}

// ToggleKeyMenu toggles the keymenu option and resizes all tabs
func (a *BufHandler) ToggleKeyMenu() bool {
	return true
}

// ShellMode opens a terminal to run a shell command
func (a *BufHandler) ShellMode() bool {
	return false
}

// CommandMode lets the user enter a command
func (a *BufHandler) CommandMode() bool {
	return false
}

// ToggleOverwriteMode lets the user toggle the text overwrite mode
func (a *BufHandler) ToggleOverwriteMode() bool {
	return false
}

// Escape leaves current mode
func (a *BufHandler) Escape() bool {
	return false
}

// Quit this will close the current tab or view that is open
func (a *BufHandler) Quit() bool {
	screen.Screen.Fini()
	os.Exit(0)
	return false
}

// QuitAll quits the whole editor; all splits and tabs
func (a *BufHandler) QuitAll() bool {
	return false
}

// AddTab adds a new tab with an empty buffer
func (a *BufHandler) AddTab() bool {
	return true
}

// PreviousTab switches to the previous tab in the tab list
func (a *BufHandler) PreviousTab() bool {
	return false
}

// NextTab switches to the next tab in the tab list
func (a *BufHandler) NextTab() bool {
	return false
}

// VSplitBinding opens an empty vertical split
func (a *BufHandler) VSplitBinding() bool {
	return false
}

// HSplitBinding opens an empty horizontal split
func (a *BufHandler) HSplitBinding() bool {
	return false
}

// Unsplit closes all splits in the current tab except the active one
func (a *BufHandler) Unsplit() bool {
	return false
}

// NextSplit changes the view to the next split
func (a *BufHandler) NextSplit() bool {
	return false
}

// PreviousSplit changes the view to the previous split
func (a *BufHandler) PreviousSplit() bool {
	return false
}

var curMacro []interface{}
var recordingMacro bool

// ToggleMacro toggles recording of a macro
func (a *BufHandler) ToggleMacro() bool {
	return true
}

// PlayMacro plays back the most recently recorded macro
func (a *BufHandler) PlayMacro() bool {
	return true
}

// SpawnMultiCursor creates a new multiple cursor at the next occurrence of the current selection or current word
func (a *BufHandler) SpawnMultiCursor() bool {
	return false
}

// SpawnMultiCursorSelect adds a cursor at the beginning of each line of a selection
func (a *BufHandler) SpawnMultiCursorSelect() bool {
	return false
}

// MouseMultiCursor is a mouse action which puts a new cursor at the mouse position
func (a *BufHandler) MouseMultiCursor(e *tcell.EventMouse) bool {
	return false
}

// SkipMultiCursor moves the current multiple cursor to the next available position
func (a *BufHandler) SkipMultiCursor() bool {
	return false
}

// RemoveMultiCursor removes the latest multiple cursor
func (a *BufHandler) RemoveMultiCursor() bool {
	return false
}

// RemoveAllMultiCursors removes all cursors except the base cursor
func (a *BufHandler) RemoveAllMultiCursors() bool {
	return false
}
