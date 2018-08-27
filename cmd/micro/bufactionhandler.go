package main

import (
	"time"

	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/tcell"
)

type BufKeyAction func(*BufActionHandler) bool
type BufMouseAction func(*BufActionHandler, *tcell.EventMouse) bool

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

// The BufActionHandler connects the buffer and the window
// It provides a cursor (or multiple) and defines a set of actions
// that can be taken on the buffer
// The ActionHandler can access the window for necessary info about
// visual positions for mouse clicks and scrolling
type BufActionHandler struct {
	Buf *buffer.Buffer
	Win *Window

	cursors []*buffer.Cursor
	Cursor  *buffer.Cursor // the active cursor

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

func NewBufActionHandler(buf *buffer.Buffer, win *Window) *BufActionHandler {
	a := new(BufActionHandler)
	a.Buf = buf
	a.Win = win

	a.cursors = []*buffer.Cursor{&buffer.Cursor{
		Buf: buf,
		Loc: buf.StartCursor,
	}}

	buf.SetCursors(a.cursors)
	return a
}

// HandleEvent executes the tcell event properly
// TODO: multiple actions bound to one key
func (a *BufActionHandler) HandleEvent(event tcell.Event) {
	switch e := event.(type) {
	case *tcell.EventKey:
		ke := KeyEvent{
			code: e.Key(),
			mod:  e.Modifiers(),
			r:    e.Rune(),
		}
		if action, ok := BufKeyBindings[ke]; ok {
			action(a)
		}
	case *tcell.EventMouse:
		me := MouseEvent{
			btn: e.Buttons(),
			mod: e.Modifiers(),
		}
		if action, ok := BufMouseBindings[me]; ok {
			action(a, e)
		}
	}
}

var BufKeyActions = map[string]BufKeyAction{
	"CursorUp":               (*BufActionHandler).CursorUp,
	"CursorDown":             (*BufActionHandler).CursorDown,
	"CursorPageUp":           (*BufActionHandler).CursorPageUp,
	"CursorPageDown":         (*BufActionHandler).CursorPageDown,
	"CursorLeft":             (*BufActionHandler).CursorLeft,
	"CursorRight":            (*BufActionHandler).CursorRight,
	"CursorStart":            (*BufActionHandler).CursorStart,
	"CursorEnd":              (*BufActionHandler).CursorEnd,
	"SelectToStart":          (*BufActionHandler).SelectToStart,
	"SelectToEnd":            (*BufActionHandler).SelectToEnd,
	"SelectUp":               (*BufActionHandler).SelectUp,
	"SelectDown":             (*BufActionHandler).SelectDown,
	"SelectLeft":             (*BufActionHandler).SelectLeft,
	"SelectRight":            (*BufActionHandler).SelectRight,
	"WordRight":              (*BufActionHandler).WordRight,
	"WordLeft":               (*BufActionHandler).WordLeft,
	"SelectWordRight":        (*BufActionHandler).SelectWordRight,
	"SelectWordLeft":         (*BufActionHandler).SelectWordLeft,
	"DeleteWordRight":        (*BufActionHandler).DeleteWordRight,
	"DeleteWordLeft":         (*BufActionHandler).DeleteWordLeft,
	"SelectLine":             (*BufActionHandler).SelectLine,
	"SelectToStartOfLine":    (*BufActionHandler).SelectToStartOfLine,
	"SelectToEndOfLine":      (*BufActionHandler).SelectToEndOfLine,
	"ParagraphPrevious":      (*BufActionHandler).ParagraphPrevious,
	"ParagraphNext":          (*BufActionHandler).ParagraphNext,
	"InsertNewline":          (*BufActionHandler).InsertNewline,
	"InsertSpace":            (*BufActionHandler).InsertSpace,
	"Backspace":              (*BufActionHandler).Backspace,
	"Delete":                 (*BufActionHandler).Delete,
	"InsertTab":              (*BufActionHandler).InsertTab,
	"Save":                   (*BufActionHandler).Save,
	"SaveAll":                (*BufActionHandler).SaveAll,
	"SaveAs":                 (*BufActionHandler).SaveAs,
	"Find":                   (*BufActionHandler).Find,
	"FindNext":               (*BufActionHandler).FindNext,
	"FindPrevious":           (*BufActionHandler).FindPrevious,
	"Center":                 (*BufActionHandler).Center,
	"Undo":                   (*BufActionHandler).Undo,
	"Redo":                   (*BufActionHandler).Redo,
	"Copy":                   (*BufActionHandler).Copy,
	"Cut":                    (*BufActionHandler).Cut,
	"CutLine":                (*BufActionHandler).CutLine,
	"DuplicateLine":          (*BufActionHandler).DuplicateLine,
	"DeleteLine":             (*BufActionHandler).DeleteLine,
	"MoveLinesUp":            (*BufActionHandler).MoveLinesUp,
	"MoveLinesDown":          (*BufActionHandler).MoveLinesDown,
	"IndentSelection":        (*BufActionHandler).IndentSelection,
	"OutdentSelection":       (*BufActionHandler).OutdentSelection,
	"OutdentLine":            (*BufActionHandler).OutdentLine,
	"Paste":                  (*BufActionHandler).Paste,
	"PastePrimary":           (*BufActionHandler).PastePrimary,
	"SelectAll":              (*BufActionHandler).SelectAll,
	"OpenFile":               (*BufActionHandler).OpenFile,
	"Start":                  (*BufActionHandler).Start,
	"End":                    (*BufActionHandler).End,
	"PageUp":                 (*BufActionHandler).PageUp,
	"PageDown":               (*BufActionHandler).PageDown,
	"SelectPageUp":           (*BufActionHandler).SelectPageUp,
	"SelectPageDown":         (*BufActionHandler).SelectPageDown,
	"HalfPageUp":             (*BufActionHandler).HalfPageUp,
	"HalfPageDown":           (*BufActionHandler).HalfPageDown,
	"StartOfLine":            (*BufActionHandler).StartOfLine,
	"EndOfLine":              (*BufActionHandler).EndOfLine,
	"ToggleHelp":             (*BufActionHandler).ToggleHelp,
	"ToggleKeyMenu":          (*BufActionHandler).ToggleKeyMenu,
	"ToggleRuler":            (*BufActionHandler).ToggleRuler,
	"JumpLine":               (*BufActionHandler).JumpLine,
	"ClearStatus":            (*BufActionHandler).ClearStatus,
	"ShellMode":              (*BufActionHandler).ShellMode,
	"CommandMode":            (*BufActionHandler).CommandMode,
	"ToggleOverwriteMode":    (*BufActionHandler).ToggleOverwriteMode,
	"Escape":                 (*BufActionHandler).Escape,
	"Quit":                   (*BufActionHandler).Quit,
	"QuitAll":                (*BufActionHandler).QuitAll,
	"AddTab":                 (*BufActionHandler).AddTab,
	"PreviousTab":            (*BufActionHandler).PreviousTab,
	"NextTab":                (*BufActionHandler).NextTab,
	"NextSplit":              (*BufActionHandler).NextSplit,
	"PreviousSplit":          (*BufActionHandler).PreviousSplit,
	"Unsplit":                (*BufActionHandler).Unsplit,
	"VSplit":                 (*BufActionHandler).VSplitBinding,
	"HSplit":                 (*BufActionHandler).HSplitBinding,
	"ToggleMacro":            (*BufActionHandler).ToggleMacro,
	"PlayMacro":              (*BufActionHandler).PlayMacro,
	"Suspend":                (*BufActionHandler).Suspend,
	"ScrollUp":               (*BufActionHandler).ScrollUpAction,
	"ScrollDown":             (*BufActionHandler).ScrollDownAction,
	"SpawnMultiCursor":       (*BufActionHandler).SpawnMultiCursor,
	"SpawnMultiCursorSelect": (*BufActionHandler).SpawnMultiCursorSelect,
	"RemoveMultiCursor":      (*BufActionHandler).RemoveMultiCursor,
	"RemoveAllMultiCursors":  (*BufActionHandler).RemoveAllMultiCursors,
	"SkipMultiCursor":        (*BufActionHandler).SkipMultiCursor,
	"JumpToMatchingBrace":    (*BufActionHandler).JumpToMatchingBrace,

	// This was changed to InsertNewline but I don't want to break backwards compatibility
	"InsertEnter": (*BufActionHandler).InsertNewline,
}
var BufMouseActions = map[string]BufMouseAction{
	"MousePress":       (*BufActionHandler).MousePress,
	"MouseMultiCursor": (*BufActionHandler).MouseMultiCursor,
}
