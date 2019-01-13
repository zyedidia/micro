package action

import (
	"time"

	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/display"
	"github.com/zyedidia/micro/cmd/micro/util"
	"github.com/zyedidia/tcell"
)

type BufKeyAction func(*BufHandler) bool
type BufMouseAction func(*BufHandler, *tcell.EventMouse) bool

var BufKeyBindings map[Event]BufKeyAction
var BufKeyStrings map[Event]string
var BufMouseBindings map[MouseEvent]BufMouseAction

func init() {
	BufKeyBindings = make(map[Event]BufKeyAction)
	BufKeyStrings = make(map[Event]string)
	BufMouseBindings = make(map[MouseEvent]BufMouseAction)
}

// BufMapKey maps a key event to an action
func BufMapKey(k Event, action string) {
	if f, ok := BufKeyActions[action]; ok {
		BufKeyStrings[k] = action
		BufKeyBindings[k] = f
	} else {
		util.TermMessage("Error:", action, "does not exist")
	}
}

// BufMapMouse maps a mouse event to an action
func BufMapMouse(k MouseEvent, action string) {
	if f, ok := BufMouseActions[action]; ok {
		BufMouseBindings[k] = f
	} else if f, ok := BufKeyActions[action]; ok {
		// allowed to map mouse buttons to key actions
		BufKeyStrings[k] = action
		BufKeyBindings[k] = f
		// ensure we don't double bind a key
		delete(BufMouseBindings, k)
	} else {
		util.TermMessage("Error:", action, "does not exist")
	}
}

// The BufHandler connects the buffer and the window
// It provides a cursor (or multiple) and defines a set of actions
// that can be taken on the buffer
// The ActionHandler can access the window for necessary info about
// visual positions for mouse clicks and scrolling
type BufHandler struct {
	display.Window

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

	// Last search stores the last successful search for FindNext and FindPrev
	lastSearch string
	// Should the current multiple cursor selection search based on word or
	// based on selection (false for selection, true for word)
	multiWord bool

	splitID uint64
}

func NewBufHandler(buf *buffer.Buffer, win display.Window) *BufHandler {
	h := new(BufHandler)
	h.Buf = buf
	h.Window = win

	h.cursors = []*buffer.Cursor{buffer.NewCursor(buf, buf.StartCursor)}
	h.Cursor = h.cursors[0]
	h.mouseReleased = true

	buf.SetCursors(h.cursors)
	return h
}

func (h *BufHandler) ID() uint64 {
	return h.splitID
}

func (h *BufHandler) Name() string {
	return h.Buf.GetName()
}

// HandleEvent executes the tcell event properly
// TODO: multiple actions bound to one key
func (h *BufHandler) HandleEvent(event tcell.Event) {
	switch e := event.(type) {
	case *tcell.EventRaw:
		re := RawEvent{
			esc: e.EscSeq(),
		}
		h.DoKeyEvent(re)
	case *tcell.EventKey:
		ke := KeyEvent{
			code: e.Key(),
			mod:  e.Modifiers(),
			r:    e.Rune(),
		}

		done := h.DoKeyEvent(ke)
		if !done && e.Key() == tcell.KeyRune {
			h.DoRuneInsert(e.Rune())
		}
	case *tcell.EventMouse:
		switch e.Buttons() {
		case tcell.ButtonNone:
			// Mouse event with no click
			if !h.mouseReleased {
				// Mouse was just released

				mx, my := e.Position()
				mouseLoc := h.GetMouseLoc(buffer.Loc{X: mx, Y: my})

				// Relocating here isn't really necessary because the cursor will
				// be in the right place from the last mouse event
				// However, if we are running in a terminal that doesn't support mouse motion
				// events, this still allows the user to make selections, except only after they
				// release the mouse

				if !h.doubleClick && !h.tripleClick {
					h.Cursor.Loc = mouseLoc
					h.Cursor.SetSelectionEnd(h.Cursor.Loc)
					h.Cursor.CopySelection("primary")
				}
				h.mouseReleased = true
			}
		}

		me := MouseEvent{
			btn: e.Buttons(),
			mod: e.Modifiers(),
		}
		h.DoMouseEvent(me, e)
	}
	h.Buf.MergeCursors()

	// Display any gutter messages for this line
	c := h.Buf.GetActiveCursor()
	none := true
	for _, m := range h.Buf.Messages {
		if c.Y == m.Start.Y || c.Y == m.End.Y {
			InfoBar.GutterMessage(m.Msg)
			none = false
			break
		}
	}
	if none && InfoBar.HasGutter {
		InfoBar.ClearGutter()
	}
}

