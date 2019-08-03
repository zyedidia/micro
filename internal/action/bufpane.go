package action

import (
	"log"
	"strings"
	"time"

	luar "layeh.com/gopher-luar"

	lua "github.com/yuin/gopher-lua"
	"github.com/zyedidia/micro/internal/buffer"
	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/micro/internal/display"
	ulua "github.com/zyedidia/micro/internal/lua"
	"github.com/zyedidia/micro/internal/screen"
	"github.com/zyedidia/tcell"
)

type BufKeyAction func(*BufPane) bool
type BufMouseAction func(*BufPane, *tcell.EventMouse) bool

var BufKeyBindings map[Event]BufKeyAction
var BufKeyStrings map[Event]string
var BufMouseBindings map[MouseEvent]BufMouseAction

func init() {
	BufKeyBindings = make(map[Event]BufKeyAction)
	BufKeyStrings = make(map[Event]string)
	BufMouseBindings = make(map[MouseEvent]BufMouseAction)
}

func LuaAction(fn string) func(*BufPane) bool {
	luaFn := strings.Split(fn, ".")
	plName, plFn := luaFn[0], luaFn[1]
	pl := config.FindPlugin(plName)
	return func(h *BufPane) bool {
		val, err := pl.Call(plFn, luar.New(ulua.L, h))
		if err != nil {
			screen.TermMessage(err)
		}
		if v, ok := val.(lua.LBool); !ok {
			return false
		} else {
			return bool(v)
		}
	}
}

// BufMapKey maps a key event to an action
func BufMapKey(k Event, action string) {
	if strings.HasPrefix(action, "command:") {
		action = strings.SplitN(action, ":", 2)[1]
		BufKeyStrings[k] = action
		BufKeyBindings[k] = CommandAction(action)
	} else if strings.HasPrefix(action, "command-edit:") {
		action = strings.SplitN(action, ":", 2)[1]
		BufKeyStrings[k] = action
		BufKeyBindings[k] = CommandEditAction(action)
	} else if strings.HasPrefix(action, "lua:") {
		action = strings.SplitN(action, ":", 2)[1]
		BufKeyStrings[k] = action
		BufKeyBindings[k] = LuaAction(action)
	} else if f, ok := BufKeyActions[action]; ok {
		BufKeyStrings[k] = action
		BufKeyBindings[k] = f
	} else {
		screen.TermMessage("Error:", action, "does not exist")
	}
}

// BufMapMouse maps a mouse event to an action
func BufMapMouse(k MouseEvent, action string) {
	if f, ok := BufMouseActions[action]; ok {
		BufMouseBindings[k] = f
	} else {
		delete(BufMouseBindings, k)
		BufMapKey(k, action)
	}
}

// The BufPane connects the buffer and the window
// It provides a cursor (or multiple) and defines a set of actions
// that can be taken on the buffer
// The ActionHandler can access the window for necessary info about
// visual positions for mouse clicks and scrolling
type BufPane struct {
	display.BWindow

	Buf *buffer.Buffer

	Cursor *buffer.Cursor // the active cursor

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

	// remember original location of a search in case the search is canceled
	searchOrig buffer.Loc
}

func NewBufPane(buf *buffer.Buffer, win display.BWindow) *BufPane {
	h := new(BufPane)
	h.Buf = buf
	h.BWindow = win

	h.Cursor = h.Buf.GetActiveCursor()
	h.mouseReleased = true

	config.RunPluginFn("onBufPaneOpen", luar.New(ulua.L, h))

	return h
}

func NewBufPaneFromBuf(buf *buffer.Buffer) *BufPane {
	w := display.NewBufWindow(0, 0, 0, 0, buf)
	return NewBufPane(buf, w)
}

// PluginCB calls all plugin callbacks with a certain name and
// displays an error if there is one and returns the aggregrate
// boolean response
func (h *BufPane) PluginCB(cb string) bool {
	b, err := config.RunPluginFnBool(cb, luar.New(ulua.L, h))
	if err != nil {
		screen.TermMessage(err)
	}
	return b
}

// PluginCBRune is the same as PluginCB but also passes a rune to
// the plugins
func (h *BufPane) PluginCBRune(cb string, r rune) bool {
	b, err := config.RunPluginFnBool(cb, luar.New(ulua.L, h), luar.New(ulua.L, string(r)))
	if err != nil {
		screen.TermMessage(err)
	}
	return b
}

