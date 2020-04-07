# Default Keys

Below are simple charts of the default hotkeys and their functions. For more
information about binding custom hotkeys or changing default bindings, please
run `> help keybindings`

Please remember that *all* keys here are rebindable! If you don't like it, you
can change it!

### Power user

| Key       | Description of function                                                                           |
|---------- |-------------------------------------------------------------------------------------------------- |
| Ctrl+E    | Open a command prompt for running commands (see `> help commands` for a list of valid commands).  |
| Tab       | In command prompt, it will autocomplete if possible.                                              |
| Ctrl+B    | Run a shell command (this will close micro while your command executes).                          |

### Navigation

| Key                         | Description of function                                                                   |
|---------------------------- |------------------------------------------------------------------------------------------ |
| Arrows                      | Move the cursor around                                                                    |
| Shift+arrows                | Move and select text                                                                      |
| Alt(Ctrl on Mac)+LeftArrow  | Move to the beginning of the current line                                                 |
| Alt(Ctrl on Mac)+RightArrow | Move to the end of the current line                                                       |
| Home                        | Move to the beginning of text on the current line                                                 |
| End                         | Move to the end of the current line                                                       |
| Ctrl(Alt on Mac)+LeftArrow  | Move cursor one word left                                                                 |
| Ctrl(Alt on Mac)+RightArrow | Move cursor one word right                                                                |
| Alt+{                       | Move cursor to previous empty line, or beginning of document                              |
| Alt+}                       | Move cursor to next empty line, or end of document                                        |
| PageUp                      | Move cursor up one page                                                                   |
| PageDown                    | Move cursor down one page                                                                 |
| Ctrl+Home or Ctrl+UpArrow   | Move cursor to start of document                                                          |
| Ctrl+End or Ctrl+DownArrow  | Move cursor to end of document                                                            |
| Ctrl+L                      | Jump to a line in the file (prompts with #)                                               |
| Ctrl+W                      | Cycle between splits in the current tab (use `> vsplit` or `> hsplit` to create a split)  |

### Tabs

| Key     | Description of function   |
|-------- |-------------------------  |
| Ctrl+T  | Open a new tab            |
| Alt+,   | Previous tab              |
| Alt+.   | Next tab                  |

### Find Operations

| Key       | Description of function                   |
|---------- |------------------------------------------ |
| Ctrl+F    | Find (opens prompt)                       |
| Ctrl+N    | Find next instance of current search      |
| Ctrl+P    | Find previous instance of current search  |

### File Operations

| Key       | Description of function                                           |
|---------- |------------------------------------------------------------------ |
| Ctrl+Q    | Close current file (quits micro if this is the last file open)    |
| Ctrl+O    | Open a file (prompts for filename)                                |
| Ctrl+S    | Save current file                                                 |

### Text operations

| Key                                 | Description of function                   |
|------------------------------------ |------------------------------------------ |
| Ctrl(Alt on Mac)+Shift+RightArrow   | Select word right                         |
| Ctrl(Alt on Mac)+Shift+LeftArrow    | Select word left                          |
| Alt(Ctrl on Mac)+Shift+LeftArrow    | Select to start of current line           |
| Alt(Ctrl on Mac)+Shift+RightArrow   | Select to end of current line             |
| Shift+Home                          | Select to start of current line           |
| Shift+End                           | Select to end of current line             |
| Ctrl+Shift+UpArrow                  | Select to start of file                   |
| Ctrl+Shift+DownArrow                | Select to end of file                     |
| Ctrl+X                              | Cut selected text                         |
| Ctrl+C                              | Copy selected text                        |
| Ctrl+V                              | Paste                                     |
| Ctrl+K                              | Cut current line                          |
| Ctrl+D                              | Duplicate current line                    |
| Ctrl+Z                              | Undo                                      |
| Ctrl+Y                              | Redo                                      |
| Alt+UpArrow                         | Move current line or selected lines up    |
| Alt+DownArrow                       | Move current line of selected lines down  |
| Alt+Backspace or Alt+Ctrl+H         | Delete word left                          |
| Ctrl+A                              | Select all                                |

### Macros

| Key       | Description of function                                                           |
|---------- |---------------------------------------------------------------------------------- |
| Ctrl+U    | Toggle macro recording (press Ctrl+U to start recording and press again to stop)  |
| Ctrl+J    | Run latest recorded macro                                                         |

### Multiple cursors

| Key               | Description of function                                                                       |
|------------------ |---------------------------------------------------------------------------------------------- |
| Alt+N             | Create new multiple cursor from selection (will select current word if no current selection)  |
| AltShiftUp        | Spawn a new cursor on the line above the current one                                          |
| AltShiftDown      | Spawn a new cursor on the line below the current one                                          |
| Alt+P             | Remove latest multiple cursor                                                                 |
| Alt+C             | Remove all multiple cursors (cancel)                                                          |
| Alt+X             | Skip multiple cursor selection                                                                |
| Alt+M             | Spawn a new cursor at the beginning of every line in the current selection                    |
| Ctrl+MouseLeft    | Place a multiple cursor at any location                                                       |

### Other

| Key       | Description of function                                                               |
|---------- |-------------------------------------------------------------------------------------- |
| Ctrl+G    | Open help file                                                                        |
| Ctrl+H    | Backspace (old terminals do not support the backspace key and use Ctrl+H instead)     |
| Ctrl+R    | Toggle the line number ruler                                                          |

### Emacs style actions

| Key       | Description of function   |
|---------- |-------------------------- |
| Alt+F     | Next word                 |
| Alt+B     | Previous word             |
| Alt+A     | Move to start of line     |
| Alt+E     | Move to end of line       |

### Function keys.

Warning! The function keys may not work in all terminals! 

| Key   | Description of function   |
|------ |-------------------------- |
| F1    | Open help                 |
| F2    | Save                      |
| F3    | Find                      |
| F4    | Quit                      |
| F7    | Find                      |
| F10   | Quit                      |
