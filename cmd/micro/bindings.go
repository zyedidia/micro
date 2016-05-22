package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/zyedidia/clipboard"
	"github.com/zyedidia/tcell"
)

var bindings map[Key]func(*View) bool

// The Key struct holds the data for a keypress (keycode + modifiers)
type Key struct {
	keyCode   tcell.Key
	modifiers tcell.ModMask
	r         rune
}

// InitBindings initializes the keybindings for micro
func InitBindings() {
	bindings = make(map[Key]func(*View) bool)

	actions := map[string]func(*View) bool{
		"CursorUp":            (*View).CursorUp,
		"CursorDown":          (*View).CursorDown,
		"CursorLeft":          (*View).CursorLeft,
		"CursorRight":         (*View).CursorRight,
		"CursorStart":         (*View).CursorStart,
		"CursorEnd":           (*View).CursorEnd,
		"SelectToStart":       (*View).SelectToStart,
		"SelectToEnd":         (*View).SelectToEnd,
		"SelectUp":            (*View).SelectUp,
		"SelectDown":          (*View).SelectDown,
		"SelectLeft":          (*View).SelectLeft,
		"SelectRight":         (*View).SelectRight,
		"WordRight":           (*View).WordRight,
		"WordLeft":            (*View).WordLeft,
		"SelectWordRight":     (*View).SelectWordRight,
		"SelectWordLeft":      (*View).SelectWordLeft,
		"SelectToStartOfLine": (*View).SelectToStartOfLine,
		"SelectToEndOfLine":   (*View).SelectToEndOfLine,
		"InsertEnter":         (*View).InsertEnter,
		"InsertSpace":         (*View).InsertSpace,
		"Backspace":           (*View).Backspace,
		"Delete":              (*View).Delete,
		"InsertTab":           (*View).InsertTab,
		"Save":                (*View).Save,
		"Find":                (*View).Find,
		"FindNext":            (*View).FindNext,
		"FindPrevious":        (*View).FindPrevious,
		"Undo":                (*View).Undo,
		"Redo":                (*View).Redo,
		"Copy":                (*View).Copy,
		"Cut":                 (*View).Cut,
		"CutLine":             (*View).CutLine,
		"DuplicateLine":       (*View).DuplicateLine,
		"Paste":               (*View).Paste,
		"SelectAll":           (*View).SelectAll,
		"OpenFile":            (*View).OpenFile,
		"Start":               (*View).Start,
		"End":                 (*View).End,
		"PageUp":              (*View).PageUp,
		"PageDown":            (*View).PageDown,
		"HalfPageUp":          (*View).HalfPageUp,
		"HalfPageDown":        (*View).HalfPageDown,
		"StartOfLine":         (*View).StartOfLine,
		"EndOfLine":           (*View).EndOfLine,
		"ToggleRuler":         (*View).ToggleRuler,
		"JumpLine":            (*View).JumpLine,
		"ClearStatus":         (*View).ClearStatus,
	}

	keys := map[string]Key{
		"Up":             Key{tcell.KeyUp, tcell.ModNone, 0},
		"Down":           Key{tcell.KeyDown, tcell.ModNone, 0},
		"Right":          Key{tcell.KeyRight, tcell.ModNone, 0},
		"Left":           Key{tcell.KeyLeft, tcell.ModNone, 0},
		"AltUp":          Key{tcell.KeyUp, tcell.ModAlt, 0},
		"AltDown":        Key{tcell.KeyDown, tcell.ModAlt, 0},
		"AltLeft":        Key{tcell.KeyLeft, tcell.ModAlt, 0},
		"AltRight":       Key{tcell.KeyRight, tcell.ModAlt, 0},
		"CtrlUp":         Key{tcell.KeyUp, tcell.ModCtrl, 0},
		"CtrlDown":       Key{tcell.KeyDown, tcell.ModCtrl, 0},
		"CtrlLeft":       Key{tcell.KeyLeft, tcell.ModCtrl, 0},
		"CtrlRight":      Key{tcell.KeyRight, tcell.ModCtrl, 0},
		"ShiftUp":        Key{tcell.KeyUp, tcell.ModShift, 0},
		"ShiftDown":      Key{tcell.KeyDown, tcell.ModShift, 0},
		"ShiftLeft":      Key{tcell.KeyLeft, tcell.ModShift, 0},
		"ShiftRight":     Key{tcell.KeyRight, tcell.ModShift, 0},
		"AltShiftUp":     Key{tcell.KeyUp, tcell.ModShift | tcell.ModAlt, 0},
		"AltShiftDown":   Key{tcell.KeyDown, tcell.ModShift | tcell.ModAlt, 0},
		"AltShiftLeft":   Key{tcell.KeyLeft, tcell.ModShift | tcell.ModAlt, 0},
		"AltShiftRight":  Key{tcell.KeyRight, tcell.ModShift | tcell.ModAlt, 0},
		"CtrlShiftUp":    Key{tcell.KeyUp, tcell.ModShift | tcell.ModCtrl, 0},
		"CtrlShiftDown":  Key{tcell.KeyDown, tcell.ModShift | tcell.ModCtrl, 0},
		"CtrlShiftLeft":  Key{tcell.KeyLeft, tcell.ModShift | tcell.ModCtrl, 0},
		"CtrlShiftRight": Key{tcell.KeyRight, tcell.ModShift | tcell.ModCtrl, 0},
		"UpLeft":         Key{tcell.KeyUpLeft, tcell.ModNone, 0},
		"UpRight":        Key{tcell.KeyUpRight, tcell.ModNone, 0},
		"DownLeft":       Key{tcell.KeyDownLeft, tcell.ModNone, 0},
		"DownRight":      Key{tcell.KeyDownRight, tcell.ModNone, 0},
		"Center":         Key{tcell.KeyCenter, tcell.ModNone, 0},
		"PgUp":           Key{tcell.KeyPgUp, tcell.ModNone, 0},
		"PgDn":           Key{tcell.KeyPgDn, tcell.ModNone, 0},
		"Home":           Key{tcell.KeyHome, tcell.ModNone, 0},
		"End":            Key{tcell.KeyEnd, tcell.ModNone, 0},
		"Insert":         Key{tcell.KeyInsert, tcell.ModNone, 0},
		"Delete":         Key{tcell.KeyDelete, tcell.ModNone, 0},
		"Help":           Key{tcell.KeyHelp, tcell.ModNone, 0},
		"Exit":           Key{tcell.KeyExit, tcell.ModNone, 0},
		"Clear":          Key{tcell.KeyClear, tcell.ModNone, 0},
		"Cancel":         Key{tcell.KeyCancel, tcell.ModNone, 0},
		"Print":          Key{tcell.KeyPrint, tcell.ModNone, 0},
		"Pause":          Key{tcell.KeyPause, tcell.ModNone, 0},
		"Backtab":        Key{tcell.KeyBacktab, tcell.ModNone, 0},
		"F1":             Key{tcell.KeyF1, tcell.ModNone, 0},
		"F2":             Key{tcell.KeyF2, tcell.ModNone, 0},
		"F3":             Key{tcell.KeyF3, tcell.ModNone, 0},
		"F4":             Key{tcell.KeyF4, tcell.ModNone, 0},
		"F5":             Key{tcell.KeyF5, tcell.ModNone, 0},
		"F6":             Key{tcell.KeyF6, tcell.ModNone, 0},
		"F7":             Key{tcell.KeyF7, tcell.ModNone, 0},
		"F8":             Key{tcell.KeyF8, tcell.ModNone, 0},
		"F9":             Key{tcell.KeyF9, tcell.ModNone, 0},
		"F10":            Key{tcell.KeyF10, tcell.ModNone, 0},
		"F11":            Key{tcell.KeyF11, tcell.ModNone, 0},
		"F12":            Key{tcell.KeyF12, tcell.ModNone, 0},
		"F13":            Key{tcell.KeyF13, tcell.ModNone, 0},
		"F14":            Key{tcell.KeyF14, tcell.ModNone, 0},
		"F15":            Key{tcell.KeyF15, tcell.ModNone, 0},
		"F16":            Key{tcell.KeyF16, tcell.ModNone, 0},
		"F17":            Key{tcell.KeyF17, tcell.ModNone, 0},
		"F18":            Key{tcell.KeyF18, tcell.ModNone, 0},
		"F19":            Key{tcell.KeyF19, tcell.ModNone, 0},
		"F20":            Key{tcell.KeyF20, tcell.ModNone, 0},
		"F21":            Key{tcell.KeyF21, tcell.ModNone, 0},
		"F22":            Key{tcell.KeyF22, tcell.ModNone, 0},
		"F23":            Key{tcell.KeyF23, tcell.ModNone, 0},
		"F24":            Key{tcell.KeyF24, tcell.ModNone, 0},
		"F25":            Key{tcell.KeyF25, tcell.ModNone, 0},
		"F26":            Key{tcell.KeyF26, tcell.ModNone, 0},
		"F27":            Key{tcell.KeyF27, tcell.ModNone, 0},
		"F28":            Key{tcell.KeyF28, tcell.ModNone, 0},
		"F29":            Key{tcell.KeyF29, tcell.ModNone, 0},
		"F30":            Key{tcell.KeyF30, tcell.ModNone, 0},
		"F31":            Key{tcell.KeyF31, tcell.ModNone, 0},
		"F32":            Key{tcell.KeyF32, tcell.ModNone, 0},
		"F33":            Key{tcell.KeyF33, tcell.ModNone, 0},
		"F34":            Key{tcell.KeyF34, tcell.ModNone, 0},
		"F35":            Key{tcell.KeyF35, tcell.ModNone, 0},
		"F36":            Key{tcell.KeyF36, tcell.ModNone, 0},
		"F37":            Key{tcell.KeyF37, tcell.ModNone, 0},
		"F38":            Key{tcell.KeyF38, tcell.ModNone, 0},
		"F39":            Key{tcell.KeyF39, tcell.ModNone, 0},
		"F40":            Key{tcell.KeyF40, tcell.ModNone, 0},
		"F41":            Key{tcell.KeyF41, tcell.ModNone, 0},
		"F42":            Key{tcell.KeyF42, tcell.ModNone, 0},
		"F43":            Key{tcell.KeyF43, tcell.ModNone, 0},
		"F44":            Key{tcell.KeyF44, tcell.ModNone, 0},
		"F45":            Key{tcell.KeyF45, tcell.ModNone, 0},
		"F46":            Key{tcell.KeyF46, tcell.ModNone, 0},
		"F47":            Key{tcell.KeyF47, tcell.ModNone, 0},
		"F48":            Key{tcell.KeyF48, tcell.ModNone, 0},
		"F49":            Key{tcell.KeyF49, tcell.ModNone, 0},
		"F50":            Key{tcell.KeyF50, tcell.ModNone, 0},
		"F51":            Key{tcell.KeyF51, tcell.ModNone, 0},
		"F52":            Key{tcell.KeyF52, tcell.ModNone, 0},
		"F53":            Key{tcell.KeyF53, tcell.ModNone, 0},
		"F54":            Key{tcell.KeyF54, tcell.ModNone, 0},
		"F55":            Key{tcell.KeyF55, tcell.ModNone, 0},
		"F56":            Key{tcell.KeyF56, tcell.ModNone, 0},
		"F57":            Key{tcell.KeyF57, tcell.ModNone, 0},
		"F58":            Key{tcell.KeyF58, tcell.ModNone, 0},
		"F59":            Key{tcell.KeyF59, tcell.ModNone, 0},
		"F60":            Key{tcell.KeyF60, tcell.ModNone, 0},
		"F61":            Key{tcell.KeyF61, tcell.ModNone, 0},
		"F62":            Key{tcell.KeyF62, tcell.ModNone, 0},
		"F63":            Key{tcell.KeyF63, tcell.ModNone, 0},
		"F64":            Key{tcell.KeyF64, tcell.ModNone, 0},
		"CtrlSpace":      Key{tcell.KeyCtrlSpace, tcell.ModCtrl, 0},
		"CtrlA":          Key{tcell.KeyCtrlA, tcell.ModCtrl, 0},
		"CtrlB":          Key{tcell.KeyCtrlB, tcell.ModCtrl, 0},
		"CtrlC":          Key{tcell.KeyCtrlC, tcell.ModCtrl, 0},
		"CtrlD":          Key{tcell.KeyCtrlD, tcell.ModCtrl, 0},
		"CtrlE":          Key{tcell.KeyCtrlE, tcell.ModCtrl, 0},
		"CtrlF":          Key{tcell.KeyCtrlF, tcell.ModCtrl, 0},
		"CtrlG":          Key{tcell.KeyCtrlG, tcell.ModCtrl, 0},
		"CtrlH":          Key{tcell.KeyCtrlH, tcell.ModCtrl, 0},
		"CtrlI":          Key{tcell.KeyCtrlI, tcell.ModCtrl, 0},
		"CtrlJ":          Key{tcell.KeyCtrlJ, tcell.ModCtrl, 0},
		"CtrlK":          Key{tcell.KeyCtrlK, tcell.ModCtrl, 0},
		"CtrlL":          Key{tcell.KeyCtrlL, tcell.ModCtrl, 0},
		"CtrlM":          Key{tcell.KeyCtrlM, tcell.ModCtrl, 0},
		"CtrlN":          Key{tcell.KeyCtrlN, tcell.ModCtrl, 0},
		"CtrlO":          Key{tcell.KeyCtrlO, tcell.ModCtrl, 0},
		"CtrlP":          Key{tcell.KeyCtrlP, tcell.ModCtrl, 0},
		"CtrlQ":          Key{tcell.KeyCtrlQ, tcell.ModCtrl, 0},
		"CtrlR":          Key{tcell.KeyCtrlR, tcell.ModCtrl, 0},
		"CtrlS":          Key{tcell.KeyCtrlS, tcell.ModCtrl, 0},
		"CtrlT":          Key{tcell.KeyCtrlT, tcell.ModCtrl, 0},
		"CtrlU":          Key{tcell.KeyCtrlU, tcell.ModCtrl, 0},
		"CtrlV":          Key{tcell.KeyCtrlV, tcell.ModCtrl, 0},
		"CtrlW":          Key{tcell.KeyCtrlW, tcell.ModCtrl, 0},
		"CtrlX":          Key{tcell.KeyCtrlX, tcell.ModCtrl, 0},
		"CtrlY":          Key{tcell.KeyCtrlY, tcell.ModCtrl, 0},
		"CtrlZ":          Key{tcell.KeyCtrlZ, tcell.ModCtrl, 0},
		"CtrlLeftSq":     Key{tcell.KeyCtrlLeftSq, tcell.ModCtrl, 0},
		"CtrlBackslash":  Key{tcell.KeyCtrlBackslash, tcell.ModCtrl, 0},
		"CtrlRightSq":    Key{tcell.KeyCtrlRightSq, tcell.ModCtrl, 0},
		"CtrlCarat":      Key{tcell.KeyCtrlCarat, tcell.ModCtrl, 0},
		"CtrlUnderscore": Key{tcell.KeyCtrlUnderscore, tcell.ModCtrl, 0},
		"Backspace":      Key{tcell.KeyBackspace, tcell.ModNone, 0},
		"Tab":            Key{tcell.KeyTab, tcell.ModNone, 0},
		"Esc":            Key{tcell.KeyEsc, tcell.ModNone, 0},
		"Escape":         Key{tcell.KeyEscape, tcell.ModNone, 0},
		"Enter":          Key{tcell.KeyEnter, tcell.ModNone, 0},
		"Backspace2":     Key{tcell.KeyBackspace2, tcell.ModNone, 0},
	}

	var parsed map[string]string
	defaults := DefaultBindings()

	filename := configDir + "/bindings.json"
	if _, e := os.Stat(filename); e == nil {
		input, err := ioutil.ReadFile(filename)
		if err != nil {
			TermMessage("Error reading bindings.json file: " + err.Error())
			return
		}

		err = json.Unmarshal(input, &parsed)
		if err != nil {
			TermMessage("Error reading bindings.json:", err.Error())
		}
	}

	for k, v := range defaults {
		if strings.Contains(k, "Alt-") {
			key := Key{tcell.KeyRune, tcell.ModAlt, rune(k[len(k)-1])}
			bindings[key] = actions[v]
		} else {
			bindings[keys[k]] = actions[v]
		}
	}
	for k, v := range parsed {
		if strings.Contains(k, "Alt-") {
			key := Key{tcell.KeyRune, tcell.ModAlt, rune(k[len(k)])}
			bindings[key] = actions[v]
		} else {
			bindings[keys[k]] = actions[v]
		}
	}
}