func (h *BufPane) OpenBuffer(b *buffer.Buffer) {
	h.Buf.Close()
	h.Buf = b
	h.BWindow.SetBuffer(b)
	h.Cursor = b.GetActiveCursor()
	h.Resize(h.GetView().Width, h.GetView().Height)
	v := new(display.View)
	h.SetView(v)
	h.Relocate()
	// Set mouseReleased to true because we assume the mouse is not being pressed when
	// the editor is opened
	h.mouseReleased = true
	// Set isOverwriteMode to false, because we assume we are in the default mode when editor
	// is opened
	h.isOverwriteMode = false
	h.lastClickTime = time.Time{}
}

func (h *BufPane) ID() uint64 {
	return h.splitID
}

func (h *BufPane) SetID(i uint64) {
	h.splitID = i
}

func (h *BufPane) Name() string {
	return h.Buf.GetName()
}

// HandleEvent executes the tcell event properly
// TODO: multiple actions bound to one key
func (h *BufPane) HandleEvent(event tcell.Event) {
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
func (h *BufPane) DoKeyEvent(e Event) bool {
	if action, ok := BufKeyBindings[e]; ok {
		estr := BufKeyStrings[e]
		if estr != "InsertTab" {
			h.Buf.HasSuggestions = false
		}
		for _, s := range MultiActions {
			if s == estr {
				cursors := h.Buf.GetCursors()
				for _, c := range cursors {
					h.Buf.SetCurCursor(c.Num)
					h.Cursor = c
					if !h.PluginCB("pre" + estr) {
						// canceled by plugin
						continue
					}
					rel := action(h)
					if h.PluginCB("on"+estr) && rel {
						h.Relocate()
					}
				}
				return true
			}
		}
		if !h.PluginCB("pre" + estr) {
			return false
		}
		rel := action(h)
		log.Println("calling on", estr)
		if h.PluginCB("on"+estr) && rel {
			h.Relocate()
		}
		return true
	}
	return false
}

func (h *BufPane) HasKeyEvent(e Event) bool {
	_, ok := BufKeyBindings[e]
	return ok
}

// DoMouseEvent executes a mouse event by finding the action it is bound
// to and executing it
func (h *BufPane) DoMouseEvent(e MouseEvent, te *tcell.EventMouse) bool {
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
func (h *BufPane) DoRuneInsert(r rune) {
	cursors := h.Buf.GetCursors()
	for _, c := range cursors {
		// Insert a character
		h.Buf.SetCurCursor(c.Num)
		if !h.PluginCBRune("preRune", r) {
			continue
		}
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
		h.PluginCBRune("onRune", r)
	}
}

func (h *BufPane) VSplitBuf(buf *buffer.Buffer) {
	e := NewBufPaneFromBuf(buf)
	e.splitID = MainTab().GetNode(h.splitID).VSplit(h.Buf.Settings["splitright"].(bool))
	MainTab().Panes = append(MainTab().Panes, e)
	MainTab().Resize()
	MainTab().SetActive(len(MainTab().Panes) - 1)
}
func (h *BufPane) HSplitBuf(buf *buffer.Buffer) {
	e := NewBufPaneFromBuf(buf)
	e.splitID = MainTab().GetNode(h.splitID).HSplit(h.Buf.Settings["splitbottom"].(bool))
	MainTab().Panes = append(MainTab().Panes, e)
	MainTab().Resize()
	MainTab().SetActive(len(MainTab().Panes) - 1)
}
func (h *BufPane) Close() {
	h.Buf.Close()
}

// BufKeyActions contains the list of all possible key actions the bufhandler could execute
var BufKeyActions = map[string]BufKeyAction{
	"CursorUp":               (*BufPane).CursorUp,
	"CursorDown":             (*BufPane).CursorDown,
	"CursorPageUp":           (*BufPane).CursorPageUp,
	"CursorPageDown":         (*BufPane).CursorPageDown,
	"CursorLeft":             (*BufPane).CursorLeft,
	"CursorRight":            (*BufPane).CursorRight,
	"CursorStart":            (*BufPane).CursorStart,
	"CursorEnd":              (*BufPane).CursorEnd,
	"SelectToStart":          (*BufPane).SelectToStart,
	"SelectToEnd":            (*BufPane).SelectToEnd,
	"SelectUp":               (*BufPane).SelectUp,
	"SelectDown":             (*BufPane).SelectDown,
	"SelectLeft":             (*BufPane).SelectLeft,
	"SelectRight":            (*BufPane).SelectRight,
	"WordRight":              (*BufPane).WordRight,
	"WordLeft":               (*BufPane).WordLeft,
	"SelectWordRight":        (*BufPane).SelectWordRight,
	"SelectWordLeft":         (*BufPane).SelectWordLeft,
	"DeleteWordRight":        (*BufPane).DeleteWordRight,
	"DeleteWordLeft":         (*BufPane).DeleteWordLeft,
	"SelectLine":             (*BufPane).SelectLine,
	"SelectToStartOfLine":    (*BufPane).SelectToStartOfLine,
	"SelectToEndOfLine":      (*BufPane).SelectToEndOfLine,
	"ParagraphPrevious":      (*BufPane).ParagraphPrevious,
	"ParagraphNext":          (*BufPane).ParagraphNext,
	"InsertNewline":          (*BufPane).InsertNewline,
	"Backspace":              (*BufPane).Backspace,
	"Delete":                 (*BufPane).Delete,
	"InsertTab":              (*BufPane).InsertTab,
	"Save":                   (*BufPane).Save,
	"SaveAll":                (*BufPane).SaveAll,
	"SaveAs":                 (*BufPane).SaveAs,
	"Find":                   (*BufPane).Find,
	"FindNext":               (*BufPane).FindNext,
	"FindPrevious":           (*BufPane).FindPrevious,
	"Center":                 (*BufPane).Center,
	"Undo":                   (*BufPane).Undo,
	"Redo":                   (*BufPane).Redo,
	"Copy":                   (*BufPane).Copy,
	"Cut":                    (*BufPane).Cut,
	"CutLine":                (*BufPane).CutLine,
	"DuplicateLine":          (*BufPane).DuplicateLine,
	"DeleteLine":             (*BufPane).DeleteLine,
	"MoveLinesUp":            (*BufPane).MoveLinesUp,
	"MoveLinesDown":          (*BufPane).MoveLinesDown,
	"IndentSelection":        (*BufPane).IndentSelection,
	"OutdentSelection":       (*BufPane).OutdentSelection,
	"OutdentLine":            (*BufPane).OutdentLine,
	"Paste":                  (*BufPane).Paste,
	"PastePrimary":           (*BufPane).PastePrimary,
	"SelectAll":              (*BufPane).SelectAll,
	"OpenFile":               (*BufPane).OpenFile,
	"Start":                  (*BufPane).Start,
	"End":                    (*BufPane).End,
	"PageUp":                 (*BufPane).PageUp,
	"PageDown":               (*BufPane).PageDown,
	"SelectPageUp":           (*BufPane).SelectPageUp,
	"SelectPageDown":         (*BufPane).SelectPageDown,
	"HalfPageUp":             (*BufPane).HalfPageUp,
	"HalfPageDown":           (*BufPane).HalfPageDown,
	"StartOfLine":            (*BufPane).StartOfLine,
	"EndOfLine":              (*BufPane).EndOfLine,
	"ToggleHelp":             (*BufPane).ToggleHelp,
	"ToggleKeyMenu":          (*BufPane).ToggleKeyMenu,
	"ToggleRuler":            (*BufPane).ToggleRuler,
	"ClearStatus":            (*BufPane).ClearStatus,
	"ShellMode":              (*BufPane).ShellMode,
	"CommandMode":            (*BufPane).CommandMode,
	"ToggleOverwriteMode":    (*BufPane).ToggleOverwriteMode,
	"Escape":                 (*BufPane).Escape,
	"Quit":                   (*BufPane).Quit,
	"QuitAll":                (*BufPane).QuitAll,
	"AddTab":                 (*BufPane).AddTab,
	"PreviousTab":            (*BufPane).PreviousTab,
	"NextTab":                (*BufPane).NextTab,
	"NextSplit":              (*BufPane).NextSplit,
	"PreviousSplit":          (*BufPane).PreviousSplit,
	"Unsplit":                (*BufPane).Unsplit,
	"VSplit":                 (*BufPane).VSplitAction,
	"HSplit":                 (*BufPane).HSplitAction,
	"ToggleMacro":            (*BufPane).ToggleMacro,
	"PlayMacro":              (*BufPane).PlayMacro,
	"Suspend":                (*BufPane).Suspend,
	"ScrollUp":               (*BufPane).ScrollUpAction,
	"ScrollDown":             (*BufPane).ScrollDownAction,
	"SpawnMultiCursor":       (*BufPane).SpawnMultiCursor,
	"SpawnMultiCursorSelect": (*BufPane).SpawnMultiCursorSelect,
	"RemoveMultiCursor":      (*BufPane).RemoveMultiCursor,
	"RemoveAllMultiCursors":  (*BufPane).RemoveAllMultiCursors,
	"SkipMultiCursor":        (*BufPane).SkipMultiCursor,
	"JumpToMatchingBrace":    (*BufPane).JumpToMatchingBrace,

	// This was changed to InsertNewline but I don't want to break backwards compatibility
	"InsertEnter": (*BufPane).InsertNewline,
}

// BufMouseActions contains the list of all possible mouse actions the bufhandler could execute
var BufMouseActions = map[string]BufMouseAction{
	"MousePress":       (*BufPane).MousePress,
	"MouseMultiCursor": (*BufPane).MouseMultiCursor,
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
