package action

import (
	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/display"
	"github.com/zyedidia/micro/cmd/micro/info"
)

type EditPane struct {
	display.Window
	Handler
}

type InfoPane struct {
	display.Window
	Handler
	*info.InfoBuf
}

func NewBufEditPane(x, y, width, height int, b *buffer.Buffer) *EditPane {
	e := new(EditPane)
	// TODO: can probably replace editpane with bufhandler entirely
	w := display.NewBufWindow(x, y, width, height, b)
	e.Window = w
	e.Handler = NewBufHandler(b, w)

	return e
}

func NewInfoBar() *InfoPane {
	e := new(InfoPane)
	ib := info.NewBuffer()
	w := display.NewInfoWindow(ib)
	e.Window = w
	e.Handler = NewBufHandler(ib.Buffer, w)
	e.InfoBuf = ib

	return e
}
