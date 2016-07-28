# Micro help text

Micro is a terminal-based text editor that aims to be easy to use and intuitive, 
while also taking advantage of the full capabilities of modern terminals.

### Accessing more help

Micro has a built-in help system much like Vim's (although less extensive).

To use it, press CtrlE to access command mode and type in help followed by a topic.
Typing help followed by nothing will open this page.

Here are the possible help topics that you can read:

* keybindings: Gives a full list of the default keybindings as well as how to rebind them
* commands: Gives a list of all the commands and what they do
* options: Gives a list of all the options you can customize
* plugins: Explains how micro's plugin system works and how to create your own plugins
* colors: Explains micro's colorscheme and syntax highlighting engine and how to create your
  own colorschemes or add new languages to the engine

For example to open the help page on plugins you would press CtrlE and type `help plugins`.

### Usage

Once you have built the editor, simply start it by running 
`micro path/to/file.txt` or simply `micro` to open an empty buffer.

Micro also supports creating buffers from stdin:

```
$ ifconfig | micro
```

You can move the cursor around with the arrow keys and mouse.