// DoKeyEvent executes a key event by finding the action it is bound
// to and executing it (possibly multiple times for multiple cursors)
func (h *BufHandler) DoKeyEvent(e Event) bool {
	if action, ok := BufKeyBindings[e]; ok {
		estr := BufKeyStrings[e]
		for _, s := range MultiActions {
			if s == estr {
				cursors := h.Buf.GetCursors()
				for _, c := range cursors {
					h.Buf.SetCurCursor(c.Num)
					h.Cursor = c
					if action(h) {
						h.Relocate()
					}
				}
				return true
			}
		}
		if action(h) {
			h.Relocate()
		}
		return true
	}
	return false
}

func (h *BufHandler) HasKeyEvent(e Event) bool {
	_, ok := BufKeyBindings[e]
	return ok
}

// DoMouseEvent executes a mouse event by finding the action it is bound
// to and executing it
func (h *BufHandler) DoMouseEvent(e MouseEvent, te *tcell.EventMouse) bool {
	if action, ok := BufMouseBindings[e]; ok {
		if action(h, te) {
			h.Relocate()
		}
		return true
	} else if h.HasKeyEvent(e) {
		return h.DoKeyEvent(e)
	}
	return false
}

// DoRuneInsert inserts a given rune into the current buffer
// (possibly multiple times for multiple cursors)
func (h *BufHandler) DoRuneInsert(r rune) {
	cursors := h.Buf.GetCursors()
	for _, c := range cursors {
		// Insert a character
		if c.HasSelection() {
			c.DeleteSelection()
			c.ResetSelection()
		}

		if h.isOverwriteMode {
			next := c.Loc
			next.X++
			h.Buf.Replace(c.Loc, next, string(r))
		} else {
			h.Buf.Insert(c.Loc, string(r))
		}
	}
}

func (h *BufHandler) VSplitBuf(buf *buffer.Buffer) {
	e := NewBufEditPane(0, 0, 0, 0, buf)
	e.splitID = MainTab().GetNode(h.splitID).VSplit(h.Buf.Settings["splitright"].(bool))
	MainTab().Panes = append(MainTab().Panes, e)
	MainTab().Resize()
	MainTab().SetActive(len(MainTab().Panes) - 1)
}
func (h *BufHandler) HSplitBuf(buf *buffer.Buffer) {
	e := NewBufEditPane(0, 0, 0, 0, buf)
	e.splitID = MainTab().GetNode(h.splitID).HSplit(h.Buf.Settings["splitbottom"].(bool))
	MainTab().Panes = append(MainTab().Panes, e)
	MainTab().Resize()
	MainTab().SetActive(len(MainTab().Panes) - 1)
}

// BufKeyActions contains the list of all possible key actions the bufhandler could execute
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
	"VSplit":                 (*BufHandler).VSplitAction,
	"HSplit":                 (*BufHandler).HSplitAction,
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

// BufMouseActions contains the list of all possible mouse actions the bufhandler could execute
var BufMouseActions = map[string]BufMouseAction{
	"MousePress":       (*BufHandler).MousePress,
	"MouseMultiCursor": (*BufHandler).MouseMultiCursor,
}

// MultiActions is a list of actions that should be executed multiple
// times if there are multiple cursors (one per cursor)
// Generally actions that modify global editor state like quitting or
// saving should not be included in this list
var MultiActions = []string{
	"CursorUp",
	"CursorDown",
	"CursorPageUp",
	"CursorPageDown",
	"CursorLeft",
	"CursorRight",
	"CursorStart",
	"CursorEnd",
	"SelectToStart",
	"SelectToEnd",
	"SelectUp",
	"SelectDown",
	"SelectLeft",
	"SelectRight",
	"WordRight",
	"WordLeft",
	"SelectWordRight",
	"SelectWordLeft",
	"DeleteWordRight",
	"DeleteWordLeft",
	"SelectLine",
	"SelectToStartOfLine",
	"SelectToEndOfLine",
	"ParagraphPrevious",
	"ParagraphNext",
	"InsertNewline",
	"Backspace",
	"Delete",
	"InsertTab",
	"FindNext",
	"FindPrevious",
	"Cut",
	"CutLine",
	"DuplicateLine",
	"DeleteLine",
	"MoveLinesUp",
	"MoveLinesDown",
	"IndentSelection",
	"OutdentSelection",
	"OutdentLine",
	"Paste",
	"PastePrimary",
	"SelectPageUp",
	"SelectPageDown",
	"StartOfLine",
	"EndOfLine",
	"JumpToMatchingBrace",
}
