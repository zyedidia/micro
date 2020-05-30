# Options

Micro stores all of the user configuration in its configuration directory.

Micro uses `$MICRO_CONFIG_HOME` as the configuration directory. If this
environment variable is not set, it uses `$XDG_CONFIG_HOME/micro` instead. If
that environment variable is not set, it uses `~/.config/micro` as the
configuration directory. In the documentation, we use `~/.config/micro` to
refer to the configuration directory (even if it may in fact be somewhere else
if you have set either of the above environment variables).

Here are the available options:

* `autoindent`: when creating a new line, use the same indentation as the 
   previous line.

	default value: `true`

* `autosave`: automatically save the buffer every n seconds, where n is the
   value of the autosave option. Also when quitting on a modified buffer, micro
   will automatically save and quit. Be warned, this option saves the buffer
   without prompting the user, so data may be overwritten. If this option is
   set to `0`, no autosaving is performed.

    default value: `0`

* `autosu`: When a file is saved that the user doesn't have permission to
   modify, micro will ask if the user would like to use super user
   privileges to save the file. If this option is enabled, micro will
   automatically attempt to use super user privileges to save without
   asking the user.

    default value: `false`

* `backup`: micro will automatically keep backups of all open buffers. Backups
   are stored in `~/.config/micro/backups` and are removed when the buffer is
   closed cleanly. In the case of a system crash or a micro crash, the contents
   of the buffer can be recovered automatically by opening the file that was
   being edited before the crash, or manually by searching for the backup in
   the backup directory. Backups are made in the background when a buffer is
   modified and the latest backup is more than 8 seconds old, or when micro
   detects a crash. It is highly recommended that you leave this feature
   enabled.

    default value: `true`

* `basename`: in the infobar and tabbar, show only the basename of the file
   being edited rather than the full path.

    default value: `false`

* `colorcolumn`: if this is not set to 0, it will display a column at the
  specified column. This is useful if you want column 80 to be highlighted
  special for example.

	default value: `0`

* `colorscheme`: loads the colorscheme stored in 
   $(configDir)/colorschemes/`option`.micro, This setting is `global only`.

	default value: `default`

    Note that the default colorschemes (default, solarized, and solarized-tc)
    are not located in configDir, because they are embedded in the micro
    binary.

	The colorscheme can be selected from all the files in the 
	~/.config/micro/colorschemes/ directory. Micro comes by default with three
	colorschemes:

	You can read more about micro's colorschemes in the `colors` help topic
	(`help colors`).

* `cursorline`: highlight the line that the cursor is on in a different color
   (the color is defined by the colorscheme you are using).

	default value: `true`

* `diffgutter`: display diff indicators before lines.

	default value: `false`

* `divchars`: specifies the "divider" characters used for the dividing line
   between vertical/horizontal splits. The first character is for vertical
   dividers, and the second is for horizontal dividers. By default, for
   horizontal splits the statusline serves as a divider, but if the statusline
   is disabled the horizontal divider character will be used.

    default value: `|-`

* `divreverse`: colorschemes provide the color (foreground and background) for
   the characters displayed in split dividers. With this option enabled, the
   colors specified by the colorscheme will be reversed (foreground and
   background colors swapped).

    default value: `true`

* `encoding`: the encoding to open and save files with. Supported encodings
   are listed at https://www.w3.org/TR/encoding/.

    default value: `utf-8`

* `eofnewline`: micro will automatically add a newline to the end of the
   file if one does not exist.

	default value: `true`

* `fastdirty`: this determines what kind of algorithm micro uses to determine
   if a buffer is modified or not. When `fastdirty` is on, micro just uses a
   boolean `modified` that is set to `true` as soon as the user makes an edit.
   This is fast, but can be inaccurate. If `fastdirty` is off, then micro will
   hash the current buffer against a hash of the original file (created when
   the buffer was loaded). This is more accurate but obviously more resource
   intensive. This option will be automatically disabled if the file size
   exceeds 50KB.

	default value: `false`

