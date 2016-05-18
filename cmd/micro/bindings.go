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

var bindings map[tcell.Key]func(*View) bool

// InitBindings initializes the keybindings for micro
func InitBindings() {
	bindings = make(map[tcell.Key]func(*View) bool)

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
		"Shell":               (*View).Shell,
		"ClearStatus":         (*View).ClearStatus,
	}

	keys := map[string]tcell.Key{
		"Up":             tcell.KeyUp,
		"Down":           tcell.KeyDown,
		"Right":          tcell.KeyRight,
		"Left":           tcell.KeyLeft,
		"AltUp":          tcell.KeyAltUp,
		"AltDown":        tcell.KeyAltDown,
		"AltLeft":        tcell.KeyAltLeft,
		"AltRight":       tcell.KeyAltRight,
		"CtrlUp":         tcell.KeyCtrlUp,
		"CtrlDown":       tcell.KeyCtrlDown,
		"CtrlLeft":       tcell.KeyCtrlLeft,
		"CtrlRight":      tcell.KeyCtrlRight,
		"ShiftUp":        tcell.KeyShiftUp,
		"ShiftDown":      tcell.KeyShiftDown,
		"ShiftLeft":      tcell.KeyShiftLeft,
		"ShiftRight":     tcell.KeyShiftRight,
		"AltShiftUp":     tcell.KeyAltShiftUp,
		"AltShiftDown":   tcell.KeyAltShiftDown,
		"AltShiftLeft":   tcell.KeyAltShiftLeft,
		"AltShiftRight":  tcell.KeyAltShiftRight,
		"CtrlShiftUp":    tcell.KeyCtrlShiftUp,
		"CtrlShiftDown":  tcell.KeyCtrlShiftDown,
		"CtrlShiftLeft":  tcell.KeyCtrlShiftLeft,
		"CtrlShiftRight": tcell.KeyCtrlShiftRight,
		"UpLeft":         tcell.KeyUpLeft,
		"UpRight":        tcell.KeyUpRight,
		"DownLeft":       tcell.KeyDownLeft,
		"DownRight":      tcell.KeyDownRight,
		"Center":         tcell.KeyCenter,
		"PgUp":           tcell.KeyPgUp,
		"PgDn":           tcell.KeyPgDn,
		"Home":           tcell.KeyHome,
		"End":            tcell.KeyEnd,
		"Insert":         tcell.KeyInsert,
		"Delete":         tcell.KeyDelete,
		"Help":           tcell.KeyHelp,
		"Exit":           tcell.KeyExit,
		"Clear":          tcell.KeyClear,
		"Cancel":         tcell.KeyCancel,
		"Print":          tcell.KeyPrint,
		"Pause":          tcell.KeyPause,
		"Backtab":        tcell.KeyBacktab,
		"F1":             tcell.KeyF1,
		"F2":             tcell.KeyF2,
		"F3":             tcell.KeyF3,
		"F4":             tcell.KeyF4,
		"F5":             tcell.KeyF5,
		"F6":             tcell.KeyF6,
		"F7":             tcell.KeyF7,
		"F8":             tcell.KeyF8,
		"F9":             tcell.KeyF9,
		"F10":            tcell.KeyF10,
		"F11":            tcell.KeyF11,
		"F12":            tcell.KeyF12,
		"F13":            tcell.KeyF13,
		"F14":            tcell.KeyF14,
		"F15":            tcell.KeyF15,
		"F16":            tcell.KeyF16,
		"F17":            tcell.KeyF17,
		"F18":            tcell.KeyF18,
		"F19":            tcell.KeyF19,
		"F20":            tcell.KeyF20,
		"F21":            tcell.KeyF21,
		"F22":            tcell.KeyF22,
		"F23":            tcell.KeyF23,
		"F24":            tcell.KeyF24,
		"F25":            tcell.KeyF25,
		"F26":            tcell.KeyF26,
		"F27":            tcell.KeyF27,
		"F28":            tcell.KeyF28,
		"F29":            tcell.KeyF29,
		"F30":            tcell.KeyF30,
		"F31":            tcell.KeyF31,
		"F32":            tcell.KeyF32,
		"F33":            tcell.KeyF33,
		"F34":            tcell.KeyF34,
		"F35":            tcell.KeyF35,
		"F36":            tcell.KeyF36,
		"F37":            tcell.KeyF37,
		"F38":            tcell.KeyF38,
		"F39":            tcell.KeyF39,
		"F40":            tcell.KeyF40,
		"F41":            tcell.KeyF41,
		"F42":            tcell.KeyF42,
		"F43":            tcell.KeyF43,
		"F44":            tcell.KeyF44,
		"F45":            tcell.KeyF45,
		"F46":            tcell.KeyF46,
		"F47":            tcell.KeyF47,
		"F48":            tcell.KeyF48,
		"F49":            tcell.KeyF49,
		"F50":            tcell.KeyF50,
		"F51":            tcell.KeyF51,
		"F52":            tcell.KeyF52,
		"F53":            tcell.KeyF53,
		"F54":            tcell.KeyF54,
		"F55":            tcell.KeyF55,
		"F56":            tcell.KeyF56,
		"F57":            tcell.KeyF57,
		"F58":            tcell.KeyF58,
		"F59":            tcell.KeyF59,
		"F60":            tcell.KeyF60,
		"F61":            tcell.KeyF61,
		"F62":            tcell.KeyF62,
		"F63":            tcell.KeyF63,
		"F64":            tcell.KeyF64,
		"CtrlSpace":      tcell.KeyCtrlSpace,
		"CtrlA":          tcell.KeyCtrlA,
		"CtrlB":          tcell.KeyCtrlB,
		"CtrlC":          tcell.KeyCtrlC,
		"CtrlD":          tcell.KeyCtrlD,
		"CtrlE":          tcell.KeyCtrlE,
		"CtrlF":          tcell.KeyCtrlF,
		"CtrlG":          tcell.KeyCtrlG,
		"CtrlH":          tcell.KeyCtrlH,
		"CtrlI":          tcell.KeyCtrlI,
		"CtrlJ":          tcell.KeyCtrlJ,
		"CtrlK":          tcell.KeyCtrlK,
		"CtrlL":          tcell.KeyCtrlL,
		"CtrlM":          tcell.KeyCtrlM,
		"CtrlN":          tcell.KeyCtrlN,
		"CtrlO":          tcell.KeyCtrlO,
		"CtrlP":          tcell.KeyCtrlP,
		"CtrlQ":          tcell.KeyCtrlQ,
		"CtrlR":          tcell.KeyCtrlR,
		"CtrlS":          tcell.KeyCtrlS,
		"CtrlT":          tcell.KeyCtrlT,
		"CtrlU":          tcell.KeyCtrlU,
		"CtrlV":          tcell.KeyCtrlV,
		"CtrlW":          tcell.KeyCtrlW,
		"CtrlX":          tcell.KeyCtrlX,
		"CtrlY":          tcell.KeyCtrlY,
		"CtrlZ":          tcell.KeyCtrlZ,
		"CtrlLeftSq":     tcell.KeyCtrlLeftSq,
		"CtrlBackslash":  tcell.KeyCtrlBackslash,
		"CtrlRightSq":    tcell.KeyCtrlRightSq,
		"CtrlCarat":      tcell.KeyCtrlCarat,
		"CtrlUnderscore": tcell.KeyCtrlUnderscore,
		"Backspace":      tcell.KeyBackspace,
		"Tab":            tcell.KeyTab,
		"Esc":            tcell.KeyEsc,
		"Escape":         tcell.KeyEscape,
		"Enter":          tcell.KeyEnter,
		"Space":          tcell.KeySpace,
		"Backspace2":     tcell.KeyBackspace2,
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
		"CtrlB":          "Shell",
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

// Shell launches a shell prompt
func (v *View) Shell() bool {
	input, canceled := messenger.Prompt("$ ")
	if !canceled {
		HandleShellCommand(input, v, true)
	}
	return false

}

// ClearStatus erases stale messages from the statusbar
func (v *View) ClearStatus() bool {
	messenger.Message("")
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

// None is no action
func None() bool {
	return false
}
