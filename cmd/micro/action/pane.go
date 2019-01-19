package action

import (
	"github.com/zyedidia/micro/cmd/micro/display"
)

type Pane interface {
	Handler
	display.Window
	ID() uint64
	SetID(i uint64)
	Name() string
	Close()
}
