package action

import (
	"log"

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
	t.Panes[t.active].HandleEvent(event)
}

func (t *TabPane) Resize() {
	for _, p := range t.Panes {
		log.Println(p.splitID)
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
