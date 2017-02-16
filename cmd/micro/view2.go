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

	// We need to add to the line offset if there are gutter messages
	var hasGutterMessages bool
	for _, v := range v.messages {
		if len(v) > 0 {
			hasGutterMessages = true
		}
	}
	if hasGutterMessages {
		v.lineNumOffset += 2
	}

	if v.x != 0 {
		// One space for the extra split divider
		v.lineNumOffset++
	}

	xOffset := v.x + v.lineNumOffset

	height := v.Height
	width := v.Width
	left := v.leftCol
	top := v.Topline

	v.cellview.Draw(v.Buf, top, height, left, width)

	screenX := v.x
	for lineN, line := range v.cellview.lines {
		screenX = v.x
		curLineN := v.Topline + lineN

		if v.x != 0 {
			// Draw the split divider
			screen.SetContent(screenX, lineN, '|', nil, defStyle.Reverse(true))
			screenX++
		}
		lineNumStyle := defStyle
		if v.Buf.Settings["ruler"] == true {
			// Write the line number
			if style, ok := colorscheme["line-number"]; ok {
				lineNumStyle = style
			}
			if style, ok := colorscheme["current-line-number"]; ok {
				if curLineN == v.Cursor.Y && tabs[curTab].CurView == v.Num && !v.Cursor.HasSelection() {
					lineNumStyle = style
				}
			}

			lineNum := strconv.Itoa(curLineN + 1)

			// Write the spaces before the line number if necessary
			for i := 0; i < maxLineNumLength-len(lineNum); i++ {
				screen.SetContent(screenX, lineN, ' ', nil, lineNumStyle)
				screenX++
			}
			// Write the actual line number
			for _, ch := range lineNum {
				screen.SetContent(screenX, lineN, ch, nil, lineNumStyle)
				screenX++
			}

			// Write the extra space
			screen.SetContent(screenX, lineN, ' ', nil, lineNumStyle)
			screenX++
		}

		var lastChar *Char
		for i, char := range line {
			if char != nil {
				if tabs[curTab].CurView == v.Num && !v.Cursor.HasSelection() &&
					v.Cursor.Y == char.realLoc.Y && v.Cursor.X == char.realLoc.X {
					screen.ShowCursor(xOffset+char.visualLoc.X, char.visualLoc.Y)
				}
				screen.SetContent(xOffset+char.visualLoc.X, char.visualLoc.Y, char.char, nil, char.style)
				if i == len(line)-1 {
					lastChar = char
				}
			}
		}

		if lastChar != nil {
			if tabs[curTab].CurView == v.Num && !v.Cursor.HasSelection() &&
				v.Cursor.Y == lastChar.realLoc.Y && v.Cursor.X == lastChar.realLoc.X+1 {
				screen.ShowCursor(xOffset+lastChar.visualLoc.X+1, lastChar.visualLoc.Y)
			}
		} else if len(line) == 0 {
			if tabs[curTab].CurView == v.Num && !v.Cursor.HasSelection() &&
				v.Cursor.Y == curLineN {
				screen.ShowCursor(xOffset, lineN)
			}
		}
	}
}
