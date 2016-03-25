# Micro

[![Build Status](https://travis-ci.org/zyedidia/micro.svg?branch=master)](https://travis-ci.org/zyedidia/micro)
[![Go Report Card](http://goreportcard.com/badge/zyedidia/micro)](http://goreportcard.com/report/zyedidia/micro)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/zyedidia/micro/blob/master/LICENSE)

> Micro is a work in progress, not suitable for use yet.

Micro is a command line text editor that aims to be easy to use and intuitive, while also taking advantage of the full capabilities
of modern terminals.

# Features

* Easy to use
* Common keybindings (ctrl-s, ctrl-c, ctrl-v, ctrl-z...)
* Extremely good mouse support
* True color support
* Cross platform
* Fast and efficient
* Syntax highlighting (in over 75 languages!)

Not all of this is implemented yet -- see [progress](#progress)

# Installation

Installation is simple. For now you must build from source, although in the future binaries will be provided.

Make sure your `GOPATH` is set.

```
$ git clone https://github.com/zyedidia/micro
$ cd micro
$ make
```

This will build micro and put the binary in the current directory. It will also install syntax highlighting files to `~/.micro/syntax`.

Alternatively you can use `make install` instead of `make` if you want the binary to be added to you `GOBIN` (make sure that it is set).

# Progress

Micro is very much a work in progress right now. To see what has and hasn't been done yet, see the [todolist](todolist.md)