* `fileformat`: this determines what kind of line endings micro will use for
   the file. UNIX line endings are just `\n` (linefeed) whereas dos line
   endings are `\r\n` (carriage return + linefeed). The two possible values for
   this option are `unix` and `dos`. The fileformat will be automatically
   detected (when you open an existing file) and displayed on the statusline,
   but this option is useful if you would like to change the line endings or if
   you are starting a new file.

	default value: `unix`

* `filetype`: sets the filetype for the current buffer. Set this option to
  `off` to completely disable filetype detection.

	default value: `unknown`. This will be automatically overridden depending
    on the file you open.

* `ignorecase`: perform case-insensitive searches.

	default value: `false`

* `indentchar`: sets the indentation character.

	default value: ` ` (space)

* `infobar`: enables the line at the bottom of the editor where messages are
   printed. This option is `global only`.

	default value: `true`

* `keepautoindent`: when using autoindent, whitespace is added for you. This
   option determines if when you move to the next line without any insertions
   the whitespace that was added should be deleted to remove trailing
   whitespace.  By default, the autoindent whitespace is deleted if the line
   was left empty.

	default value: `false`

* `keymenu`: display the nano-style key menu at the bottom of the screen. Note
   that ToggleKeyMenu is bound to `Alt-g` by default and this is displayed in
   the statusline. To disable this, simply by `Alt-g` to `UnbindKey`.

	default value: `false`

* `matchbrace`: underline matching braces for '()', '{}', '[]' when the cursor
   is on a brace character.

    default value: `true`

* `mkparents`: if a file is opened on a path that does not exist, the file
   cannot be saved because the parent directories don't exist. This option lets
   micro automatically create the parent directories in such a situation.

    default value: `false`

* `mouse`: mouse support. When mouse support is disabled,
   usually the terminal will be able to access mouse events which can be useful
   if you want to copy from the terminal instead of from micro (if over ssh for
   example, because the terminal has access to the local clipboard and micro
   does not).

	default value: `true`

* `paste`: treat characters sent from the terminal in a single chunk as a paste
   event rather than a series of manual key presses. If you are pasting using
   the terminal keybinding (not Ctrl-v, which is micro's default paste
   keybinding) then it is a good idea to enable this option during the paste
   and disable once the paste is over. See `> help copypaste` for details about
   copying and pasting in a terminal environment.

    default value: `false`

* `parsecursor`: if enabled, this will cause micro to parse filenames such as
   file.txt:10:5 as requesting to open `file.txt` with the cursor at line 10
   and column 5. The column number can also be dropped to open the file at a
   given line and column 0. Note that with this option enabled it is not possible
   to open a file such as `file.txt:10:5`, where `:10:5` is part of the filename.
   It is also possible to open a file with a certain cursor location by using the
   `+LINE,COL` flag syntax. See `micro -help` for the command line options.

    default value: `false`

* `pluginchannels`: list of URLs pointing to plugin channels for downloading and
   installing plugins. A plugin channel consists of a json file with links to
   plugin repos, which store information about plugin versions and download URLs.
   By default, this option points to the official plugin channel hosted on GitHub
   at https://github.com/micro-editor/plugin-channel.

    default value: `https://raw.githubusercontent.com/micro-editor/plugin-channel
                    /master/channel.json`

* `pluginrepos`: a list of links to plugin repositories.

    default value: ``

* `readonly`: when enabled, disallows edits to the buffer. It is recommended
   to only ever set this option locally using `setlocal`.

    default value: `false`

* `rmtrailingws`: micro will automatically trim trailing whitespaces at ends of
   lines.

	default value: `false`

* `ruler`: display line numbers.

	default value: `true`

* `relativeruler`: make line numbers display relatively. If set to true, all lines except
	for the line that the cursor is located will display the distance from the 
	cursor's line. 

	default value: `false` 

* `savecursor`: remember where the cursor was last time the file was opened and
   put it there when you open the file again. Information is saved to
   `~/.config/micro/buffers/`

	default value: `false`

* `savehistory`: remember command history between closing and re-opening
   micro. Information is saved to `~/.config/micro/buffers/history`.

    default value: `true`

* `saveundo`: when this option is on, undo is saved even after you close a file
   so if you close and reopen a file, you can keep undoing. Information is
   saved to `~/.config/micro/buffers/`.

	default value: `false`

