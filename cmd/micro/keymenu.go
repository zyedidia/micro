package main

// DisplayKeyMenu displays the nano-style key menu at the bottom of the screen
func DisplayKeyMenu() {
	w, h := screen.Size()

	bot := h - 3

	display := []string{"^Q Quit, ^S Save, ^O Open, ^G Help, ^E Command Bar, ^K Cut Line", "^F Find, ^Z Undo, ^Y Redo, ^A Select All, ^D Duplicate Line, ^T New Tab"}

	for y := 0; y < len(display); y++ {
		for x := 0; x < w; x++ {
			if x < len(display[y]) {
				screen.SetContent(x, bot+y, rune(display[y][x]), nil, defStyle)
			} else {
				screen.SetContent(x, bot+y, ' ', nil, defStyle)
			}
		}
	}
}
