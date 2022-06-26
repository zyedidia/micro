### Mouse

Micro has very good mouse support that works pretty intuitively and it's very
likely that you have already figured out most of the things you can do with 
the mouse in micro. However, for the sake of completeness, here is a short 
explanation of the actions that can be performed by the mouse with a default
configuration of micro. In the future, mouse actions may be extensible through
plugins or expanded further by us, so check this file or micro's changelog
every so often for new information.

### Default editing

By default, the mouse can be used to move the cursor around as the cursor will
move to the location of a click. In addition, you can left-click and drag to
select text, as in any other text editor. You can also scroll using the scroll
wheel of your mouse to move up and down the document, although the cursor will
not move until you click an area to move to, so if you press any navigation 
buttons ( arrows, page up, page down, etc. ) the screen will scroll back to 
where the cursor is. 

### Default navigation

Currently, the mouse can also be used to navigate the tabbar when a user has 
more than one tab open. Left-clicking a tab's name will move you to that tab,
right-clicking will close a tab and all splits contained within. ( Micro will
still give you a prompt to save any unsaved splits contained in a tab, so you
needn't worry about accidentally losing work this way. ) Left-clicking on "<"
will display tabs to the left of what can be displayed on the screen and similarly
left-clicking on ">" will display tabs to the right of what can be displayed.
We are working on implementing a feature to allow use of the scroll-wheel on 
the tab bar to scroll to tabs that are out of sight to the left and right, but
it is currently unfinished. Clicking on an area belong to a different split than
your current one will switch to that split.

### Troubleshooting

* Micro doesn't respond to my mouse inputs: Check first to see if any other 
  applications that you know support mouse input on terminal do. If not, please
  try a terminal emulator. If you are running in the Linux console ( /dev/ttyX )
  please note that for mouse support there, micro would have to built with 
  `gpm` support and you would have to have `gpm` installed properly. There are
  no current plans to add gpm support to tcell or micro as it would require 
  making bindings to a native C library. If other mouse supported programs work
  in your terminal or micro's mouse support has worked fine in the past but now
  no longer does, please put in an issue on Github.
* I can't see the cursor: Try pressing a navigation button ( the arrows, for 
  instance ) to recenter the screen on the cursor. If you have multiple splits
  open, be sure that you are in the split you want to edit. ( Press the button
  you have bound to switching between splits or try clicking where you want the
  cursor to move to. )
