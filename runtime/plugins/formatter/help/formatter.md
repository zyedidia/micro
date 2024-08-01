# Formatter

This plugin provides a way for the user to configure formatters for their codes.
These formatters can be defined by the `filetype`, by a `regex` in the `filepath`,
or by the `operating system`.
These formatters can be run when `saving the file`, by a `keybinding`, or by the
command `> format <name>` where `<name>` is the name of the formatter

Formatter settings can be for any type of file (javascript, go, python), but you
need to have the cli's you want to use as formatter installed.
For example, if you want to configure a formatter for python files and decide to
use [blue], you would need to have it installed on your machine.

Here is an example with [blue]:

```lua
function init()
  formatter.setup({
    { cmd = 'blue %f', filetypes = { 'python' } },
  })
end
```

## formatter.setup(formatters)

The `formatter.setup` function is used to register your formatters, it receives
a table containing the information for each formatter.
And just like in the example above, you must call this function in the `init()`
callback of the `initlua` plugin.

For more details about `initlua`, run `> help tutorial` and see the `Configuration with Lua` topic.
See `> help plugins` to learn more about lua plugins.

## `> format`

The `> format` command is available to format the file with all possible formatters.

You can use the `Alt-f` key shortcut for this.

You can run a single formatter using its `name` as an argument to the `format` command.

```sh
> format <name>
```

## Formatter

### cmd

`type`: **string**, `required`

---

`cmd` is required and must be a string.

Its value will be the executed command.

```lua
function init()
  formatter.setup({
    {
      cmd = 'goimports -w %f',
      filetypes = { 'go' },
    },
  })
end
```

The symbol `%f` will be replaced by the file name at run time,
the same behavior will be applied to `args`.

### filetypes

`type`: **string[]**, `required`

---

`filetypes` is required and must be a table of string.

These are the types of files on which the formatter will be executed.

```lua
function init()
  formatter.setup({
    {
      cmd = 'clang-format -i %f',
      filetypes = { 'c', 'c++', 'csharp' },
    },
  })
end
```