// DefaultBindings returns a map containing micro's default keybindings
func DefaultBindings() map[string]string {
	return map[string]string{
		"Up":             "CursorUp",
		"Down":           "CursorDown",
		"Right":          "CursorRight",
		"Left":           "CursorLeft",
		"ShiftUp":        "SelectUp",
		"ShiftDown":      "SelectDown",
		"ShiftLeft":      "SelectLeft",
		"ShiftRight":     "SelectRight",
		"AltLeft":        "WordLeft",
		"AltRight":       "WordRight",
		"AltShiftRight":  "SelectWordRight",
		"AltShiftLeft":   "SelectWordLeft",
		"CtrlLeft":       "StartOfLine",
		"CtrlRight":      "EndOfLine",
		"CtrlShiftLeft":  "SelectToStartOfLine",
		"CtrlShiftRight": "SelectToEndOfLine",
		"CtrlUp":         "CursorStart",
		"CtrlDown":       "CursorEnd",
		"CtrlShiftUp":    "SelectToStart",
		"CtrlShiftDown":  "SelectToEnd",
		"Enter":          "InsertEnter",
		"Space":          "InsertSpace",
		"Backspace":      "Backspace",
		"Backspace2":     "Backspace",
		"Tab":            "InsertTab",
		"CtrlO":          "OpenFile",
		"CtrlS":          "Save",
		"CtrlF":          "Find",
		"CtrlN":          "FindNext",
		"CtrlP":          "FindPrevious",
		"CtrlZ":          "Undo",
		"CtrlY":          "Redo",
		"CtrlC":          "Copy",
		"CtrlX":          "Cut",
		"CtrlK":          "CutLine",
		"CtrlD":          "DuplicateLine",
		"CtrlV":          "Paste",
		"CtrlA":          "SelectAll",
		"Home":           "Start",
		"End":            "End",
		"PgUp":           "PageUp",
		"PgDn":           "PageDown",
		// Find alternative key
		// "CtrlU":          "HalfPageUp",
		// "CtrlD":          "HalfPageDown",
		"CtrlR":  "ToggleRuler",
		"CtrlL":  "JumpLine",
		"Delete": "Delete",
		"Esc":    "ClearStatus",

		// Emacs-style keybindings
		"Alt-f": "WordRight",
		"Alt-b": "WordLeft",
		"Alt-a": "StartOfLine",
		"Alt-e": "EndOfLine",
		"Alt-p": "CursorUp",
		"Alt-n": "CursorDown",
	}
}

