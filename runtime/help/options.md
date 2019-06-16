# Options

Micro stores all of the user configuration in its configuration directory.

Micro uses the `$XDG_CONFIG_HOME/micro` as the configuration directory. As per
the XDG spec, if `$XDG_CONFIG_HOME` is not set, `~/.config/micro` is used as 
the config directory.

Here are the options that you can set:

* `autoindent`: when creating a new line, use the same indentation as the 
   previous line.

	default value: `true`

* `autosave`: micro will save the buffer every 8 seconds automatically. Micro
   also will automatically save and quit when you exit without asking. Be
   careful when using this feature, because you might accidentally save a file,
   overwriting what was there before.

	default value: `false`

* `basename`: in the infobar, show only the basename of the file being edited
   rather than the full path.

    default value: `false`

* `colorcolumn`: if this is not set to 0, it will display a column at the
  specified column. This is useful if you want column 80 to be highlighted
  special for example.

	default value: `0`

* `colorscheme`: loads the colorscheme stored in 
   $(configDir)/colorschemes/`option`.micro, This setting is `global only`.

	default value: `default`

	Note that the default colorschemes (default, solarized, and solarized-tc)
	are not located in configDir, because they are embedded in the micro binary.

	The colorscheme can be selected from all the files in the 
	~/.config/micro/colorschemes/ directory. Micro comes by default with three
	colorschemes:

	You can read more about micro's colorschemes in the `colors` help topic
	(`help colors`).

* `cursorline`: highlight the line that the cursor is on in a different color
   (the color is defined by the colorscheme you are using).

	default value: `true`

* `encoding`: the encoding to open and save files with. Supported encodings
   are listed at https://www.w3.org/TR/encoding/.

    default value: `utf-8`

* `eofnewline`: micro will automatically add a newline to the file.

	default value: `false`

* `fastdirty`: this determines what kind of algorithm micro uses to determine if
   a buffer is modified or not. When `fastdirty` is on, micro just uses a
   boolean `modified` that is set to `true` as soon as the user makes an edit.
   This is fast, but can be inaccurate. If `fastdirty` is off, then micro will
   hash the current buffer against a hash of the original file (created when the
   buffer was loaded). This is more accurate but obviously more resource
   intensive. This option is only for people who really care about having
   accurate modified status.

	default value: `true`

* `fileformat`: this determines what kind of line endings micro will use for the
   file. UNIX line endings are just `\n` (linefeed) whereas dos line endings are
   `\r\n` (carriage return + linefeed). The two possible values for this option
   are `unix` and `dos`. The fileformat will be automatically detected (when you
   open an existing file) and displayed on the statusline, but this option is
   useful if you would like to change the line endings or if you are starting a
   new file.

	default value: `unix`

* `filetype`: sets the filetype for the current buffer. This setting is
   `local only`.

	default value: this will be automatically set depending on the file you have
	open

* `ignorecase`: perform case-insensitive searches.

	default value: `false`

* `indentchar`: sets the indentation character.

	default value: ` ` (space)

* `infobar`: enables the line at the bottom of the editor where messages are
   printed. This option is `global only`.

	default value: `true`

* `keepautoindent`: when using autoindent, whitespace is added for you. This
   option determines if when you move to the next line without any insertions
   the whitespace that was added should be deleted to remove trailing whitespace.
   By default, the autoindent whitespace is deleted if the line was left empty.

	default value: `false`

* `keymenu`: display the nano-style key menu at the bottom of the screen. Note
   that ToggleKeyMenu is bound to `Alt-g` by default and this is displayed in
   the statusline. To disable this, simply by `Alt-g` to `UnbindKey`.

	default value: `false`

* `matchbrace`: highlight matching braces for '()', '{}', '[]'

    default value: `false`

* `matchbraceleft`: when matching a closing brace, should matching match the
   brace directly under the cursor, or the character to the left? only matters
   if `matchbrace` is true

    default value: `false`

