# Keybindings

Micro has a plethora of hotkeys that make it easy and powerful to use and all
hotkeys are fully customizable to your liking.

Custom keybindings are stored internally in micro if changed with the `> bind`
command or you can also be added in the file `~/.config/micro/bindings.json` as
discussed below. For a list of the default keybindings in the json format used
by micro, please see the end of this file. For a more user-friendly list with
explanations of what the default hotkeys are and what they do, please see
`> help defaultkeys` (a json formatted list of default keys is included
at the end of this document).

If `~/.config/micro/bindings.json` does not exist, you can simply create it.
Micro will know what to do with it.

You can use the alt keys + arrows to move word by word. Ctrl left and right
move the cursor to the start and end of the line, and ctrl up and down move the
cursor the start and end of the buffer.

You can hold shift with all of these movement actions to select while moving.

## Rebinding keys

The bindings may be rebound using the `~/.config/micro/bindings.json` file.
Each key is bound to an action.

For example, to bind `Ctrl-y` to undo and `Ctrl-z` to redo, you could put the
following in the `bindings.json` file.

```json
{
	"CtrlY": "Undo",
	"CtrlZ": "Redo"
}
```

In addition to editing your `~/.config/micro/bindings.json`, you can run
`>bind <keycombo> <action>` For a list of bindable actions, see below.

You can also chain commands when rebinding. For example, if you want Alt-s to
save and quit you can bind it like so:

```json
{
    "Alt-s": "Save,Quit"
}
```

Each action will return a success flag. Actions can be chained such that
the chain only continues when there are successes, or failures, or either.
The `,` separator will always chain to the next action. The `|` separator
will abort the chain if the action preceding it succeeds, and the `&` will
abort the chain if the action preceding it fails. For example, in the default
bindings, tab is bound as

```
"Tab": "Autocomplete|IndentSelection|InsertTab"
```

This means that if the `Autocomplete` action is successful, the chain will
abort. Otherwise, it will try `IndentSelection`, and if that fails too, it
will execute `InsertTab`.

## Binding commands

You can also bind a key to execute a command in command mode (see 
`help commands`). Simply prepend the binding with `command:`. For example:

```json
{
    "Alt-p": "command:pwd"
}
```

Now when you press `Alt-p` the `pwd` command will be executed which will show
your working directory in the infobar.

You can also bind an "editable" command with `command-edit:`. This means that 
micro won't immediately execute the command when you press the binding, but
instead just place the string in the infobar in command mode. For example, 
you could rebind `CtrlG` to `> help`:

```json
{
    "CtrlG": "command-edit:help "
}
```

Now when you press `CtrlG`, `help` will appear in the command bar and your
cursor will be placed after it (note the space in the json that controls the
cursor placement).

## Binding raw escape sequences

Only read this section if you are interested in binding keys that aren't on the 
list of supported keys for binding.

One of the drawbacks of using a terminal-based editor is that the editor must
get all of its information about key events through the terminal. The terminal
sends these events in the form of escape sequences often (but not always)
starting with `0x1b`. 

For example, if micro reads `\x1b[1;5D`, on most terminals this will mean the
user pressed CtrlLeft.

