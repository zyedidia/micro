package main

import "strconv"

func (v *View) DisplayView() {
	if v.Type == vtLog {
		// Log views should always follow the cursor...
		v.Relocate()
	}

	// We need to know the string length of the largest line number
	// so we can pad appropriately when displaying line numbers
	maxLineNumLength := len(strconv.Itoa(v.Buf.NumLines))

	if v.Buf.Settings["ruler"] == true {
		// + 1 for the little space after the line number
		v.lineNumOffset = maxLineNumLength + 1
	} else {
		v.lineNumOffset = 0
	}

	height := v.Height
	width := v.Width
	left := v.leftCol
	top := v.Topline

	v.cellview.Draw(v.Buf, top, height, left, width)

	for _, line := range v.cellview.lines {
		for _, char := range line {
			if char != nil {
				screen.SetContent(char.visualLoc.X, char.visualLoc.Y, char.char, nil, char.style)
			}
		}
	}
}
