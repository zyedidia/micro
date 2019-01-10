package display

import (
	"log"

	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/screen"
)

type TabWindow struct {
	Names   []string
	Active  int
	width   int
	hscroll int
	y       int
}

func NewTabWindow(w int, y int) *TabWindow {
	tw := new(TabWindow)
	tw.width = w
	tw.y = y
	return tw
}

func (w *TabWindow) Display() {
	x := -w.hscroll

	draw := func(r rune, n int) {
		for i := 0; i < n; i++ {
			screen.Screen.SetContent(x, w.y, r, nil, config.DefStyle.Reverse(true))
			x++
			log.Println(x)
		}
	}

	for i, n := range w.Names {
		if i == w.Active {
			draw('[', 1)
		}
		for _, c := range n {
			draw(c, 1)
		}
		if i == w.Active {
			draw(']', 1)
			draw(' ', 3)
		} else {
			draw(' ', 4)
		}
		if x >= w.width {
			break
		}
	}

	if x < w.width {
		draw(' ', w.width-x)
	}
}
