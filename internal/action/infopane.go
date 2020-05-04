package action

import (
	"bytes"
	"strings"

	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/display"
	"github.com/zyedidia/micro/v2/internal/info"
	"github.com/zyedidia/micro/v2/internal/util"
	"github.com/zyedidia/tcell"
)

type InfoKeyAction func(*InfoPane)

type InfoPane struct {
	*BufPane
	*info.InfoBuf
}

func NewInfoPane(ib *info.InfoBuf, w display.BWindow, tab *Tab) *InfoPane {
	ip := new(InfoPane)
	ip.InfoBuf = ib
	ip.BufPane = NewBufPane(ib.Buffer, w, tab)

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
			mod:  e.Modifiers(),
			r:    e.Rune(),
		}

		done := h.DoKeyEvent(ke)
		hasYN := h.HasYN
		if e.Key() == tcell.KeyRune && hasYN {
			if e.Rune() == 'y' && hasYN {
				h.YNResp = true
				h.DonePrompt(false)
			} else if e.Rune() == 'n' && hasYN {
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
	done := false
	if action, ok := BufKeyBindings[e]; ok {
		estr := BufKeyStrings[e]
		for _, s := range InfoNones {
			if s == estr {
				return false
			}
		}
		for s, a := range InfoOverrides {
			// TODO this is a hack and really we should have support
			// for having binding overrides for different buffers
			if strings.HasPrefix(estr, s) {
				done = true
				a(h)
				break
			}
		}
		if !done {
			done = action(h.BufPane)
		}
	}
	return done
}

// InfoNones is a list of actions that should have no effect when executed
// by an infohandler
var InfoNones = []string{
	"Save",
	"SaveAll",
	"SaveAs",
	"Find",
	"FindNext",
	"FindPrevious",
	"Center",
	"DuplicateLine",
	"MoveLinesUp",
	"MoveLinesDown",
	"OpenFile",
	"Start",
	"End",
	"PageUp",
	"PageDown",
	"SelectPageUp",
	"SelectPageDown",
	"HalfPageUp",
	"HalfPageDown",
	"ToggleHelp",
	"ToggleKeyMenu",
	"ToggleDiffGutter",
	"ToggleRuler",
	"JumpLine",
	"ClearStatus",
	"ShellMode",
	"CommandMode",
	"AddTab",
	"PreviousTab",
	"NextTab",
	"NextSplit",
	"PreviousSplit",
	"Unsplit",
	"VSplit",
	"HSplit",
	"ToggleMacro",
	"PlayMacro",
	"Suspend",
	"ScrollUp",
	"ScrollDown",
	"SpawnMultiCursor",
	"SpawnMultiCursorSelect",
	"RemoveMultiCursor",
	"RemoveAllMultiCursors",
	"SkipMultiCursor",
}

// InfoOverrides is the list of actions which have been overridden
// by the infohandler
var InfoOverrides = map[string]InfoKeyAction{
	"CursorUp":      (*InfoPane).CursorUp,
	"CursorDown":    (*InfoPane).CursorDown,
	"InsertNewline": (*InfoPane).InsertNewline,
	"Autocomplete":  (*InfoPane).Autocomplete,
	"Escape":        (*InfoPane).Escape,
	"Quit":          (*InfoPane).Quit,
	"QuitAll":       (*InfoPane).QuitAll,
}

// CursorUp cycles history up
func (h *InfoPane) CursorUp() {
	h.UpHistory(h.History[h.PromptType])
}

// CursorDown cycles history down
func (h *InfoPane) CursorDown() {
	h.DownHistory(h.History[h.PromptType])
}

// Autocomplete begins autocompletion
func (h *InfoPane) Autocomplete() {
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

// InsertNewline completes the prompt
func (h *InfoPane) InsertNewline() {
	if !h.HasYN {
		h.DonePrompt(false)
	}
}

// Quit cancels the prompt
func (h *InfoPane) Quit() {
	h.DonePrompt(true)
}

// QuitAll cancels the prompt
func (h *InfoPane) QuitAll() {
	h.DonePrompt(true)
}

// Escape cancels the prompt
func (h *InfoPane) Escape() {
	h.DonePrompt(true)
}
