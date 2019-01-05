package display

import (
	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/micro/cmd/micro/views"
)

type UIWindow struct {
	root *views.Node
}

func NewUIWindow(n *views.Node) *UIWindow {
	uw := new(UIWindow)
	uw.root = n
	return uw
}

func (w *UIWindow) drawNode(n *views.Node) {
	cs := n.Children()
	for i, c := range cs {
		if c.IsLeaf() && c.Kind == views.STVert {
			if i != len(cs)-1 {
				for h := 0; h < c.H; h++ {
					screen.Screen.SetContent(c.X+c.W, c.Y+h, '|', nil, config.DefStyle.Reverse(true))
				}
			}
		} else {
			w.drawNode(c)
		}
	}
}

func (w *UIWindow) Display() {
	w.drawNode(w.root)
}

func (w *UIWindow) Clear()         {}
func (w *UIWindow) Relocate() bool { return false }
func (w *UIWindow) GetView() *View { return nil }
func (w *UIWindow) SetView(*View)  {}
func (w *UIWindow) GetMouseLoc(vloc buffer.Loc) buffer.Loc {
	var mouseLoc func(*views.Node) buffer.Loc
	mouseLoc = func(n *views.Node) buffer.Loc {
		cs := n.Children()
		for i, c := range cs {
			if c.IsLeaf() && c.Kind == views.STVert {
				if i != len(cs)-1 {
					if vloc.X == c.X+c.W {
						return vloc
					}
				}
			} else {
				return mouseLoc(c)
			}
		}
		return buffer.Loc{}
	}
	return mouseLoc(w.root)
}
func (w *UIWindow) Resize(width, height int) {}
func (w *UIWindow) SetActive(b bool)         {}
