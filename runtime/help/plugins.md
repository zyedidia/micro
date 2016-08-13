# Plugins

Micro supports creating plugins with a simple Lua system. Every plugin has a
main script which is run at startup which should be placed in 
`~/.config/micro/plugins/pluginName/pluginName.lua`.

There are a number of callback functions which you can create in your
plugin to run code at times other than startup. The naming scheme is
`onAction()`. For example a function which is run every time the user saves
the buffer would be:

```lua
function onSave()
    ...
    return false
end
```

All available actions are listed in the keybindings section of the help.

These functions should also return a boolean specifying whether the view
should be relocated to the cursor or not after the action is complete.

Note that these callbacks occur after the action has been completed. If you
want a callback before the action is executed, use `preAction()`. In this case
the boolean returned specifies whether or not the action should be executed
after the lua code completes.

---

There are a number of functions and variables that are available to you in
oder to access the inner workings of micro. Here is a list (the type signatures
for functions are given using Go's type system):

* OS: variable which gives the OS micro is currently running on (this is the same
as Go's GOOS variable, so `darwin`, `windows`, `linux`, `freebsd`...)

* tabs: a list of all the tabs currently in use

* curTab: the index of the current tabs in the tabs list

* messenger: lets you send messages to the user or create prompts

* GetOption(name string): returns the value of the requested option

* AddOption(name string, value interface{}): sets the given option with the given
value (`interface{}` means any type in Go).

* BindKey(key, action string): binds `key` to `action`.

* MakeCommand(name, function string): creates a command with `name` which will
call `function` when executed.

* CurView(): returns the current view

* HandleCommand(cmd string): runs the given command

* HandleShellCommand(shellCmd string, interactive bool): runs the given shell
command

// Used for asynchronous jobs
* JobStart(cmd string, onStdout, onStderr, onExit string, userargs ...string):
Starts running the given shell command in the background. `onStdout` `onStderr` and `onExit`
are callbacks to lua functions which will be called when the given actions happen
to the background process.
`userargs` are the arguments which will get passed to the callback functions

* JobSend(cmd *exec.Cmd, data string): send a string into the stdin of the job process

* JobStop(cmd *exec.Cmd): kill a job

This may seem like a small list of available functions but some of the objects
returned by the functions have many methods. `CurView()` returns a view object
which has all the actions which you can call. For example `CurView():Save()`.