// CursorUp moves the cursor up
func (v *View) CursorUp() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.SetLoc(v.Cursor.curSelection[0])
		v.Cursor.ResetSelection()
	}
	v.Cursor.Up()
	return true
}

// CursorDown moves the cursor down
func (v *View) CursorDown() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.SetLoc(v.Cursor.curSelection[1])
		v.Cursor.ResetSelection()
	}
	v.Cursor.Down()
	return true
}

// CursorLeft moves the cursor left
func (v *View) CursorLeft() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.SetLoc(v.Cursor.curSelection[0])
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Left()
	}
	return true
}

// CursorRight moves the cursor right
func (v *View) CursorRight() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.SetLoc(v.Cursor.curSelection[1] - 1)
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Right()
	}
	return true
}

// WordRight moves the cursor one word to the right
func (v *View) WordRight() bool {
	v.Cursor.WordRight()
	return true
}

// WordLeft moves the cursor one word to the left
func (v *View) WordLeft() bool {
	v.Cursor.WordLeft()
	return true
}

// SelectUp selects up one line
func (v *View) SelectUp() bool {
	loc := v.Cursor.Loc()
	if !v.Cursor.HasSelection() {
		v.Cursor.origSelection[0] = loc
	}
	v.Cursor.Up()
	v.Cursor.SelectTo(v.Cursor.Loc())
	return true
}

