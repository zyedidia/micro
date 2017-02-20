#Default Keys

Below are simple charts of the default hotkeys and their functions.
For more information about binding custom hotkeys or changing
default bindings, please run `>help keybindings`

Please remember that *all* keys here are rebindable!
If you don't like it, you can change it!

(We are not responsible for you forgetting what you bind keys to.
 Do not open an issue because you forgot your keybindings.)

#Power user
+--------+---------------------------------------------------------+
| Ctrl+E | Switch to the micro command prompt to run a command.    |
|        | (See `>help commands` for a list of commands. )         |
+--------+---------------------------------------------------------+
| Ctrl+B | Run shell commands in micro's current working directory.|
+--------+---------------------------------------------------------+

#Navigation

+--------+---------------------------------------------------------+
| Key    | Description of function                                 |
|--------+---------------------------------------------------------+
| Arrows | Move the cursor around your current document.           |
|        | (Yes this is rebindable to the vim keys if you want.)   |
+--------+---------------------------------------------------------+
| Shift+ | Move and select text.                                   |
| Arrows |                                                         |
+--------+---------------------------------------------------------+
| Home   | Move to the beginning of the current line. (Naturally.) |
+--------+---------------------------------------------------------+
| End    | Move to the end of the current line.                    |
+--------+---------------------------------------------------------+
| PageUp | Move cursor up lines quickly.                           |
+--------+---------------------------------------------------------+
| PageDn | Move cursor down lines quickly.                         |
+--------+---------------------------------------------------------+
| Ctrl+L | Jump to line in current file. ( Prompts for line # )    |
+--------+---------------------------------------------------------+
| Ctrl+W | Move between splits open in current tab.                |
|        | (See vsplit and hsplit in `>help commands`)             |
+--------+---------------------------------------------------------+
| Ctrl+T | Open a new tab.                                         |
+--------+---------------------------------------------------------+
| Alt+,  | Move to the previous tab in the tablist.                |
|        | (This works like moving between file buffers in nano)   |
+--------+---------------------------------------------------------+
| Alt+.  | Move to the next tab in the tablist.                    |
+--------+---------------------------------------------------------+

#Find Operations

+--------+---------------------------------------------------------+
| Ctrl+F | Find text in current file. ( Prompts for text to find.) |
+--------+---------------------------------------------------------+
| Ctrl+N | Find next instance of current search in current file.   |
+--------+---------------------------------------------------------+
| Ctrl+P | Find prev instance of current search in current file.   |
+--------+---------------------------------------------------------+

#File Operations

+--------+---------------------------------------------------------+
| Ctrl+Q | Close current file. ( Quits micro if last file open. )  |
+--------+---------------------------------------------------------+
| Ctrl+O | Open a file. ( Prompts you to input filename. )         |
+--------+---------------------------------------------------------+
| Ctrl+S | Save current file.                                      |
+--------+---------------------------------------------------------+

#Text operations

+--------+---------------------------------------------------------+
| Ctrl+A | Select all text in current file.                        |
+--------+---------------------------------------------------------+
| Ctrl+X | Cut selected text.                                      |
+--------+---------------------------------------------------------+
| Ctrl+C | Copy selected text.                                     |
+--------+---------------------------------------------------------+
| Ctrl+V | Paste selected text.                                    |
+--------+---------------------------------------------------------+
| Ctrl+K | Cut current line. ( Can then be pasted with Ctrl+V)     |
+--------+---------------------------------------------------------+
| Ctrl+D | Duplicate current line.                                 |
+--------+---------------------------------------------------------+
| Ctrl+Z | Undo actions.                                           |
+--------+---------------------------------------------------------+
| Ctrl+Y | Redo actions.                                           |
+--------+---------------------------------------------------------+

#Other
+--------+---------------------------------------------------------+
| Ctrl+G | Open the help file.                                     |
+--------+---------------------------------------------------------+
| Ctrl+H | Alternate backspace.                                    |
|        | (Some old terminals don't support the Backspace key .)  |
+--------+---------------------------------------------------------+
| Ctrl+R | Toggle the line number ruler. ( On the lefthand side.)  |
+--------+---------------------------------------------------------+

#Emacs style actions

+--------+---------------------------------------------------------+
| Alt+F  | Move to the end of the next word. (To the next space.)  |
+--------+---------------------------------------------------------+
| Alt+B  | Move to the beginning of the previous word.             |
+--------+---------------------------------------------------------+
| Alt+A  | Alternate Home key. ( Move to beginning of line. )      |
+--------+---------------------------------------------------------+
| Alt+E  | Alternate End key. ( Move to the end of line.)          |
+--------+---------------------------------------------------------+
| Alt+P  | Move cursor up. ( Same as up key. )                     |
+--------+---------------------------------------------------------+
| Alt+N  | Move cursor down. ( Same as down key. )                 |
+--------+---------------------------------------------------------+

#Function keys.

Warning! The function keys may not work in all terminals! 
+--------+---------------------------------------------------------+
| F1     | Open help.                                              |
+--------+---------------------------------------------------------+
| F2     | Save current file.                                      |
+--------+---------------------------------------------------------+
| F3     | Find in current file. ( Same as Ctrl+F )                |
+--------+---------------------------------------------------------+
| F4     | Close current file. (Quit if only file.)                |
+--------+---------------------------------------------------------+
| F7     | Find in current file. (Same as Ctrl+F)                  |
+--------+---------------------------------------------------------+
| F10    | Close current file.                                     |
+--------+---------------------------------------------------------+

#Macros

Micro supports the use of keyboard macros. Simply press Ctrl+U to
begin recording a macro and press Ctrl+U to stop recording.

Press Ctrl+J to run your recorded macro. 