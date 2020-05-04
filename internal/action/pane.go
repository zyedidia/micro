package action

import (
	"github.com/zyedidia/micro/v2/internal/display"
)

type Pane interface {
	Handler
	display.Window
	ID() uint64
	SetID(i uint64)
	Name() string
	Close()
	SetTab(t *Tab)
	Tab() *Tab
}
