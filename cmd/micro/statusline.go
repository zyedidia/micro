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
	y := sline.view.Height + sline.view.y

	file := sline.view.Buf.GetName()
	if file == "" {
		file = "No name"
	}

	// If the buffer is dirty (has been modified) write a little '+'
	if sline.view.Buf.IsModified {
		file += " +"
	}

	// Add one to cursor.x and cursor.y because (0,0) is the top left,
	// but users will be used to (1,1) (first line,first column)
	// We use GetVisualX() here because otherwise we get the column number in runes
	// so a '\t' is only 1, when it should be tabSize
	columnNum := strconv.Itoa(sline.view.Cursor.GetVisualX() + 1)
	lineNum := strconv.Itoa(sline.view.Cursor.Y + 1)

	file += " (" + lineNum + "," + columnNum + ")"

	// Add the filetype
	file += " " + sline.view.Buf.FileType()

	rightText := helpBinding + " for help "
	if sline.view.Type == vtHelp {
		rightText = helpBinding + " to close help "
	}

	statusLineStyle := defStyle.Reverse(true)
	if style, ok := colorscheme["statusline"]; ok {
		statusLineStyle = style
	}

	// Maybe there is a unicode filename?
	fileRunes := []rune(file)
	viewX := sline.view.x
	if viewX != 0 {
		screen.SetContent(viewX, y, ' ', nil, statusLineStyle)
		viewX++
	}
	for x := 0; x < sline.view.Width; x++ {
		if x < len(fileRunes) {
			screen.SetContent(viewX+x, y, fileRunes[x], nil, statusLineStyle)
		} else if x >= sline.view.Width-len(rightText) && x < len(rightText)+sline.view.Width-len(rightText) {
			screen.SetContent(viewX+x, y, []rune(rightText)[x-sline.view.Width+len(rightText)], nil, statusLineStyle)
		} else {
			screen.SetContent(viewX+x, y, ' ', nil, statusLineStyle)
		}
	}
}
