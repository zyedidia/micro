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
	are not located in configDir, because they are embedded in the micro binary

	The colorscheme can be selected from all the files in the 
	~/.config/micro/colorschemes/ directory. Micro comes by default with three
	colorschemes:

    You can read more about micro's colorschemes in the `colors` help topic
    (`help colors`).

* `tabsize`: sets the tab size to `option`

	default value: `4`

* `indentchar`: sets the indentation character

	default value: ` `

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