// SelectDown selects down one line
func (v *View) SelectDown() bool {
	loc := v.Cursor.Loc()
	if !v.Cursor.HasSelection() {
		v.Cursor.origSelection[0] = loc
	}
	v.Cursor.Down()
	v.Cursor.SelectTo(v.Cursor.Loc())
	return true
}

// SelectLeft selects the character to the left of the cursor
func (v *View) SelectLeft() bool {
	loc := v.Cursor.Loc()
	count := v.Buf.Len() - 1
	if loc > count {
		loc = count
	}
	if !v.Cursor.HasSelection() {
		v.Cursor.origSelection[0] = loc
	}
	v.Cursor.Left()
	v.Cursor.SelectTo(v.Cursor.Loc())
	return true
}

// SelectRight selects the character to the right of the cursor
func (v *View) SelectRight() bool {
	loc := v.Cursor.Loc()
	count := v.Buf.Len() - 1
	if loc > count {
		loc = count
	}
	if !v.Cursor.HasSelection() {
		v.Cursor.origSelection[0] = loc
	}
	v.Cursor.Right()
	v.Cursor.SelectTo(v.Cursor.Loc())
	return true
}

// SelectWordRight selects the word to the right of the cursor
func (v *View) SelectWordRight() bool {
	loc := v.Cursor.Loc()
	if !v.Cursor.HasSelection() {
		v.Cursor.origSelection[0] = loc
	}
	v.Cursor.WordRight()
	v.Cursor.SelectTo(v.Cursor.Loc())
	return true
}

