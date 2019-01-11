package action

import (
	"github.com/zyedidia/clipboard"
	"github.com/zyedidia/micro/cmd/micro/display"
	"github.com/zyedidia/micro/cmd/micro/shell"
	"github.com/zyedidia/tcell"
	"github.com/zyedidia/terminal"
)

type TermHandler struct {
	*shell.Terminal
	display.Window

	mouseReleased bool
}

// HandleEvent handles a tcell event by forwarding it to the terminal emulator
// If the event is a mouse event and the program running in the emulator
// does not have mouse support, the emulator will support selections and
// copy-paste
func (t *TermHandler) HandleEvent(event tcell.Event) {
	if e, ok := event.(*tcell.EventKey); ok {
		if t.Status == shell.TTDone {
			switch e.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlQ, tcell.KeyEnter:
				t.Close()
			default:
			}
		}
		if e.Key() == tcell.KeyCtrlC && t.HasSelection() {
			clipboard.WriteAll(t.GetSelection(t.GetView().Width), "clipboard")
			InfoBar.Message("Copied selection to clipboard")
		} else if t.Status != shell.TTDone {
			t.WriteString(event.EscSeq())
		}
	} else if e, ok := event.(*tcell.EventMouse); !ok || t.State.Mode(terminal.ModeMouseMask) {
		t.WriteString(event.EscSeq())
	} else {
		x, y := e.Position()
		v := t.GetView()
		x -= v.X
		y += v.Y

		if e.Buttons() == tcell.Button1 {
			if !t.mouseReleased {
				// drag
				t.Selection[1].X = x
				t.Selection[1].Y = y
			} else {
				t.Selection[0].X = x
				t.Selection[0].Y = y
				t.Selection[1].X = x
				t.Selection[1].Y = y
			}

			t.mouseReleased = false
		} else if e.Buttons() == tcell.ButtonNone {
			if !t.mouseReleased {
				t.Selection[1].X = x
				t.Selection[1].Y = y
			}
			t.mouseReleased = true
		}
	}
}
