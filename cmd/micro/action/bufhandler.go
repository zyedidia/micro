package action

import (
	"time"

	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/tcell"
)

type BufKeyAction func(*BufHandler) bool
type BufMouseAction func(*BufHandler, *tcell.EventMouse) bool

var BufKeyBindings map[KeyEvent]BufKeyAction
var BufMouseBindings map[MouseEvent]BufMouseAction

func init() {
	BufKeyBindings = make(map[KeyEvent]BufKeyAction)
	BufMouseBindings = make(map[MouseEvent]BufMouseAction)
}

func BufMapKey(k KeyEvent, action string) {
	BufKeyBindings[k] = BufKeyActions[action]
}
func BufMapMouse(k MouseEvent, action string) {
	BufMouseBindings[k] = BufMouseActions[action]
}

// The BufHandler connects the buffer and the window
// It provides a cursor (or multiple) and defines a set of actions
// that can be taken on the buffer
// The ActionHandler can access the window for necessary info about
// visual positions for mouse clicks and scrolling
type BufHandler struct {
	Buf *buffer.Buffer

	cursors []*buffer.Cursor
	Cursor  *buffer.Cursor // the active cursor

	StartLine int // Vertical scrolling
	StartCol  int // Horizontal scrolling

	// Since tcell doesn't differentiate between a mouse release event
	// and a mouse move event with no keys pressed, we need to keep
	// track of whether or not the mouse was pressed (or not released) last event to determine
	// mouse release events
	mouseReleased bool

	// We need to keep track of insert key press toggle
	isOverwriteMode bool
	// This stores when the last click was
	// This is useful for detecting double and triple clicks
	lastClickTime time.Time
	lastLoc       buffer.Loc

	// lastCutTime stores when the last ctrl+k was issued.
	// It is used for clearing the clipboard to replace it with fresh cut lines.
	lastCutTime time.Time

	// freshClip returns true if the clipboard has never been pasted.
	freshClip bool

	// Was the last mouse event actually a double click?
	// Useful for detecting triple clicks -- if a double click is detected
	// but the last mouse event was actually a double click, it's a triple click
	doubleClick bool
	// Same here, just to keep track for mouse move events
	tripleClick bool
}

func NewBufHandler(buf *buffer.Buffer) *BufHandler {
	h := new(BufHandler)
	h.Buf = buf

	h.cursors = []*buffer.Cursor{&buffer.Cursor{
		Buf: buf,
		Loc: buf.StartCursor,
	}}
	h.Cursor = h.cursors[0]

	buf.SetCursors(h.cursors)
	return h
}

// HandleEvent executes the tcell event properly
// TODO: multiple actions bound to one key
func (h *BufHandler) HandleEvent(event tcell.Event) {
	switch e := event.(type) {
	case *tcell.EventKey:
		ke := KeyEvent{
			code: e.Key(),
			mod:  e.Modifiers(),
			r:    e.Rune(),
		}
		h.DoKeyEvent(ke)
	case *tcell.EventMouse:
		me := MouseEvent{
			btn: e.Buttons(),
			mod: e.Modifiers(),
		}
		h.DoMouseEvent(me, e)
	}
}

func (h *BufHandler) DoKeyEvent(e KeyEvent) bool {
	if action, ok := BufKeyBindings[e]; ok {
		action(h)
		return true
	}
	return false
}

func (h *BufHandler) DoMouseEvent(e MouseEvent, te *tcell.EventMouse) bool {
	if action, ok := BufMouseBindings[e]; ok {
		action(h, te)
		return true
	}
	return false
}

