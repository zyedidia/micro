# Colors

This help page aims to cover two aspects of micro's syntax highlighting engine:

- How to create colorschemes and use them.
- How to create syntax files to add to the list of languages micro can highlight.

## Colorschemes

To change your colorscheme, press Ctrl-E in micro to bring up the command
prompt, and type:

```
set colorscheme monokai
```

(or whichever colorscheme you choose).

Micro comes with a number of colorschemes by default. Modern terminals tend to
have three different kinds of color support. The most common is 256 color where
the terminal provides 256 standardized colors (except the first 16 may be configured
by the user). A 256-color theme requires a terminal with 256 color support and
is the most portable.

A 16-color theme uses the 16 user-configurable colors (or 16 default colors on
old terminals). These colorschemes are guranteed to work, but won't look great
unless the 16 colors are configured to the user's liking. Using a 16-color theme
will also preserve the terminal's theme because the terminal usually uses its 16
colors for prompts or other coloring.

Some terminals support "true color" with 16 million colors (using standard RGB values).
There is no one standard for this color support among terminals so this method
is not guaranteed to work. Usually truecolor must also be enabled by the user. The
colorschemes using true color will look exactly as intended. If true color is not
supported, a true color colorscheme will approximate its colors to 256-color.

Here is the list of colorschemes:

### 256 color

These should work and look nice in most terminals. I recommend these
themes the most.

* `monokai` (also the `default` colorscheme)
* `zenburn`
* `gruvbox`
* `darcula`
* `twilight`
* `railscast`
* `bubblegum`

### 16 color

These may vary widely based on the 16 colors selected for your terminal.

* `simple`
* `solarized` (must have the solarized color palette in your terminal to use this colorscheme properly)
* `cmc-16`
* `cmc-paper`
* `geany`

### True color

These require terminals that support true color and require `MICRO_TRUECOLOR=1` (this is an environment variable).

* `solarized-tc`: this is the solarized colorscheme for true color.
* `atom-dark-tc`: this colorscheme is based off of Atom's "dark" colorscheme.
* `cmc-tc`: A true colour variant of the cmc theme.  It requires true color to
  look its best. Use cmc-16 if your terminal doesn't support true color.
* `gruvbox-tc`: The true color version of the gruvbox colorscheme
* `github-tc`: The true color version of the Github colorscheme

## Creating a Colorscheme

Micro's colorschemes are also extremely simple to create. The default ones can
be found [here](https://github.com/zyedidia/micro/tree/master/runtime/colorschemes).

They are only about 18-30 lines in total.

Basically to create the colorscheme you need to link highlight groups with
actual colors. This is done using the `color-link` command.

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

Color terminals usually have 16 colors that are preset by the user. This means
that you cannot depend on those colors always being the same. You can use those
colors with the names `black, red, green, yellow, blue, magenta, cyan, white`
and the bright variants of each one (brightblack, brightred...).

Then you can use the terminals 256 colors by using their numbers 1-256 (numbers
1-16 will refer to the named colors).

If the user's terminal supports true color, then you can also specify colors
exactly using their hex codes. If the terminal is not true color but micro is
told to use a true color colorscheme it will attempt to map the colors to the 
available 256 colors.

Generally colorschemes which require true color terminals to look good are
marked with a `-tc` suffix and colorschemes which supply a white background are
marked with a `-paper` suffix.

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
* statusline (Color of the statusline)
* tabbar (Color of the tabbar that lists open files)
* indent-char (Color of the character which indicates tabs if the option is
  enabled)
* line-number
* gutter-error
* gutter-warning
* cursor-line
* current-line-number
* color-column
* ignore
* divider (Color of the divider between vertical splits)

Colorschemes must be placed in the `~/.config/micro/colorschemes` directory to
be used.

---

In addition to the main colorscheme groups, there are subgroups that you can
specify by adding `.subgroup` to the group. If you're creating your own custom
syntax files, you can make use of your own subgroups.

If micro can't match the subgroup, it'll default to the root group, so  it's
safe and recommended to use subgroups in your custom syntax files.

For example if `constant.string` is found in your colorscheme, micro will us
that for highlighting strings. If it's not found, it will use constant instead.
Micro tries to match the largest set of groups it can find in the colorscheme
definitions, so if, for examle `constant.bool.true` is found then micro will use
that. If `constant.bool.true` is not found but `constant.bool` is found micro
will use `constant.bool`. If not, it uses `constant`. 

Here's a list of subgroups used in micro's built-in syntax files.

* comment.bright (Some filetypes have distinctions between types of comments)
* constant.bool
* constant.bool.true
* constant.bool.false
* constant.number 
* constant.specialChar
* constant.string
* constant.string.url 
* identifier.class (Also used for functions)
* identifier.macro
* identifier.var
* preproc.shebang (The #! at the beginning of a file that tells the os what
  script interpreter to use)
* symbol.brackets (`{}()[]` and sometimes `<>`)
* symbol.operator (Color operator symbols differently)
* symbol.tag (For html tags, among other things)
* type.keyword (If you want a special highlight for keywords like `private`)

In the future, plugins may also be able to use color groups for styling.


## Syntax files

The syntax files is written in yaml-format and specify how to highlight
languages.

Micro's builtin syntax highlighting tries very hard to be sane, sensible and
provide ample coverage of the meaningful elements of a language. Micro has
syntax files built in for over 100 languages now! However, there may be 
situations where you find Micro's highlighting to be insufficient or not to your
liking. The good news is that you can create your own syntax files, and place them
in  `~/.config/micro/syntax` and Micro will use those instead.

### Filetype definition

You must start the syntax file by declaring the filetype:

```
filetype: go
```

#### Detect definition

Then you must provide information about how to detect the filetype:

```
detect:
    filename: "\\.go$"
```

Micro will match this regex against a given filename to detect the filetype. You
may also provide an optional `header` regex that will check the first line of
the file. For example:

```
detect:
    filename: "\\.ya?ml$"
    header: "%YAML"
```

#### Syntax rules

Next you must provide the syntax highlighting rules. There are two types of
rules: patterns and regions. A pattern is matched on a single line and usually a
single word as well. A region highlights between two patterns over multiple
lines and may have rules of its own inside the region.

Here are some example patterns in Go:

```
rules:
    - special: "\\b(break|case|continue|default|go|goto|range|return)\\b"
    - statement: "\\b(else|for|if|switch)\\b"
    - preproc: "\\b(package|import|const|var|type|struct|func|go|defer|iota)\\b"
```

The order of patterns does matter as patterns lower in the file will overwrite
the ones defined above them.

And here are some example regions for Go:

```
- constant.string:
    start: "\""
    end: "\""
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

Notice how the regions may contain rules inside of them. Any inner rules that
are matched are then skipped when searching for the end of the region. For
example, when highlighting `"foo \" bar"`, since `\"` is matched by an inner
rule in the region, it is skipped. Likewise for `"foo \\" bar`, since `\\` is
matched by an inner rule, it is skipped, and then the `"` is found and the
string ends at the correct place.

You may also explicitly mark skip regexes if you don't want them to be
highlighted. For example:

```
- constant.string:
    start: "\""
    end: "\""
    skip: "\\."
    rules: []
```

#### Includes

You may also include rules from other syntax files as embedded languages. For
example, the following is possible for html:

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
