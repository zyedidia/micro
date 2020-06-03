package action

import (
	"strings"
	"time"

	luar "layeh.com/gopher-luar"

	lua "github.com/yuin/gopher-lua"
	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/display"
	ulua "github.com/zyedidia/micro/v2/internal/lua"
	"github.com/zyedidia/micro/v2/internal/screen"
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
	if len(luaFn) <= 1 {
		return nil
	}
	plName, plFn := luaFn[0], luaFn[1]
	pl := config.FindPlugin(plName)
	if pl == nil {
		return nil
	}
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
	BufKeyStrings[k] = action
	var actionfns []func(*BufPane) bool
	var names []string
	var types []byte
	for i := 0; ; i++ {
		if action == "" {
			break
		}

		// TODO: fix problem when complex bindings have these
		// characters (escape them?)
		idx := strings.IndexAny(action, "&|,")
		a := action
		if idx >= 0 {
			a = action[:idx]
			types = append(types, action[idx])
			action = action[idx+1:]
		} else {
			types = append(types, ' ')
			action = ""
		}

		var afn func(*BufPane) bool
		if strings.HasPrefix(a, "command:") {
			a = strings.SplitN(a, ":", 2)[1]
			afn = CommandAction(a)
			names = append(names, "")
		} else if strings.HasPrefix(a, "command-edit:") {
			a = strings.SplitN(a, ":", 2)[1]
			afn = CommandEditAction(a)
			names = append(names, "")
		} else if strings.HasPrefix(a, "lua:") {
			a = strings.SplitN(a, ":", 2)[1]
			afn = LuaAction(a)
			if afn == nil {
				screen.TermMessage("Lua Error:", a, "does not exist")
				continue
			}
			split := strings.SplitN(a, ".", 2)
			if len(split) > 1 {
				a = strings.Title(split[0]) + strings.Title(split[1])
			} else {
				a = strings.Title(a)
			}

			names = append(names, a)
		} else if f, ok := BufKeyActions[a]; ok {
			afn = f
			names = append(names, a)
		} else {
			screen.TermMessage("Error in bindings: action", a, "does not exist")
			continue
		}
		actionfns = append(actionfns, afn)
	}
	BufKeyBindings[k] = func(h *BufPane) bool {
		cursors := h.Buf.GetCursors()
		success := true
		for i, a := range actionfns {
			for j, c := range cursors {
				if c == nil {
					continue
				}
				h.Buf.SetCurCursor(c.Num)
				h.Cursor = c
				if i == 0 || (success && types[i-1] == '&') || (!success && types[i-1] == '|') || (types[i-1] == ',') {
					success = h.execAction(a, names[i], j)
				} else {
					break
				}
			}
			// if the action changed the current pane, update the reference
			h = MainTab().CurPane()
		}
		return true
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

// BufUnmap unmaps a key or mouse event from any action
func BufUnmap(k Event) {
	delete(BufKeyBindings, k)
	delete(BufKeyStrings, k)

	switch e := k.(type) {
	case MouseEvent:
		delete(BufMouseBindings, e)
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
	lastSearch      string
	lastSearchRegex bool
	// Should the current multiple cursor selection search based on word or
	// based on selection (false for selection, true for word)
	multiWord bool

	splitID uint64
	tab     *Tab

	// remember original location of a search in case the search is canceled
	searchOrig buffer.Loc
}

func NewBufPane(buf *buffer.Buffer, win display.BWindow, tab *Tab) *BufPane {
	h := new(BufPane)
	h.Buf = buf
	h.BWindow = win
	h.tab = tab

	h.Cursor = h.Buf.GetActiveCursor()
	h.mouseReleased = true

	config.RunPluginFn("onBufPaneOpen", luar.New(ulua.L, h))

	return h
}

func NewBufPaneFromBuf(buf *buffer.Buffer, tab *Tab) *BufPane {
	w := display.NewBufWindow(0, 0, 0, 0, buf)
	return NewBufPane(buf, w, tab)
}

func (h *BufPane) SetTab(t *Tab) {
	h.tab = t
}

func (h *BufPane) Tab() *Tab {
	return h.tab
}

func (h *BufPane) ResizePane(size int) {
	n := h.tab.GetNode(h.splitID)
	n.ResizeSplit(size)
	h.tab.Resize()
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
	n := h.Buf.GetName()
	if h.Buf.Modified() {
		n += " +"
	}
	return n
}

// HandleEvent executes the tcell event properly
func (h *BufPane) HandleEvent(event tcell.Event) {
	if h.Buf.ExternallyModified() && !h.Buf.ReloadDisabled {
		InfoBar.YNPrompt("The file on disk has changed. Reload file? (y,n,esc)", func(yes, canceled bool) {
			if canceled {
				h.Buf.DisableReload()
			}
			if !yes || canceled {
				h.Buf.UpdateModTime()
			} else {
				h.Buf.ReOpen()
			}
		})

	}

	switch e := event.(type) {
	case *tcell.EventRaw:
		re := RawEvent{
			esc: e.EscSeq(),
		}
		h.DoKeyEvent(re)
	case *tcell.EventPaste:
		h.paste(e.Text())
		h.Relocate()
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
		cancel := false
		switch e.Buttons() {
		case tcell.Button1:
			_, my := e.Position()
			if h.Buf.Settings["statusline"].(bool) && my >= h.GetView().Y+h.GetView().Height-1 {
				cancel = true
			}
		case tcell.ButtonNone:
			// Mouse event with no click
			if !h.mouseReleased {
				// Mouse was just released

				// mx, my := e.Position()
				// mouseLoc := h.LocFromVisual(buffer.Loc{X: mx, Y: my})

				// we could finish the selection based on the release location as described
				// below but when the mouse click is within the scroll margin this will
				// cause a scroll and selection even for a simple mouse click which is
				// not good
				// for terminals that don't support mouse motion events, selection via
				// the mouse won't work but this is ok

				// Relocating here isn't really necessary because the cursor will
				// be in the right place from the last mouse event
				// However, if we are running in a terminal that doesn't support mouse motion
				// events, this still allows the user to make selections, except only after they
				// release the mouse

				// if !h.doubleClick && !h.tripleClick {
				// 	h.Cursor.SetSelectionEnd(h.Cursor.Loc)
				// }
				if h.Cursor.HasSelection() {
					h.Cursor.CopySelection("primary")
				}
				h.mouseReleased = true
			}
		}

		if !cancel {
			me := MouseEvent{
				btn: e.Buttons(),
				mod: e.Modifiers(),
			}
			h.DoMouseEvent(me, e)
		}
	}
	h.Buf.MergeCursors()

	if h.IsActive() {
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
}

// DoKeyEvent executes a key event by finding the action it is bound
// to and executing it (possibly multiple times for multiple cursors)
func (h *BufPane) DoKeyEvent(e Event) bool {
	if action, ok := BufKeyBindings[e]; ok {
		return action(h)
	}
	return false
}

func (h *BufPane) execAction(action func(*BufPane) bool, name string, cursor int) bool {
	if name != "Autocomplete" && name != "CycleAutocompleteBack" {
		h.Buf.HasSuggestions = false
	}

	_, isMulti := MultiActions[name]
	if (!isMulti && cursor == 0) || isMulti {
		if h.PluginCB("pre" + name) {
			success := action(h)
			success = success && h.PluginCB("on"+name)

			if isMulti {
				if recording_macro {
					if name != "ToggleMacro" && name != "PlayMacro" {
						curmacro = append(curmacro, action)
					}
				}
			}

			return success
		}
	}

	return false
}

func (h *BufPane) completeAction(action string) {
	h.PluginCB("on" + action)
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
		h.Cursor = c
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
		if recording_macro {
			curmacro = append(curmacro, r)
		}
		h.Relocate()
		h.PluginCBRune("onRune", r)
	}
}

func (h *BufPane) VSplitIndex(buf *buffer.Buffer, right bool) *BufPane {
	e := NewBufPaneFromBuf(buf, h.tab)
	e.splitID = MainTab().GetNode(h.splitID).VSplit(right)
	MainTab().Panes = append(MainTab().Panes, e)
	MainTab().Resize()
	MainTab().SetActive(len(MainTab().Panes) - 1)
	return e
}
func (h *BufPane) HSplitIndex(buf *buffer.Buffer, bottom bool) *BufPane {
	e := NewBufPaneFromBuf(buf, h.tab)
	e.splitID = MainTab().GetNode(h.splitID).HSplit(bottom)
	MainTab().Panes = append(MainTab().Panes, e)
	MainTab().Resize()
	MainTab().SetActive(len(MainTab().Panes) - 1)
	return e
}

func (h *BufPane) VSplitBuf(buf *buffer.Buffer) *BufPane {
	return h.VSplitIndex(buf, h.Buf.Settings["splitright"].(bool))
}
func (h *BufPane) HSplitBuf(buf *buffer.Buffer) *BufPane {
	return h.HSplitIndex(buf, h.Buf.Settings["splitbottom"].(bool))
}
func (h *BufPane) Close() {
	h.Buf.Close()
}

func (h *BufPane) SetActive(b bool) {
	h.BWindow.SetActive(b)
	if b {
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

}

// BufKeyActions contains the list of all possible key actions the bufhandler could execute
var BufKeyActions = map[string]BufKeyAction{
	"CursorUp":                  (*BufPane).CursorUp,
	"CursorDown":                (*BufPane).CursorDown,
	"CursorPageUp":              (*BufPane).CursorPageUp,
	"CursorPageDown":            (*BufPane).CursorPageDown,
	"CursorLeft":                (*BufPane).CursorLeft,
	"CursorRight":               (*BufPane).CursorRight,
	"CursorStart":               (*BufPane).CursorStart,
	"CursorEnd":                 (*BufPane).CursorEnd,
	"SelectToStart":             (*BufPane).SelectToStart,
	"SelectToEnd":               (*BufPane).SelectToEnd,
	"SelectUp":                  (*BufPane).SelectUp,
	"SelectDown":                (*BufPane).SelectDown,
	"SelectLeft":                (*BufPane).SelectLeft,
	"SelectRight":               (*BufPane).SelectRight,
	"WordRight":                 (*BufPane).WordRight,
	"WordLeft":                  (*BufPane).WordLeft,
	"SelectWordRight":           (*BufPane).SelectWordRight,
	"SelectWordLeft":            (*BufPane).SelectWordLeft,
	"DeleteWordRight":           (*BufPane).DeleteWordRight,
	"DeleteWordLeft":            (*BufPane).DeleteWordLeft,
	"SelectLine":                (*BufPane).SelectLine,
	"SelectToStartOfLine":       (*BufPane).SelectToStartOfLine,
	"SelectToStartOfText":       (*BufPane).SelectToStartOfText,
	"SelectToStartOfTextToggle": (*BufPane).SelectToStartOfTextToggle,
	"SelectToEndOfLine":         (*BufPane).SelectToEndOfLine,
	"ParagraphPrevious":         (*BufPane).ParagraphPrevious,
	"ParagraphNext":             (*BufPane).ParagraphNext,
	"InsertNewline":             (*BufPane).InsertNewline,
	"Backspace":                 (*BufPane).Backspace,
	"Delete":                    (*BufPane).Delete,
	"InsertTab":                 (*BufPane).InsertTab,
	"Save":                      (*BufPane).Save,
	"SaveAll":                   (*BufPane).SaveAll,
	"SaveAs":                    (*BufPane).SaveAs,
	"Find":                      (*BufPane).Find,
	"FindLiteral":               (*BufPane).FindLiteral,
	"FindNext":                  (*BufPane).FindNext,
	"FindPrevious":              (*BufPane).FindPrevious,
	"Center":                    (*BufPane).Center,
	"Undo":                      (*BufPane).Undo,
	"Redo":                      (*BufPane).Redo,
	"Copy":                      (*BufPane).Copy,
	"CopyLine":                  (*BufPane).CopyLine,
	"Cut":                       (*BufPane).Cut,
	"CutLine":                   (*BufPane).CutLine,
	"DuplicateLine":             (*BufPane).DuplicateLine,
	"DeleteLine":                (*BufPane).DeleteLine,
	"MoveLinesUp":               (*BufPane).MoveLinesUp,
	"MoveLinesDown":             (*BufPane).MoveLinesDown,
	"IndentSelection":           (*BufPane).IndentSelection,
	"OutdentSelection":          (*BufPane).OutdentSelection,
	"Autocomplete":              (*BufPane).Autocomplete,
	"CycleAutocompleteBack":     (*BufPane).CycleAutocompleteBack,
	"OutdentLine":               (*BufPane).OutdentLine,
	"IndentLine":                (*BufPane).IndentLine,
	"Paste":                     (*BufPane).Paste,
	"PastePrimary":              (*BufPane).PastePrimary,
	"SelectAll":                 (*BufPane).SelectAll,
	"OpenFile":                  (*BufPane).OpenFile,
	"Start":                     (*BufPane).Start,
	"End":                       (*BufPane).End,
	"PageUp":                    (*BufPane).PageUp,
	"PageDown":                  (*BufPane).PageDown,
	"SelectPageUp":              (*BufPane).SelectPageUp,
	"SelectPageDown":            (*BufPane).SelectPageDown,
	"HalfPageUp":                (*BufPane).HalfPageUp,
	"HalfPageDown":              (*BufPane).HalfPageDown,
	"StartOfText":               (*BufPane).StartOfText,
	"StartOfTextToggle":         (*BufPane).StartOfTextToggle,
	"StartOfLine":               (*BufPane).StartOfLine,
	"EndOfLine":                 (*BufPane).EndOfLine,
	"ToggleHelp":                (*BufPane).ToggleHelp,
	"ToggleKeyMenu":             (*BufPane).ToggleKeyMenu,
	"ToggleDiffGutter":          (*BufPane).ToggleDiffGutter,
	"ToggleRuler":               (*BufPane).ToggleRuler,
	"ClearStatus":               (*BufPane).ClearStatus,
	"ShellMode":                 (*BufPane).ShellMode,
	"CommandMode":               (*BufPane).CommandMode,
	"ToggleOverwriteMode":       (*BufPane).ToggleOverwriteMode,
	"Escape":                    (*BufPane).Escape,
	"Quit":                      (*BufPane).Quit,
	"QuitAll":                   (*BufPane).QuitAll,
	"AddTab":                    (*BufPane).AddTab,
	"PreviousTab":               (*BufPane).PreviousTab,
	"NextTab":                   (*BufPane).NextTab,
	"NextSplit":                 (*BufPane).NextSplit,
	"PreviousSplit":             (*BufPane).PreviousSplit,
	"Unsplit":                   (*BufPane).Unsplit,
	"VSplit":                    (*BufPane).VSplitAction,
	"HSplit":                    (*BufPane).HSplitAction,
	"ToggleMacro":               (*BufPane).ToggleMacro,
	"PlayMacro":                 (*BufPane).PlayMacro,
	"Suspend":                   (*BufPane).Suspend,
	"ScrollUp":                  (*BufPane).ScrollUpAction,
	"ScrollDown":                (*BufPane).ScrollDownAction,
	"SpawnMultiCursor":          (*BufPane).SpawnMultiCursor,
	"SpawnMultiCursorUp":        (*BufPane).SpawnMultiCursorUp,
	"SpawnMultiCursorDown":      (*BufPane).SpawnMultiCursorDown,
	"SpawnMultiCursorSelect":    (*BufPane).SpawnMultiCursorSelect,
	"RemoveMultiCursor":         (*BufPane).RemoveMultiCursor,
	"RemoveAllMultiCursors":     (*BufPane).RemoveAllMultiCursors,
	"SkipMultiCursor":           (*BufPane).SkipMultiCursor,
	"JumpToMatchingBrace":       (*BufPane).JumpToMatchingBrace,
	"JumpLine":                  (*BufPane).JumpLine,
	"None":                      (*BufPane).None,

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
var MultiActions = map[string]bool{
	"CursorUp":                  true,
	"CursorDown":                true,
	"CursorPageUp":              true,
	"CursorPageDown":            true,
	"CursorLeft":                true,
	"CursorRight":               true,
	"CursorStart":               true,
	"CursorEnd":                 true,
	"SelectToStart":             true,
	"SelectToEnd":               true,
	"SelectUp":                  true,
	"SelectDown":                true,
	"SelectLeft":                true,
	"SelectRight":               true,
	"WordRight":                 true,
	"WordLeft":                  true,
	"SelectWordRight":           true,
	"SelectWordLeft":            true,
	"DeleteWordRight":           true,
	"DeleteWordLeft":            true,
	"SelectLine":                true,
	"SelectToStartOfLine":       true,
	"SelectToStartOfText":       true,
	"SelectToStartOfTextToggle": true,
	"SelectToEndOfLine":         true,
	"ParagraphPrevious":         true,
	"ParagraphNext":             true,
	"InsertNewline":             true,
	"Backspace":                 true,
	"Delete":                    true,
	"InsertTab":                 true,
	"FindNext":                  true,
	"FindPrevious":              true,
	"CopyLine":                  true,
	"Cut":                       true,
	"CutLine":                   true,
	"DuplicateLine":             true,
	"DeleteLine":                true,
	"MoveLinesUp":               true,
	"MoveLinesDown":             true,
	"IndentSelection":           true,
	"OutdentSelection":          true,
	"OutdentLine":               true,
	"IndentLine":                true,
	"Paste":                     true,
	"PastePrimary":              true,
	"SelectPageUp":              true,
	"SelectPageDown":            true,
	"StartOfLine":               true,
	"StartOfText":               true,
	"StartOfTextToggle":         true,
	"EndOfLine":                 true,
	"JumpToMatchingBrace":       true,
}
