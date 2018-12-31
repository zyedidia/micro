package main

import (
	"github.com/zyedidia/micro/cmd/micro/action"
	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/display"
)

type EditPane struct {
	display.Window
	action.Handler
}

func NewBufEditPane(x, y, width, height int, b *buffer.Buffer) *EditPane {
	e := new(EditPane)
	// TODO: can probably replace editpane with bufhandler entirely
	w := display.NewBufWindow(x, y, width, height, b)
	e.Window = w
	e.Handler = action.NewBufHandler(b, w)

	return e
}
