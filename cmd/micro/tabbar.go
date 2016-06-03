package main

func DisplayTabBar() {
	str := ""
	for i, v := range views {
		if i == mainView {
			str += "["
		} else {
			str += " "
		}
		str += v.Buf.Name
		if i == mainView {
			str += "]"
		} else {
			str += " "
		}
		str += " "
	}

	tabBarStyle := defStyle.Reverse(true)
	if style, ok := colorscheme["tabbar"]; ok {
		tabBarStyle = style
	}

	// Maybe there is a unicode filename?
	fileRunes := []rune(str)
	w, _ := screen.Size()
	for x := 0; x < w; x++ {
		if x < len(fileRunes) {
			screen.SetContent(x, 0, fileRunes[x], nil, tabBarStyle)
		} else {
			screen.SetContent(x, 0, ' ', nil, tabBarStyle)
		}
	}
}
