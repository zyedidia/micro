# Micro help text

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

* Home:     Go to beginning of file
* End:      Go to end of file

* Ctrl-r:   Toggle line numbers

The buffer bindings may be rebound using the `~/.config/micro/bindings.json` file. Each key is bound to an action.

For example, to bind `Ctrl-y` to undo and `Ctrl-z` to redo, you could put the following in the `bindings.json` file.

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

`replace "search" "value"`: This will replace `search` with `value`.
Note that `search` must be a valid regex.  If one of the arguments
does not have any spaces in it, you may omit the quotes.

`set option value`: sets the option to value. Please see the next section for a list of options you can set.

`run sh-command`: runs the given shell command in the background. The command's output will be displayed
in one line when it finishes running.

### Options

Micro stores all of the user configuration in its configuration directory.

Micro uses the `$XDG_CONFIG_HOME/micro` as the configuration directory. As per the XDG spec,
if `$XDG_CONFIG_HOME` is not set, `~/.config/micro` is used as the config directory.

Here are the options that you can set:

`colorscheme`: loads the colorscheme stored in $(configDir)/colorschemes/`option`.micro
	default value: `default`
	Note that the default colorschemes (default, solarized, and solarized-tc) are not located in configDir,
	because they are embedded in the micro binary

`tabsize`: sets the tab size to `option`
	default value: `4`

`syntax`: turns syntax on or off
	default value: `on`

`tabsToSpaces`: use spaces instead of tabs
	default value: `off`

`autoindent`: when creating a new line use the same indentation as the previous line
    default value: `on`

`ruler`: display line numbers
    default value: `on`

`gofmt`: Run `gofmt` whenever the file is saved (this only applies to `.go` files)
    default value: `off`

`goimports`: run `goimports` whenever the file is saved (this only applies to `.go` files)
    default value: `off`

In the future, the `gofmt` and `goimports` will be refactored using a plugin system. However,
currently they just make it easier to program micro in micro.
