# Micro

[![Build Status](https://travis-ci.org/zyedidia/micro.svg?branch=master)](https://travis-ci.org/zyedidia/micro)
![Go Report Card](https://goreportcard.com/badge/github.com/zyedidia/micro)
[![Join the chat at https://gitter.im/zyedidia/micro](https://badges.gitter.im/zyedidia/micro.svg)](https://gitter.im/zyedidia/micro?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/zyedidia/micro/blob/master/LICENSE)

> Micro is very much a work in progress

Micro is a terminal-based text editor that aims to be easy to use and intuitive, while also taking advantage of the full capabilities
of modern terminals. It comes as one single, batteries-included, static binary with no dependencies, and you can download and use it right now.

Here is a picture of micro editing its source code.

![Screenshot](./screenshot.png)

# Features

* Easy to use and to install
* No dependencies or external files are needed -- just the binary you can download further down the page
* Common keybindings (ctrl-s, ctrl-c, ctrl-v, ctrl-z...)
    * Keybindings can be rebound to your liking
* Extremely good mouse support
* Cross platform (It should work on all the platforms Go runs on)
* Plugin system (plugins are written in Lua)
* Syntax highlighting (for over [75 languages](runtime/syntax)!)
* Colorscheme support
* True color support (set the `MICRO_TRUECOLOR` env variable to 1 to enable it)
* Sane defaults
* Copy and paste with the system clipboard
* Small and simple
* Easily configurable
* Common editor things such as undo/redo, line numbers, unicode support...

# Installation

This section gives instructions for how to simply install micro using the prebuilt binaries, or building from source.

You can also install micro with a few package managers (on OSX, Arch Linux, and CRUX). 
See the [wiki page](https://github.com/zyedidia/micro/wiki/Installing-Micro) for details.

### Prebuilt binaries

To easily install micro on any of the operating systems listed below, just download the tar file, 
extract it, and run the binary inside. It's as simple as that!

Micro has no released version, instead these binaries are compiled every night and you can find the
commit it was compiled with by running `micro -version`.

[You can find the binaries in the nightly build release](https://github.com/zyedidia/micro/releases/tag/nightly)

To run the micro binary just run `./micro` (you probably want to place the binary on your `$PATH` for ease of use).

### Building from source

Make sure that you have Go version 1.4 or greater.

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

You can also use the mouse to manipulate the text. Simply clicking and dragging will select text. You can also double click
to enable word selection, and triple click to enable line selection.

You can run `$ micro -version` to get the version number. Since there is no release, this just gives you the
commit hash. The version is unknown if you built with `go get`, instead use `make install` or `make` to get a binary
with a version number defined.

### Help text

See the [help text](./runtime/help/help.md) for information about keybindings, editor commands, colorschemes and
configuration options.

# Contributing

If you find any bugs, please report them! I am also happy to accept pull requests from anyone.