For many key chords though, the terminal won't send any escape code or will
send an escape code already in use. For example for `CtrlBackspace`, my
terminal sends `\u007f` (note this doesn't start with `0x1b`), which it also
sends for `Backspace` meaning micro can't bind `CtrlBackspace`.

However, some terminals do allow you to bind keys to send specific escape
sequences you define. Then from micro you can directly bind those escape
sequences to actions. For example, to bind `CtrlBackspace` you can instruct
your terminal to send `\x1bctrlback` and then bind it in `bindings.json`:

```json
{
    "\u001bctrlback": "DeleteWordLeft"
}
```

Here are some instructions for sending raw escapes in different terminals

### iTerm2

In iTerm2, you can do this in  `Preferences->Profiles->Keys` then click the
`+`, input your keybinding, and for the `Action` select `Send Escape Sequence`.
For the above example your would type `ctrlback` into the box (the `\x1b`) is
automatically sent by iTerm2.

### Linux using loadkeys

You can do this in linux using the loadkeys program.

Coming soon!

## Unbinding keys

It is also possible to disable any of the default key bindings by use of the 
`None` action in the user's `bindings.json` file.

## Bindable actions and bindable keys

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
SelectToStartOfText
SelectToStartOfTextToggle
WordRight
WordLeft
SelectWordRight
SelectWordLeft
MoveLinesUp
MoveLinesDown
DeleteWordRight
DeleteWordLeft
SelectLine
SelectToStartOfLine
SelectToEndOfLine
InsertNewline
InsertSpace
Backspace
Delete
Center
InsertTab
Save
SaveAll
SaveAs
Find
FindLiteral
FindNext
FindPrevious
Undo
Redo
Copy
CopyLine
Cut
CutLine
DuplicateLine
DeleteLine
IndentSelection
OutdentSelection
OutdentLine
IndentLine
Paste
SelectAll
OpenFile
Start
End
PageUp
PageDown
SelectPageUp
SelectPageDown
HalfPageUp
HalfPageDown
StartOfLine
EndOfLine
StartOfText
StartOfTextToggle
ParagraphPrevious
ParagraphNext
ToggleHelp
ToggleDiffGutter
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
Suspend (Unix only)
ScrollUp
ScrollDown
SpawnMultiCursor
SpawnMultiCursorUp
SpawnMultiCursorDown
SpawnMultiCursorSelect
RemoveMultiCursor
RemoveAllMultiCursors
SkipMultiCursor
None
JumpToMatchingBrace
Autocomplete
```

The `StartOfTextToggle` and `SelectToStartOfTextToggle` actions toggle between
jumping to the start of the text (first) and start of the line.

You can also bind some mouse actions (these must be bound to mouse buttons)

```
MousePress
MouseMultiCursor
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
OldBackspace
Tab
Esc
Escape
Enter
```

You can also bind some mouse buttons (they may be bound to normal actions or
mouse actions)

```
MouseLeft
MouseMiddle
MouseRight
MouseWheelUp
MouseWheelDown
MouseWheelLeft
MouseWheelRight
```

# Default keybinding configuration.

A select few keybindings are different on MacOS compared to other
operating systems. This is because different OSes have different
conventions for text editing defaults.

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
    "AltLeft":        "WordLeft", (Mac)
    "AltRight":       "WordRight", (Mac)
    "AltUp":          "MoveLinesUp",
    "AltDown":        "MoveLinesDown",
    "CtrlShiftRight": "SelectWordRight",
    "CtrlShiftLeft":  "SelectWordLeft",
    "AltLeft":        "StartOfTextToggle",
    "AltRight":       "EndOfLine",
    "AltShiftRight":  "SelectWordRight", (Mac)
    "AltShiftLeft":   "SelectWordLeft", (Mac)
    "CtrlLeft":       "StartOfText", (Mac)
    "CtrlRight":      "EndOfLine", (Mac)
    "AltShiftLeft":   "SelectToStartOfTextToggle",
    "CtrlShiftLeft":  "SelectToStartOfTextToggle", (Mac)
    "ShiftHome":      "SelectToStartOfTextToggle",
    "AltShiftRight":  "SelectToEndOfLine",
    "CtrlShiftRight": "SelectToEndOfLine", (Mac)
    "ShiftEnd":       "SelectToEndOfLine",
    "CtrlUp":         "CursorStart",
    "CtrlDown":       "CursorEnd",
    "CtrlShiftUp":    "SelectToStart",
    "CtrlShiftDown":  "SelectToEnd",
    "Alt-{":          "ParagraphPrevious",
    "Alt-}":          "ParagraphNext",
    "Enter":          "InsertNewline",
    "CtrlH":          "Backspace",
    "Backspace":      "Backspace",
    "Alt-CtrlH":      "DeleteWordLeft",
    "Alt-Backspace":  "DeleteWordLeft",
    "Tab":            "Autocomplete|IndentSelection|InsertTab",
    "Backtab":        "OutdentSelection|OutdentLine",
    "CtrlO":          "OpenFile",
    "CtrlS":          "Save",
    "CtrlF":          "Find",
    "CtrlN":          "FindNext",
    "CtrlP":          "FindPrevious",
    "CtrlZ":          "Undo",
    "CtrlY":          "Redo",
    "CtrlC":          "CopyLine|Copy",
    "CtrlX":          "Cut",
    "CtrlK":          "CutLine",
    "CtrlD":          "DuplicateLine",
    "CtrlV":          "Paste",
    "CtrlA":          "SelectAll",
    "CtrlT":          "AddTab",
    "Alt,":           "PreviousTab",
    "Alt.":           "NextTab",
    "Home":           "StartOfText",
    "End":            "EndOfLine",
    "CtrlHome":       "CursorStart",
    "CtrlEnd":        "CursorEnd",
    "PageUp":         "CursorPageUp",
    "PageDown":       "CursorPageDown",
    "CtrlPageUp":     "PreviousTab",
    "CtrlPageDown":   "NextTab",
    "CtrlG":          "ToggleHelp",
    "Alt-g":          "ToggleKeyMenu",
    "CtrlR":          "ToggleRuler",
    "CtrlL":          "command-edit:goto ",
    "Delete":         "Delete",
    "CtrlB":          "ShellMode",
    "CtrlQ":          "Quit",
    "CtrlE":          "CommandMode",
    "CtrlW":          "NextSplit",
    "CtrlU":          "ToggleMacro",
    "CtrlJ":          "PlayMacro",
    "Insert":         "ToggleOverwriteMode",

    // Emacs-style keybindings
    "Alt-f": "WordRight",
    "Alt-b": "WordLeft",
    "Alt-a": "StartOfLine",
    "Alt-e": "EndOfLine",

    // Integration with file managers
    "F2":  "Save",
    "F3":  "Find",
    "F4":  "Quit",
    "F7":  "Find",
    "F10": "Quit",
    "Esc": "Escape",

    // Mouse bindings
    "MouseWheelUp":   "ScrollUp",
    "MouseWheelDown": "ScrollDown",
    "MouseLeft":      "MousePress",
    "MouseMiddle":    "PastePrimary",
    "Ctrl-MouseLeft": "MouseMultiCursor",

    "Alt-n":        "SpawnMultiCursor",
    "AltShiftUp":   "SpawnMultiCursorUp",
    "AltShiftDown": "SpawnMultiCursorDown",
    "Alt-m":        "SpawnMultiCursorSelect",
    "Alt-p":        "RemoveMultiCursor",
    "Alt-c":        "RemoveAllMultiCursors",
    "Alt-x":        "SkipMultiCursor",
}
```

## Final notes

Note: On some old terminal emulators and on Windows machines, `CtrlH` should be
used for backspace.

Additionally, alt keys can be bound by using `Alt-key`. For example `Alt-a` or
`Alt-Up`. Micro supports an optional `-` between modifiers like `Alt` and 
`Ctrl` so `Alt-a` could be rewritten as `Alta` (case matters for alt bindings).
This is why in the default keybindings you can see `AltShiftLeft` instead of
`Alt-ShiftLeft` (they are equivalent).

Please note that terminal emulators are strange applications and micro only
receives key events that the terminal decides to send. Some terminal emulators
may not send certain events even if this document says micro can receive the
event. To see exactly what micro receives from the terminal when you press a
key, run the `> raw` command.