// SelectWordLeft selects the word to the left of the cursor
func (v *View) SelectWordLeft() bool {
	loc := v.Cursor.Loc()
	if !v.Cursor.HasSelection() {
		v.Cursor.origSelection[0] = loc
	}
	v.Cursor.WordLeft()
	v.Cursor.SelectTo(v.Cursor.Loc())
	return true
}

// StartOfLine moves the cursor to the start of the line
func (v *View) StartOfLine() bool {
	v.Cursor.Start()
	return true
}

// EndOfLine moves the cursor to the end of the line
func (v *View) EndOfLine() bool {
	v.Cursor.End()
	return true
}

// SelectToStartOfLine selects to the start of the current line
func (v *View) SelectToStartOfLine() bool {
	loc := v.Cursor.Loc()
	if !v.Cursor.HasSelection() {
		v.Cursor.origSelection[0] = loc
	}
	v.Cursor.Start()
	v.Cursor.SelectTo(v.Cursor.Loc())
	return true
}

// SelectToEndOfLine selects to the end of the current line
func (v *View) SelectToEndOfLine() bool {
	loc := v.Cursor.Loc()
	if !v.Cursor.HasSelection() {
		v.Cursor.origSelection[0] = loc
	}
	v.Cursor.End()
	v.Cursor.SelectTo(v.Cursor.Loc())
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (v *View) CursorStart() bool {
	v.Cursor.x = 0
	v.Cursor.y = 0
	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (v *View) CursorEnd() bool {
	v.Cursor.SetLoc(v.Buf.Len())
	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (v *View) SelectToStart() bool {
	loc := v.Cursor.Loc()
	if !v.Cursor.HasSelection() {
		v.Cursor.origSelection[0] = loc
	}
	v.CursorStart()
	v.Cursor.SelectTo(0)
	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (v *View) SelectToEnd() bool {
	loc := v.Cursor.Loc()
	if !v.Cursor.HasSelection() {
		v.Cursor.origSelection[0] = loc
	}
	v.CursorEnd()
	v.Cursor.SelectTo(v.Buf.Len())
	return true
}

// InsertSpace inserts a space
func (v *View) InsertSpace() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.Buf.Insert(v.Cursor.Loc(), " ")
	v.Cursor.Right()
	return true
}

// InsertEnter inserts a newline plus possible some whitespace if autoindent is on
func (v *View) InsertEnter() bool {
	// Insert a newline
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	v.Buf.Insert(v.Cursor.Loc(), "\n")
	ws := GetLeadingWhitespace(v.Buf.Lines[v.Cursor.y])
	v.Cursor.Right()

	if settings["autoindent"].(bool) {
		v.Buf.Insert(v.Cursor.Loc(), ws)
		for i := 0; i < len(ws); i++ {
			v.Cursor.Right()
		}
	}
	v.Cursor.lastVisualX = v.Cursor.GetVisualX()
	return true
}

// Backspace deletes the previous character
func (v *View) Backspace() bool {
	// Delete a character
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else if v.Cursor.Loc() > 0 {
		// We have to do something a bit hacky here because we want to
		// delete the line by first moving left and then deleting backwards
		// but the undo redo would place the cursor in the wrong place
		// So instead we move left, save the position, move back, delete
		// and restore the position

		// If the user is using spaces instead of tabs and they are deleting
		// whitespace at the start of the line, we should delete as if its a
		// tab (tabSize number of spaces)
		lineStart := v.Buf.Lines[v.Cursor.y][:v.Cursor.x]
		tabSize := int(settings["tabsize"].(float64))
		if settings["tabsToSpaces"].(bool) && IsSpaces(lineStart) && len(lineStart) != 0 && len(lineStart)%tabSize == 0 {
			loc := v.Cursor.Loc()
			v.Cursor.SetLoc(loc - tabSize)
			cx, cy := v.Cursor.x, v.Cursor.y
			v.Cursor.SetLoc(loc)
			v.Buf.Remove(loc-tabSize, loc)
			v.Cursor.x, v.Cursor.y = cx, cy
		} else {
			v.Cursor.Left()
			cx, cy := v.Cursor.x, v.Cursor.y
			v.Cursor.Right()
			loc := v.Cursor.Loc()
			v.Buf.Remove(loc-1, loc)
			v.Cursor.x, v.Cursor.y = cx, cy
		}
	}
	v.Cursor.lastVisualX = v.Cursor.GetVisualX()
	return true
}

// Delete deletes the next character
func (v *View) Delete() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else {
		loc := v.Cursor.Loc()
		if loc < v.Buf.Len() {
			v.Buf.Remove(loc, loc+1)
		}
	}
	return true
}

// InsertTab inserts a tab or spaces
func (v *View) InsertTab() bool {
	// Insert a tab
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	if settings["tabsToSpaces"].(bool) {
		tabSize := int(settings["tabsize"].(float64))
		v.Buf.Insert(v.Cursor.Loc(), Spaces(tabSize))
		for i := 0; i < tabSize; i++ {
			v.Cursor.Right()
		}
	} else {
		v.Buf.Insert(v.Cursor.Loc(), "\t")
		v.Cursor.Right()
	}
	return true
}

// Save the buffer to disk
func (v *View) Save() bool {
	// If this is an empty buffer, ask for a filename
	if v.Buf.Path == "" {
		filename, canceled := messenger.Prompt("Filename: ")
		if !canceled {
			v.Buf.Path = filename
			v.Buf.Name = filename
		} else {
			return true
		}
	}
	err := v.Buf.Save()
	if err != nil {
		messenger.Error(err.Error())
	} else {
		messenger.Message("Saved " + v.Buf.Path)
	}
	return true
}

// Find opens a prompt and searches forward for the input
func (v *View) Find() bool {
	if v.Cursor.HasSelection() {
		searchStart = v.Cursor.curSelection[1]
	} else {
		searchStart = ToCharPos(v.Cursor.x, v.Cursor.y, v.Buf)
	}
	BeginSearch()
	return true
}

// FindNext searches forwards for the last used search term
func (v *View) FindNext() bool {
	if v.Cursor.HasSelection() {
		searchStart = v.Cursor.curSelection[1]
	} else {
		searchStart = ToCharPos(v.Cursor.x, v.Cursor.y, v.Buf)
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, true)
	return true
}

// FindPrevious searches backwards for the last used search term
func (v *View) FindPrevious() bool {
	if v.Cursor.HasSelection() {
		searchStart = v.Cursor.curSelection[0]
	} else {
		searchStart = ToCharPos(v.Cursor.x, v.Cursor.y, v.Buf)
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, false)
	return true
}

// Undo undoes the last action
func (v *View) Undo() bool {
	v.Buf.Undo()
	messenger.Message("Undid action")
	return true
}

// Redo redoes the last action
func (v *View) Redo() bool {
	v.Buf.Redo()
	messenger.Message("Redid action")
	return true
}

// Copy the selection to the system clipboard
func (v *View) Copy() bool {
	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection())
		v.freshClip = true
		messenger.Message("Copied selection")
	}
	return true
}

