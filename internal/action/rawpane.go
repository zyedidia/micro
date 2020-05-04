package action

import (
	"fmt"
	"reflect"

	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/display"
	"github.com/zyedidia/tcell"
)

type RawPane struct {
	*BufPane
}

func NewRawPaneFromWin(b *buffer.Buffer, win display.BWindow, tab *Tab) *RawPane {
	rh := new(RawPane)
	rh.BufPane = NewBufPane(b, win, tab)

	return rh
}

func NewRawPane(tab *Tab) *RawPane {
	b := buffer.NewBufferFromString("", "", buffer.BTRaw)
	w := display.NewBufWindow(0, 0, 0, 0, b)
	return NewRawPaneFromWin(b, w, tab)
}

func (h *RawPane) HandleEvent(event tcell.Event) {
	switch e := event.(type) {
	case *tcell.EventKey:
		if e.Key() == tcell.KeyCtrlQ {
			h.Quit()
		}
	}

	h.Buf.Insert(h.Cursor.Loc, reflect.TypeOf(event).String()[7:])
	h.Buf.Insert(h.Cursor.Loc, fmt.Sprintf(": %q\n", event.EscSeq()))
	h.Relocate()
}
