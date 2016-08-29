# ![Micro](./assets/logo.png)

[![Build Status](https://travis-ci.org/zyedidia/micro.svg?branch=master)](https://travis-ci.org/zyedidia/micro)
[![Go Report Card](https://goreportcard.com/badge/github.com/zyedidia/micro)](https://goreportcard.com/report/github.com/zyedidia/micro)
[![Join the chat at https://gitter.im/zyedidia/micro](https://badges.gitter.im/zyedidia/micro.svg)](https://gitter.im/zyedidia/micro?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/zyedidia/micro/blob/master/LICENSE)

Micro is a terminal-based text editor that aims to be easy to use and intuitive, while also taking advantage of the full capabilities
of modern terminals. It comes as one single, batteries-included, static binary with no dependencies, and you can download and use it right now.

As the name indicates, micro aims to be somewhat of a successor to the nano editor by being easy to install and use in a pinch, but micro also aims to be
enjoyable to use full time, whether you work in the terminal because you prefer it (like me), or because you need to (over ssh).

Here is a picture of micro editing its source code.

![Screenshot](./assets/micro-solarized.png)

To see more screenshots of micro, showcasing all of the default colorschemes, see [here](http://zbyedidia.webfactional.com/micro/screenshots.html).

# Features

* Easy to use and to install
* No dependencies or external files are needed -- just the binary you can download further down the page
* Common keybindings (ctrl-s, ctrl-c, ctrl-v, ctrl-z...)
    * Keybindings can be rebound to your liking
* Sane defaults
    * You shouldn't have to configure much out of the box (and it is extremely easy to configure)
* Extremely good mouse support
    * This means mouse dragging to create a selection, double click to select by word, and triple click to select by line
* Cross platform (It should work on all the platforms Go runs on)
    * Note that while Windows is supported, there are still some bugs that need to be worked out
* Plugin system (plugins are written in Lua)
* Persistent undo
* Automatic linting and error notifications
* Syntax highlighting (for over [75 languages](runtime/syntax)!)
* Colorscheme support
    * By default, micro comes with 16, 256, and true color themes.
* True color support (set the `MICRO_TRUECOLOR` env variable to 1 to enable it)
* Copy and paste with the system clipboard
* Small and simple
* Easily configurable
* Common editor things such as undo/redo, line numbers, unicode support...

Although not yet implemented, I hope to add more features such as autocompletion, and multiple cursors in the future.

# Installation

To install micro, you can download a prebuilt binary, or you can build it from source.

Please note that micro uses the amazing [tcell library](https://github.com/gdamore/tcell), but this
means that micro is restricted to the platforms tcell supports. As a result, micro does not support
Plan9, NaCl, and Cygwin (although this may change in the future).

If you want more information about ways to install micro, see this [wiki page](https://github.com/zyedidia/micro/wiki/Installing-Micro)

### Prebuilt binaries

All you need to install micro is one file, the binary itself. It's as simple as that!

You can download the correct binary for your operating system from the [releases](https://github.com/zyedidia/micro/releases).

On that page you'll see the nightly release, which contains binaries for micro which are built every night,
and you'll see all the stable releases with the corresponding binaries.

If you'd like to see more information after installing micro, run `micro -version`.

### Building from source

If your operating system does not have binary, but does run Go, you can build from source.

Make sure that you have Go version 1.5 or greater (Go 1.4 will work if your version supports CGO).

```sh
go get -u github.com/zyedidia/micro/...
```

### Linux clipboard support

On Linux, clipboard support requires 'xclip' or 'xsel' command to be installed.

For Ubuntu:

```sh
sudo apt-get install xclip
```

If you don't have xclip or xsel, micro will use an internal clipboard for copy and paste, but it won't work with external applications.

### Windows colors

Many of the Windows terminals don't support more than 16 colors, which means
that micro's default colorscheme won't look very good. You can either set
the colorscheme to `simple`, or download a better terminal emulator, like
mintty or cmder.

# Usage

Once you have built the editor, simply start it by running `micro path/to/file.txt` or simply `micro` to open an empty buffer.

Micro also supports creating buffers from `stdin`:

```sh
ifconfig | micro
```

You can move the cursor around with the arrow keys and mouse.

You can also use the mouse to manipulate the text. Simply clicking and dragging
will select text. You can also double click to enable word selection, and triple
click to enable line selection.

# Documentation and Help

Micro has a built-in help system which you can access by pressing `CtrlE` and typing `help`. Additionally, you can
view the help files online [here](https://github.com/zyedidia/micro/tree/master/runtime/help).

I also recommend reading the [tutorial](https://github.com/zyedidia/micro/tree/master/runtime/help/tutorial.md) for
a brief introduction to the more powerful features micro offers.

# Contributing

If you find any bugs, please report them! I am also happy to accept pull requests from anyone.

You can use the Github issue tracker to report bugs, ask questions, or suggest new features.
