### Options

Micro stores all of the user configuration in its configuration directory.

Micro uses the `$XDG_CONFIG_HOME/micro` as the configuration directory. As per
the XDG spec, if `$XDG_CONFIG_HOME` is not set, `~/.config/micro` is used as 
the config directory.

Here are the options that you can set:

* `colorscheme`: loads the colorscheme stored in 
   $(configDir)/colorschemes/`option`.micro
   This setting is `global only`.

	default value: `default`
	Note that the default colorschemes (default, solarized, and solarized-tc)
	are not located in configDir, because they are embedded in the micro binary.

	The colorscheme can be selected from all the files in the 
	~/.config/micro/colorschemes/ directory. Micro comes by default with three
	colorschemes:

    You can read more about micro's colorschemes in the `colors` help topic
    (`help colors`).

* `colorcolumn`: if this is not set to 0, it will display a column at the specified
   column. This is useful if you want column 80 to be highlighted special for example.

	default value: `0`

* `eofnewline`: micro will automatically add a newline to the file.

	default value: `false`

* `rmtrailingws`: micro will automatically trim trailing whitespaces at eol.

	default value: `false`

* `tabsize`: sets the tab size to `option`

	default value: `4`

* `indentchar`: sets the indentation character

	default value: ` `

* `infobar`: enables the line at the bottom of the editor where messages are printed.
   This option is `global only`.

	default value: `on`

* `filetype`: sets the filetype for the current buffer. This setting is `local only`

    default value: this will be automatically set depending on the file you have open

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

	default value: `on`

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

* `softwrap`: should micro wrap lines that are too long to fit on the screen

    default value: `off`

* `splitRight`: when a vertical split is created, should it be created to the right of
   the current split?

    default value: `on`

* `splitRight`: when a horizontal split is created, should it be created below the
   current split?

    default value: `on`

* `autosave`: micro will save the buffer every 8 seconds automatically.
   Micro also will automatically save and quit when you exit without asking.
   Be careful when using this feature, because you might accidentally save a file,
   overwriting what was there before.

	default value: `off`

* `pluginchannels`: contains all the channels micro's plugin manager will search
   for plugins in. A channel is simply a list of 'repository' json files which contain
   metadata about the given plugin. See the `Plugin Manager` section of the `plugins` help topic
   for more information.

    default value: `https://github.com/micro-editor/plugin-channel`

* `pluginrepos`: contains all the 'repositories' micro's plugin manager will search for
   plugins in. A repository consists of a `repo.json` file which contains metadata for a
   single plugin.

    default value: ` `

---

Default plugin options:

* `autoclose`: Automatically close `{}` `()` `[]` `""` `''`. Provided by the `autoclose` plugin

	default value: `on`

* `linter`: Automatically lint when the file is saved. Provided by the `linter` plugin

	default value: `on`

Any option you set in the editor will be saved to the file 
~/.config/micro/settings.json so, in effect, your configuration file will be 
created for you. If you'd like to take your configuration with you to another
machine, simply copy the settings.json to the other machine.

# Global and local settings

You can set these settings either globally or locally. Locally means that the setting
won't be saved to `~/.config/micro/settings.json` and that it will only be set in
the current buffer. Setting an option globally is the default, and will set the option
in all buffers.

The `colorscheme` option is global only, and the `filetype` option is local only. To
set an option locally, use `setlocal` instead of `set`.

In the `settings.json` file you can also put set options locally by specifying a glob.
Here is an example which has `tabstospaces` on for all files except Go files, and
`tabsize` 4 for all files except Ruby files:

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

As you can see it is quite easy to set options locally using the `settings.json` file.
