# Keybindings

Here are the default keybindings in json format. You can rebind them to your liking, following the same format.

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
    "AltUp":          "MoveLinesUp",
    "AltDown":        "MoveLinesDown",
    "CtrlLeft":       "StartOfLine",
    "CtrlRight":      "EndOfLine",
    "CtrlShiftLeft":  "SelectToStartOfLine",
    "CtrlShiftRight": "SelectToEndOfLine",
    "CtrlUp":         "CursorStart",
    "CtrlDown":       "CursorEnd",
    "CtrlShiftUp":    "SelectToStart",
    "CtrlShiftDown":  "SelectToEnd",
    "Enter":          "InsertNewline",
    "Space":          "InsertSpace",
    "CtrlH":          "Backspace",
    "Backspace":      "Backspace",
    "Alt-CtrlH":      "DeleteWordLeft",
    "Alt-Backspace":  "DeleteWordLeft",
    "Tab":            "IndentSelection,InsertTab",
    "Backtab":        "OutdentSelection",
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
    "CtrlT":          "AddTab",
    "CtrlRightSq":    "PreviousTab",
    "CtrlBackslash":  "NextTab",
    "Home":           "StartOfLine",
    "End":            "EndOfLine",
    "CtrlHome":       "CursorStart",
    "CtrlEnd":        "CursorEnd",
    "PageUp":         "CursorPageUp",
    "PageDown":       "CursorPageDown",
    "CtrlG":          "ToggleHelp",
    "CtrlR":          "ToggleRuler",
    "CtrlL":          "JumpLine",
    "Delete":         "Delete",
    "CtrlB":          "ShellMode",
    "CtrlQ":          "Quit",
    "CtrlE":          "CommandMode",
    "CtrlW":          "NextSplit",
    "CtrlU":          "ToggleMacro",
    "CtrlJ":          "PlayMacro",

    // Emacs-style keybindings
    "Alt-f": "WordRight",
    "Alt-b": "WordLeft",
    "Alt-a": "StartOfLine",
    "Alt-e": "EndOfLine",
    "Alt-p": "CursorUp",
    "Alt-n": "CursorDown",

    // Integration with file managers
    "F1":  "ToggleHelp",
    "F2":  "Save",
    "F4":  "Quit",
    "F7":  "Find",
    "F10": "Quit",
    "Esc": "Escape",
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

# Bindable actions and bindable keys

The list of default keybindings contains most of the possible actions and keys
which you can use, but not all of them. Here is a full list of both.

Full list of possible actions:

```
CursorUp
CursorDown
CursorPageUp
CursorPageDown
CursorLeft
CursorRight
CursorStart
CursorEnd
SelectToStart
SelectToEnd
SelectUp
SelectDown
SelectLeft
SelectRight
WordRight
WordLeft
SelectWordRight
SelectWordLeft
MoveLinesUp
MoveLinesDown
DeleteWordRight
DeleteWordLeft
SelectToStartOfLine
SelectToEndOfLine
InsertNewline
InsertSpace
Backspace
Delete
Center
InsertTab
Save
SaveAs
Find
FindNext
FindPrevious
Undo
Redo
Copy
Cut
CutLine
DuplicateLine
DeleteLine
IndentSelection
OutdentSelection
Paste
SelectAll
OpenFile
Start
End
PageUp
PageDown
HalfPageUp
HalfPageDown
StartOfLine
EndOfLine
ToggleHelp
ToggleRuler
JumpLine
ClearStatus
ShellMode
CommandMode
Quit
QuitAll
AddTab
PreviousTab
NextTab
NextSplit
Unsplit
VSplit
HSplit
PreviousSplit
ToggleMacro
PlayMacro
NextLoc
PrevLoc
```

Here is the list of all possible keys you can bind:

```
Up
Down
Right
Left
UpLeft
UpRight
DownLeft
DownRight
Center
PageUp
PageDown
Home
End
Insert
Delete
Help
Exit
Clear
Cancel
Print
Pause
Backtab
F1
F2
F3
F4
F5
F6
F7
F8
F9
F10
F11
F12
F13
F14
F15
F16
F17
F18
F19
F20
F21
F22
F23
F24
F25
F26
F27
F28
F29
F30
F31
F32
F33
F34
F35
F36
F37
F38
F39
F40
F41
F42
F43
F44
F45
F46
F47
F48
F49
F50
F51
F52
F53
F54
F55
F56
F57
F58
F59
F60
F61
F62
F63
F64
CtrlSpace
CtrlA
CtrlB
CtrlC
CtrlD
CtrlE
CtrlF
CtrlG
CtrlH
CtrlI
CtrlJ
CtrlK
CtrlL
CtrlM
CtrlN
CtrlO
CtrlP
CtrlQ
CtrlR
CtrlS
CtrlT
CtrlU
CtrlV
CtrlW
CtrlX
CtrlY
CtrlZ
CtrlLeftSq
CtrlBackslash
CtrlRightSq
CtrlCarat
CtrlUnderscore
Backspace
Tab
Esc
Escape
Enter
```

Note: On some old terminal emulators and on Windows machines, `CtrlH` should be used
for backspace.

Additionally, alt keys can be bound by using `Alt-key`. For example `Alt-a`
or `Alt-Up`. Micro supports an optional `-` between modifiers like `Alt` and `Ctrl`
so `Alt-a` could be rewritten as `Alta` (case matters for alt bindings). This is
why in the default keybindings you can see `AltShiftLeft` instead of `Alt-ShiftLeft` 
(they are equivalent).
