package display

import (
	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/micro/v2/internal/util"
	"github.com/zyedidia/micro/v2/internal/views"
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
	dividerStyle := config.DefStyle
	if style, ok := config.Colorscheme["divider"]; ok {
		dividerStyle = style
	}

	divchars := config.GetGlobalOption("divchars").(string)
	if util.CharacterCountInString(divchars) != 2 {
		divchars = "|-"
	}

	divchar, combc, _ := util.DecodeCharacterInString(divchars)

	divreverse := config.GetGlobalOption("divreverse").(bool)
	if divreverse {
		dividerStyle = dividerStyle.Reverse(true)
	}

	for i, c := range cs {
		if c.Kind == views.STVert {
			if i != len(cs)-1 {
				for h := 0; h < c.H; h++ {
					screen.SetContent(c.X+c.W, c.Y+h, divchar, combc, dividerStyle)
				}
			}
		}
		w.drawNode(c)
	}
}

func (w *UIWindow) Display() {
	w.drawNode(w.root)
}

func (w *UIWindow) GetMouseSplitNode(vloc buffer.Loc) *views.Node {
	var mouseLoc func(*views.Node) *views.Node
	mouseLoc = func(n *views.Node) *views.Node {
		cs := n.Children()
		for i, c := range cs {
			if c.Kind == views.STVert {
				if i != len(cs)-1 {
					if vloc.X == c.X+c.W && vloc.Y >= c.Y && vloc.Y < c.Y+c.H {
						return c
					}
				}
			} else if c.Kind == views.STHoriz {
				if i != len(cs)-1 {
					if vloc.Y == c.Y+c.H-1 && vloc.X >= c.X && vloc.X < c.X+c.W {
						return c
					}
				}
			}
		}
		for _, c := range cs {
			m := mouseLoc(c)
			if m != nil {
				return m
			}
		}
		return nil
	}
	return mouseLoc(w.root)
}
func (w *UIWindow) Resize(width, height int) {}
func (w *UIWindow) SetActive(b bool)         {}
