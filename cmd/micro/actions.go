package main

import (
	"os"

	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/tcell"
)

// MousePress is the event that should happen when a normal click happens
// This is almost always bound to left click
func (a *BufActionHandler) MousePress(e *tcell.EventMouse) bool {
	return false
}

// ScrollUpAction scrolls the view up
func (a *BufActionHandler) ScrollUpAction() bool {
	return false
}

// ScrollDownAction scrolls the view up
func (a *BufActionHandler) ScrollDownAction() bool {
	return false
}

// Center centers the view on the cursor
func (a *BufActionHandler) Center() bool {
	return true
}

// CursorUp moves the cursor up
func (a *BufActionHandler) CursorUp() bool {
	a.Cursor.Deselect(true)
	a.Cursor.Up()
	return true
}

// CursorDown moves the cursor down
func (a *BufActionHandler) CursorDown() bool {
	a.Cursor.Deselect(true)
	a.Cursor.Down()
	return true
}

// CursorLeft moves the cursor left
func (a *BufActionHandler) CursorLeft() bool {
	a.Cursor.Deselect(true)
	a.Cursor.Left()
	return true
}

// CursorRight moves the cursor right
func (a *BufActionHandler) CursorRight() bool {
	a.Cursor.Deselect(true)
	a.Cursor.Right()
	return true
}

// WordRight moves the cursor one word to the right
func (a *BufActionHandler) WordRight() bool {
	return true
}

// WordLeft moves the cursor one word to the left
func (a *BufActionHandler) WordLeft() bool {
	return true
}

// SelectUp selects up one line
func (a *BufActionHandler) SelectUp() bool {
	return true
}

// SelectDown selects down one line
func (a *BufActionHandler) SelectDown() bool {
	return true
}

// SelectLeft selects the character to the left of the cursor
func (a *BufActionHandler) SelectLeft() bool {
	return true
}

// SelectRight selects the character to the right of the cursor
func (a *BufActionHandler) SelectRight() bool {
	return true
}

// SelectWordRight selects the word to the right of the cursor
func (a *BufActionHandler) SelectWordRight() bool {
	return true
}

// SelectWordLeft selects the word to the left of the cursor
func (a *BufActionHandler) SelectWordLeft() bool {
	return true
}

// StartOfLine moves the cursor to the start of the line
func (a *BufActionHandler) StartOfLine() bool {
	return true
}

// EndOfLine moves the cursor to the end of the line
func (a *BufActionHandler) EndOfLine() bool {
	return true
}

// SelectLine selects the entire current line
func (a *BufActionHandler) SelectLine() bool {
	return true
}

// SelectToStartOfLine selects to the start of the current line
func (a *BufActionHandler) SelectToStartOfLine() bool {
	return true
}

// SelectToEndOfLine selects to the end of the current line
func (a *BufActionHandler) SelectToEndOfLine() bool {
	return true
}

// ParagraphPrevious moves the cursor to the previous empty line, or beginning of the buffer if there's none
func (a *BufActionHandler) ParagraphPrevious() bool {
	return true
}

// ParagraphNext moves the cursor to the next empty line, or end of the buffer if there's none
func (a *BufActionHandler) ParagraphNext() bool {
	return true
}

