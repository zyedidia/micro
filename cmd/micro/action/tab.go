package action

import (
	"github.com/zyedidia/micro/cmd/micro/views"
	"github.com/zyedidia/tcell"
)

var MainTab *TabPane

type TabPane struct {
	*views.Node
	Panes  []*EditPane
	active int
}

func (t *TabPane) HandleEvent(event tcell.Event) {
	switch e := event.(type) {
	case *tcell.EventMouse:
		switch e.Buttons() {
		case tcell.Button1:
			mx, my := e.Position()

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

func (t *TabPane) Resize() {
	for _, p := range t.Panes {
		v := t.GetNode(p.splitID).GetView()
		pv := p.GetView()
		pv.X, pv.Y = v.X, v.Y
		p.SetView(pv)
		p.Resize(v.W, v.H)
	}
}

func (t *TabPane) CurPane() *EditPane {
	return t.Panes[t.active]
}