var BufKeyActions = map[string]BufKeyAction{
	"CursorUp":               (*BufHandler).CursorUp,
	"CursorDown":             (*BufHandler).CursorDown,
	"CursorPageUp":           (*BufHandler).CursorPageUp,
	"CursorPageDown":         (*BufHandler).CursorPageDown,
	"CursorLeft":             (*BufHandler).CursorLeft,
	"CursorRight":            (*BufHandler).CursorRight,
	"CursorStart":            (*BufHandler).CursorStart,
	"CursorEnd":              (*BufHandler).CursorEnd,
	"SelectToStart":          (*BufHandler).SelectToStart,
	"SelectToEnd":            (*BufHandler).SelectToEnd,
	"SelectUp":               (*BufHandler).SelectUp,
	"SelectDown":             (*BufHandler).SelectDown,
	"SelectLeft":             (*BufHandler).SelectLeft,
	"SelectRight":            (*BufHandler).SelectRight,
	"WordRight":              (*BufHandler).WordRight,
	"WordLeft":               (*BufHandler).WordLeft,
	"SelectWordRight":        (*BufHandler).SelectWordRight,
	"SelectWordLeft":         (*BufHandler).SelectWordLeft,
	"DeleteWordRight":        (*BufHandler).DeleteWordRight,
	"DeleteWordLeft":         (*BufHandler).DeleteWordLeft,
	"SelectLine":             (*BufHandler).SelectLine,
	"SelectToStartOfLine":    (*BufHandler).SelectToStartOfLine,
	"SelectToEndOfLine":      (*BufHandler).SelectToEndOfLine,
	"ParagraphPrevious":      (*BufHandler).ParagraphPrevious,
	"ParagraphNext":          (*BufHandler).ParagraphNext,
	"InsertNewline":          (*BufHandler).InsertNewline,
	"InsertSpace":            (*BufHandler).InsertSpace,
	"Backspace":              (*BufHandler).Backspace,
	"Delete":                 (*BufHandler).Delete,
	"InsertTab":              (*BufHandler).InsertTab,
	"Save":                   (*BufHandler).Save,
	"SaveAll":                (*BufHandler).SaveAll,
	"SaveAs":                 (*BufHandler).SaveAs,
	"Find":                   (*BufHandler).Find,
	"FindNext":               (*BufHandler).FindNext,
	"FindPrevious":           (*BufHandler).FindPrevious,
	"Center":                 (*BufHandler).Center,
	"Undo":                   (*BufHandler).Undo,
	"Redo":                   (*BufHandler).Redo,
	"Copy":                   (*BufHandler).Copy,
	"Cut":                    (*BufHandler).Cut,
	"CutLine":                (*BufHandler).CutLine,
	"DuplicateLine":          (*BufHandler).DuplicateLine,
	"DeleteLine":             (*BufHandler).DeleteLine,
	"MoveLinesUp":            (*BufHandler).MoveLinesUp,
	"MoveLinesDown":          (*BufHandler).MoveLinesDown,
	"IndentSelection":        (*BufHandler).IndentSelection,
	"OutdentSelection":       (*BufHandler).OutdentSelection,
	"OutdentLine":            (*BufHandler).OutdentLine,
	"Paste":                  (*BufHandler).Paste,
	"PastePrimary":           (*BufHandler).PastePrimary,
	"SelectAll":              (*BufHandler).SelectAll,
	"OpenFile":               (*BufHandler).OpenFile,
	"Start":                  (*BufHandler).Start,
	"End":                    (*BufHandler).End,
	"PageUp":                 (*BufHandler).PageUp,
	"PageDown":               (*BufHandler).PageDown,
	"SelectPageUp":           (*BufHandler).SelectPageUp,
	"SelectPageDown":         (*BufHandler).SelectPageDown,
	"HalfPageUp":             (*BufHandler).HalfPageUp,
	"HalfPageDown":           (*BufHandler).HalfPageDown,
	"StartOfLine":            (*BufHandler).StartOfLine,
	"EndOfLine":              (*BufHandler).EndOfLine,
	"ToggleHelp":             (*BufHandler).ToggleHelp,
	"ToggleKeyMenu":          (*BufHandler).ToggleKeyMenu,
	"ToggleRuler":            (*BufHandler).ToggleRuler,
	"JumpLine":               (*BufHandler).JumpLine,
	"ClearStatus":            (*BufHandler).ClearStatus,
	"ShellMode":              (*BufHandler).ShellMode,
	"CommandMode":            (*BufHandler).CommandMode,
	"ToggleOverwriteMode":    (*BufHandler).ToggleOverwriteMode,
	"Escape":                 (*BufHandler).Escape,
	"Quit":                   (*BufHandler).Quit,
	"QuitAll":                (*BufHandler).QuitAll,
	"AddTab":                 (*BufHandler).AddTab,
	"PreviousTab":            (*BufHandler).PreviousTab,
	"NextTab":                (*BufHandler).NextTab,
	"NextSplit":              (*BufHandler).NextSplit,
	"PreviousSplit":          (*BufHandler).PreviousSplit,
	"Unsplit":                (*BufHandler).Unsplit,
	"VSplit":                 (*BufHandler).VSplitBinding,
	"HSplit":                 (*BufHandler).HSplitBinding,
	"ToggleMacro":            (*BufHandler).ToggleMacro,
	"PlayMacro":              (*BufHandler).PlayMacro,
	"Suspend":                (*BufHandler).Suspend,
	"ScrollUp":               (*BufHandler).ScrollUpAction,
	"ScrollDown":             (*BufHandler).ScrollDownAction,
	"SpawnMultiCursor":       (*BufHandler).SpawnMultiCursor,
	"SpawnMultiCursorSelect": (*BufHandler).SpawnMultiCursorSelect,
	"RemoveMultiCursor":      (*BufHandler).RemoveMultiCursor,
	"RemoveAllMultiCursors":  (*BufHandler).RemoveAllMultiCursors,
	"SkipMultiCursor":        (*BufHandler).SkipMultiCursor,
	"JumpToMatchingBrace":    (*BufHandler).JumpToMatchingBrace,

	// This was changed to InsertNewline but I don't want to break backwards compatibility
	"InsertEnter": (*BufHandler).InsertNewline,
}
var BufMouseActions = map[string]BufMouseAction{
	"MousePress":       (*BufHandler).MousePress,
	"MouseMultiCursor": (*BufHandler).MouseMultiCursor,
}
