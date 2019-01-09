package action

import (
	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/display"
	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/micro/cmd/micro/views"
	"github.com/zyedidia/tcell"
)

var MainTab *TabPane

type TabPane struct {
	*views.Node
	display.Window
	Panes  []*EditPane
	active int

	resizing *views.Node // node currently being resized
}

func (t *TabPane) HandleEvent(event tcell.Event) {
	switch e := event.(type) {
	case *tcell.EventResize:
		w, h := screen.Screen.Size()
		InfoBar.Resize(w, h-1)
		t.Node.Resize(w, h-1)
		t.Resize()
	case *tcell.EventMouse:
		switch e.Buttons() {
		case tcell.Button1:
			mx, my := e.Position()

			resizeID := t.GetMouseLoc(buffer.Loc{mx, my}).X
			if t.resizing != nil {
				var size int
				if t.resizing.Kind == views.STVert {
					size = mx - t.resizing.X
				} else {
					size = my - t.resizing.Y + 1
				}
				t.resizing.ResizeSplit(size)
				t.Resize()
				return
			}

			if resizeID != -1 {
				t.resizing = t.GetNode(uint64(resizeID))
				return
			}

			for i, p := range t.Panes {
				v := p.GetView()
				inpane := mx >= v.X && mx < v.X+v.Width && my >= v.Y && my < v.Y+v.Height
				if inpane {
					t.active = i
					p.SetActive(true)
				} else {
					p.SetActive(false)
				}
			}
		case tcell.ButtonNone:
			t.resizing = nil
		}

	}
	t.Panes[t.active].HandleEvent(event)
}

func (t *TabPane) SetActive(i int) {
	t.active = i
	for j, p := range t.Panes {
		if j == i {
			p.SetActive(true)
		} else {
			p.SetActive(false)
		}
	}
}

func (t *TabPane) GetPane(splitid uint64) int {
	for i, p := range t.Panes {
		if p.splitID == splitid {
			return i
		}
	}
	return 0
}

func (t *TabPane) RemovePane(i int) {
	copy(t.Panes[i:], t.Panes[i+1:])
	t.Panes[len(t.Panes)-1] = nil // or the zero value of T
	t.Panes = t.Panes[:len(t.Panes)-1]
}

func (t *TabPane) Resize() {
	for i, p := range t.Panes {
		n := t.GetNode(p.splitID)
		pv := p.GetView()
		offset := 0
		if i != 0 {
			offset = 1
		}
		pv.X, pv.Y = n.X+offset, n.Y
		p.SetView(pv)
		p.Resize(n.W-offset, n.H)
	}
}

func (t *TabPane) CurPane() *EditPane {
	return t.Panes[t.active]
}