* `mouse`: whether to enable mouse support. When mouse support is disabled,
   usually the terminal will be able to access mouse events which can be useful
   if you want to copy from the terminal instead of from micro (if over ssh for
   example, because the terminal has access to the local clipboard and micro
   does not).

	default value: `true`

* `pluginchannels`: contains all the channels micro's plugin manager will search
   for plugins in. A channel is simply a list of 'repository' json files which
   contain metadata about the given plugin. See the `Plugin Manager` section of
   the `plugins` help topic for more information.

	default value: `https://github.com/micro-editor/plugin-channel`

* `pluginrepos`: contains all the 'repositories' micro's plugin manager will
   search for plugins in. A repository consists of a `repo.json` file which
   contains metadata for a single plugin.

	default value: ` `

* `rmtrailingws`: micro will automatically trim trailing whitespaces at eol.

	default value: `false`

* `ruler`: display line numbers.

	default value: `true`

* `savecursor`: remember where the cursor was last time the file was opened and
   put it there when you open the file again.

	default value: `false`

* `savehistory`: remember command history between closing and re-opening
   micro.

    default value: `true`

* `saveundo`: when this option is on, undo is saved even after you close a file
   so if you close and reopen a file, you can keep undoing.

	default value: `false`

* `scrollbar`: display a scroll bar

    default value: `false`

* `scrollmargin`: amount of lines you would like to see above and below the
   cursor.

	default value: `3`

* `scrollspeed`: amount of lines to scroll for one scroll event.

	default value: `2`

* `smartpaste`: should micro add leading whitespace when pasting multiple lines?
   This will attempt to preserve the current indentation level when pasting an
   unindented block.

	default value: `true`

* `softwrap`: should micro wrap lines that are too long to fit on the screen.

	default value: `false`

* `splitbottom`: when a horizontal split is created, should it be created below
   the current split?

	default value: `true`

* `splitright`: when a vertical split is created, should it be created to the
   right of the current split?

	default value: `true`

* `statusformatl`: format string definition for the left-justified part of the
   statusline. Special directives should be placed inside `$()`. Special
   directives include: `filename`, `modified`, `line`, `col`, `opt`, `bind`.
   The `opt` and `bind` directives take either an option or an action afterward
   and fill in the value of the option or the key bound to the action.

    default value: `$(filename) $(modified)($(line),$(col)) $(opt:filetype)
                    $(opt:fileformat) $(opt:encoding)`

* `statusformatl`: format string definition for the left-justified part of the
   statusline.

    default value: `$(bind:ToggleKeyMenu): show bindings, $(bind:ToggleHelp):
                    toggle help`

* `statusline`: display the status line at the bottom of the screen.

	default value: `true`

* `sucmd`: specifies the super user command. On most systems this is "sudo" but
   on BSD it can be "doas." This option can be customized and is only used when
   saving with su.

	default value: `sudo`

* `syntax`: turns syntax on or off.

	default value: `true`

* `tabmovement`: navigate spaces at the beginning of lines as if they are tabs
   (e.g. move over 4 spaces at once). This option only does anything if
   `tabstospaces` is on.

	default value: `false`

* `tabsize`: sets the tab size to `option`

	default value: `4`

* `tabstospaces`: use spaces instead of tabs

	default value: `false`

* `termtitle`: defines whether or not your terminal's title will be set by micro
   when opened.

	default value: `false`

* `useprimary` (only useful on *nix): defines whether or not micro will use the
   primary clipboard to copy selections in the background. This does not affect
   the normal clipboard using Ctrl-C and Ctrl-V.

	default value: `true`

---

Default plugin options:

* `autoclose`: automatically close `{}` `()` `[]` `""` `''`. Provided by the
   `autoclose` plugin

	default value: `true`

* `ftoptions`: by default, micro will set some options based on the filetype. At
   the moment, micro will use tabs for makefiles and spaces for python and yaml
   files regardless of your settings. If you would like to disable this behavior
   turn this option off.

	default value: `true`

* `linter`: Automatically lint when the file is saved. Provided by the `linter`
   plugin.

	default value: `true`

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
