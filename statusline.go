package main

import (
	"github.com/gdamore/tcell"
	"strconv"
)

// Statusline represents the blue line at the bottom of the
// editor that gives information about the buffer
type Statusline struct {
	v *View
}

// Display draws the statusline to the screen
func (sl *Statusline) Display() {
	y := sl.v.height

	file := sl.v.buf.name
	if file == "" {
		file = "Untitled"
	}
	if sl.v.buf.text != sl.v.buf.savedText {
		file += " +"
	}
	file += " (" + strconv.Itoa(sl.v.cursor.y+1) + "," + strconv.Itoa(sl.v.cursor.GetVisualX()+1) + ")"

	statusLineStyle := tcell.StyleDefault.Reverse(true)

	for x := 0; x < sl.v.width; x++ {
		if x < Count(file) {
			sl.v.s.SetContent(x, y, []rune(file)[x], nil, statusLineStyle)
		} else {
			sl.v.s.SetContent(x, y, ' ', nil, statusLineStyle)
		}
	}
}
