# [![micro logo](./assets/logo.png)](https://micro-editor.github.io)

[![Build Status](https://travis-ci.org/zyedidia/micro.svg?branch=master)](https://travis-ci.org/zyedidia/micro)
[![Go Report Card](https://goreportcard.com/badge/github.com/zyedidia/micro)](https://goreportcard.com/report/github.com/zyedidia/micro)
[![Join the chat at https://gitter.im/zyedidia/micro](https://badges.gitter.im/zyedidia/micro.svg)](https://gitter.im/zyedidia/micro?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/zyedidia/micro/blob/master/LICENSE)
[![Snap Status](https://build.snapcraft.io/badge/zyedidia/micro.svg)](https://build.snapcraft.io/user/zyedidia/micro)

**micro** is a terminal-based text editor that aims to be easy to use and intuitive, while also taking advantage of the capabilities
of modern terminals. It comes as a single, batteries-included, static binary with no dependencies; you can download and use it right away.

As its name indicates, micro aims to be somewhat of a successor to the nano editor by being easy to install and use.
It strives to be enjoyable as a full-time editor for people who prefer to work in a terminal, or those who regularly edit files over SSH.

Here is a picture of micro editing its source code.

![Screenshot](./assets/micro-solarized.png)

To see more screenshots of micro, showcasing all of the default color schemes, see [here](http://zbyedidia.webfactional.com/micro/screenshots.html).

## Table of Contents

- [Features](#features)
- [Installation](#installation)
  - [Prebuilt binaries](#prebuilt-binaries)
  - [Package Managers](#package-managers)
  - [Building from source](#building-from-source)
  - [macOS terminal](#macos-terminal)
  - [Linux clipboard support](#linux-clipboard-support)
  - [Colors and syntax highlighting](#colors-and-syntax-highlighting)
  - [Plan9, Cygwin](#plan9-cygwin)
- [Usage](#usage)
- [Documentation and Help](#documentation-and-help)
- [Contributing](#contributing)

## Features

- Easy to use and install.
- No dependencies or external files are needed — just the binary you can download further down the page.
- Multiple cursors.
- Common keybindings (<kbd>Ctrl+S</kbd>, <kbd>Ctrl+C</kbd>, <kbd>Ctrl+V</kbd>, <kbd>Ctrl+Z</kbd>, …).
  - Keybindings can be rebound to your liking.
- Sane defaults.
  - You shouldn't have to configure much out of the box (and it is extremely easy to configure).
- Splits and tabs.
- nano-like menu to help you remember the keybindings.
- Extremely good mouse support.
  - This means mouse dragging to create a selection, double click to select by word, and triple click to select by line.
- Cross-platform (it should work on all the platforms Go runs on).
  - Note that while Windows is supported, there are still some bugs that need to be worked out.
- Plugin system (plugins are written in Lua).
  - micro has a built-in plugin manager to automatically install, remove, and update all your plugins.
- Persistent undo.
- Automatic linting and error notifications
- Syntax highlighting for over [90 languages](runtime/syntax).
- Color scheme support.
  - By default, micro comes with 16, 256, and true color themes.
- True color support (set the `MICRO_TRUECOLOR` environment variable to 1 to enable it).
- Snippets.
  - The snippet plugin can be installed with `> plugin install snippets`.
- Copy and paste with the system clipboard.
- Small and simple.
- Easily configurable.
- Macros.
- Common editor features such as undo/redo, line numbers, Unicode support, soft wrapping, …

Although not yet implemented, I hope to add more features such as autocompletion ([#174](https://github.com/zyedidia/micro/issues/174)) or a tree view ([#249](https://github.com/zyedidia/micro/issues/249)) in the future.

## Installation

To install micro, you can download a [prebuilt binary](https://github.com/zyedidia/micro/releases), or you can build it from source.

If you want more information about ways to install micro, see this [wiki page](https://github.com/zyedidia/micro/wiki/Installing-Micro).

### Prebuilt binaries

All you need to install micro is one file, the binary itself. It's as simple as that!

Download the binary from the [releases](https://github.com/zyedidia/micro/releases) page.

On that page you'll see the nightly release, which contains binaries for micro which are built every night,
and you'll see all the stable releases with the corresponding binaries.

If you'd like to see more information after installing micro, run `micro -version`.

### Installation script

There is a script which can install micro for you by downloading the latest prebuilt binary. You can find it at <https://getmic.ro>.

Then you can easily install micro:

```bash
curl https://getmic.ro | bash
```

The script will install the micro binary to the current directory. See its [GitHub repository](https://github.com/benweissmann/getmic.ro) for more information.

### Package managers

You can install micro using Homebrew on Mac:

```
brew install micro
```

On Windows, you can install micro through [Chocolatey](https://chocolatey.org/) or [Scoop](https://github.com/lukesampson/scoop):

```
choco install micro
```

or

```
scoop install micro
```

On Linux, you can install micro through [snap](https://snapcraft.io/docs/core/install)

```
snap install micro --classic
```

On OpenBSD, micro is available in the ports tree. It is also available as a binary package.

```
pkg_add -v micro
```

### Building from source

If your operating system does not have a binary release, but does run Go, you can build from source.

Make sure that you have Go version 1.5 or greater; Go 1.4 will only work if your version supports cgo.

```
go get -d github.com/zyedidia/micro/cmd/micro
cd $GOPATH/src/github.com/zyedidia/micro
make install
```

The binary will then be installed to `$GOPATH/bin` (or your `$GOBIN`).

Please make sure that when you are working with micro's code, you are working on your `GOPATH`.

You can install directly with `go get` (`go get -u github.com/zyedidia/micro/cmd/micro`) but this is not recommended because it doesn't build micro with version information which is useful for the plugin manager.

### macOS terminal

If you are using macOS, you should consider using [iTerm2](http://iterm2.com/) instead of the default terminal (Terminal.app). The iTerm2 terminal has much better mouse support as well as better handling of key events. For best keybinding behavior, choose `xterm defaults` under `Preferences->Profiles->Keys->Load Preset`. The newest versions also support true color.

### Linux clipboard support

On Linux, clipboard support requires the `xclip` or `xsel` commands to be installed.

For Ubuntu:

```sh
sudo apt-get install xclip
```

If you don't have `xclip` or `xsel`, micro will use an internal clipboard for copy and paste, but it won't work with external applications.

### Colors and syntax highlighting

If you open micro and it doesn't seem like syntax highlighting is working, this is probably because
you are using a terminal which does not support 256 color mode. Try changing the color scheme to `simple`
by pressing <kbd>Ctrl+E</kbd> in micro and typing `set colorscheme simple`.

If you are using the default Ubuntu terminal, to enable 256 make sure your `TERM` variable is set
to `xterm-256color`.

Many of the Windows terminals don't support more than 16 colors, which means
that micro's default color scheme won't look very good. You can either set
the color scheme to `simple`, or download a better terminal emulator, like
mintty.

### Plan9, Cygwin

Please note that micro uses the amazing [tcell library](https://github.com/gdamore/tcell), but this
means that micro is restricted to the platforms tcell supports. As a result, micro does not support
Plan9, and Cygwin (although this may change in the future). micro also doesn't support NaCl (which is deprecated anyway).

## Usage

Once you have built the editor, simply start it by running `micro path/to/file.txt` or simply `micro` to open an empty buffer.

micro also supports creating buffers from `stdin`:

```sh
ifconfig | micro
```

You can move the cursor around with the arrow keys and mouse.

You can also use the mouse to manipulate the text. Simply clicking and dragging
will select text. You can also double click to enable word selection, and triple
click to enable line selection.

## Documentation and Help

micro has a built-in help system which you can access by pressing <kbd>Ctrl+E</kbd> and typing `help`. Additionally, you can
view the help files here:

- [main help](https://github.com/zyedidia/micro/tree/master/runtime/help/help.md)
- [keybindings](https://github.com/zyedidia/micro/tree/master/runtime/help/keybindings.md)
- [commands](https://github.com/zyedidia/micro/tree/master/runtime/help/commands.md)
- [colors](https://github.com/zyedidia/micro/tree/master/runtime/help/colors.md)
- [options](https://github.com/zyedidia/micro/tree/master/runtime/help/options.md)
- [plugins](https://github.com/zyedidia/micro/tree/master/runtime/help/plugins.md)

I also recommend reading the [tutorial](https://github.com/zyedidia/micro/tree/master/runtime/help/tutorial.md) for
a brief introduction to the more powerful configuration features micro offers.

## Contributing

If you find any bugs, please report them! I am also happy to accept pull requests from anyone.

You can use the [GitHub issue tracker](https://github.com/zyedidia/micro/issues)
to report bugs, ask questions, or suggest new features.

For a more informal setting to discuss the editor, you can join the [Gitter chat](https://gitter.im/zyedidia/micro).
