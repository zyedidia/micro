# Plugins

Micro supports creating plugins with a simple Lua system. Plugins are
folders containing Lua files and possibly other source files placed
in `~/.config/micro/plug`. The plugin directory (within `plug`) should
contain at least one Lua file and an `info.json` file. The info file
provides additional information such as the name of the plugin, the
plugin's website, dependencies, etc... Here is an example info file
from the go plugin, which has the following file structure:

```
~/.config/micro/plug/go-plugin/
    go.lua
    repo.json
    help/
        go-plugin.md
```

The `go.lua` file contains the main code for the plugin, though the
code may be distributed across multiple Lua files. The `info.json`
file contains information about the plugin such as the website,
description, version, and any requirements. Plugins may also
have additional files which can be added to micro's runtime files,
of which there are 5 types:

* Colorschemes
* Syntax files
* Help files
* Plugin files
* Syntax header files

In most cases, a plugin will want to add help files, but in certain
cases a plugin may also want to add colorschemes or syntax files. It
is unlikely for a plugin to need to add plugin files at runtime or
syntax header files. No directory structure is enforced but keeping
runtime files in their own directories is good practice.

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

The `bp` variable is a reference to the bufpane the action is being executed
within.  This is almost always the current bufpane.

All available actions are listed in the keybindings section of the help.

These functions should also return a boolean specifying whether the bufpane
should be relocated to the cursor or not after the action is complete.

## Accessing micro functions

Some of micro's internal information is exposed in the form of packages which
can be imported by Lua plugins. A package can be imported in Lua and a value
within it can be accessed using the following syntax:

```lua
local micro = import("micro")
micro.Log("Hello")
```

The packages and functions are listed below (in Go type signatures):

* `micro`
    - `TermMessage(msg interface{}...)`: temporarily close micro and print a
       message

    - `TermError(filename string, lineNum int, err string)`: temporarily close
       micro and print an error formatted as `filename, lineNum: err`.

    - `InfoBar()`: return the infobar BufPane object.

    - `Log(msg interface{}...)`: write a message to `log.txt` (requires
       `-debug` flag, or binary built with `build-dbg`).

    - `SetStatusInfoFn(fn string)`: register the given lua function as
       accessible from the statusline formatting options.

    - `CurPane() *BufPane`: returns the current BufPane, or nil if the
       current pane is not a BufPane.

    - `CurTab() *Tab`: returns the current tab.
* `micro/config`
	- `MakeCommand(name string, action func(bp *BufPane, args[]string),
                   completer buffer.Completer)`:
       create a command with the given name, and lua callback function when
       the command is run. A completer may also be given to specify how
       autocompletion should work with the custom command.

	- `FileComplete`: autocomplete using files in the current directory
	- `HelpComplete`: autocomplete using names of help documents
	- `OptionComplete`: autocomplete using names of options
	- `OptionValueComplete`: autocomplete using names of options, and valid
       values afterwards
	- `NoComplete`: no autocompletion suggestions

	- `TryBindKey(k, v string, overwrite bool) (bool, error)`: bind the key
       `k` to the string `v` in the `bindings.json` file.  If `overwrite` is
       true, this will overwrite any existing binding to key `k`. Returns true
       if the binding was made, and a possible error (for example writing to
       `bindings.json` can cause an error).

	- `Reload()`: reload configuration files.

	- `AddRuntimeFileFromMemory(filetype RTFiletype, filename, data string)`:
       add a runtime file to the `filetype` runtime filetype, with name
       `filename` and data `data`.

	- `AddRuntimeFilesFromDirectory(plugin string, filetype RTFiletype,
                                    directory, pattern string)`:
       add runtime files for the given plugin with the given RTFiletype from
       a directory within the plugin root. Only adds files that match the
       pattern using Go's `filepath.Match`

	- `AddRuntimeFile(plugin string, filetype RTFiletype, filepath string)`:
       add a given file inside the plugin root directory as a runtime file
       to the given RTFiletype category.

	- `ListRuntimeFiles(fileType RTFiletype) []string`: returns a list of
       names of runtime files of the given type.

	- `ReadRuntimeFile(fileType RTFiletype, name string) string`: returns the
       contents of a given runtime file.

	- `NewRTFiletype() int`: creates a new RTFiletype, and returns its value.

	- `RTColorscheme`: runtime files for colorschemes.
	- `RTSyntax`: runtime files for syntax files.
	- `RTHelp`: runtime files for help documents.
	- `RTPlugin`: runtime files for plugin source code.

	- `RegisterCommonOption(pl string, name string, defaultvalue interface{})`:
       registers a new option with for the given plugin. The name of the
       option will be `pl.name`, and will have the given default value. Since
       this registers a common option, the option will be modifiable on a
       per-buffer basis, while also having a global value (in the
       GlobalSettings map).

	- `RegisterGlobalOption(pl string, name string, defaultvalue interface{})`:
       same as `RegisterCommonOption` but the option cannot be modified
       locally to each buffer.

	- `GetGlobalOption(name string) interface{}`: returns the value of a
       given plugin in the `GlobalSettings` map.

	- `SetGlobalOption(option, value string) error`: sets an option to a
       given value. Same as using the `> set` command. This will parse the
       value to the actual value type.

	- `SetGlobalOptionNative(option string, value interface{}) error`: sets
       an option to a given value, where the type of value is the actual
       type of the value internally.
