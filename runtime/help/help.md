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

These are the default keybindings, along with their actions.


#### Editor bindings

* Ctrl-q:   Quit
* Ctrl-e:   Execute a command
* Ctrl-g:   Toggle help text
* Ctrl-b:   Run a shell command

#### Buffer bindings

* Ctrl-s:   Save
* Ctrl-o:   Open file
* Ctrl-z:   Undo
* Ctrl-y:   Redo
* Ctrl-f:   Find
* Ctrl-n:   Find next
* Ctrl-p:   Find previous
* Ctrl-a:   Select all
* Ctrl-c:   Copy
* Ctrl-x:   Cut
* Ctrl-k:   Cut line
* Ctrl-v:   Paste
* Ctrl-u:   Half page up
* Ctrl-d:   Half page down
* PageUp:   Page up
* PageDown: Page down
* Home:     Go to beginning of line
* End:      Go to end of line
* Ctrl-r:   Toggle line numbers

You can use the alt keys + arrows to move word by word.
Ctrl left and right move the cursor to the start and end of the line, and
ctrl up and down move the cursor the start and end of the buffer.

You can hold shift with all of these movement actions to select while moving.

The buffer bindings may be rebound using the `~/.config/micro/bindings.json` 
file. Each key is bound to an action.

For example, to bind `Ctrl-y` to undo and `Ctrl-z` to redo, you could put the 
following in the `bindings.json` file.

```json
{
    "CtrlY": "Undo",
    "CtrlZ": "Redo"
}
```

Here are the defaults:

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
		"CtrlV":          "Paste",
		"CtrlA":          "SelectAll",
		"Home":           "Start",
		"End":            "End",
		"PgUp":           "PageUp",
		"PgDn":           "PageDown",
		"CtrlU":          "HalfPageUp",
		"CtrlD":          "HalfPageDown",
		"CtrlR":          "ToggleRuler",
		"Delete":         "Delete"
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


* `tabsize`: sets the tab size to `option`

	default value: `4`

* `syntax`: turns syntax on or off

	default value: `on`

* `tabsToSpaces`: use spaces instead of tabs

	default value: `off`

* `autoindent`: when creating a new line use the same indentation as the 
   previous line

    default value: `on`

* `ruler`: display line numbers

    default value: `on`

* `statusline`: display the status line at the bottom of the screen

    default value: `on`

* `scrollspeed`: amount of lines to scroll

	default value: `2`

---

Default plugin options:

* `linter`: lint languages on save (supported languages are C, D, Go, Java,
   Javascript, Lua). Provided by the `linter` plugin.

    default value: `on`

* `goimports`: Run goimports on save. Provided by the `go` plugin.

    default value: `off`

* `gofmt`: Run gofmt on save. Provided by the `go` plugin.

    default value: `on`

Any option you set in the editor will be saved to the file 
~/.config/micro/settings.json so, in effect, your configuration file will be 
created for you. If you'd like to take your configuration with you to another
machine, simply copy the settings.json to the other machine.
