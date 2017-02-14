package main

func (v *View) DisplayView() {
	if v.Type == vtLog {
		// Log views should always follow the cursor...
		v.Relocate()
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
