# Possible commands

You can execute an editor command by pressing `Ctrl-e` followed by the command.
Here are the possible commands that you can use.

* `quit`: Quits micro.

* `save filename?`: Saves the current buffer. If the filename is provided it will
   'save as' the filename.

* `replace "search" "value" flags`: This will replace `search` with `value`. 
   The `flags` are optional.
   At this point, there is only one flag: `-c`, which enables `check` mode 
   which asks if you'd like to perform the replacement each time.

   Note that `search` must be a valid regex.  If one of the arguments
   does not have any spaces in it, you may omit the quotes.

* `set option value`: sets the option to value. See the `options` help topic
   for a list of options you can set.

* `setlocal option value`: sets the option to value locally (only in the current
   buffer).

* `show option`: shows the current value of the given option.

* `eval "expression"`: Evaluates a Lua expression. Note that micro will not
   print anything so you should use `messenger:Message(...)` to display a
   value.

* `run sh-command`: runs the given shell command in the background. The 
   command's output will be displayed in one line when it finishes running.

* `bind key action`: creates a keybinding from key to action. See the sections on
   keybindings above for more info about what keys and actions are available.

* `vsplit filename`: opens a vertical split with `filename`. If no filename is
   provided, a vertical split is opened with an empty buffer.

* `hsplit filename`: same as `vsplit` but opens a horizontal split instead of
   a vertical split.

* `tab filename`: opens the given file in a new tab.

* `tabswitch tab`: This command will switch to the specified tab.
   The `tab` can either be a tab number, or a name of a tab.
                     

* `log`: opens a log of all messages and debug statements.

* `plugin install plugin_name`: installs the given plugin.

* `plugin remove plugin_name`: removes the given plugin.

* `plugin list`: lists all installed plugins.

* `plugin update`: updates all installed plugins.

* `plugin search plugin_name`: searches for the given plugin.
   Note that you can find a list of all available plugins at
   github.com/micro-editor/plugin-channel.

   You can also see more information about the plugin manager
   in the `Plugin Manager` section of the `plugins` help topic.

* `plugin available`: list plugins available for download (this includes
   any plugins that may be already installed).

* `reload`: reloads all runtime files.

* `cd path`: Change the working directory to the given `path`.

* `pwd`: Print the current working directory.

* `open filename`: Open a file in the current buffer.

---

The following commands are provided by the default plugins:

* `lint`: Lint the current file for errors.
