package main

import (
	"github.com/gdamore/tcell"
	"strconv"
)

type Statusline struct {
	v *View
}

func (sl *Statusline) display() {
	y := sl.v.height

	file := sl.v.buf.name
	if file == "" {
		file = "Untitled"
	}
	if sl.v.buf.text != sl.v.buf.savedText {
		file += " +"
	}
	file += " (" + strconv.Itoa(sl.v.cursor.y+1) + "," + strconv.Itoa(sl.v.cursor.getVisualX()+1) + ")"

	statusLineStyle := tcell.StyleDefault.Background(tcell.ColorNavy).Foreground(tcell.ColorBlack)

	for x := 0; x < sl.v.width; x++ {
		if x < count(file) {
			sl.v.s.SetContent(x, y, []rune(file)[x], nil, statusLineStyle)
		} else {
			sl.v.s.SetContent(x, y, ' ', nil, statusLineStyle)
		}
	}
}