// CutLine cuts the current line to the clipboard
func (v *View) CutLine() bool {
	v.Cursor.SelectLine()
	if v.freshClip == true {
		if v.Cursor.HasSelection() {
			if clip, err := clipboard.ReadAll(); err != nil {
				messenger.Error(err)
			} else {
				clipboard.WriteAll(clip + v.Cursor.GetSelection())
			}
		}
	} else if time.Since(v.lastCutTime)/time.Second > 10*time.Second || v.freshClip == false {
		v.Copy()
	}
	v.freshClip = true
	v.lastCutTime = time.Now()
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()
	messenger.Message("Cut line")
	return true
}

// Cut the selection to the system clipboard
func (v *View) Cut() bool {
	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection())
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		v.freshClip = true
		messenger.Message("Cut selection")
	}
	return true
}

// DuplicateLine duplicates the current line
func (v *View) DuplicateLine() bool {
	v.Cursor.End()
	v.Buf.Insert(v.Cursor.Loc(), "\n"+v.Buf.Lines[v.Cursor.y])
	v.Cursor.Right()
	messenger.Message("Duplicated line")
	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (v *View) Paste() bool {
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	clip, _ := clipboard.ReadAll()
	v.Buf.Insert(v.Cursor.Loc(), clip)
	v.Cursor.SetLoc(v.Cursor.Loc() + Count(clip))
	v.freshClip = false
	messenger.Message("Pasted clipboard")
	return true
}

// SelectAll selects the entire buffer
func (v *View) SelectAll() bool {
	v.Cursor.curSelection[0] = 0
	v.Cursor.curSelection[1] = v.Buf.Len()
	// Put the cursor at the beginning
	v.Cursor.x = 0
	v.Cursor.y = 0
	return true
}

// OpenFile opens a new file in the buffer
func (v *View) OpenFile() bool {
	if v.CanClose("Continue? (yes, no, save) ") {
		filename, canceled := messenger.Prompt("File to open: ")
		if canceled {
			return true
		}
		home, _ := homedir.Dir()
		filename = strings.Replace(filename, "~", home, 1)
		file, err := ioutil.ReadFile(filename)

		if err != nil {
			messenger.Error(err.Error())
			return true
		}
		buf := NewBuffer(string(file), filename)
		v.OpenBuffer(buf)
	}
	return true
}

// Start moves the viewport to the start of the buffer
func (v *View) Start() bool {
	v.Topline = 0
	return false
}

// End moves the viewport to the end of the buffer
func (v *View) End() bool {
	if v.height > v.Buf.NumLines {
		v.Topline = 0
	} else {
		v.Topline = v.Buf.NumLines - v.height
	}
	return false
}

// PageUp scrolls the view up a page
func (v *View) PageUp() bool {
	if v.Topline > v.height {
		v.ScrollUp(v.height)
	} else {
		v.Topline = 0
	}
	return false
}

// PageDown scrolls the view down a page
func (v *View) PageDown() bool {
	if v.Buf.NumLines-(v.Topline+v.height) > v.height {
		v.ScrollDown(v.height)
	} else if v.Buf.NumLines >= v.height {
		v.Topline = v.Buf.NumLines - v.height
	}
	return false
}

// HalfPageUp scrolls the view up half a page
func (v *View) HalfPageUp() bool {
	if v.Topline > v.height/2 {
		v.ScrollUp(v.height / 2)
	} else {
		v.Topline = 0
	}
	return false
}

// HalfPageDown scrolls the view down half a page
func (v *View) HalfPageDown() bool {
	if v.Buf.NumLines-(v.Topline+v.height) > v.height/2 {
		v.ScrollDown(v.height / 2)
	} else {
		if v.Buf.NumLines >= v.height {
			v.Topline = v.Buf.NumLines - v.height
		}
	}
	return false
}

// ToggleRuler turns line numbers off and on
func (v *View) ToggleRuler() bool {
	if settings["ruler"] == false {
		settings["ruler"] = true
		messenger.Message("Enabled ruler")
	} else {
		settings["ruler"] = false
		messenger.Message("Disabled ruler")
	}
	return false
}

// JumpLine jumps to a line and moves the view accordingly.
func (v *View) JumpLine() bool {
	// Prompt for line number
	linestring, canceled := messenger.Prompt("Jump to line # ")
	if canceled {
		return false
	}
	lineint, err := strconv.Atoi(linestring)
	lineint = lineint - 1 // fix offset
	if err != nil {
		messenger.Error(err) // return errors
		return false
	}
	// Move cursor and view if possible.
	if lineint < v.Buf.NumLines {
		v.Cursor.x = 0
		v.Cursor.y = lineint
		return true
	}
	messenger.Error("Only ", v.Buf.NumLines, " lines to jump")
	return false
}

// ClearStatus clears the messenger bar
func (v *View) ClearStatus() bool {
	messenger.Message("")
	return false
}

// None is no action
func None() bool {
	return false
}
