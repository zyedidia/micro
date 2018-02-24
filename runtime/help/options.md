# Options

Micro stores all of the user configuration in its configuration directory.

Micro uses the `$XDG_CONFIG_HOME/micro` as the configuration directory. As per
the XDG spec, if `$XDG_CONFIG_HOME` is not set, `~/.config/micro` is used as 
the config directory.

Here are the options that you can set:

Option | Default Value | Description
--- | --- | ---
`autoindent` | `true` | When creating a new line use the same indentation as the previous line.
`autosave` | `false` | micro will save the buffer every 8 seconds automatically. Micro will also automatically save and quit when you exit without asking. Be careful when using this feature, because you might accidentally save a file, overwriting what was there before.
`basename` | `false` | in the infobar, show only the basename of the file being edited rather than the full path.
`colorcolumn` | `0` | if this is not set to 0, it will display a column at the specified column. This is useful if you want column 80 to be highlighted special for example.
`colorscheme` | `default` |  loads the colorscheme stored in `$(configDir)/colorschemes/option`.micro, This setting is `global only`. Note that the default colorschemes (default, solarized, and solarized-tc) are not located in configDir, because they are embedded in the micro binary. The colorscheme can be selected from all the files in the `~/.config/micro/colorschemes/` directory. You can read more about micro's colorschemes in the `colors` help topic (`help colors`).
`cursorline` | `true` |  highlight the line that the cursor is on in a different color (the color is defined by the colorscheme you are using).
`eofnewline` | `false` |  micro will automatically add a newline to the file.
`fastdirty` | `true` |  this determines what kind of algorithm micro uses to determine if a buffer is modified or not. When `fastdirty` is on, micro just uses a boolean `modified` that is set to `true` as soon as the user makes an edit. This is fast, but can be inaccurate. If `fastdirty` is off, then micro will hash the current buffer against a hash of the original file (created when the buffer was loaded). This is more accurate but obviously more resource intensive. This option is only for people who really care about having accurate modified status.
`fileformat` | `unix` |  this determines what kind of line endings micro will use for the file. UNIX line endings are just `\n` (lf) whereas dos line endings are `\r\n` (crlf). The two possible values for this option are `unix` and `dos`. The fileformat will be automatically detected and displayed on the statusline but this option is useful if you would like to change the line endings or if you are starting a new file.
`filetype` | Automatically set depending on the file you have open |  sets the filetype for the current buffer. This setting is `local only`.
`ignorecase` | `false` |  perform case-insensitive searches.
`indentchar` | ` ` |  sets the indentation character.
`infobar` | `true` |  enables the line at the bottom of the editor where messages are printed. This option is `global only`.
`keepautoindent` | `false` |  when using autoindent, whitespace is added for you. This option determines if when you move to the next line without any insertions the whitespace that was added should be deleted. By default the autoindent whitespace is deleted if the line was left empty.
`keymenu` | `false` |  display the nano-style key menu at the bottom of the screen. Note that ToggleKeyMenu is bound to `Alt-g` by default and this is displayed in the statusline. To disable this, simply by `Alt-g` to `UnbindKey`.
`mouse` | `true` |  whether to enable mouse support. When mouse support is disabled, usually the terminal will be able to access mouse events which can be useful if you want to copy from the terminal instead of from micro (if over ssh for example, because the terminal has access to the local clipboard and micro does not).
`pluginchannels` | `ttps://github.com/micro-editor/plugin-channel` |  contains all the channels micro's plugin manager will search for plugins in. A channel is simply a list of 'repository' json files which contain metadata about the given plugin. See the `Plugin Manager` section of the `plugins` help topic for more information.
`pluginrepos` | ` ` |  contains all the 'repositories' micro's plugin manager will search for plugins in. A repository consists of a `repo.json` file which contains metadata for a single plugin.
`rmtrailingws` | `false` |  micro will automatically trim trailing whitespaces at eol.
`ruler` | `true` |  display line numbers.
`savecursor` | `false` |  remember where the cursor was last time the file was opened and put it there when you open the file again.
`savehistory` | `true` |  remember command history between closing and re-opening micro.
`saveundo` | `false` |  when this option is on, undo is saved even after you close a file so if you close and reopen a file, you can keep undoing.
`scrollbar` | `false` |  display a scroll bar
`scrollmargin | `3` |  amount of lines you would like to see above and below the cursor.
`scrollspeed` | `2` |  amount of lines to scroll for one scroll event.
`softwrap` | `false` |  should micro wrap lines that are too long to fit on the screen.
`splitbottom` | `true` |  when a horizontal split is created, should it be created below the current split?
`splitright` | `true` |  when a vertical split is created, should it be created to the right of the current split?
`statusline` | `true` |  display the status line at the bottom of the screen.
`matchbrace` | `false` |  highlight matching braces for '()', '{}', '[]'
`syntax` | `true` |  turns syntax on or off.
`sucmd` | `sudo` |  specifies the super user command. On most systems this is "sudo" but on BSD it can be "doas." This option can be customized and is only used when saving with su.
`tabmovement` | `false` |  navigate spaces at the beginning of lines as if they are tabs (e.g. move over 4 spaces at once). This option only does anything if `tabstospaces` is on.
`tabsize` | `4` |  sets the tab size to `option`
`tabstospaces`: | `false` | use spaces instead of tabs
`termtitle` | `false` |  defines whether or not your terminal's title will be set by micro when opened.
`useprimary` | `true` | (only useful on *nix) defines whether or not micro will use the primary clipboard to copy selections in the background. This does not affect the normal clipboard using Ctrl-C and Ctrl-V.

---

Default plugin options:

Option | Default Value | Description
--- | --- | ---
`autoclose` | `true` |  automatically close `{}` `()` `[]` `""` `''`. Provided by the `autoclose` plugin.
`ftoptions` | `true` |  by default, micro will set some options based on the filetype. At the moment, micro will use tabs for makefiles and spaces for python and yaml files regardless of your settings. If you would like to disable this behavior turn this option off.
`linter` | `true` |  Automatically lint when the file is saved. Provided by the `linter` plugin.

Any option you set in the editor will be saved to the file 
~/.config/micro/settings.json so, in effect, your configuration file will be 
created for you. If you'd like to take your configuration with you to another
machine, simply copy the settings.json to the other machine.


## Global and local settings

You can set these settings either globally or locally. Locally means that the
setting won't be saved to `~/.config/micro/settings.json` and that it will only
be set in the current buffer. Setting an option globally is the default, and
will set the option in all buffers.

The `colorscheme` option is global only, and the `filetype` option is local
only. To set an option locally, use `setlocal` instead of `set`.

In the `settings.json` file you can also put set options locally by specifying either
a glob or a filetype. Here is an example which has `tabstospaces` on for all files except Go
files, and `tabsize` 4 for all files except Ruby files:

```json
{
	"ft:go": {
		"tabstospaces": false
	},
	"ft:ruby": {
		"tabsize": 2
	},
	"tabstospaces": true,
	"tabsize": 4
}
```

Or similarly you can match with globs:

```json
{
	"*.go": {
		"tabstospaces": false
	},
	"*.rb": {
		"tabsize": 2
	},
	"tabstospaces": true,
	"tabsize": 4
}
```