// Retab changes all tabs to spaces or all spaces to tabs depending
// on the user's settings
func (a *BufActionHandler) Retab() bool {
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (a *BufActionHandler) CursorStart() bool {
	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (a *BufActionHandler) CursorEnd() bool {
	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (a *BufActionHandler) SelectToStart() bool {
	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (a *BufActionHandler) SelectToEnd() bool {
	return true
}

// InsertSpace inserts a space
func (a *BufActionHandler) InsertSpace() bool {
	return true
}

// InsertNewline inserts a newline plus possible some whitespace if autoindent is on
func (a *BufActionHandler) InsertNewline() bool {
	return true
}

// Backspace deletes the previous character
func (a *BufActionHandler) Backspace() bool {
	return true
}

// DeleteWordRight deletes the word to the right of the cursor
func (a *BufActionHandler) DeleteWordRight() bool {
	return true
}

// DeleteWordLeft deletes the word to the left of the cursor
func (a *BufActionHandler) DeleteWordLeft() bool {
	return true
}

// Delete deletes the next character
func (a *BufActionHandler) Delete() bool {
	return true
}

// IndentSelection indents the current selection
func (a *BufActionHandler) IndentSelection() bool {
	return false
}

// OutdentLine moves the current line back one indentation
func (a *BufActionHandler) OutdentLine() bool {
	return true
}

// OutdentSelection takes the current selection and moves it back one indent level
func (a *BufActionHandler) OutdentSelection() bool {
	return false
}

// InsertTab inserts a tab or spaces
func (a *BufActionHandler) InsertTab() bool {
	return true
}

// SaveAll saves all open buffers
func (a *BufActionHandler) SaveAll() bool {
	return false
}

// Save the buffer to disk
func (a *BufActionHandler) Save() bool {
	return false
}

// SaveAs saves the buffer to disk with the given name
func (a *BufActionHandler) SaveAs() bool {
	return false
}

// Find opens a prompt and searches forward for the input
func (a *BufActionHandler) Find() bool {
	return true
}

// FindNext searches forwards for the last used search term
func (a *BufActionHandler) FindNext() bool {
	return true
}

// FindPrevious searches backwards for the last used search term
func (a *BufActionHandler) FindPrevious() bool {
	return true
}

// Undo undoes the last action
func (a *BufActionHandler) Undo() bool {
	return true
}

// Redo redoes the last action
func (a *BufActionHandler) Redo() bool {
	return true
}

// Copy the selection to the system clipboard
func (a *BufActionHandler) Copy() bool {
	return true
}

// CutLine cuts the current line to the clipboard
func (a *BufActionHandler) CutLine() bool {
	return true
}

// Cut the selection to the system clipboard
func (a *BufActionHandler) Cut() bool {
	return true
}

// DuplicateLine duplicates the current line or selection
func (a *BufActionHandler) DuplicateLine() bool {
	return true
}

// DeleteLine deletes the current line
func (a *BufActionHandler) DeleteLine() bool {
	return true
}

// MoveLinesUp moves up the current line or selected lines if any
func (a *BufActionHandler) MoveLinesUp() bool {
	return true
}

// MoveLinesDown moves down the current line or selected lines if any
func (a *BufActionHandler) MoveLinesDown() bool {
	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (a *BufActionHandler) Paste() bool {
	return true
}

// PastePrimary pastes from the primary clipboard (only use on linux)
func (a *BufActionHandler) PastePrimary() bool {
	return true
}

// JumpToMatchingBrace moves the cursor to the matching brace if it is
// currently on a brace
func (a *BufActionHandler) JumpToMatchingBrace() bool {
	return true
}

// SelectAll selects the entire buffer
func (a *BufActionHandler) SelectAll() bool {
	return true
}

// OpenFile opens a new file in the buffer
func (a *BufActionHandler) OpenFile() bool {
	return false
}

// Start moves the viewport to the start of the buffer
func (a *BufActionHandler) Start() bool {
	return false
}

// End moves the viewport to the end of the buffer
func (a *BufActionHandler) End() bool {
	return false
}

// PageUp scrolls the view up a page
func (a *BufActionHandler) PageUp() bool {
	return false
}

// PageDown scrolls the view down a page
func (a *BufActionHandler) PageDown() bool {
	return false
}

// SelectPageUp selects up one page
func (a *BufActionHandler) SelectPageUp() bool {
	return true
}

// SelectPageDown selects down one page
func (a *BufActionHandler) SelectPageDown() bool {
	return true
}

// CursorPageUp places the cursor a page up
func (a *BufActionHandler) CursorPageUp() bool {
	return true
}

// CursorPageDown places the cursor a page up
func (a *BufActionHandler) CursorPageDown() bool {
	return true
}

// HalfPageUp scrolls the view up half a page
func (a *BufActionHandler) HalfPageUp() bool {
	return false
}

// HalfPageDown scrolls the view down half a page
func (a *BufActionHandler) HalfPageDown() bool {
	return false
}

// ToggleRuler turns line numbers off and on
func (a *BufActionHandler) ToggleRuler() bool {
	return false
}

// JumpLine jumps to a line and moves the view accordingly.
func (a *BufActionHandler) JumpLine() bool {
	return false
}

// ClearStatus clears the messenger bar
func (a *BufActionHandler) ClearStatus() bool {
	return false
}

// ToggleHelp toggles the help screen
func (a *BufActionHandler) ToggleHelp() bool {
	return true
}

// ToggleKeyMenu toggles the keymenu option and resizes all tabs
func (a *BufActionHandler) ToggleKeyMenu() bool {
	return true
}

// ShellMode opens a terminal to run a shell command
func (a *BufActionHandler) ShellMode() bool {
	return false
}

// CommandMode lets the user enter a command
func (a *BufActionHandler) CommandMode() bool {
	return false
}

// ToggleOverwriteMode lets the user toggle the text overwrite mode
func (a *BufActionHandler) ToggleOverwriteMode() bool {
	return false
}

// Escape leaves current mode
func (a *BufActionHandler) Escape() bool {
	return false
}

// Quit this will close the current tab or view that is open
func (a *BufActionHandler) Quit() bool {
	screen.Screen.Fini()
	os.Exit(0)
	return false
}

// QuitAll quits the whole editor; all splits and tabs
func (a *BufActionHandler) QuitAll() bool {
	return false
}

// AddTab adds a new tab with an empty buffer
func (a *BufActionHandler) AddTab() bool {
	return true
}

// PreviousTab switches to the previous tab in the tab list
func (a *BufActionHandler) PreviousTab() bool {
	return false
}

// NextTab switches to the next tab in the tab list
func (a *BufActionHandler) NextTab() bool {
	return false
}

// VSplitBinding opens an empty vertical split
func (a *BufActionHandler) VSplitBinding() bool {
	return false
}

// HSplitBinding opens an empty horizontal split
func (a *BufActionHandler) HSplitBinding() bool {
	return false
}

// Unsplit closes all splits in the current tab except the active one
func (a *BufActionHandler) Unsplit() bool {
	return false
}

// NextSplit changes the view to the next split
func (a *BufActionHandler) NextSplit() bool {
	return false
}

// PreviousSplit changes the view to the previous split
func (a *BufActionHandler) PreviousSplit() bool {
	return false
}

var curMacro []interface{}
var recordingMacro bool

// ToggleMacro toggles recording of a macro
func (a *BufActionHandler) ToggleMacro() bool {
	return true
}

// PlayMacro plays back the most recently recorded macro
func (a *BufActionHandler) PlayMacro() bool {
	return true
}

// SpawnMultiCursor creates a new multiple cursor at the next occurrence of the current selection or current word
func (a *BufActionHandler) SpawnMultiCursor() bool {
	return false
}

// SpawnMultiCursorSelect adds a cursor at the beginning of each line of a selection
func (a *BufActionHandler) SpawnMultiCursorSelect() bool {
	return false
}

// MouseMultiCursor is a mouse action which puts a new cursor at the mouse position
func (a *BufActionHandler) MouseMultiCursor(e *tcell.EventMouse) bool {
	return false
}

// SkipMultiCursor moves the current multiple cursor to the next available position
func (a *BufActionHandler) SkipMultiCursor() bool {
	return false
}

// RemoveMultiCursor removes the latest multiple cursor
func (a *BufActionHandler) RemoveMultiCursor() bool {
	return false
}

// RemoveAllMultiCursors removes all cursors except the base cursor
func (a *BufActionHandler) RemoveAllMultiCursors() bool {
	return false
}
