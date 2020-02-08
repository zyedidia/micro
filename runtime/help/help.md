# Micro help text

Micro is a terminal-based text editor that aims to be easy to use and
intuitive, while also taking advantage of the full capabilities of modern
terminals.

To open the command bar, press CtrlE. This enables a `>` prompt for typing
commands. From now on when the documentation says to run a command such as `>
help`, this means press CtrlE and type `help` (and press enter to execute the
command).

For a list of the default keybindings run `> help defaultkeys`.
For more information on keybindings see `> help keybindings`.

## Quick-start

Press Ctrl-q to quit, and Ctrl-s to save. Press CtrlE to start typing commands
and you can see which commands are available by pressing tab, or by viewing the
help topic `> help commands`.

Move the cursor around with the mouse or the arrow keys. Run
`> help defaultkeys` to  get a quick, easy overview of the default hotkeys and
what they do. For more info on rebinding keys, see type `> help keybindings`.

If the colorscheme doesn't look good, you can change it with
`> set colorscheme ...`. You can press tab to see the available colorschemes,
or see more information about colorschemes and syntax highlighting with `> help
colors`.

Press Ctrl-w to move between splits, and type `> vsplit filename` or
`> hsplit filename` to open a new split.

## Accessing more help

Micro has a built-in help system which can be accessed with the `help` command.

To use it, press CtrlE to access command mode and type in `help` followed by a
topic. Typing `help` followed by nothing will open this page.

Here are the possible help topics that you can read:

* tutorial: A brief tutorial which gives an overview of all the other help
  topics
* keybindings: Gives a full list of the default keybindings as well as how to
  rebind them
* defaultkeys: Gives a more straight-forward list of the hotkey commands and
  what they do.
* commands: Gives a list of all the commands and what they do
* options: Gives a list of all the options you can customize
* plugins: Explains how micro's plugin system works and how to create your own
  plugins
* colors: Explains micro's colorscheme and syntax highlighting engine and how
  to create your own colorschemes or add new languages to the engine

For example, to open the help page on plugins you would run `> help plugins`.

I recommend looking at the `tutorial` help file because it is short for each
section and gives concrete examples of how to use the various configuration
options in micro. However, it does not give the in-depth documentation that the
other topics provide.