You can write patterns to use instead of exact types.
Just set the [`domatch`](#domatch) field to `true`, from then on every string
within `filetypes` will be a pattern. A [golang regular expression] to be more
specific.

```lua
function init()
  formatter.setup({
    {
      cmd = 'raco fmt -i %f',
      filetypes = { '\\.rkt$' },
      domatch = true,
    },
  })
end
```

### args

`type` **string|string[]**, `default`: **''**

---

List which arguments will be passed to `cmd` when the formatter is run.

```lua
function init()
  formatter.setup({
    {
      cmd = 'rustfmt',
      args = '+nightly %f',
      filetypes = { 'rust' },
    },
    {
      cmd = 'stylua %f',
      args = {
        '--column-width=120',
        '--quote-style=ForceSingle',
        '--line-endings=Unix',
        '--indent-type=Spaces',
        '--call-parentheses=Always',
        '--indent-width=2',
      },
      filetypes = { 'lua' },
    },
  })
end
```

In the example above, both the rustfmt args and the stylua args are valid.

The `args` is optional.
You can pass the arguments inside `cmd`.

```lua
function init()
  formatter.setup({
    {
      cmd = 'zig fmt %f',
      filetypes = { 'zig' },
    },
  })
end
```

We can also write a formatter configuration file:

```toml
# ~/.config/micro/stylua.toml

column_width = 120
line_endings = 'Unix'
indent_type = 'Spaces'
indent_width = 2
quote_style = 'ForceSingle'
call_parentheses = 'Always'
collapse_simple_statement = 'Never'

[sort_requires]
enabled = true
```

And link it in `init.lua`:

```lua
local config = import('micro/config')

local filepath = import('path/filepath')

function init()
  formatter.setup({
    {
      cmd = 'stylua',
      args = { '%f', '--config', filepath.Join(config.ConfigDir, 'stylua.toml') },
      filetypes = { 'lua' },
    },
  })
end
```

We are using lua/go, so let your imagination go to the lua.

### name

`type`: **string?**, `optional`

---

Define the name of the command to be executed in the command bar `(Ctrl-e)` with `> format <name>`.

If no `name` is specified, the first `cmd` string is used to define a name for the formatter.

```lua
function init()
  formatter.setup({
    {
      cmd = 'python -m json.tool',
      args = '--sort-keys --no-ensure-ascii --indent 4 %f %f',
      filetypes = { 'json' },
    },
  })
end
```

In the example above, the name of the formatter will be `python` and you use this
name to format it using the command `> format python` in the json files.

If you want to define another name for the formatter, use the `name` field for that.

```lua
function init()
  formatter.setup({
    {
      name = 'json-fmt',
      cmd = 'python -m json.tool',
      args = '--sort-keys --no-ensure-ascii --indent 4 %f %f',
      filetypes = { 'json' },
    },
  })
end
```

So we have the same command but it must be called by `> format json-fmt`.

### bind

`type`: **string**, `optional`

---

Creates a keybinding for the formatter.

```lua
function init()
  formatter.setup({
    {
      cmd = 'crystal tool format %f',
      bind = 'Alt-l'
      filetypes = { 'crystal' },
    },
  })
end
```

So the `Alt-l` key shortcut will run this formatter.

### onSave

`type`: **boolean**, `default`: **false**

---

If `true` the formatter will be executed when the file is saved.

```lua
function init()
  formatter.setup({
    {
      cmd = 'gofmt -w %f',
      onSave = true,
      filetypes = { 'go' },
    },
  })
end
```

### os

`type`: **string[]**, `default`: **{}**

---

It represents a list of operating systems on which this formatter is supported or not.

```lua
function init()
  formatter.setup({
    {
      cmd = 'rubocop --fix-layout --safe --autocorrect %f',
      filetypes = { 'ruby' },
      os = { 'linux' },
    },
  })
end
```

What defines whether the `os` field is a list of compatible operating systems or
not is the [`whitelist`](#whitelist) field.

```lua
function init()
  formatter.setup({
    {
      cmd = 'mix format %f',
      filetypes = { 'elixir' },
      os = { 'windows' },
      whitelist = true,
    },
  })
end
```

Choices for `os`:

- `android`
- `darwin`
- `dragonfly`
- `freebsd`
- `illumos`
- `ios`
- `js`
- `linux`
- `netbsd`
- `openbsd`
- `plan9`
- `solaris`
- `wasip1`
- `windows`

### whitelist

`type`: **boolean**, `default` **false**

---

`whitelist` is of type boolean and by default its value is `false`.

- If `true` all operating systems within the [`os`](#os) field are considered compatible
  with the formatter.
- If `false` all operating systems within the [`os`](#os) field are considered not
  compatible with the formatter.

### domatch

`type`: **boolean**, `default`: **false**

---

`domatch` is of type boolean and by default its value is `false`.

- If `true` the matches with the files will be done with the function
  `regexp.MatchString(pattern, filename)` where `filename` would be the name of the
  file and `pattern` would be the default defined in [`filetypes`](#filetypes).
- If `false` is an exact match with the file type.

### callback

`type`: **func(b Buffer): boolean**, `optional`

---

Function to be called before executing the formatter, if it returns `false` the formatter
will be canceled.

The type of `callback` would be `func(buf: Buffer): boolean` where `Buffer` would
be a [micro buffer].

```lua
function init()
  formatter.setup({
    {
      cmd = 'taplo fmt %f',
      filetypes = { 'toml' },
      callback = function(buf)
        return buf.Settings['foo'] == nil
      end,
    },
  })
end
```

[micro buffer]: https://pkg.go.dev/github.com/zyedidia/micro/v2@v2.0.13/internal/buffer#Buffer
[blue]: https://blue.readthedocs.io
[golang regular expression]: https://zetcode.com/golang/regex/
