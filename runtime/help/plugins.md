# Plugins

Micro supports creating plugins with a simple Lua system. Plugins are
folders containing Lua files and possibly other source files placed
in `~/.config/micro/plug`. The plugin directory (within `plug`) should
contain at least one Lua file and an `info.json` file. The info file
provides additional information such as the name of the plugin, the
plugin's website, dependencies, etc... Here is an example info file
from the go plugin, which has the following file structure:

```
~/.config/micro/plug/go-plugin
    go.lua
    info.json
```

info.json:
```
{
    "name": "go",
    "description": "Go formatting and tool support",
    "website": "https://github.com/micro-editor/go-plugin",
	"install": "https://github.com/micro-editor/go-plugin",
    "version": "1.0.0",
    "require": [
        "micro >= 2.0.0"
    ]
}
```

All fields are simply interpreted as strings, so the version does not
need to be a semantic version, and the dependencies are also only
meant to be parsed by humans. The name should be an identifier, and
the website should point to a valid website. The install field should
provide info about installing the plugin, or point to a website that
provides information.

## Lua callbacks

Plugins use Lua but also have access to many functions both from micro
and from the Go standard library. Many callbacks are also defined which
are called when certain events happen. Here is the list of callbacks
which micro defines:

* `init()`: this function should be used for your plugin initialization.

* `onBufferOpen(buf)`: runs when a buffer is opened. The input contains
   the buffer object.

* `onBufPaneOpen(bufpane)`: runs when a bufpane is opened. The input
   contains the bufpane object.

* `onAction(bufpane)`: runs when `Action` is triggered by the user, where
   `Action` is a bindable action (see `> help keybindings`). A bufpane
   is passed as input and the function should return a boolean defining
   whether the view should be relocated after this action is performed.

* `preAction(bufpane)`: runs immediately before `Action` is triggered
   by the user. Returns a boolean which defines whether the action should
   be canceled.

For example a function which is run every time the user saves the buffer
would be:

```lua
function onSave(bp)
    ...
    return false
end
```

The `bp` variable is a reference to the bufpane the action is being executed within.
This is almost always the current bufpane.

All available actions are listed in the keybindings section of the help.

For callbacks to mouse actions, you are also given the event info:

```lua
function onMousePress(view, event)
    local x, y = event:Position()

    return false
end
```

These functions should also return a boolean specifying whether the bufpane should
be relocated to the cursor or not after the action is complete.

## Accessing micro functions

Some of micro's internal information is exposed in the form of packages which
can be imported by Lua plugins. A package can be imported in Lua and a value
within it can be accessed using the following syntax:

```lua
local micro = import("micro")
micro.Log("Hello")
```

The packages and functions are listed below:

* `micro`
    - `TermMessage(msg interface{}...)`
    - `TermError()`
    - `InfoBar()`
    - `Log(msg interface{}...)`
    - `SetStatusInfoFn`
* `micro/config`
    - `MakeCommand`
    - `FileComplete`
    - `HelpComplete`
    - `OptionComplete`
    - `OptionValueComplete`
    - `NoComplete`
    - `TryBindKey`
    - `Reload`
    - `AddRuntimeFilesFromDirectory`
    - `AddRuntimeFileFromMemory`
    - `AddRuntimeFile`
    - `ListRuntimeFiles`
    - `ReadRuntimeFile`
    - `RTColorscheme`
    - `RTSyntax`
    - `RTHelp`
    - `RTPlugin`
    - `RegisterCommonOption`
    - `RegisterGlobalOption`
* `micro/shell`
    - `ExecCommand`
    - `RunCommand`
    - `RunBackgroundShell`
    - `RunInteractiveShell`
    - `JobStart`
    - `JobSpawn`
    - `JobStop`
    - `JobStop`
    - `RunTermEmulator`
    - `TermEmuSupported`
* `micro/buffer`
    - `NewMessage`
    - `NewMessageAtLine`
    - `MTInfo`
    - `MTWarning`
    - `MTError`
    - `Loc`
    - `BTDefault`
    - `BTLog`
    - `BTRaw`
    - `BTInfo`
    - `NewBufferFromFile`
    - `ByteOffset`
* `micro/util`
    - `RuneAt`
    - `GetLeadingWhitespace`
    - `IsWordChar`


This may seem like a small list of available functions but some of the objects
returned by the functions have many methods. The Lua plugin may access any
public methods of an object returned by any of the functions above. For example,
with a BufPane object called `bp`, you could called the `Save` function in Lua
with `bp:Save()`.

Note that Lua uses the `:` syntax to call a function rather than Go's `.` syntax.

```go
micro.InfoBar().Message()
```

turns to

```lua
micro.InfoBar():Message()
```

## Accessing the Go standard library

It is possible for your lua code to access many of the functions in the Go
standard library.

Simply import the package you'd like and then you can use it. For example:

```lua
local ioutil = import("io/ioutil")
local fmt = import("fmt")
local micro = import("micro")

local data, err = ioutil.ReadFile("SomeFile.txt")

if err ~= nil then
    micro.InfoBar():Error("Error reading file: SomeFile.txt")
else
    -- Data is returned as an array of bytes
    -- Using Sprintf will convert it to a string
    local str = fmt.Sprintf("%s", data)

    -- Do something with the file you just read!
    -- ...
end
```

Here are the packages from the Go standard library that you can access.
Nearly all functions from these packages are supported. For an exact
list of which functions are supported you can look through `lua.go`
(which should be easy to understand).

```
fmt
io
io/ioutil
net
math
math/rand
os
runtime
path
filepath
strings
regexp
errors
time
```

For documentation for each of these functions, you can simply look
through the Go standard library documentation.

## Adding help files, syntax files, or colorschemes in your plugin

You can use the `AddRuntimeFile(name, type, path string)` function to add
various kinds of files to your plugin. For example, if you'd like to add a help
topic to your plugin called `test`, you would create a `test.md` file, and call
the function:

```lua
AddRuntimeFile("test", "help", "test.md")
```

Use `AddRuntimeFilesFromDirectory(name, type, dir, pattern)` to add a number of
files to the runtime. To read the content of a runtime file use
`ReadRuntimeFile(fileType, name string)` or `ListRuntimeFiles(fileType string)`
for all runtime files.

## Autocomplete command arguments

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

## Default plugins

For examples of plugins, see the default `autoclose` and `linter` plugins
(stored in the normal micro core repo under `runtime/plugins`) as well as any
plugins that are stored in the official channel
[here](https://github.com/micro-editor/plugin-channel).

## Plugin Manager

Micro also has a built in plugin manager which you can invoke with the
`> plugin ...` command.

For the valid commands you can use, see the `commands` help topic.

The manager fetches plugins from the channels (which is simply a list of plugin
metadata) which it knows about. By default, micro only knows about the official
channel which is located at github.com/micro-editor/plugin-channel but you can
add your own third-party channels using the `pluginchannels` option and you can
directly link third-party plugins to allow installation through the plugin
manager with the `pluginrepos` option.

If you'd like to publish a plugin you've made as an official plugin, you should
upload your plugin online (to Github preferably) and add a `repo.json` file.
This file will contain the metadata for your plugin. Here is an example:

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

Then open a pull request at github.com/micro-editor/plugin-channel adding a link
to the raw `repo.json` that is in your plugin repository. To make updating the
plugin work, the first line of your plugins lua code should contain the version
of the plugin. (Like this: `VERSION = "1.0.0"`) Please make sure to use
[semver](http://semver.org/) for versioning.
