package action

import (
	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/display"
	"github.com/zyedidia/micro/cmd/micro/info"
)

type Pane interface {
	Handler
	display.Window
	ID() uint64
	Name() string
	Close()
}

type EditPane struct {
	display.BWindow
	*BufHandler
}

type InfoPane struct {
	display.BWindow
	*InfoHandler
	*info.InfoBuf
}

func NewBufEditPane(x, y, width, height int, b *buffer.Buffer) *EditPane {
	e := new(EditPane)
	// TODO: can probably replace editpane with bufhandler entirely
	w := display.NewBufWindow(x, y, width, height, b)
	e.BWindow = w
	e.BufHandler = NewBufHandler(b, w)

	return e
}

func NewInfoBar() *InfoPane {
	e := new(InfoPane)
	ib := info.NewBuffer()
	w := display.NewInfoWindow(ib)
	e.BWindow = w
	e.InfoHandler = NewInfoHandler(ib, w)
	e.InfoBuf = ib

	return e
}
