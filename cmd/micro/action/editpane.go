package action

import (
	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/display"
	"github.com/zyedidia/micro/cmd/micro/info"
	"github.com/zyedidia/micro/cmd/micro/views"
)

type EditPane struct {
	display.Window
	*BufHandler
}

type InfoPane struct {
	display.Window
	*InfoHandler
	*info.InfoBuf
}

func NewBufEditPane(x, y, width, height int, b *buffer.Buffer) *EditPane {
	e := new(EditPane)
	// TODO: can probably replace editpane with bufhandler entirely
	w := display.NewBufWindow(x, y, width, height, b)
	e.Window = w
	e.BufHandler = NewBufHandler(b, w)

	return e
}

func NewTabPane(width, height int, b *buffer.Buffer) *TabPane {
	t := new(TabPane)
	t.Node = views.NewRoot(0, 0, width, height)

	e := NewBufEditPane(0, 0, width, height, b)
	e.splitID = t.ID()

	t.Panes = append(t.Panes, e)
	return t
}

func NewInfoBar() *InfoPane {
	e := new(InfoPane)
	ib := info.NewBuffer()
	w := display.NewInfoWindow(ib)
	e.Window = w
	e.InfoHandler = NewInfoHandler(ib, w)
	e.InfoBuf = ib

	return e
}
