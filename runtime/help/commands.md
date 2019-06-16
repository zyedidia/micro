# Possible commands

You can execute an editor command by pressing `Ctrl-e` followed by the command.
Here are the possible commands that you can use.

* `quit`: Quits micro.

* `save filename?`: Saves the current buffer. If the filename is provided it
  will 'save as' the filename.

* `replace "search" "value" flags`: This will replace `search` with `value`. 
   The `flags` are optional. Possible flags are:
   * `-a`: Replace all occurrences at once
   * `-l`: Do a literal search instead of a regex search

   Note that `search` must be a valid regex (unless `-l` is passed). If one 
   of the arguments does not have any spaces in it, you may omit the quotes.

* `replaceall "search" "value"`: This will replace `search` with `value` without
   user confirmation.

	See `replace` command for more information.

* `set option value`: sets the option to value. See the `options` help topic for
   a list of options you can set.

* `setlocal option value`: sets the option to value locally (only in the current
   buffer).

* `show option`: shows the current value of the given option.

* `eval "expression"`: Evaluates a Lua expression. Note that micro will not
   print anything so you should use `messenger:Message(...)` to display a value.

* `run sh-command`: runs the given shell command in the background. The 
   command's output will be displayed in one line when it finishes running.

* `bind key action`: creates a keybinding from key to action. See the sections
   on keybindings above for more info about what keys and actions are available.

* `vsplit filename`: opens a vertical split with `filename`. If no filename is
   provided, a vertical split is opened with an empty buffer.

* `hsplit filename`: same as `vsplit` but opens a horizontal split instead of a
   vertical split.

* `tab filename`: opens the given file in a new tab.

* `tabswitch tab`: This command will switch to the specified tab. The `tab` can
   either be a tab number, or a name of a tab.
					 
* `log`: opens a log of all messages and debug statements.

* `plugin install plugin_name`: installs the given plugin.

* `plugin remove plugin_name`: removes the given plugin.

* `plugin list`: lists all installed plugins.

* `plugin update`: updates all installed plugins.

* `plugin search plugin_name`: searches for the given plugin. Note that you can
   find a list of all available plugins at
   github.com/micro-editor/plugin-channel.

   You can also see more information about the plugin manager in the
   `Plugin Manager` section of the `plugins` help topic.

* `plugin available`: list plugins available for download (this includes any
   plugins that may be already installed).

* `reload`: reloads all runtime files.

* `cd path`: Change the working directory to the given `path`.

* `pwd`: Print the current working directory.

* `open filename`: Open a file in the current buffer.

* `reset option`: resets the given option to its default value

* `retab`: Replaces all leading tabs with spaces or leading spaces with tabs
   depending on the value of `tabstospaces`.

* `raw`: Micro will open a new tab and show the escape sequence for every event
   it receives from the terminal. This shows you what micro actually sees from
   the terminal and helps you see which bindings aren't possible and why. This
   is most useful for debugging keybindings.

* `showkey`: Show the action(s) bound to a given key. For example
   running `> showkey CtrlC` will display `main.(*View).Copy`. Unfortuately
   showkey does not work well for keys bound to plugin actions. For those
   it just shows "LuaFunctionBinding."

---

The following commands are provided by the default plugins:

* `lint`: Lint the current file for errors.

# Command Parsing

When running a command, you can use extra syntax that micro will expand before
running the command. To use an argument with a space in it, simply put it in
quotes. You can also use environment variables in the command bar and they
will be expanded to their value. Finally, you can put an expression in backticks
and it will be evaluated by the shell beforehand.