* `micro/shell`
	- `ExecCommand(name string, arg ...string) (string, error)`: runs an
       executable with the given arguments, and pipes the output (stderr
       and stdout) of the executable to an internal buffer, which is
       returned as a string, along with a possible error.

	- `RunCommand(input string) (string, error)`: same as `ExecCommand`,
       except this uses micro's argument parser to parse the arguments from
       the input. For example `cat 'hello world.txt' file.txt`, will pass
       two arguments in the `ExecCommand` argument list (quoting arguments
       will preserve spaces).

	- `RunBackgroundShell(input string) (func() string, error)`: returns a
       function that will run the given shell command and return its output.

	- `RunInteractiveShell(input string, wait bool, getOutput bool)
                          (string, error)`:
       temporarily closes micro and runs the given command in the terminal.
       If `wait` is true, micro will wait for the user to press enter before
       returning to text editing. If `getOutput` is true, micro redirect
       stdout from the command to the returned string.

	- `JobStart(cmd string, onStdout, onStderr,
                onExit func(string, []interface{}), userargs ...interface{})
                *exec.Cmd`:
       Starts a background job by running the shell on the given command
       (using `sh -c`). Three callbacks can be provided which will be called
       when the command generates stdout, stderr, or exits. The userargs will
       be passed to the callbacks, along with the output as the first
       argument of the callback.

	- `JobSpawn(cmd string, cmdArgs []string, onStdout, onStderr,
                onExit func(string, []interface{}), userargs ...interface{})
                *exec.Cmd`:
       same as `JobStart`, except doesn't run the command through the shell
       and instead takes as inputs the list of arguments.

	- `JobStop(cmd *exec.Cmd)`: kills a job.
	- `JobSend(cmd *exec.Cmd, data string)`: sends some data to a job's stdin.

	- `RunTermEmulator(h *BufPane, input string, wait bool, getOutput bool,
                       callback func(out string, userargs []interface{}),
                       userargs []interface{}) error`:
       starts a terminal emulator from a given BufPane with the input command.
       If `wait` is true it will wait for the user to exit by pressing enter
       once the executable has terminated and if `getOutput` is true it will
       redirect the stdout of the process to a pipe which will be passed to
       the callback which is a function that takes a string and a list of
       optional user arguments. This function returns an error on systems
       where the terminal emulator is not supported.

	- `TermEmuSupported`: true on systems where the terminal emulator is
       supported and false otherwise. Supported systems:
        * Linux
        * MacOS
        * Dragonfly
        * OpenBSD
        * FreeBSD