* `scrollbar`: display a scroll bar

    default value: `false`

* `scrollmargin`: margin at which the view starts scrolling when the cursor
   approaches the edge of the view.

	default value: `3`

* `scrollspeed`: amount of lines to scroll for one scroll event.

	default value: `2`

* `smartpaste`: add leading whitespace when pasting multiple lines.
   This will attempt to preserve the current indentation level when pasting an
   unindented block.

	default value: `true`

* `softwrap`: wrap lines that are too long to fit on the screen.

	default value: `false`

* `splitbottom`: when a horizontal split is created, create it below the
   current split.

	default value: `true`

* `splitright`: when a vertical split is created, create it to the right of the
   current split.

	default value: `true`

* `statusformatl`: format string definition for the left-justified part of the
   statusline. Special directives should be placed inside `$()`. Special
   directives include: `filename`, `modified`, `line`, `col`, `opt`, `bind`.
   The `opt` and `bind` directives take either an option or an action afterward
   and fill in the value of the option or the key bound to the action.

    default value: `$(filename) $(modified)($(line),$(col)) $(status.paste)|
                    ft:$(opt:filetype) | $(opt:fileformat) | $(opt:encoding)`

* `statusformatr`: format string definition for the right-justified part of the
   statusline.

    default value: `$(bind:ToggleKeyMenu): bindings, $(bind:ToggleHelp): help`

* `statusline`: display the status line at the bottom of the screen.

	default value: `true`

* `sucmd`: specifies the super user command. On most systems this is "sudo" but
   on BSD it can be "doas." This option can be customized and is only used when
   saving with su.

	default value: `sudo`

* `syntax`: enables syntax highlighting.

	default value: `true`

* `tabmovement`: navigate spaces at the beginning of lines as if they are tabs
   (e.g. move over 4 spaces at once). This option only does anything if
   `tabstospaces` is on.

	default value: `false`

* `tabsize`: the size in spaces that a tab character should be displayed with.

	default value: `4`

* `tabstospaces`: use spaces instead of tabs.

	default value: `false`

* `useprimary` (only useful on unix): defines whether or not micro will use the
   primary clipboard to copy selections in the background. This does not affect
   the normal clipboard using Ctrl-c and Ctrl-v.

	default value: `true`

* `xterm`: micro will assume that the terminal it is running in conforms to
  `xterm-256color` regardless of what the `$TERM` variable actually contains.
   Enabling this option may cause unwanted effects if your terminal in fact
   does not conform to the `xterm-256color` standard.

    Default value: `false`

---

Plugin options: all plugins come with a special option to enable or disable
them. The option is a boolean with the same name as the plugin itself.

By default, the following plugins are provided, each with an option to enable
or disable them:

* `autoclose`: automatically closes brackets, quotes, etc...
* `comment`: provides automatic commenting for a number of languages
* `ftoptions`: alters some default options depending on the filetype
* `linter`: provides extensible linting for many languages
* `literate`: provides advanced syntax highlighting for the Literate
   programming tool.
* `status`: provides some extensions to the status line (integration with
   Git and more).
* `diff`: integrates the `diffgutter` option with Git. If you are in a Git
   directory, the diff gutter will show changes with respect to the most
   recent Git commit rather than the diff since opening the file.

Any option you set in the editor will be saved to the file
~/.config/micro/settings.json so, in effect, your configuration file will be 
created for you. If you'd like to take your configuration with you to another
machine, simply copy the settings.json to the other machine.

## Global and local settings

You can set these settings either globally or locally. Locally means that the
setting won't be saved to `~/.config/micro/settings.json` and that it will only
be set in the current buffer. Setting an option globally is the default, and
will set the option in all buffers. Use the `setlocal` command to set an option
locally rather than globally.

The `colorscheme` option is global only, and the `filetype` option is local
only. To set an option locally, use `setlocal` instead of `set`.

In the `settings.json` file you can also put set options locally by specifying
either a glob or a filetype. Here is an example which has `tabstospaces` on for
all files except Go files, and `tabsize` 4 for all files except Ruby files:

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
