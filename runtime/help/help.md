# Micro help text

Micro is a terminal-based text editor that aims to be easy to use and intuitive, 
while also taking advantage of the full capabilities of modern terminals.

### Usage

Once you have built the editor, simply start it by running 
`micro path/to/file.txt` or simply `micro` to open an empty buffer.

Micro also supports creating buffers from stdin:

```
$ ifconfig | micro
```

You can move the cursor around with the arrow keys and mouse.

### Keybindings

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
	"Tab":            "InsertTab",
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

### Possible commands

You can execute an editor command by pressing `Ctrl-e` followed by the command.
Here are the possible commands that you can use.

* `quit`: Quits micro.
* `save`: Saves the current buffer.

* `replace "search" "value" flags`: This will replace `search` with `value`. 
   The `flags` are optional.
   At this point, there is only one flag: `c`, which enables `check` mode 
   which asks if you'd like to perform the replacement each time

   Note that `search` must be a valid regex.  If one of the arguments
   does not have any spaces in it, you may omit the quotes.

* `set option value`: sets the option to value. Please see the next section for
   a list of options you can set.

* `run sh-command`: runs the given shell command in the background. The 
   command's output will be displayed in one line when it finishes running.

* `bind key action`: creates a keybinding from key to action. See the sections on
   keybindings above for more info about what keys and actions are available.

### Options

Micro stores all of the user configuration in its configuration directory.

Micro uses the `$XDG_CONFIG_HOME/micro` as the configuration directory. As per
the XDG spec, if `$XDG_CONFIG_HOME` is not set, `~/.config/micro` is used as 
the config directory.

Here are the options that you can set:

* `colorscheme`: loads the colorscheme stored in 
   $(configDir)/colorschemes/`option`.micro

	default value: `default`
	Note that the default colorschemes (default, solarized, and solarized-tc)
	are not located in configDir, because they are embedded in the micro binary

	The colorscheme can be selected from all the files in the 
	~/.config/micro/colorschemes/ directory. Micro comes by default with three
	colorschemes:

	* default: this is the default colorscheme.

	* solarized: this is the solarized colorscheme (used in the screenshot). 
	  You should have the solarized color palette in your terminal to use it.

	* solarized-tc: this is the solarized colorscheme for true color, just 
	  make sure your terminal supports true color before using it and that the 
	  MICRO_TRUECOLOR environment variable is set to 1 before starting micro.

	* monokai-tc: this is the monokai colorscheme. It requires true color to
	  look perfect, but the 256 color approximation looks good as well.

	* atom-dark-tc: this colorscheme is based off of Atom's "dark" colorscheme.
	  It requires true color to look good.


* `tabsize`: sets the tab size to `option`

	default value: `4`

* `indentchar`: sets the indentation character

	default value: ` `

* `ignorecase`: perform case-insensitive searches

	default value: `off`

* `syntax`: turns syntax on or off

	default value: `on`

* `tabstospaces`: use spaces instead of tabs

	default value: `off`

* `autoindent`: when creating a new line use the same indentation as the 
   previous line

	default value: `on`

* `cursorline`: highlight the line that the cursor is on in a different color
   (the color is defined by the colorscheme you are using)

	default value: `off`

* `ruler`: display line numbers

	default value: `on`

* `statusline`: display the status line at the bottom of the screen

	default value: `on`

* `savecursor`: remember where the cursor was last time the file was opened and
   put it there when you open the file again

	default value: `off`

* `saveundo`: when this option is on, undo is saved even after you close a file
   so if you close and reopen a file, you can keep undoing

	default value: `off`

* `scrollmargin`: amount of lines you would like to see above and below the cursor

	default value: `3`

* `scrollspeed`: amount of lines to scroll for one scroll event

	default value: `2`

---

Default plugin options:

* `linter`: lint languages on save (supported languages are C, D, Go, Java,
   Javascript, Lua). Provided by the `linter` plugin.

	default value: `on`

* `autoclose`: Automatically close `{}` `()` `[]` `""` `''`. Provided by the autoclose plugin

	default value: `on`

* `goimports`: Run goimports on save. Provided by the `go` plugin.

	default value: `off`

* `gofmt`: Run gofmt on save. Provided by the `go` plugin.

	default value: `on`

Any option you set in the editor will be saved to the file 
~/.config/micro/settings.json so, in effect, your configuration file will be 
created for you. If you'd like to take your configuration with you to another
machine, simply copy the settings.json to the other machine.
