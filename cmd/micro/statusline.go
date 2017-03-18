package main

import (
	"strconv"
	"time"
)

// Statusline represents the information line at the bottom
// of each view
// It gives information such as filename, whether the file has been
// modified, filetype, cursor location
type Statusline struct {
	view *View
	pluginline string
}

// Display draws the statusline to the screen
func (sline *Statusline) Display() {
	// We'll draw the line at the lowest line in the view
	y := sline.view.Height + sline.view.y

	file := ""

	if globalSettings["showclock"].(bool) {
		t := time.Now()
		curtime := "["
		if globalSettings["12hourclock"].(bool) { 
			if t.Hour() > 12 {
				if t.Hour()-12 >= 10 {curtime += strconv.Itoa(t.Hour()-12) + ":" } else {
					curtime += "0" + strconv.Itoa(t.Hour()-12) + ":"
				} 
			} else {
				if t.Hour()+1 >= 10 {curtime += strconv.Itoa(t.Hour()+1) + ":" } else {
					curtime += "0" + strconv.Itoa(t.Hour()) + ":"
				}
			} 
			if t.Minute() < 10 {curtime += "0" + strconv.Itoa(t.Minute()) } else {
				curtime += strconv.Itoa(t.Minute())
			}
		}
		if !globalSettings["12hourclock"].(bool) { 
			if t.Hour() < 10 { curtime += "0" + strconv.Itoa(t.Hour()) + ":"  } else {
				curtime += strconv.Itoa(t.Hour()) + ":" 
			}
			if t.Minute() < 10 { curtime += "0" + strconv.Itoa(t.Minute()) } else {
				curtime += strconv.Itoa(t.Minute())
			}
		}
		if globalSettings["showseconds"].(bool) {
			if t.Second() < 10 { curtime += ":0" + strconv.Itoa(t.Second()) } else {
				curtime += ":" + strconv.Itoa(t.Second())
			}
		}
		curtime += "] "
		file += curtime
	}
	
	file += sline.view.Buf.GetName()

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

	//Support for modifying the statusline
	file += " " +  sline.pluginline + pluginsline

	rightText := ""
	if len(helpBinding) > 0 {
		rightText = helpBinding + " for help "
		if sline.view.Type == vtHelp {
			rightText = helpBinding + " to close help "
		}
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
