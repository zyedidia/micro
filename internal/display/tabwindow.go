package display

import (
	"unicode/utf8"

	"github.com/zyedidia/micro/internal/buffer"
	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/micro/internal/screen"
	"github.com/zyedidia/micro/internal/util"
)

type TabWindow struct {
	Names   []string
	active  int
	Y       int
	width   int
	hscroll int
}

func NewTabWindow(w int, y int) *TabWindow {
	tw := new(TabWindow)
	tw.width = w
	tw.Y = y
	return tw
}

func (w *TabWindow) GetMouseLoc(vloc buffer.Loc) int {
	x := -w.hscroll

	for i, n := range w.Names {
		x++
		s := utf8.RuneCountInString(n)
		if vloc.Y == w.Y && vloc.X < x+s {
			return i
		}
		x += s
		x += 3
		if x >= w.width {
			break
		}
	}
	return -1
}

func (w *TabWindow) Scroll(amt int) {
	w.hscroll += amt
	w.hscroll = util.Clamp(w.hscroll, 0, w.TotalSize()-w.width)
}

func (w *TabWindow) TotalSize() int {
	sum := 2
	for _, n := range w.Names {
		sum += utf8.RuneCountInString(n) + 4
	}
	return sum - 4
}

func (w *TabWindow) Active() int {
	return w.active
}

func (w *TabWindow) SetActive(a int) {
	w.active = a
	x := 2
	s := w.TotalSize()
	for i, n := range w.Names {
		c := utf8.RuneCountInString(n)
		if i == a {
			if x+c >= w.hscroll+w.width {
				w.hscroll = util.Clamp(x+c+1-w.width, 0, s-w.width)
			} else if x < w.hscroll {
				w.hscroll = util.Clamp(x-4, 0, s-w.width)
			}
			break
		}
		x += c + 4
	}
}

// TODO: handle files with character width >=2

func (w *TabWindow) Display() {
	x := -w.hscroll
	done := false

	draw := func(r rune, n int) {
		for i := 0; i < n; i++ {
			if x == w.width-1 && !done {
				screen.Screen.SetContent(w.width-1, w.Y, '>', nil, config.DefStyle.Reverse(true))
				x++
				break
			} else if x == 0 && w.hscroll > 0 {
				screen.Screen.SetContent(0, w.Y, '<', nil, config.DefStyle.Reverse(true))
			} else if x >= 0 && x < w.width {
				screen.Screen.SetContent(x, w.Y, r, nil, config.DefStyle.Reverse(true))
			}
			x++
		}
	}

	for i, n := range w.Names {
		if i == w.active {
			draw('[', 1)
		} else {
			draw(' ', 1)
		}
		for _, c := range n {
			draw(c, 1)
		}
		if i == len(w.Names)-1 {
			done = true
		}
		if i == w.active {
			draw(']', 1)
			draw(' ', 2)
		} else {
			draw(' ', 3)
		}
		if x >= w.width {
			break
		}
	}

	if x < w.width {
		draw(' ', w.width-x)
	}
}
