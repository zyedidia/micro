# QuickMenu Plugin

The quickmenu plugin is a slightly more involved example of what micro's new
overlay system can do.

The plugin exposes a palette-like quickmenu that can be used to quickly find
files by name (via 'find') or by content (via 'grep').

It exposes two new commands, and a single global option.

Commands:
* `quicksearch`: Opens the find-by-name menu.
* `quickopen`: Opens the find-by-contents menu.

By default, quicksearch will be bound to `Alt-f`, and quickopen to `Alt-o`

Options:
* `quickmenu.newtab`: when a file is opened via the quickmenu, it will be opened
   in a new tab.

    default value: `true`
