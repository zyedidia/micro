# Colors

This help page aims to cover two aspects of micro's syntax highlighting engine:

- How to create colorschemes and use them
- How to create syntax files to add to the list of languages micro can highlight

### Colorschemes

Micro comes with a number of colorschemes by default. Here is the list:

* default: this is the simplest colorscheme. It uses 16 colors which are
  set by your terminal

* solarized: this is the solarized colorscheme. 
  You should have the solarized color palette in your terminal to use it.

* solarized-tc: this is the solarized colorscheme for true color, just 
  make sure your terminal supports true color before using it and that the 
  MICRO_TRUECOLOR environment variable is set to 1 before starting micro.

* monokai: this is the monokai colorscheme and is micro's default colorscheme
  (as well as sublime text's).  It requires true color to
  look perfect, but the 256 color approximation looks very good as well.

* atom-dark-tc: this colorscheme is based off of Atom's "dark" colorscheme.
  It requires true color to look good.

To enable one of these colorschemes just run the command `set colorscheme solarized`.
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

Colorschemes can be placed in the `~/.config/micro/colorschemes` directory to be used.

### Syntax files

The syntax files specify how to highlight certain languages.

In progress...
