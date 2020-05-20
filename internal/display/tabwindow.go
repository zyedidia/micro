package display

import (
	runewidth "github.com/mattn/go-runewidth"
	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/micro/v2/internal/util"
)

type TabWindow struct {
	Names   []string
	active  int
	Y       int
	Width   int
	hscroll int
}

func NewTabWindow(w int, y int) *TabWindow {
	tw := new(TabWindow)
	tw.Width = w
	tw.Y = y
	return tw
}

func (w *TabWindow) Resize(width, height int) {
	w.Width = width
}

func (w *TabWindow) LocFromVisual(vloc buffer.Loc) int {
	x := -w.hscroll

	for i, n := range w.Names {
		x++
		s := util.CharacterCountInString(n)
		if vloc.Y == w.Y && vloc.X < x+s {
			return i
		}
		x += s
		x += 3
		if x >= w.Width {
			break
		}
	}
	return -1
}

func (w *TabWindow) Scroll(amt int) {
	w.hscroll += amt
	s := w.TotalSize()
	w.hscroll = util.Clamp(w.hscroll, 0, s-w.Width)

	if s-w.Width <= 0 {
		w.hscroll = 0
	}
}

func (w *TabWindow) TotalSize() int {
	sum := 2
	for _, n := range w.Names {
		sum += runewidth.StringWidth(n) + 4
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
		c := util.CharacterCountInString(n)
		if i == a {
			if x+c >= w.hscroll+w.Width {
				w.hscroll = util.Clamp(x+c+1-w.Width, 0, s-w.Width)
			} else if x < w.hscroll {
				w.hscroll = util.Clamp(x-4, 0, s-w.Width)
			}
			break
		}
		x += c + 4
	}

	if s-w.Width <= 0 {
		w.hscroll = 0
	}
}

func (w *TabWindow) Display() {
	x := -w.hscroll
	done := false

	tabBarStyle := config.DefStyle.Reverse(true)
	if style, ok := config.Colorscheme["tabbar"]; ok {
		tabBarStyle = style
	}

	draw := func(r rune, n int) {
		for i := 0; i < n; i++ {
			rw := runewidth.RuneWidth(r)
			for j := 0; j < rw; j++ {
				c := r
				if j > 0 {
					c = ' '
				}
				if x == w.Width-1 && !done {
					screen.SetContent(w.Width-1, w.Y, '>', nil, tabBarStyle)
					x++
					break
				} else if x == 0 && w.hscroll > 0 {
					screen.SetContent(0, w.Y, '<', nil, tabBarStyle)
				} else if x >= 0 && x < w.Width {
					screen.SetContent(x, w.Y, c, nil, tabBarStyle)
				}
				x++
			}
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
		if x >= w.Width {
			break
		}
	}

	if x < w.Width {
		draw(' ', w.Width-x)
	}
}
