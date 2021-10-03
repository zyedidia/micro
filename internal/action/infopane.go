package action

import (
	"bytes"

	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/display"
	"github.com/zyedidia/micro/v2/internal/info"
	"github.com/zyedidia/micro/v2/internal/util"
	"github.com/zyedidia/tcell/v2"
)

type InfoKeyAction func(*InfoPane)

var InfoBindings *KeyTree
var InfoBufBindings *KeyTree

func init() {
	InfoBindings = NewKeyTree()
	InfoBufBindings = NewKeyTree()
}

func InfoMapEvent(k Event, action string) {
	config.Bindings["command"][k.Name()] = action

	switch e := k.(type) {
	case KeyEvent, KeySequenceEvent, RawEvent:
		infoMapKey(e, action)
	case MouseEvent:
		infoMapMouse(e, action)
	}
}

func infoMapKey(k Event, action string) {
	if f, ok := InfoKeyActions[action]; ok {
		InfoBindings.RegisterKeyBinding(k, InfoKeyActionGeneral(f))
	} else if f, ok := BufKeyActions[action]; ok {
		InfoBufBindings.RegisterKeyBinding(k, BufKeyActionGeneral(f))
	}
}

func infoMapMouse(k MouseEvent, action string) {
	// TODO: map mouse
	if f, ok := BufMouseActions[action]; ok {
		InfoBufBindings.RegisterMouseBinding(k, BufMouseActionGeneral(f))
	} else {
		infoMapKey(k, action)
	}
}

func InfoKeyActionGeneral(a InfoKeyAction) PaneKeyAction {
	return func(p Pane) bool {
		a(p.(*InfoPane))
		return true
	}
}

type InfoPane struct {
	*BufPane
	*info.InfoBuf
}

func NewInfoPane(ib *info.InfoBuf, w display.BWindow, tab *Tab) *InfoPane {
	ip := new(InfoPane)
	ip.InfoBuf = ib
	ip.BufPane = NewBufPane(ib.Buffer, w, tab)
	ip.BufPane.bindings = InfoBufBindings

	return ip
}

func NewInfoBar() *InfoPane {
	ib := info.NewBuffer()
	w := display.NewInfoWindow(ib)
	return NewInfoPane(ib, w, nil)
}

func (h *InfoPane) Close() {
	h.InfoBuf.Close()
	h.BufPane.Close()
}

func (h *InfoPane) HandleEvent(event tcell.Event) {
	switch e := event.(type) {
	case *tcell.EventKey:
		ke := KeyEvent{
			code: e.Key(),
			mod:  metaToAlt(e.Modifiers()),
			r:    e.Rune(),
		}

		done := h.DoKeyEvent(ke)
		hasYN := h.HasYN
		if e.Key() == tcell.KeyRune && hasYN {
			if (e.Rune() == 'y' || e.Rune() == 'Y') && hasYN {
				h.YNResp = true
				h.DonePrompt(false)
			} else if (e.Rune() == 'n' || e.Rune() == 'N') && hasYN {
				h.YNResp = false
				h.DonePrompt(false)
			}
		}
		if e.Key() == tcell.KeyRune && !done && !hasYN {
			h.DoRuneInsert(e.Rune())
			done = true
		}
		if done && h.HasPrompt && !hasYN {
			resp := string(h.LineBytes(0))
			hist := h.History[h.PromptType]
			hist[h.HistoryNum] = resp
			if h.EventCallback != nil {
				h.EventCallback(resp)
			}
		}
	default:
		h.BufPane.HandleEvent(event)
	}
}

// DoKeyEvent executes a key event for the command bar, doing any overridden actions
func (h *InfoPane) DoKeyEvent(e KeyEvent) bool {
	action, more := InfoBindings.NextEvent(e, nil)
	if action != nil && !more {
		action(h)
		InfoBindings.ResetEvents()

		return true
	} else if action == nil && !more {
		InfoBindings.ResetEvents()
		// return false //TODO:?
	}

	if !more {
		action, more = InfoBufBindings.NextEvent(e, nil)
		if action != nil && !more {
			done := action(h.BufPane)
			InfoBufBindings.ResetEvents()
			return done
		} else if action == nil && !more {
			InfoBufBindings.ResetEvents()
		}
	}

	return more
}

// HistoryUp cycles history up
func (h *InfoPane) HistoryUp() {
	h.UpHistory(h.History[h.PromptType])
}

// HistoryDown cycles history down
func (h *InfoPane) HistoryDown() {
	h.DownHistory(h.History[h.PromptType])
}

// Autocomplete begins autocompletion
func (h *InfoPane) CommandComplete() {
	b := h.Buf
	if b.HasSuggestions {
		b.CycleAutocomplete(true)
		return
	}

	c := b.GetActiveCursor()
	l := b.LineBytes(0)
	l = util.SliceStart(l, c.X)

	args := bytes.Split(l, []byte{' '})
	cmd := string(args[0])

	if h.PromptType == "Command" {
		if len(args) == 1 {
			b.Autocomplete(CommandComplete)
		} else if action, ok := commands[cmd]; ok {
			if action.completer != nil {
				b.Autocomplete(action.completer)
			}
		}
	} else {
		// by default use filename autocompletion
		b.Autocomplete(buffer.FileComplete)
	}
}

// ExecuteCommand completes the prompt
func (h *InfoPane) ExecuteCommand() {
	if !h.HasYN {
		h.DonePrompt(false)
	}
}

// AbortCommand cancels the prompt
func (h *InfoPane) AbortCommand() {
	h.DonePrompt(true)
}

// InfoKeyActions contains the list of all possible key actions the infopane could execute
var InfoKeyActions = map[string]InfoKeyAction{
	"HistoryUp":       (*InfoPane).HistoryUp,
	"HistoryDown":     (*InfoPane).HistoryDown,
	"CommandComplete": (*InfoPane).CommandComplete,
	"ExecuteCommand":  (*InfoPane).ExecuteCommand,
	"AbortCommand":    (*InfoPane).AbortCommand,
}
