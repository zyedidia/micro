package main

import (
	"sort"

	"github.com/zyedidia/tcell"
)

type Tab struct {
	// This contains all the views in this tab
	// There is generally only one view per tab, but you can have
	// multiple views with splits
	views []*View
	// This is the current view for this tab
	curView int
	// Generally this is the name of the current view's buffer
	name string
}

func NewTabFromView(v *View) *Tab {
	t := new(Tab)
	t.views = append(t.views, v)
	t.views[0].Num = 0
	return t
}

func (t *Tab) SetNum(num int) {
	for _, v := range t.views {
		v.TabNum = num
	}
}

// CurView returns the current view
func CurView() *View {
	curTab := tabs[curTab]
	return curTab.views[curTab.curView]
}

func TabbarString() (string, map[int]int) {
	str := ""
	indicies := make(map[int]int)
	for i, t := range tabs {
		if i == curTab {
			str += "["
		} else {
			str += " "
		}
		str += t.views[t.curView].Buf.Name
		if i == curTab {
			str += "]"
		} else {
			str += " "
		}
		indicies[len(str)-1] = i + 1
		str += " "
	}
	return str, indicies
}

func TabbarHandleMouseEvent(event tcell.Event) bool {
	if len(tabs) <= 1 {
		return false
	}

	switch e := event.(type) {
	case *tcell.EventMouse:
		button := e.Buttons()
		if button == tcell.Button1 {
			x, y := e.Position()
			if y != 0 {
				return false
			}
			str, indicies := TabbarString()
			if x >= len(str) {
				return false
			}
			var tabnum int
			var keys []int
			for k := range indicies {
				keys = append(keys, k)
			}
			sort.Ints(keys)
			for _, k := range keys {
				if x <= k {
					tabnum = indicies[k] - 1
					break
				}
			}
			curTab = tabnum
			return true
		}
	}

	return false
}

func DisplayTabs() {
	if len(tabs) <= 1 {
		return
	}

	str, _ := TabbarString()

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
