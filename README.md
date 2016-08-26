# ![Micro](./assets/logo.png)

[![Build Status](https://travis-ci.org/zyedidia/micro.svg?branch=master)](https://travis-ci.org/zyedidia/micro)
![Go Report Card](https://goreportcard.com/badge/github.com/zyedidia/micro)
[![Join the chat at https://gitter.im/zyedidia/micro](https://badges.gitter.im/zyedidia/micro.svg)](https://gitter.im/zyedidia/micro?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/zyedidia/micro/blob/master/LICENSE)

> Micro is still a work in progress

Micro is a terminal-based text editor that aims to be easy to use and intuitive, while also taking advantage of the full capabilities
of modern terminals. It comes as one single, batteries-included, static binary with no dependencies, and you can download and use it right now.

Here is a picture of micro editing its source code.

![Screenshot](./assets/micro-solarized.png)

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
* Plugin system (plugins are written in Lua)
* Automatic linting and error notifications
* Syntax highlighting (for over [75 languages](runtime/syntax)!)
* Colorscheme support
    * By default, micro comes with 16, 256, and true color themes.
* True color support (set the `MICRO_TRUECOLOR` env variable to 1 to enable it)
* Copy and paste with the system clipboard
* Small and simple
* Easily configurable
* Common editor things such as undo/redo, line numbers, unicode support...

# Installation

To install micro, you can download a prebuilt binary, or you can build it from source.

You can also install micro with a few package managers (on OSX, Arch Linux, and CRUX). 
See this [wiki page](https://github.com/zyedidia/micro/wiki/Installing-Micro) for details.

Please note that micro uses the amazing [tcell library](https://github.com/gdamore/tcell), but this
means that micro is restricted to the platforms tcell supports. As a result, micro does not support
Plan9, NaCl, and Cygwin (although this may change in the future).

### Prebuilt binaries

All you need to install micro is one file, the binary itself. It's as simple as that!

You can download the correct binary for your operating system from the list in the [nightly build release](https://github.com/zyedidia/micro/releases/tag/nightly).

Micro has no released version, instead these binaries are compiled every night and you can find the
commit they were compiled with by running `micro -version`.

If your operating system does not have binary, but does run Go, you can build from source.

### Building from source

Make sure that you have Go version 1.5 or greater (Go 1.4 will work for the systems like support CGO then).

```sh
go get -u github.com/zyedidia/micro/...
```

### Clipboard support

On Linux, clipboard support requires 'xclip' or 'xsel' command to be installed.

For Ubuntu:

```sh
sudo apt-get install xclip
```

If you don't have xclip or xsel, micro will use an internal clipboard for copy and paste, but it won't work with external applications.

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

# Contributing

If you find any bugs, please report them! I am also happy to accept pull requests from anyone.

You can use the Github issue tracker to report bugs, ask questions, or suggest new features.
