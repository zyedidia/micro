# Keybindings

Here are the default keybindings in json form which is also how
you can rebind them to your liking.

```json
{
	"Up":             "CursorUp",
	"Down":           "CursorDown",
	"Right":          "CursorRight",
	"Left":           "CursorLeft",
	"ShiftUp":        "SelectUp",
	"ShiftDown":      "SelectDown",
	"ShiftLeft":      "SelectLeft",
	"ShiftRight":     "SelectRight",
	"AltLeft":        "WordLeft",
	"AltRight":       "WordRight",
	"AltShiftRight":  "SelectWordRight",
	"AltShiftLeft":   "SelectWordLeft",
	"CtrlLeft":       "StartOfLine",
	"CtrlRight":      "EndOfLine",
	"CtrlShiftLeft":  "SelectToStartOfLine",
	"CtrlShiftRight": "SelectToEndOfLine",
	"CtrlUp":         "CursorStart",
	"CtrlDown":       "CursorEnd",
	"CtrlShiftUp":    "SelectToStart",
	"CtrlShiftDown":  "SelectToEnd",
	"Enter":          "InsertEnter",
	"Space":          "InsertSpace",
	"Backspace":      "Backspace",
	"Backspace2":     "Backspace",
	"Alt-Backspace":  "DeleteWordLeft",
	"Alt-Backspace2": "DeleteWordLeft",
	"Tab":            "InsertTab,IndentSelection",
	"CtrlO":          "OpenFile",
	"CtrlS":          "Save",
	"CtrlF":          "Find",
	"CtrlN":          "FindNext",
	"CtrlP":          "FindPrevious",
	"CtrlZ":          "Undo",
	"CtrlY":          "Redo",
	"CtrlC":          "Copy",
	"CtrlX":          "Cut",
	"CtrlK":          "CutLine",
	"CtrlD":          "DuplicateLine",
	"CtrlV":          "Paste",
	"CtrlA":          "SelectAll",
	"CtrlT":          "AddTab"
	"CtrlRightSq":    "PreviousTab",
	"CtrlBackslash":  "NextTab",
	"Home":           "Start",
	"End":            "End",
	"PageUp":         "CursorPageUp",
	"PageDown":       "CursorPageDown",
	"CtrlG":          "ToggleHelp",
	"CtrlR":          "ToggleRuler",
	"CtrlL":          "JumpLine",
	"Delete":         "Delete",
	"Esc":            "ClearStatus",
	"CtrlB":          "ShellMode",
	"CtrlQ":          "Quit",
	"CtrlE":          "CommandMode",
	"CtrlW":          "NextSplit",
	
	// Emacs-style keybindings
	"Alt-f": "WordRight",
	"Alt-b": "WordLeft",
	"Alt-a": "StartOfLine",
	"Alt-e": "EndOfLine",
	"Alt-p": "CursorUp",
	"Alt-n": "CursorDown"
}
```

You can use the alt keys + arrows to move word by word.
Ctrl left and right move the cursor to the start and end of the line, and
ctrl up and down move the cursor the start and end of the buffer.

You can hold shift with all of these movement actions to select while moving.

# Rebinding keys

The bindings may be rebound using the `~/.config/micro/bindings.json`
file. Each key is bound to an action.

For example, to bind `Ctrl-y` to undo and `Ctrl-z` to redo, you could put the
following in the `bindings.json` file.

```json
{
	"CtrlY": "Undo",
	"CtrlZ": "Redo"
}
```

You can also chain commands when rebinding. For example, if you want Alt-s to save
and quit you can bind it like so:

```json
{
    "Alt-s": "Save,Quit"
}
```
