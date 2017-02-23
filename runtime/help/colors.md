# Colors

This help page aims to cover two aspects of micro's syntax highlighting engine:

- How to create colorschemes and use them
- How to create syntax files to add to the list of languages micro can highlight

### Colorschemes

Micro comes with a number of colorschemes by default. Here is the list:

* simple: this is the simplest colorscheme. It uses 16 colors which are
  set by your terminal

* mc: A 16-color theme based on the look and feel of GNU Midnight Commander.
  This will look great used in conjunction with Midnight Commander.
  
* nano: A 16-color theme loosely based on GNU nano's syntax highlighting.   
  
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

* cmc-16: A very nice 16-color theme. Written by contributor CaptainMcClellan
  (Collin Warren.) Licensed under the same license as the rest of the themes.

* cmc-paper: Basically cmc-16, but on a white background. ( Actually light grey on most 
  ANSI (16-color) terminals.)

* cmc-tc: A true colour variant of the cmc theme. 
  It requires true color to look its best. Use cmc-16 if your terminal doesn't support true color.

* codeblocks: A colorscheme based on the Code::Blocks IDE's default syntax highlighting.

* codeblocks-paper: Same as codeblocks, but on a white background. ( Actually light grey. )

* github-tc: A colorscheme based on Github's syntax highlighting. Requires true color to look its best.

* paper-tc: A nice minimalist theme with a light background, good for editing documents on.
  Requires true color to look its best. Not to be confused with `-paper` suffixed themes.

* geany: Colorscheme 

* geany-alt-tc: Based on an alternate theme bundled with geany. 

* flamepoint-tc: A fire inspired, high intensity true color theme written by CaptainMcClellan.
  As with all the other `-tc` suffixed themes, it looks its best on a

To enable one of these colorschemes just press CtrlE in micro and type `set colorscheme solarized`.
(or whichever one you choose). You can also use `set colorscheme monochrome` if you'd prefer
to have just the terminal's default foreground and background colors. 
Note: This provides no syntax highlighting!

See `help gimmickcolors` for a list of some true colour themes that are more 
just for fun than for serious use. ( Though feel free if you want! )

---

### Creating a Colorscheme

Micro's colorschemes are also extremely simple to create. The default ones can be found
[here](https://github.com/zyedidia/micro/tree/master/runtime/colorschemes).

They are only about 18-30 lines in total.

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

Generally colorschemes which require true color terminals to look good are marked with a `-tc` suffix
and colorschemes which supply a white background are marked with a `-paper` suffix.

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
* tabbar ( color of the tabbar that lists open files.)
* indent-char (color of the character which indicates tabs if the option is enabled)
* line-number
* gutter-error
* gutter-warning
* cursor-line
* current-line-number
* color-column
* ignore

Colorschemes must be placed in the `~/.config/micro/colorschemes` directory to be used.

---

In addition to the main colorscheme groups, there are subgroups that you can
specify by adding `.subgroup` to the group. If you're creating your own
custom syntax files, you can make use of your own subgroups.

If micro can't match the subgroup, it'll default to the root group, so 
it's safe and recommended to use subgroups in your custom syntax files.

For example if `constant.string` is found in your colorscheme, micro will
use that for highlighting strings. If it's not found, it will use constant 
instead. Micro tries to match the largest set of groups it can find in the
colorscheme definitions, so if, for examle `constant.bool.true` is found then
micro will use that. If `constant.bool.true` is not found but `constant.bool`
is found micro will use `constant.bool`. If not, it uses `constant`. 

Here's a list of subgroups used in micro's built-in syntax files.

* comment.bright ( Some filetypes have distinctions between types of comments.)
* constant.bool
* constant.bool.true
* constant.bool.false
* constant.number 
* constant.specialChar
* constant.string
* constant.string.url 
* identifier.class ( Also used for functions. )
* identifier.macro
* identifier.var
* symbol.brackets ( {}()[] and sometimes <> )
* symbol.operator ( Color operator symbols differently. )
* symbol.tag ( For html tags, among other things.)
* type.keyword ( If you want a special highlight for keywords like `private` )

In the future, plugins may also be able to use color groups for styling.

### Syntax files

The syntax files specify how to highlight certain languages.

Micro's builtin syntax highlighting tries very hard to be sane, sensible
and provide ample coverage of the meaningful elements of a language. Micro has
syntax files built int for over 100 languages now. However, there may be 
situations where you find Micro's highlighting to be insufficient or not to
your liking. Good news is you can create syntax files (.micro extension), place them in 
`~/.config/micro/syntax` and Micro will use those instead.

The first statement in a syntax file will probably the syntax statement. This tells micro
what language the syntax file is for and how to detect a file in that language.

Essentially, it's just

```
syntax "Name of language" "\.extension$"
```

For the extension, micro will just compare that regex to the filename and if it matches then it
will use the syntax rules defined in the remainder of the file.

There is also a possibility to use a header statement which is a regex that micro will compare
with the first line of the file. This is almost only used for shebangs at the top of shell scripts
which don't have any extension (see sh.micro for an example).

---

The rest of a syntax file is very simple and is essentially a list of regexes specifying how to highlight
different expressions.

It is recommended that when creating a syntax file you use the colorscheme groups (see above) to
highlight different expressions. You may also hard code colors, but that may not look good depending
on what terminal colorscheme the user has installed.

Here is an example to highlight comments (expressions starting with `//`):

```
color comment "//.*"
```

This will highlight the regex `//.*` in the color that the user's colorscheme has linked to the comment
group.

Note that this regex only matches the current line. Here is an example for multiline comments (`/* comment */`):

```
color comment start="/\*" end="\*/"
```

Note: The format of syntax files will be changing with the view refactor.
If this help file still retains this note but the syntax files are yaml
please open an issue.
