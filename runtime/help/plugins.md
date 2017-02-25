# Plugins

Micro supports creating plugins with a simple Lua system. Every plugin has a
main script which is run at startup which should be placed in 
`~/.config/micro/plugins/pluginName/pluginName.lua`.

There are a number of callback functions which you can create in your
plugin to run code at times other than startup. The naming scheme is
`onAction(view)`. For example a function which is run every time the user saves
the buffer would be:

```lua
function onSave(view)
    ...
    return false
end
```

The `view` variable is a reference to the view the action is being executed on.
This is almost always the current view, which you can get with `CurView()` as well.

All available actions are listed in the keybindings section of the help.

These functions should also return a boolean specifying whether the view
should be relocated to the cursor or not after the action is complete.

Note that these callbacks occur after the action has been completed. If you
want a callback before the action is executed, use `preAction()`. In this case
the boolean returned specifies whether or not the action should be executed
after the lua code completes.

Another useful callback to know about which is not a action is
`onViewOpen(view)` which is called whenever a new view is opened and the new
view is passed in. This is useful for setting local options based on the filetype,
for example turning off `tabstospaces` only for Go files when they are opened.

---

There are a number of functions and variables that are available to you in
order to access the inner workings of micro. Here is a list (the type signatures
for functions are given using Go's type system):

* `OS`: variable which gives the OS micro is currently running on (this is the same
as Go's GOOS variable, so `darwin`, `windows`, `linux`, `freebsd`...)

* `configDir`: contains the path to the micro configuration files

* `tabs`: a list of all the tabs currently in use

* `curTab`: the index of the current tabs in the tabs list

* `messenger`: lets you send messages to the user or create prompts

* `NewBuffer(text, path string) *Buffer`: creates a new buffer from a given reader with a given path

* `GetLeadingWhitespace() bool`: returns the leading whitespace of the given string

* `IsWordChar(str string) bool`: returns whether or not the string is a 'word character'

* `RuneStr(r rune) string`: returns a string containing the given rune

* `Loc(x, y int) Loc`: returns a new `Loc` struct

* `JoinPaths(dir... string) string` combines multiple directories to a full path

* `DirectoryName(path string)` returns all but the last element of path ,typically the path's directory

* `GetOption(name string)`: returns the value of the requested option

* `AddOption(name string, value interface{})`: sets the given option with the given
   value (`interface{}` means any type in Go)

* `SetOption(option, value string)`: sets the given option to the value. This will
   set the option globally, unless it is a local only option.

* `SetLocalOption(option, value string, view *View)`: sets the given option to
   the value locally in the given buffer

* `BindKey(key, action string)`: binds `key` to `action`

* `MakeCommand(name, function string, completions ...Completion)`: 
   creates a command with `name` which will call `function` when executed.
   Use 0 for completions to get NoCompletion.

* `MakeCompletion(function string)`:
   creates a `Completion` to use with `MakeCommand`

* `CurView()`: returns the current view

* `HandleCommand(cmd string)`: runs the given command

* `HandleShellCommand(shellCmd string, interactive bool, waitToClose bool)`: runs the given shell
   command. The `interactive` bool specifies whether the command should run in the background. The
   `waitToClose` bool only applies if `interactive` is true and means that it should wait before
   returning to the editor.

* `ToCharPos(loc Loc, buf *Buffer) int`: returns the character position of a given x, y location

* `Reload`: (Re)load everything

* `ByteOffset(loc Loc, buf *Buffer) int`: exactly like `ToCharPos` except it it counts bytes instead of runes

* `JobSpawn(cmdName string, cmdArgs []string, onStdout, onStderr, onExit string, userargs ...string)`:
   Starts running the given process in the background. `onStdout` `onStderr` and `onExit`
   are callbacks to lua functions which will be called when the given actions happen
   to the background process.
   `userargs` are the arguments which will get passed to the callback functions

* `JobStart(cmd string, onStdout, onStderr, onExit string, userargs ...string)`:
   Starts running the given shell command in the background. Note that the command execute
   is first parsed by a shell when using this command. It is executed with `sh -c`.

* `JobSend(cmd *exec.Cmd, data string)`: send a string into the stdin of the job process

* `JobStop(cmd *exec.Cmd)`: kill a job

This may seem like a small list of available functions but some of the objects
returned by the functions have many methods. `CurView()` returns a view object
which has all the actions which you can call. For example `CurView():Save(false)`.
You can see the full list of possible actions in the keybindings help topic.
The boolean on all the actions indicates whether or not the lua callbacks should
be run. I would recommend generally sticking to false when making a plugin to
avoid recursive problems, for example if you call `CurView():Save(true)` in `onSave()`.
Just use `CurView():Save(false)` so that it won't call `onSave()` again.

Using the view object, you can also access the buffer associated with that view
by using `CurView().Buf`, which lets you access the `FileType`, `Path`, `Name`...

The possible methods which you can call using the `messenger` variable are:

* `messenger.Message(msg ...interface{})`
* `messenger.Error(msg ...interface{})`
* `messenger.YesNoPrompt(prompt string) (bool, bool)`
* `messenger.Prompt(prompt, historyType string, completionType Completion) (string, bool)`

If you want a standard prompt, just use `messenger.Prompt(prompt, "", 0)`

# Adding help files, syntax files, or colorschemes in your plugin

You can use the `AddRuntimeFile(name, type, path string)` function to add various kinds of
files to your plugin. For example, if you'd like to add a help topic to your plugin
called `test`, you would create a `test.md` file, and call the function:

```lua
AddRuntimeFile("test", "help", "test.md")
```

Use `AddRuntimeFilesFromDirectory(name, type, dir, pattern)` to add a number of files
to the runtime.
To read the content of a runtime file use `ReadRuntimeFile(fileType, name string)`
or `ListRuntimeFiles(fileType string)` for all runtime files.

# Autocomplete command arguments

See this example to learn how to use `MakeCompletion` and `MakeCommand`

```lua
local function StartsWith(String,Start)
  String = String:upper()
  Start = Start:upper() 
  return string.sub(String,1,string.len(Start))==Start
end

function complete(input)
  local allCompletions = {"Hello", "World", "Foo", "Bar"}
  local result = {}
   
  for i,v in pairs(allCompletions) do
  if StartsWith(v, input) then
       table.insert(result, v)
     end
   end
   return result
end

function foo(arg)
  messenger:Message(arg)
end

MakeCommand("foo", "example.foo", MakeCompletion("example.complete"))
```

# Default plugins

For examples of plugins, see the default `autoclose` and `linter` plugins
(stored in the normal micro core repo under `runtime/plugins`) as well as
any plugins that are stored in the official channel [here](https://github.com/micro-editor/plugin-channel).

# Plugin Manager

Micro also has a built in plugin manager which you can invoke with the `> plugin ...` command.

For the valid commands you can use, see the `command` help topic.

The manager fetches plugins from the channels (which is simply a list of plugin metadata)
which it knows about. By default, micro only knows about the official channel which is located
at github.com/micro-editor/plugin-channel but you can add your own third-party channels using
the `pluginchannels` option and you can directly link third-party plugins to allow installation
through the plugin manager with the `pluginrepos` option.

If you'd like to publish a plugin you've made as an official plugin, you should upload your
plugin online (to Github preferably) and add a `repo.json` file. This file will contain the
metadata for your plugin. Here is an example:

```json
[{
  "Name": "pluginname",
  "Description": "Here is a nice concise description of my plugin",
  "Tags": ["python", "linting"],
  "Versions": [
    {
      "Version": "1.0.0",
      "Url": "https://github.com/user/plugin/archive/v1.0.0.zip",
      "Require": {
        "micro": ">=1.0.3"
      }
    }
  ]
}]
```

Then open a pull request at github.com/micro-editor/plugin-channel adding a link to the
raw `repo.json` that is in your plugin repository.
To make updating the plugin work, the first line of your plugins lua code should contain the version of the plugin. (Like this: `VERSION = "1.0.0"`)
Please make sure to use [semver](http://semver.org/) for versioning.
