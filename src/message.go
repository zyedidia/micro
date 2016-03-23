package main

import (
	"github.com/zyedidia/tcell"
)

var (
	curMessage string
	curStyle   tcell.Style
)

func Message(msg string) {
	curMessage = msg
	curStyle = tcell.StyleDefault

	if _, ok := colorscheme["message"]; ok {
		curStyle = colorscheme["message"]
	}
}

func Error(msg string) {
	curMessage = msg
	curStyle = tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorRed)

	if _, ok := colorscheme["error-message"]; ok {
		curStyle = colorscheme["error-message"]
	}
}

func DisplayMessage(s tcell.Screen) {
	_, h := s.Size()

	runes := []rune(curMessage)
	for x := 0; x < len(runes); x++ {
		s.SetContent(x, h-1, runes[x], nil, curStyle)
	}
}
