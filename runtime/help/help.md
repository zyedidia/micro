# Micro help text

Micro is a terminal-based text editor that aims to be easy to use and intuitive, 
while also taking advantage of the full capabilities of modern terminals.

*Press CtrlQ to quit, and CtrlS to save.*

If you want to see all the keybindings press CtrlE and type `help keybindings`.

See the next section for more information about documentation and help.

### Accessing more help

Micro has a built-in help system much like Vim's (although less extensive).

To use it, press CtrlE to access command mode and type in `help` followed by a topic.
Typing `help` followed by nothing will open this page.

Here are the possible help topics that you can read:

* tutorial: A brief tutorial which gives an overview of all the other help topics
* keybindings: Gives a full list of the default keybindings as well as how to rebind them
* commands: Gives a list of all the commands and what they do
* options: Gives a list of all the options you can customize
* plugins: Explains how micro's plugin system works and how to create your own plugins
* colors: Explains micro's colorscheme and syntax highlighting engine and how to create your
  own colorschemes or add new languages to the engine

For example, to open the help page on plugins you would press CtrlE and type `help plugins`.

I recommend looking at the `tutorial` help file because it is short for each section and
gives concrete examples of how to use the various configuration options in micro. However,
it does not give the in-depth documentation that the other topics provide.
