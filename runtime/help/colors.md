# Colors

This help page aims to cover two aspects of micro's syntax highlighting engine:

- How to create colorschemes and use them
- How to create syntax files to add to the list of languages micro can highlight

### Colorschemes

Micro comes with a number of colorschemes by default. Here is the list:

* simple: this is the simplest colorscheme. It uses 16 colors which are
  set by your terminal

* monokai: this is the monokai colorscheme; you may recognize it as
  Sublime Text's default colorscheme. It requires true color to
  look perfect, but the 256 color approximation looks very good as well.
  It's also the default colorscheme.

* zenburn: The 'zenburn' colorscheme and works well with 256 color terminals

* solarized: this is the solarized colorscheme.
  You should have the solarized color palette in your terminal to use it.

* solarized-tc: this is the solarized colorscheme for true color; just
  make sure your terminal supports true color before using it and that the
  MICRO_TRUECOLOR environment variable is set to 1 before starting micro.

* atom-dark-tc: this colorscheme is based off of Atom's "dark" colorscheme.
  It requires true color to look good.

To enable one of these colorschemes just press CtrlE in micro and type `set colorscheme solarized`.
(or whichever one you choose).

---

Micro's colorschemes are also extremely simple to create. The default ones can be found
[here](https://github.com/zyedidia/micro/tree/master/runtime/colorschemes).

They are only about 18 lines in total.

Basically to create the colorscheme you need to link highlight groups with actual colors.
This is done using the `color-link` command.

For example, to highlight all comments in green, you would use the command:

```
color-link comment "green"
```

Background colors can also be specified with a comma:

```
color-link comment "green,blue"
```

This will give the comments a blue background.

If you would like no foreground you can just use a comma with nothing in front:

```
color-link comment ",blue"
```

You can also put bold, or underline in front of the color:

```
color-link comment "bold red"
```

---

There are three different ways to specify the color.

Color terminals usually have 16 colors that are preset by the user. This means that
you cannot depend on those colors always being the same. You can use those colors with
the names `black, red, green, yellow, blue, magenta, cyan, white` and the bright variants
of each one (brightblack, brightred...).

Then you can use the terminals 256 colors by using their numbers 1-256 (numbers 1-16 will
refer to the named colors).

If the user's terminal supports true color, then you can also specify colors exactly using
their hex codes. If the terminal is not true color but micro is told to use a true color colorscheme
it will attempt to map the colors to the available 256 colors.

Generally colorschemes which require true color terminals to look good are marked with a `-tc` suffix.

---

Here is a list of the colorscheme groups that you can use:

* default (color of the background and foreground for unhighlighted text)
* comment
* identifier
* constant
* statement
* symbol
* preproc
* type
* special
* underlined
* error
* todo
* statusline (color of the statusline)
* indent-char (color of the character which indicates tabs if the option is enabled)
* line-number
* gutter-error
* gutter-warning
* cursor-line
* current-line-number
* color-column

Colorschemes can be placed in the `~/.config/micro/colorschemes` directory to be used.

### Syntax files

The syntax files specify how to highlight certain languages.

Syntax files are specified in the yaml format.

#### Filetype defintion

You must start the syntax file by declaring the filetype:

```
filetype: go
```

#### Detect definition

Then you can provide information about how to detect the filetype:

```
detect:
    filename: "\\.go$"
```

Micro will match this regex against a given filename to detect the filetype. You may also
provide an optional `header` regex that will check the first line of the file. For example for yaml:

```
detect:
    filename: "\\.ya?ml$"
    header: "%YAML"
```

#### Syntax rules

Next you must provide the syntax highlighting rules. There are two types of rules: patterns and regions.
A pattern is matched on a single line and usually a single word as well. A region highlights between two
patterns over multiple lines and may have rules of its own inside the region.

Here are some example patterns in Go:

```
rules:
    - special: "\\b(break|case|continue|default|go|goto|range|return)\\b"
    - statement: "\\b(else|for|if|switch)\\b"
    - preproc: "\\b(package|import|const|var|type|struct|func|go|defer|iota)\\b"
```

The order of patterns does matter as patterns lower in the file will overwrite the ones defined above them.

And here are some example regions for Go:

```
- constant.string:
    start: "\""
    end: "(?<!\\\\)\""
    rules:
        - constant.specialChar: "%."
        - constant.specialChar: "\\\\[abfnrtv'\\\"\\\\]"
        - constant.specialChar: "\\\\([0-7]{3}|x[A-Fa-f0-9]{2}|u[A-Fa-f0-9]{4}|U[A-Fa-f0-9]{8})"

- comment:
    start: "//"
    end: "$"
    rules:
        - todo: "(TODO|XXX|FIXME):?"

- comment:
    start: "/\\*"
    end: "\\*/"
    rules:
        - todo: "(TODO|XXX|FIXME):?"
```

Notice how the regions may contain rules inside of them.

Also the regexes for region start and end may contain more complex regexes with lookahead and lookbehind,
but this is not supported for pattern regexes.

#### Includes

You may also include rules from other syntax files as embedded languages. For example, the following is possible
for html:

```
- default:
    start: "<script.*?>"
    end: "</script.*?>"
    rules:
        - include: "javascript"

- default:
    start: "<style.*?>"
    end: "</style.*?>"
    rules:
        - include: "css"
```
