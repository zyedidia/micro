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
	e.Window = display.NewBufWindow(x, y, width, height, b)
	e.Handler = action.NewBufHandler(b)

	return e
}
