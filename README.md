# Micro

[![Build Status](https://travis-ci.org/zyedidia/micro.svg?branch=master)](https://travis-ci.org/zyedidia/micro)
[![Go Report Card](http://goreportcard.com/badge/zyedidia/micro)](http://goreportcard.com/report/zyedidia/micro)
[![Join the chat at https://gitter.im/zyedidia/micro](https://badges.gitter.im/zyedidia/micro.svg)](https://gitter.im/zyedidia/micro?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/zyedidia/micro/blob/master/LICENSE)

> Micro is very much a work in progress

Micro is a terminal-based text editor that aims to be easy to use and intuitive, while also taking advantage of the full capabilities
of modern terminals.

Here is a picture of micro editing its source code.

![Screenshot](./screenshot.png)

# Features

* Easy to use
* Common keybindings (ctrl-s, ctrl-c, ctrl-v, ctrl-z...)
* Extremely good mouse support
* Cross platform
* Syntax highlighting (in over [75 languages](runtime/syntax)!)
* Colorscheme support
* True color support (set the `MICRO_TRUECOLOR` env variable to 1 to enable it)
* Search and replace
* Sane defaults
* Plugin system (plugins are written in Lua)
* Undo and redo
* Unicode support
* Copy and paste with the system clipboard
* Small and simple
* Configurable

If you'd like to see what has been implemented, and what I plan on implementing soon-ish, see the [todo list](todolist.md)

# Installation

### Homebrew

If you are on Mac, you can install micro using Homebrew:

```
brew tap zyedidia/micro
brew install --devel micro
```

Micro is devel-only for now because there is no released version.

### Prebuilt binaries
| Version | Mac | Linux 64 | Linux 32 | Linux Arm | Windows 64 | Windows 32 |
| ------- | --- |---|---|---|---|---|
| Nightly Binaries | [Mac OS X](http://zbyedidia.webfactional.com/micro/binaries/micro-osx.tar.gz) | [Linux 64](http://zbyedidia.webfactional.com/micro/binaries/micro-linux64.tar.gz) | [Linux 32](http://zbyedidia.webfactional.com/micro/binaries/micro-linux32.tar.gz) | [Linux Arm](http://zbyedidia.webfactional.com/micro/binaries/micro-linux-arm.tar.gz) | [Windows 64](http://zbyedidia.webfactional.com/micro/binaries/micro-win64.zip) | [Windows 32](http://zbyedidia.webfactional.com/micro/binaries/micro-win32.zip)

To run the micro binary just run `./bin/micro` (you may want to place the binary on your path for ease of use).

### Building from source

Micro is made in Go so you must have Go installed on your system to build it.

Make sure that you have Go version 1.4 or greater.

You can simply `go get` it.

```
go get -u github.com/zyedidia/micro/cmd/micro
```

### Clipboard support

On Linux, clipboard support requires 'xclip' or 'xsel' command to be installed. For Ubuntu:

```
$ sudo apt-get install xclip
```

If you don't have xclip or xsel, micro will use an internal clipboard for copy and paste, but it won't work with external applications.

# Usage

Once you have built the editor, simply start it by running `micro path/to/file.txt` or simply `micro` to open an empty buffer.

Micro also supports creating buffers from `stdin`:

```
$ ifconfig | micro
```

You can move the cursor around with the arrow keys and mouse.

You can also use the mouse to manipulate the text. Simply clicking and dragging will select text. You can also double click
to enable word selection, and triple click to enable line selection.

You can run `$ micro -version` to get the version number. Since there is no release, this just gives you the
commit hash. The version is unknown if you built with `go get`, instead use `make install` or `make` to get a binary
with a version number defined.

#### Help text

See the [help text](./runtime/help/help.md) for information about keybindings, editor commands, colorschemes and
configuration options.

# Contributing

If you find any bugs, please report them! I am also happy to accept pull requests from anyone.
