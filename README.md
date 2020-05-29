<img alt="micro logo" src="./assets/micro-logo.svg" width="500px"/>

[![Build Status](https://travis-ci.org/zyedidia/micro.svg?branch=master)](https://travis-ci.org/zyedidia/micro)
[![Go Report Card](https://goreportcard.com/badge/github.com/zyedidia/micro)](https://goreportcard.com/report/github.com/zyedidia/micro)
[![Release](https://img.shields.io/github/release/zyedidia/micro.svg?label=Release)](https://github.com/zyedidia/micro/releases)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/zyedidia/micro/blob/master/LICENSE)
[![Join the chat at https://gitter.im/zyedidia/micro](https://badges.gitter.im/zyedidia/micro.svg)](https://gitter.im/zyedidia/micro?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Snap Status](https://build.snapcraft.io/badge/zyedidia/micro.svg)](https://build.snapcraft.io/user/zyedidia/micro)

**micro** is a terminal-based text editor that aims to be easy to use and intuitive, while also taking advantage of the capabilities
of modern terminals. It comes as a single, batteries-included, static binary with no dependencies; you can download and use it right now!

As its name indicates, micro aims to be somewhat of a successor to the nano editor by being easy to install and use.
It strives to be enjoyable as a full-time editor for people who prefer to work in a terminal, or those who regularly edit files over SSH.

Here is a picture of micro editing its source code.

![Screenshot](./assets/micro-solarized.png)

To see more screenshots of micro, showcasing some of the default color schemes, see [here](http://zbyedidia.webfactional.com/micro/screenshots.html).
 
You can also check out the website for Micro at https://micro-editor.github.io.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
  - [Prebuilt binaries](#prebuilt-binaries)
  - [Package Managers](#package-managers)
  - [Building from source](#building-from-source)
  - [Fully static binary](#fully-static-binary)
  - [macOS terminal](#macos-terminal)
  - [Linux clipboard support](#linux-clipboard-support)
  - [Colors and syntax highlighting](#colors-and-syntax-highlighting)
  - [Plan9, Cygwin, Mingw](#plan9-cygwin-mingw)
- [Usage](#usage)
- [Documentation and Help](#documentation-and-help)
- [Contributing](#contributing)

- - -

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
  - Note that while Windows is supported Mingw/Cygwin is not (see below)
- Plugin system (plugins are written in Lua).
  - micro has a built-in plugin manager to automatically install, remove, and update plugins.
- Built-in diff gutter
- Simple autocompletion
- Persistent undo.
- Automatic linting and error notifications
- Syntax highlighting for over [130 languages](runtime/syntax).
- Color scheme support.
  - By default, micro comes with 16, 256, and true color themes.
- True color support (set the `MICRO_TRUECOLOR` environment variable to 1 to enable it).
- Copy and paste with the system clipboard.
- Small and simple.
- Easily configurable.
- Macros.
- Common editor features such as undo/redo, line numbers, Unicode support, soft wrapping, …

## Installation

To install micro, you can download a [prebuilt binary](https://github.com/zyedidia/micro/releases), or you can build it from source.

If you want more information about ways to install micro, see this [wiki page](https://github.com/zyedidia/micro/wiki/Installing-Micro).

Use `micro -version` to get the version information after installing. It is only guaranteed that you are installing the most recent
stable version if you install from the prebuilt binaries, Homebrew, or Snap.

### Prebuilt binaries

All you need to install micro is one file, the binary itself. It's as simple as that!

Download the binary from the [releases](https://github.com/zyedidia/micro/releases) page.

### Installation script

There is a script which can install micro for you by downloading the latest prebuilt binary. You can find it at <https://getmic.ro>.

You can easily install micro by running

```bash
curl https://getmic.ro | bash
```

The script will place the micro binary in the current directory. See its [GitHub repository](https://github.com/benweissmann/getmic.ro) for more information.

### Package managers

You can install micro using Homebrew on Mac:

```
brew install micro
```

**Note for Mac:** All micro keybindings use the control or alt (option) key, not the command
key. By default, macOS terminals do not forward alt key events. To fix this, please see
the section on [macOS terminals](https://github.com/zyedidia/micro#macos-terminal) further below.

On Linux, you can install micro through [snap](https://snapcraft.io/docs/core/install)

```
snap install micro --classic
```

**Note for Linux:** for interfacing with the local system clipboard, `xclip` or `xsel`
must be installed. Please see the section on [Linux clipboard support](https://github.com/zyedidia/micro#linux-clipboard-support)
further below.

Micro is also available through other package managers on Linux such as apt, dnf, AUR, Nix, and package managers
for other operating systems. These packages are not guaranteed to be up-to-date.

* Linux: Available in distro-specific package managers.
    * `apt install micro` (Ubuntu 20.04 `focal`, and Debian `unstable | testing | buster-backports`).
    * `dnf install micro` (Fedora).
    * `yay -S micro` (Arch Linux).
    * See [wiki](https://github.com/zyedidia/micro/wiki/Installing-Micro) for details about CRUX, Termux.
* Windows: [Chocolatey](https://chocolatey.org) and [Scoop](https://github.com/lukesampson/scoop).
    * `choco install micro`.
    * `scoop install micro`.
* OpenBSD: Available in the ports tree and also available as a binary package.
    * `pkd_add -v micro`.

### Building from source

If your operating system does not have a binary release, but does run Go, you can build from source.

Make sure that you have Go version 1.11 or greater and Go modules are enabled.

```
git clone https://github.com/zyedidia/micro
cd micro
make build
sudo mv micro /usr/local/bin # optional
```

The binary will be placed in the current directory and can be moved to
anywhere you like (for example `/usr/local/bin`).

The command `make install` will install the binary to `$GOPATH/bin` or `$GOBIN`.

You can install directly with `go get` (`go get github.com/zyedidia/micro/cmd/micro`) but this isn't
recommended because it doesn't build micro with version information (necessary for the plugin manager),
and doesn't disable debug mode.

### Fully static binary

By default, the micro binary will dynamically link with core system libraries (this is generally
recommended for security and portability). However, there is a fully static prebuilt binary that
is provided for amd64 as `linux-static.tar.gz`, and to build a fully static binary from source, run

```
CGO_ENABLED=0 make build
```

### macOS terminal

If you are using macOS, you should consider using [iTerm2](http://iterm2.com/) instead of the default terminal (Terminal.app). The iTerm2 terminal has much better mouse support as well as better handling of key events. For best keybinding behavior, choose `xterm defaults` under `Preferences->Profiles->Keys->Load Preset`. The newest versions also support true color.

If you still insist on using the default Mac terminal, be sure to set `Use Option key as Meta key` under
`Preferences->Profiles->Keyboard` to use <kbd>option</kbd> as <kbd>alt</kbd>.

### Linux clipboard support

On Linux, clipboard support requires:

- On X11, the `xclip` or `xsel` commands (for Ubuntu: `sudo apt install xclip`)
- On Wayland, the `wl-clipboard` command

If you don't have these commands, micro will use an internal clipboard for copy and paste, but it won't work with external applications.

### Colors and syntax highlighting

If you open micro and it doesn't seem like syntax highlighting is working, this is probably because
you are using a terminal which does not support 256 color mode. Try changing the color scheme to `simple`
by pressing <kbd>Ctrl+E</kbd> in micro and typing `set colorscheme simple`.

If you are using the default Ubuntu terminal, to enable 256 make sure your `TERM` variable is set
to `xterm-256color`.

Many of the Windows terminals don't support more than 16 colors, which means
that micro's default color scheme won't look very good. You can either set
the color scheme to `simple`, or download and configure a better terminal emulator
than the Windows default.

### Cygwin, Mingw, Plan9

Cygwin, Mingw, and Plan9 are unfortunately not officially supported. In Cygwin and Mingw, micro will often work when run using
the `winpty` utility:

```
winpty micro.exe ...
```

Micro uses the amazing [tcell library](https://github.com/gdamore/tcell), but this
means that micro is restricted to the platforms tcell supports. As a result, micro does not support
Plan9, and Cygwin (although this may change in the future). Micro also doesn't support NaCl (which is deprecated anyway).

## Usage

Once you have built the editor, start it by running `micro path/to/file.txt` or `micro` to open an empty buffer.

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

Sometimes I am unresponsive, and I apologize! If that happens, please ping me.
