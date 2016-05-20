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
		"Up":             Key{tcell.KeyUp, tcell.ModNone},
		"Down":           Key{tcell.KeyDown, tcell.ModNone},
		"Right":          Key{tcell.KeyRight, tcell.ModNone},
		"Left":           Key{tcell.KeyLeft, tcell.ModNone},
		"AltUp":          Key{tcell.KeyUp, tcell.ModAlt},
		"AltDown":        Key{tcell.KeyDown, tcell.ModAlt},
		"AltLeft":        Key{tcell.KeyLeft, tcell.ModAlt},
		"AltRight":       Key{tcell.KeyRight, tcell.ModAlt},
		"CtrlUp":         Key{tcell.KeyUp, tcell.ModCtrl},
		"CtrlDown":       Key{tcell.KeyDown, tcell.ModCtrl},
		"CtrlLeft":       Key{tcell.KeyLeft, tcell.ModCtrl},
		"CtrlRight":      Key{tcell.KeyRight, tcell.ModCtrl},
		"ShiftUp":        Key{tcell.KeyUp, tcell.ModShift},
		"ShiftDown":      Key{tcell.KeyDown, tcell.ModShift},
		"ShiftLeft":      Key{tcell.KeyLeft, tcell.ModShift},
		"ShiftRight":     Key{tcell.KeyRight, tcell.ModShift},
		"AltShiftUp":     Key{tcell.KeyUp, tcell.ModShift | tcell.ModAlt},
		"AltShiftDown":   Key{tcell.KeyDown, tcell.ModShift | tcell.ModAlt},
		"AltShiftLeft":   Key{tcell.KeyLeft, tcell.ModShift | tcell.ModAlt},
		"AltShiftRight":  Key{tcell.KeyRight, tcell.ModShift | tcell.ModAlt},
		"CtrlShiftUp":    Key{tcell.KeyUp, tcell.ModShift | tcell.ModCtrl},
		"CtrlShiftDown":  Key{tcell.KeyDown, tcell.ModShift | tcell.ModCtrl},
		"CtrlShiftLeft":  Key{tcell.KeyLeft, tcell.ModShift | tcell.ModCtrl},
		"CtrlShiftRight": Key{tcell.KeyRight, tcell.ModShift | tcell.ModCtrl},
		"UpLeft":         Key{tcell.KeyUpLeft, tcell.ModNone},
		"UpRight":        Key{tcell.KeyUpRight, tcell.ModNone},
		"DownLeft":       Key{tcell.KeyDownLeft, tcell.ModNone},
		"DownRight":      Key{tcell.KeyDownRight, tcell.ModNone},
		"Center":         Key{tcell.KeyCenter, tcell.ModNone},
		"PgUp":           Key{tcell.KeyPgUp, tcell.ModNone},
		"PgDn":           Key{tcell.KeyPgDn, tcell.ModNone},
		"Home":           Key{tcell.KeyHome, tcell.ModNone},
		"End":            Key{tcell.KeyEnd, tcell.ModNone},
		"Insert":         Key{tcell.KeyInsert, tcell.ModNone},
		"Delete":         Key{tcell.KeyDelete, tcell.ModNone},
		"Help":           Key{tcell.KeyHelp, tcell.ModNone},
		"Exit":           Key{tcell.KeyExit, tcell.ModNone},
		"Clear":          Key{tcell.KeyClear, tcell.ModNone},
		"Cancel":         Key{tcell.KeyCancel, tcell.ModNone},
		"Print":          Key{tcell.KeyPrint, tcell.ModNone},
		"Pause":          Key{tcell.KeyPause, tcell.ModNone},
		"Backtab":        Key{tcell.KeyBacktab, tcell.ModNone},
		"F1":             Key{tcell.KeyF1, tcell.ModNone},
		"F2":             Key{tcell.KeyF2, tcell.ModNone},
		"F3":             Key{tcell.KeyF3, tcell.ModNone},
		"F4":             Key{tcell.KeyF4, tcell.ModNone},
		"F5":             Key{tcell.KeyF5, tcell.ModNone},
		"F6":             Key{tcell.KeyF6, tcell.ModNone},
		"F7":             Key{tcell.KeyF7, tcell.ModNone},
		"F8":             Key{tcell.KeyF8, tcell.ModNone},
		"F9":             Key{tcell.KeyF9, tcell.ModNone},
		"F10":            Key{tcell.KeyF10, tcell.ModNone},
		"F11":            Key{tcell.KeyF11, tcell.ModNone},
		"F12":            Key{tcell.KeyF12, tcell.ModNone},
		"F13":            Key{tcell.KeyF13, tcell.ModNone},
		"F14":            Key{tcell.KeyF14, tcell.ModNone},
		"F15":            Key{tcell.KeyF15, tcell.ModNone},
		"F16":            Key{tcell.KeyF16, tcell.ModNone},
		"F17":            Key{tcell.KeyF17, tcell.ModNone},
		"F18":            Key{tcell.KeyF18, tcell.ModNone},
		"F19":            Key{tcell.KeyF19, tcell.ModNone},
		"F20":            Key{tcell.KeyF20, tcell.ModNone},
		"F21":            Key{tcell.KeyF21, tcell.ModNone},
		"F22":            Key{tcell.KeyF22, tcell.ModNone},
		"F23":            Key{tcell.KeyF23, tcell.ModNone},
		"F24":            Key{tcell.KeyF24, tcell.ModNone},
		"F25":            Key{tcell.KeyF25, tcell.ModNone},
		"F26":            Key{tcell.KeyF26, tcell.ModNone},
		"F27":            Key{tcell.KeyF27, tcell.ModNone},
		"F28":            Key{tcell.KeyF28, tcell.ModNone},
		"F29":            Key{tcell.KeyF29, tcell.ModNone},
		"F30":            Key{tcell.KeyF30, tcell.ModNone},
		"F31":            Key{tcell.KeyF31, tcell.ModNone},
		"F32":            Key{tcell.KeyF32, tcell.ModNone},
		"F33":            Key{tcell.KeyF33, tcell.ModNone},
		"F34":            Key{tcell.KeyF34, tcell.ModNone},
		"F35":            Key{tcell.KeyF35, tcell.ModNone},
		"F36":            Key{tcell.KeyF36, tcell.ModNone},
		"F37":            Key{tcell.KeyF37, tcell.ModNone},
		"F38":            Key{tcell.KeyF38, tcell.ModNone},
		"F39":            Key{tcell.KeyF39, tcell.ModNone},
		"F40":            Key{tcell.KeyF40, tcell.ModNone},
		"F41":            Key{tcell.KeyF41, tcell.ModNone},
		"F42":            Key{tcell.KeyF42, tcell.ModNone},
		"F43":            Key{tcell.KeyF43, tcell.ModNone},
		"F44":            Key{tcell.KeyF44, tcell.ModNone},
		"F45":            Key{tcell.KeyF45, tcell.ModNone},
		"F46":            Key{tcell.KeyF46, tcell.ModNone},
		"F47":            Key{tcell.KeyF47, tcell.ModNone},
		"F48":            Key{tcell.KeyF48, tcell.ModNone},
		"F49":            Key{tcell.KeyF49, tcell.ModNone},
		"F50":            Key{tcell.KeyF50, tcell.ModNone},
		"F51":            Key{tcell.KeyF51, tcell.ModNone},
		"F52":            Key{tcell.KeyF52, tcell.ModNone},
		"F53":            Key{tcell.KeyF53, tcell.ModNone},
		"F54":            Key{tcell.KeyF54, tcell.ModNone},
		"F55":            Key{tcell.KeyF55, tcell.ModNone},
		"F56":            Key{tcell.KeyF56, tcell.ModNone},
		"F57":            Key{tcell.KeyF57, tcell.ModNone},
		"F58":            Key{tcell.KeyF58, tcell.ModNone},
		"F59":            Key{tcell.KeyF59, tcell.ModNone},
		"F60":            Key{tcell.KeyF60, tcell.ModNone},
		"F61":            Key{tcell.KeyF61, tcell.ModNone},
		"F62":            Key{tcell.KeyF62, tcell.ModNone},
		"F63":            Key{tcell.KeyF63, tcell.ModNone},
		"F64":            Key{tcell.KeyF64, tcell.ModNone},
		"CtrlSpace":      Key{tcell.KeyCtrlSpace, tcell.ModCtrl},
		"CtrlA":          Key{tcell.KeyCtrlA, tcell.ModCtrl},
		"CtrlB":          Key{tcell.KeyCtrlB, tcell.ModCtrl},
		"CtrlC":          Key{tcell.KeyCtrlC, tcell.ModCtrl},
		"CtrlD":          Key{tcell.KeyCtrlD, tcell.ModCtrl},
		"CtrlE":          Key{tcell.KeyCtrlE, tcell.ModCtrl},
		"CtrlF":          Key{tcell.KeyCtrlF, tcell.ModCtrl},
		"CtrlG":          Key{tcell.KeyCtrlG, tcell.ModCtrl},
		"CtrlH":          Key{tcell.KeyCtrlH, tcell.ModCtrl},
		"CtrlI":          Key{tcell.KeyCtrlI, tcell.ModCtrl},
		"CtrlJ":          Key{tcell.KeyCtrlJ, tcell.ModCtrl},
		"CtrlK":          Key{tcell.KeyCtrlK, tcell.ModCtrl},
		"CtrlL":          Key{tcell.KeyCtrlL, tcell.ModCtrl},
		"CtrlM":          Key{tcell.KeyCtrlM, tcell.ModCtrl},
		"CtrlN":          Key{tcell.KeyCtrlN, tcell.ModCtrl},
		"CtrlO":          Key{tcell.KeyCtrlO, tcell.ModCtrl},
		"CtrlP":          Key{tcell.KeyCtrlP, tcell.ModCtrl},
		"CtrlQ":          Key{tcell.KeyCtrlQ, tcell.ModCtrl},
		"CtrlR":          Key{tcell.KeyCtrlR, tcell.ModCtrl},
		"CtrlS":          Key{tcell.KeyCtrlS, tcell.ModCtrl},
		"CtrlT":          Key{tcell.KeyCtrlT, tcell.ModCtrl},
		"CtrlU":          Key{tcell.KeyCtrlU, tcell.ModCtrl},
		"CtrlV":          Key{tcell.KeyCtrlV, tcell.ModCtrl},
		"CtrlW":          Key{tcell.KeyCtrlW, tcell.ModCtrl},
		"CtrlX":          Key{tcell.KeyCtrlX, tcell.ModCtrl},
		"CtrlY":          Key{tcell.KeyCtrlY, tcell.ModCtrl},
		"CtrlZ":          Key{tcell.KeyCtrlZ, tcell.ModCtrl},
		"CtrlLeftSq":     Key{tcell.KeyCtrlLeftSq, tcell.ModCtrl},
		"CtrlBackslash":  Key{tcell.KeyCtrlBackslash, tcell.ModCtrl},
		"CtrlRightSq":    Key{tcell.KeyCtrlRightSq, tcell.ModCtrl},
		"CtrlCarat":      Key{tcell.KeyCtrlCarat, tcell.ModCtrl},
		"CtrlUnderscore": Key{tcell.KeyCtrlUnderscore, tcell.ModCtrl},
		"Backspace":      Key{tcell.KeyBackspace, tcell.ModNone},
		"Tab":            Key{tcell.KeyTab, tcell.ModNone},
		"Esc":            Key{tcell.KeyEsc, tcell.ModNone},
		"Escape":         Key{tcell.KeyEscape, tcell.ModNone},
		"Enter":          Key{tcell.KeyEnter, tcell.ModNone},
		"Backspace2":     Key{tcell.KeyBackspace2, tcell.ModNone},
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
		bindings[keys[k]] = actions[v]
	}
	for k, v := range parsed {
		bindings[keys[k]] = actions[v]
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
		"CtrlV":          "Paste",
		"CtrlA":          "SelectAll",
		"Home":           "Start",
		"End":            "End",
		"PgUp":           "PageUp",
		"PgDn":           "PageDown",
		"CtrlU":          "HalfPageUp",
		"CtrlD":          "HalfPageDown",
		"CtrlR":          "ToggleRuler",
		"CtrlL":          "JumpLine",
		"Delete":         "Delete",
		"Esc":            "ClearStatus",
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
	// Insert a space
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.eh.Insert(v.Cursor.Loc(), " ")
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

	v.eh.Insert(v.Cursor.Loc(), "\n")
	ws := GetLeadingWhitespace(v.Buf.Lines[v.Cursor.y])
	v.Cursor.Right()

	if settings["autoindent"].(bool) {
		v.eh.Insert(v.Cursor.Loc(), ws)
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
			v.eh.Remove(loc-tabSize, loc)
			v.Cursor.x, v.Cursor.y = cx, cy
		} else {
			v.Cursor.Left()
			cx, cy := v.Cursor.x, v.Cursor.y
			v.Cursor.Right()
			loc := v.Cursor.Loc()
			v.eh.Remove(loc-1, loc)
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
			v.eh.Remove(loc, loc+1)
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
		v.eh.Insert(v.Cursor.Loc(), Spaces(tabSize))
		for i := 0; i < tabSize; i++ {
			v.Cursor.Right()
		}
	} else {
		v.eh.Insert(v.Cursor.Loc(), "\t")
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
	v.eh.Undo()
	return true
}

// Redo redoes the last action
func (v *View) Redo() bool {
	v.eh.Redo()
	return true
}

// Copy the selection to the system clipboard
func (v *View) Copy() bool {
	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection())
		v.freshClip = true
		messenger.Message("Copied selection to clipboard")
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
	return true
}

// Cut the selection to the system clipboard
func (v *View) Cut() bool {
	if v.Cursor.HasSelection() {
		clipboard.WriteAll(v.Cursor.GetSelection())
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		v.freshClip = true
	}
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
	v.eh.Insert(v.Cursor.Loc(), clip)
	v.Cursor.SetLoc(v.Cursor.Loc() + Count(clip))
	v.freshClip = false
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
	} else {
		settings["ruler"] = false
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