* `micro/buffer`
    - `NewMessage(owner string, msg string, start, end, Loc, kind MsgType)
                  *Message`:
       creates a new message with an owner over a range given by the start
       and end locations.

    - `NewMessageAtLine(owner string, msg string, line int, kindMsgType)
                        *Message`:
       creates a new message with owner, type and message at a given line.

    - `MTInfo`: info message.
    - `MTWarning`: warning message.
    - `MTError` error message.

    - `Loc(x, y int) Loc`: creates a new location struct.

    - `BTDefault`: default buffer type.
    - `BTLog`: log buffer type.
    - `BTRaw`: raw buffer type.
    - `BTInfo`: info buffer type.

    - `NewBuffer(text, path string) *Buffer`: creates a new buffer with the
       given text at a certain path.

    - `NewBufferFromFile(path string) (*Buffer, error)`: creates a new
       buffer by reading from disk at the given path.

    - `ByteOffset(pos Loc, buf *Buffer) int`: returns the byte index of the
       given position in a buffer.

    - `Log(s string)`: writes a string to the log buffer.
    - `LogBuf() *Buffer`: returns the log buffer.
* `micro/util`
    - `RuneAt(str string, idx int) string`: returns the utf8 rune at a
       given index within a string.
    - `GetLeadingWhitespace(s string) string`: returns the leading
       whitespace of a string.
    - `IsWordChar(s string) bool`: returns true if the first rune in a
       string is a word character.
    - `String(b []byte) string`: converts a byte array to a string.
    - `RuneStr(r rune) string`: converts a rune to a string.

This may seem like a small list of available functions but some of the objects
returned by the functions have many methods. The Lua plugin may access any
public methods of an object returned by any of the functions above.
Unfortunately it is not possible to list all the available functions on this
page. Please go to the internal documentation at
https://godoc.org/github.com/zyedidia/micro to see the full list of available
methods. Note that only methods of types that are available to plugins via
the functions above can be called from a plugin.  For an even more detailed
reference see the source code on Github.

For example, with a BufPane object called `bp`, you could call the `Save`
function in Lua with `bp:Save()`.

Note that Lua uses the `:` syntax to call a function rather than Go's `.`
syntax.

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

For documentation for each of these functions, see the Go standard
library documentation at https://golang.org/pkg/ (for the packages
exposed to micro plugins). The Lua standard library is also available
to plugins though it is rather small.

## Adding help files, syntax files, or colorschemes in your plugin

You can use the `AddRuntimeFile(name string, type config.RTFiletype,
                                path string)`
function to add various kinds of files to your plugin. For example, if you'd
like to add a help topic to your plugin called `test`, you would create a
`test.md` file, and call the function:

```lua
config = import("micro/config")
config.AddRuntimeFile("test", config.RTHelp, "test.md")
```

Use `AddRuntimeFilesFromDirectory(name, type, dir, pattern)` to add a number of
files to the runtime. To read the content of a runtime file use
`ReadRuntimeFile(fileType, name string)` or `ListRuntimeFiles(fileType string)`
for all runtime files. In addition, there is `AddRuntimeFileFromMemory` which
adds a runtime file based on a string that may have been constructed at
runtime.

## Default plugins

There are 6 default plugins that come pre-installed with micro. These are

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

See `> help linter`, `> help comment`, and `> help status` for additional
documentation specific to those plugins.

These are good examples for many use-cases if you are looking to write
your own plugins.

## Plugin Manager

Micro also has a built in plugin manager which you can invoke with the
`> plugin ...` command, or in the shell with `micro -plugin ...`.

For the valid commands you can use, see the `command` help topic.

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
  "Website": "https://github.com/user/plugin",
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

Then open a pull request at github.com/micro-editor/plugin-channel adding a
link to the raw `repo.json` that is in your plugin repository.

To make updating the plugin work, the first line of your plugins lua code
should contain the version of the plugin. (Like this: `VERSION = "1.0.0"`)
Please make sure to use [semver](http://semver.org/) for versioning.
