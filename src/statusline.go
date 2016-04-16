package main

import (
	"strconv"
)

// Statusline represents the information line at the bottom
// of each view
// It gives information such as filename, whether the file has been
// modified, filetype, cursor location
type Statusline struct {
	view *View
}

// Display draws the statusline to the screen
func (sline *Statusline) Display() {
	// We'll draw the line at the lowest line in the view
	y := sline.view.height

	file := sline.view.buf.name
	// If the name is empty, use 'No name'
	if file == "" {
		file = "No name"
	}

	// If the buffer is dirty (has been modified) write a little '+'
	if sline.view.buf.IsDirty() {
		file += " +"
	}

	// Add one to cursor.x and cursor.y because (0,0) is the top left,
	// but users will be used to (1,1) (first line,first column)
	// We use GetVisualX() here because otherwise we get the column number in runes
	// so a '\t' is only 1, when it should be tabSize
	columnNum := strconv.Itoa(sline.view.cursor.GetVisualX() + 1)
	lineNum := strconv.Itoa(sline.view.cursor.y + 1)

	file += " (" + lineNum + "," + columnNum + ")"

	// Add the filetype
	file += " " + sline.view.buf.filetype

	centerText := "Press Ctrl-h for help"

	statusLineStyle := defStyle.Reverse(true)
	if style, ok := colorscheme["statusline"]; ok {
		statusLineStyle = style
	}

	// Maybe there is a unicode filename?
	fileRunes := []rune(file)
	for x := 0; x < sline.view.width; x++ {
		if x < len(fileRunes) {
			screen.SetContent(x, y, fileRunes[x], nil, statusLineStyle)
		} else if x >= sline.view.width/2-len(centerText)/2 && x < len(centerText)+sline.view.width/2-len(centerText)/2 {
			screen.SetContent(x, y, []rune(centerText)[x-sline.view.width/2+len(centerText)/2], nil, statusLineStyle)
		} else {
			screen.SetContent(x, y, ' ', nil, statusLineStyle)
		}
	}
}
