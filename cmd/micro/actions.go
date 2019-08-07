package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/yuin/gopher-lua"
	"github.com/zyedidia/clipboard"
	"github.com/zyedidia/micro/cmd/micro/shellwords"
	"github.com/zyedidia/tcell"
)

// PreActionCall executes the lua pre callback if possible
func PreActionCall(funcName string, view *View, args ...interface{}) bool {
	executeAction := true
	for pl := range loadedPlugins {
		ret, err := Call(pl+".pre"+funcName, append([]interface{}{view}, args...)...)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
		if ret == lua.LFalse {
			executeAction = false
		}
	}
	return executeAction
}

// PostActionCall executes the lua plugin callback if possible
func PostActionCall(funcName string, view *View, args ...interface{}) bool {
	relocate := true
	for pl := range loadedPlugins {
		ret, err := Call(pl+".on"+funcName, append([]interface{}{view}, args...)...)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
		if ret == lua.LFalse {
			relocate = false
		}
	}
	return relocate
}

func (v *View) deselect(index int) bool {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[index]
		v.Cursor.ResetSelection()
		v.Cursor.StoreVisualX()
		return true
	}
	return false
}

// MousePress is the event that should happen when a normal click happens
// This is almost always bound to left click
func (v *View) MousePress(usePlugin bool, e *tcell.EventMouse) bool {
	if usePlugin && !PreActionCall("MousePress", v, e) {
		return false
	}

	x, y := e.Position()
	x -= v.lineNumOffset - v.leftCol + v.x
	y += v.Topline - v.y

	// This is usually bound to left click
	v.MoveToMouseClick(x, y)
	if v.mouseReleased {
		if len(v.Buf.cursors) > 1 {
			for i := 1; i < len(v.Buf.cursors); i++ {
				v.Buf.cursors[i] = nil
			}
			v.Buf.cursors = v.Buf.cursors[:1]
			v.Buf.UpdateCursors()
			v.Cursor.ResetSelection()
			v.Relocate()
		}
		if time.Since(v.lastClickTime)/time.Millisecond < doubleClickThreshold && (x == v.lastLoc.X && y == v.lastLoc.Y) {
			if v.doubleClick {
				// Triple click
				v.lastClickTime = time.Now()

				v.tripleClick = true
				v.doubleClick = false

				v.Cursor.SelectLine()
				v.Cursor.CopySelection("primary")
			} else {
				// Double click
				v.lastClickTime = time.Now()

				v.doubleClick = true
				v.tripleClick = false

				v.Cursor.SelectWord()
				v.Cursor.CopySelection("primary")
			}
		} else {
			v.doubleClick = false
			v.tripleClick = false
			v.lastClickTime = time.Now()

			v.Cursor.OrigSelection[0] = v.Cursor.Loc
			v.Cursor.CurSelection[0] = v.Cursor.Loc
			v.Cursor.CurSelection[1] = v.Cursor.Loc
		}
		v.mouseReleased = false
	} else if !v.mouseReleased {
		if v.tripleClick {
			v.Cursor.AddLineToSelection()
		} else if v.doubleClick {
			v.Cursor.AddWordToSelection()
		} else {
			v.Cursor.SetSelectionEnd(v.Cursor.Loc)
			v.Cursor.CopySelection("primary")
		}
	}

	v.lastLoc = Loc{x, y}

	if usePlugin {
		PostActionCall("MousePress", v, e)
	}
	return false
}

// ScrollUpAction scrolls the view up
func (v *View) ScrollUpAction(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ScrollUp", v) {
			return false
		}

		scrollspeed := int(v.Buf.Settings["scrollspeed"].(float64))
		v.ScrollUp(scrollspeed)

		if usePlugin {
			PostActionCall("ScrollUp", v)
		}
	}
	return false
}

// ScrollDownAction scrolls the view up
func (v *View) ScrollDownAction(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ScrollDown", v) {
			return false
		}

		scrollspeed := int(v.Buf.Settings["scrollspeed"].(float64))
		v.ScrollDown(scrollspeed)

		if usePlugin {
			PostActionCall("ScrollDown", v)
		}
	}
	return false
}

// Center centers the view on the cursor
func (v *View) Center(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Center", v) {
		return false
	}

	v.Topline = v.Cursor.Y - v.Height/2
	if v.Topline+v.Height > v.Buf.NumLines {
		v.Topline = v.Buf.NumLines - v.Height
	}
	if v.Topline < 0 {
		v.Topline = 0
	}

	if usePlugin {
		return PostActionCall("Center", v)
	}
	return true
}

// CursorUp moves the cursor up
func (v *View) CursorUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorUp", v) {
		return false
	}

	v.deselect(0)
	v.Cursor.Up()

	if usePlugin {
		return PostActionCall("CursorUp", v)
	}
	return true
}

// CursorDown moves the cursor down
func (v *View) CursorDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorDown", v) {
		return false
	}

	v.deselect(1)
	v.Cursor.Down()

	if usePlugin {
		return PostActionCall("CursorDown", v)
	}
	return true
}

// CursorLeft moves the cursor left
func (v *View) CursorLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorLeft", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
		v.Cursor.StoreVisualX()
	} else {
		tabstospaces := v.Buf.Settings["tabstospaces"].(bool)
		tabmovement := v.Buf.Settings["tabmovement"].(bool)
		if tabstospaces && tabmovement {
			tabsize := int(v.Buf.Settings["tabsize"].(float64))
			line := v.Buf.Line(v.Cursor.Y)
			if v.Cursor.X-tabsize >= 0 && line[v.Cursor.X-tabsize:v.Cursor.X] == Spaces(tabsize) && IsStrWhitespace(line[0:v.Cursor.X-tabsize]) {
				for i := 0; i < tabsize; i++ {
					v.Cursor.Left()
				}
			} else {
				v.Cursor.Left()
			}
		} else {
			v.Cursor.Left()
		}
	}

	if usePlugin {
		return PostActionCall("CursorLeft", v)
	}
	return true
}

// CursorRight moves the cursor right
func (v *View) CursorRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorRight", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1]
		v.Cursor.ResetSelection()
		v.Cursor.StoreVisualX()
	} else {
		tabstospaces := v.Buf.Settings["tabstospaces"].(bool)
		tabmovement := v.Buf.Settings["tabmovement"].(bool)
		if tabstospaces && tabmovement {
			tabsize := int(v.Buf.Settings["tabsize"].(float64))
			line := v.Buf.Line(v.Cursor.Y)
			if v.Cursor.X+tabsize < Count(line) && line[v.Cursor.X:v.Cursor.X+tabsize] == Spaces(tabsize) && IsStrWhitespace(line[0:v.Cursor.X]) {
				for i := 0; i < tabsize; i++ {
					v.Cursor.Right()
				}
			} else {
				v.Cursor.Right()
			}
		} else {
			v.Cursor.Right()
		}
	}

	if usePlugin {
		return PostActionCall("CursorRight", v)
	}
	return true
}

// WordRight moves the cursor one word to the right
func (v *View) WordRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("WordRight", v) {
		return false
	}

	v.Cursor.WordRight()

	if usePlugin {
		return PostActionCall("WordRight", v)
	}
	return true
}

// WordLeft moves the cursor one word to the left
func (v *View) WordLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("WordLeft", v) {
		return false
	}

	v.Cursor.WordLeft()

	if usePlugin {
		return PostActionCall("WordLeft", v)
	}
	return true
}

// SelectUp selects up one line
func (v *View) SelectUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectUp", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Up()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectUp", v)
	}
	return true
}

// SelectDown selects down one line
func (v *View) SelectDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectDown", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Down()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectDown", v)
	}
	return true
}

// SelectLeft selects the character to the left of the cursor
func (v *View) SelectLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectLeft", v) {
		return false
	}

	loc := v.Cursor.Loc
	count := v.Buf.End()
	if loc.GreaterThan(count) {
		loc = count
	}
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = loc
	}
	v.Cursor.Left()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectLeft", v)
	}
	return true
}

// SelectRight selects the character to the right of the cursor
func (v *View) SelectRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectRight", v) {
		return false
	}

	loc := v.Cursor.Loc
	count := v.Buf.End()
	if loc.GreaterThan(count) {
		loc = count
	}
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = loc
	}
	v.Cursor.Right()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectRight", v)
	}
	return true
}

// SelectWordRight selects the word to the right of the cursor
func (v *View) SelectWordRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectWordRight", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordRight()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectWordRight", v)
	}
	return true
}

// SelectWordLeft selects the word to the left of the cursor
func (v *View) SelectWordLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectWordLeft", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordLeft()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectWordLeft", v)
	}
	return true
}

// StartOfLine moves the cursor to the start of the line
func (v *View) StartOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("StartOfLine", v) {
		return false
	}

	v.deselect(0)

	if v.Cursor.X != 0 {
		v.Cursor.Start()
	} else {
		v.Cursor.StartOfText()
	}

	if usePlugin {
		return PostActionCall("StartOfLine", v)
	}
	return true
}

// EndOfLine moves the cursor to the end of the line
func (v *View) EndOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("EndOfLine", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.End()

	if usePlugin {
		return PostActionCall("EndOfLine", v)
	}
	return true
}

// SelectLine selects the entire current line
func (v *View) SelectLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectLine", v) {
		return false
	}

	v.Cursor.SelectLine()

	if usePlugin {
		return PostActionCall("SelectLine", v)
	}
	return true
}

// SelectToStartOfLine selects to the start of the current line
func (v *View) SelectToStartOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToStartOfLine", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Start()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectToStartOfLine", v)
	}
	return true
}

// SelectToEndOfLine selects to the end of the current line
func (v *View) SelectToEndOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToEndOfLine", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.End()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectToEndOfLine", v)
	}
	return true
}

// ParagraphPrevious moves the cursor to the previous empty line, or beginning of the buffer if there's none
func (v *View) ParagraphPrevious(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ParagraphPrevious", v) {
		return false
	}
	var line int
	for line = v.Cursor.Y; line > 0; line-- {
		if len(v.Buf.lines[line].data) == 0 && line != v.Cursor.Y {
			v.Cursor.X = 0
			v.Cursor.Y = line
			break
		}
	}
	// If no empty line found. move cursor to end of buffer
	if line == 0 {
		v.Cursor.Loc = v.Buf.Start()
	}

	if usePlugin {
		return PostActionCall("ParagraphPrevious", v)
	}
	return true
}

// ParagraphNext moves the cursor to the next empty line, or end of the buffer if there's none
func (v *View) ParagraphNext(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ParagraphNext", v) {
		return false
	}

	var line int
	for line = v.Cursor.Y; line < len(v.Buf.lines); line++ {
		if len(v.Buf.lines[line].data) == 0 && line != v.Cursor.Y {
			v.Cursor.X = 0
			v.Cursor.Y = line
			break
		}
	}
	// If no empty line found. move cursor to end of buffer
	if line == len(v.Buf.lines) {
		v.Cursor.Loc = v.Buf.End()
	}

	if usePlugin {
		return PostActionCall("ParagraphNext", v)
	}
	return true
}

// Retab changes all tabs to spaces or all spaces to tabs depending
// on the user's settings
func (v *View) Retab(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Retab", v) {
		return false
	}

	toSpaces := v.Buf.Settings["tabstospaces"].(bool)
	tabsize := int(v.Buf.Settings["tabsize"].(float64))
	dirty := false

	for i := 0; i < v.Buf.NumLines; i++ {
		l := v.Buf.Line(i)

		ws := GetLeadingWhitespace(l)
		if ws != "" {
			if toSpaces {
				ws = strings.Replace(ws, "\t", Spaces(tabsize), -1)
			} else {
				ws = strings.Replace(ws, Spaces(tabsize), "\t", -1)
			}
		}

		l = strings.TrimLeft(l, " \t")
		v.Buf.lines[i].data = []byte(ws + l)
		dirty = true
	}

	v.Buf.IsModified = dirty

	if usePlugin {
		return PostActionCall("Retab", v)
	}
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (v *View) CursorStart(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorStart", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.X = 0
	v.Cursor.Y = 0

	if usePlugin {
		return PostActionCall("CursorStart", v)
	}
	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (v *View) CursorEnd(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorEnd", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.Loc = v.Buf.End()
	v.Cursor.StoreVisualX()

	if usePlugin {
		return PostActionCall("CursorEnd", v)
	}
	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (v *View) SelectToStart(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToStart", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorStart(false)
	v.Cursor.SelectTo(v.Buf.Start())

	if usePlugin {
		return PostActionCall("SelectToStart", v)
	}
	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (v *View) SelectToEnd(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToEnd", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorEnd(false)
	v.Cursor.SelectTo(v.Buf.End())

	if usePlugin {
		return PostActionCall("SelectToEnd", v)
	}
	return true
}

// InsertSpace inserts a space
func (v *View) InsertSpace(usePlugin bool) bool {
	if usePlugin && !PreActionCall("InsertSpace", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.Buf.Insert(v.Cursor.Loc, " ")
	// v.Cursor.Right()

	if usePlugin {
		return PostActionCall("InsertSpace", v)
	}
	return true
}

// InsertNewline inserts a newline plus possible some whitespace if autoindent is on
func (v *View) InsertNewline(usePlugin bool) bool {
	if usePlugin && !PreActionCall("InsertNewline", v) {
		return false
	}

	// Insert a newline
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	ws := GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))
	cx := v.Cursor.X
	v.Buf.Insert(v.Cursor.Loc, "\n")
	// v.Cursor.Right()

	if v.Buf.Settings["autoindent"].(bool) {
		if cx < len(ws) {
			ws = ws[0:cx]
		}
		v.Buf.Insert(v.Cursor.Loc, ws)
		// for i := 0; i < len(ws); i++ {
		// 	v.Cursor.Right()
		// }

		// Remove the whitespaces if keepautoindent setting is off
		if IsSpacesOrTabs(v.Buf.Line(v.Cursor.Y-1)) && !v.Buf.Settings["keepautoindent"].(bool) {
			line := v.Buf.Line(v.Cursor.Y - 1)
			v.Buf.Remove(Loc{0, v.Cursor.Y - 1}, Loc{Count(line), v.Cursor.Y - 1})
		}
	}
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()

	if usePlugin {
		return PostActionCall("InsertNewline", v)
	}
	return true
}

// Backspace deletes the previous character
func (v *View) Backspace(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Backspace", v) {
		return false
	}

	// Delete a character
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else if v.Cursor.Loc.GreaterThan(v.Buf.Start()) {
		// We have to do something a bit hacky here because we want to
		// delete the line by first moving left and then deleting backwards
		// but the undo redo would place the cursor in the wrong place
		// So instead we move left, save the position, move back, delete
		// and restore the position

		// If the user is using spaces instead of tabs and they are deleting
		// whitespace at the start of the line, we should delete as if it's a
		// tab (tabSize number of spaces)
		lineStart := sliceEnd(v.Buf.LineBytes(v.Cursor.Y), v.Cursor.X)
		tabSize := int(v.Buf.Settings["tabsize"].(float64))
		if v.Buf.Settings["tabstospaces"].(bool) && IsSpaces(lineStart) && utf8.RuneCount(lineStart) != 0 && utf8.RuneCount(lineStart)%tabSize == 0 {
			loc := v.Cursor.Loc
			v.Buf.Remove(loc.Move(-tabSize, v.Buf), loc)
		} else {
			loc := v.Cursor.Loc
			v.Buf.Remove(loc.Move(-1, v.Buf), loc)
		}
	}
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()

	if usePlugin {
		return PostActionCall("Backspace", v)
	}
	return true
}

// DeleteWordRight deletes the word to the right of the cursor
func (v *View) DeleteWordRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteWordRight", v) {
		return false
	}

	v.SelectWordRight(false)
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	if usePlugin {
		return PostActionCall("DeleteWordRight", v)
	}
	return true
}

// DeleteWordLeft deletes the word to the left of the cursor
func (v *View) DeleteWordLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteWordLeft", v) {
		return false
	}

	v.SelectWordLeft(false)
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	if usePlugin {
		return PostActionCall("DeleteWordLeft", v)
	}
	return true
}

// Delete deletes the next character
func (v *View) Delete(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Delete", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else {
		loc := v.Cursor.Loc
		if loc.LessThan(v.Buf.End()) {
			v.Buf.Remove(loc, loc.Move(1, v.Buf))
		}
	}

	if usePlugin {
		return PostActionCall("Delete", v)
	}
	return true
}

// IndentSelection indents the current selection
func (v *View) IndentSelection(usePlugin bool) bool {
	if usePlugin && !PreActionCall("IndentSelection", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		start := v.Cursor.CurSelection[0]
		end := v.Cursor.CurSelection[1]
		if end.Y < start.Y {
			start, end = end, start
			v.Cursor.SetSelectionStart(start)
			v.Cursor.SetSelectionEnd(end)
		}

		startY := start.Y
		endY := end.Move(-1, v.Buf).Y
		endX := end.Move(-1, v.Buf).X
		tabsize := len(v.Buf.IndentString())
		for y := startY; y <= endY; y++ {
			v.Buf.Insert(Loc{0, y}, v.Buf.IndentString())
			if y == startY && start.X > 0 {
				v.Cursor.SetSelectionStart(start.Move(tabsize, v.Buf))
			}
			if y == endY {
				v.Cursor.SetSelectionEnd(Loc{endX + tabsize + 1, endY})
			}
		}
		v.Cursor.Relocate()

		if usePlugin {
			return PostActionCall("IndentSelection", v)
		}
		return true
	}
	return false
}

// OutdentLine moves the current line back one indentation
func (v *View) OutdentLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("OutdentLine", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		return false
	}

	for x := 0; x < len(v.Buf.IndentString()); x++ {
		if len(GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))) == 0 {
			break
		}
		v.Buf.Remove(Loc{0, v.Cursor.Y}, Loc{1, v.Cursor.Y})
	}
	v.Cursor.Relocate()

	if usePlugin {
		return PostActionCall("OutdentLine", v)
	}
	return true
}

// OutdentSelection takes the current selection and moves it back one indent level
func (v *View) OutdentSelection(usePlugin bool) bool {
	if usePlugin && !PreActionCall("OutdentSelection", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		start := v.Cursor.CurSelection[0]
		end := v.Cursor.CurSelection[1]
		if end.Y < start.Y {
			start, end = end, start
			v.Cursor.SetSelectionStart(start)
			v.Cursor.SetSelectionEnd(end)
		}

		startY := start.Y
		endY := end.Move(-1, v.Buf).Y
		for y := startY; y <= endY; y++ {
			for x := 0; x < len(v.Buf.IndentString()); x++ {
				if len(GetLeadingWhitespace(v.Buf.Line(y))) == 0 {
					break
				}
				v.Buf.Remove(Loc{0, y}, Loc{1, y})
			}
		}
		v.Cursor.Relocate()

		if usePlugin {
			return PostActionCall("OutdentSelection", v)
		}
		return true
	}
	return false
}

// InsertTab inserts a tab or spaces
func (v *View) InsertTab(usePlugin bool) bool {
	if usePlugin && !PreActionCall("InsertTab", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		return false
	}

	tabBytes := len(v.Buf.IndentString())
	bytesUntilIndent := tabBytes - (v.Cursor.GetVisualX() % tabBytes)
	v.Buf.Insert(v.Cursor.Loc, v.Buf.IndentString()[:bytesUntilIndent])
	// for i := 0; i < bytesUntilIndent; i++ {
	// 	v.Cursor.Right()
	// }

	if usePlugin {
		return PostActionCall("InsertTab", v)
	}
	return true
}

// SaveAll saves all open buffers
func (v *View) SaveAll(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("SaveAll", v) {
			return false
		}

		for _, t := range tabs {
			for _, v := range t.Views {
				v.Save(false)
			}
		}

		if usePlugin {
			return PostActionCall("SaveAll", v)
		}
	}
	return false
}

// Save the buffer to disk
func (v *View) Save(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("Save", v) {
			return false
		}

		if v.Type.Scratch == true {
			// We can't save any view type with scratch set. eg help and log text
			return false
		}
		_, err := os.Open(v.Buf.Path)
		fileInfo, _ := os.Stat(v.Buf.Path)

		// If this is an empty buffer, ask for a filename
		// We'll also ask for a filename if the path is a directory.
		if v.Buf.Path == "" {
			if !v.SaveAs(false) {
				return false
			}
		} else if err == nil && fileInfo.IsDir() {
			if !v.SaveAs(false) {
				return false
			}
		} else {
			v.saveToFile(v.Buf.Path)
		}

		if usePlugin {
			return PostActionCall("Save", v)
		}
	}
	return false
}

// This function saves the buffer to `filename` and changes the buffer's path and name
// to `filename` if the save is successful
func (v *View) saveToFile(filename string) {
	err := v.Buf.SaveAs(filename)

	if err != nil {
		if strings.HasSuffix(err.Error(), "permission denied") {
			choice, _ := messenger.YesNoPrompt("Permission denied. Do you want to save this file using sudo? (y,n)")
			if choice {
				err = v.Buf.SaveAsWithSudo(filename)
				if err != nil {
					messenger.Error(err.Error())
				} else {
					v.Buf.Path = filename
					v.Buf.name = filename
					messenger.Message("Saved " + filename)
				}
			}
			messenger.Reset()
			messenger.Clear()
		} else {
			messenger.Error(err.Error())
		}
	} else {
		v.Buf.Path = filename
		v.Buf.name = filename
		messenger.Message("Saved " + filename)
	}
}

// SaveAs saves the buffer to disk with the given name
func (v *View) SaveAs(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("SaveAs", v) {
			return false
		}
		// We'll refer to the path stored as part of the buffer.
		filename, canceled := messenger.Prompt("Filename: ", v.Buf.Path, "Save", NoCompletion)
		if !canceled {
			// the filename might or might not be quoted, so unquote first then join the strings.
			args, err := shellwords.Split(filename)
			filename = strings.Join(args, " ")
			if err != nil {
				messenger.Error("Error parsing arguments: ", err)
				return false
			}

			_, err = os.Open(filename)
			fileInfo, _ := os.Stat(filename)

			if err == nil {
				if fileInfo.IsDir() {
					messenger.Error(filename + " is a directory")
					return false
				} else {
					choice, _ := messenger.YesNoPrompt("Overwrite " + filename + "? (y,n)")
					messenger.Reset()
					messenger.Clear()
					if !choice {
						return false
					}
				}
			}
			v.saveToFile(filename)
		}

		if usePlugin {
			PostActionCall("SaveAs", v)
		}
	}
	return false
}

// Find opens a prompt and searches forward for the input
func (v *View) Find(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("Find", v) {
			return false
		}

		searchStr := ""
		if v.Cursor.HasSelection() {
			searchStart = v.Cursor.CurSelection[1]
			searchStart = v.Cursor.CurSelection[1]
			searchStr = v.Cursor.GetSelection()
		} else {
			searchStart = v.Cursor.Loc
		}
		BeginSearch(searchStr)

		if usePlugin {
			return PostActionCall("Find", v)
		}
	}
	return true
}

// FindNext searches forwards for the last used search term
func (v *View) FindNext(usePlugin bool) bool {
	if usePlugin && !PreActionCall("FindNext", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		searchStart = v.Cursor.CurSelection[1]
		// lastSearch = v.Cursor.GetSelection()
	} else {
		searchStart = v.Cursor.Loc
	}
	if lastSearch == "" {
		return true
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, true)

	if usePlugin {
		return PostActionCall("FindNext", v)
	}
	return true
}

// FindPrevious searches backwards for the last used search term
func (v *View) FindPrevious(usePlugin bool) bool {
	if usePlugin && !PreActionCall("FindPrevious", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		searchStart = v.Cursor.CurSelection[0]
	} else {
		searchStart = v.Cursor.Loc
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, false)

	if usePlugin {
		return PostActionCall("FindPrevious", v)
	}
	return true
}

// Undo undoes the last action
func (v *View) Undo(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Undo", v) {
		return false
	}

	if v.Buf.curCursor == 0 {
		v.Buf.clearCursors()
	}

	v.Buf.Undo()
	messenger.Message("Undid action")

	if usePlugin {
		return PostActionCall("Undo", v)
	}
	return true
}

// Redo redoes the last action
func (v *View) Redo(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Redo", v) {
		return false
	}

	if v.Buf.curCursor == 0 {
		v.Buf.clearCursors()
	}

	v.Buf.Redo()
	messenger.Message("Redid action")

	if usePlugin {
		return PostActionCall("Redo", v)
	}
	return true
}

// Copy the selection to the system clipboard
func (v *View) Copy(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("Copy", v) {
			return false
		}

		if v.Cursor.HasSelection() {
			v.Cursor.CopySelection("clipboard")
			v.freshClip = true
			messenger.Message("Copied selection")
		}

		if usePlugin {
			return PostActionCall("Copy", v)
		}
	}
	return true
}

// CutLine cuts the current line to the clipboard
func (v *View) CutLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CutLine", v) {
		return false
	}

	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	if v.freshClip == true {
		if v.Cursor.HasSelection() {
			if clip, err := clipboard.ReadAll("clipboard"); err != nil {
				messenger.Error(err)
			} else {
				clipboard.WriteAll(clip+v.Cursor.GetSelection(), "clipboard")
			}
		}
	} else if time.Since(v.lastCutTime)/time.Second > 10*time.Second || v.freshClip == false {
		v.Copy(true)
	}
	v.freshClip = true
	v.lastCutTime = time.Now()
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()
	messenger.Message("Cut line")

	if usePlugin {
		return PostActionCall("CutLine", v)
	}
	return true
}

// Cut the selection to the system clipboard
func (v *View) Cut(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Cut", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.CopySelection("clipboard")
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		v.freshClip = true
		messenger.Message("Cut selection")

		if usePlugin {
			return PostActionCall("Cut", v)
		}
		return true
	} else {
		return v.CutLine(usePlugin)
	}
}

// DuplicateLine duplicates the current line or selection
func (v *View) DuplicateLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DuplicateLine", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Buf.Insert(v.Cursor.CurSelection[1], v.Cursor.GetSelection())
	} else {
		v.Cursor.End()
		v.Buf.Insert(v.Cursor.Loc, "\n"+v.Buf.Line(v.Cursor.Y))
		// v.Cursor.Right()
	}

	messenger.Message("Duplicated line")

	if usePlugin {
		return PostActionCall("DuplicateLine", v)
	}
	return true
}

// DeleteLine deletes the current line
func (v *View) DeleteLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteLine", v) {
		return false
	}

	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()
	messenger.Message("Deleted line")

	if usePlugin {
		return PostActionCall("DeleteLine", v)
	}
	return true
}

// MoveLinesUp moves up the current line or selected lines if any
func (v *View) MoveLinesUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("MoveLinesUp", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		if v.Cursor.CurSelection[0].Y == 0 {
			messenger.Message("Can not move further up")
			return true
		}
		start := v.Cursor.CurSelection[0].Y
		end := v.Cursor.CurSelection[1].Y
		if start > end {
			end, start = start, end
		}

		v.Buf.MoveLinesUp(
			start,
			end,
		)
		v.Cursor.CurSelection[1].Y -= 1
		messenger.Message("Moved up selected line(s)")
	} else {
		if v.Cursor.Loc.Y == 0 {
			messenger.Message("Can not move further up")
			return true
		}
		v.Buf.MoveLinesUp(
			v.Cursor.Loc.Y,
			v.Cursor.Loc.Y+1,
		)
		messenger.Message("Moved up current line")
	}
	v.Buf.IsModified = true

	if usePlugin {
		return PostActionCall("MoveLinesUp", v)
	}
	return true
}

// MoveLinesDown moves down the current line or selected lines if any
func (v *View) MoveLinesDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("MoveLinesDown", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		if v.Cursor.CurSelection[1].Y >= len(v.Buf.lines) {
			messenger.Message("Can not move further down")
			return true
		}
		start := v.Cursor.CurSelection[0].Y
		end := v.Cursor.CurSelection[1].Y
		if start > end {
			end, start = start, end
		}

		v.Buf.MoveLinesDown(
			start,
			end,
		)
		messenger.Message("Moved down selected line(s)")
	} else {
		if v.Cursor.Loc.Y >= len(v.Buf.lines)-1 {
			messenger.Message("Can not move further down")
			return true
		}
		v.Buf.MoveLinesDown(
			v.Cursor.Loc.Y,
			v.Cursor.Loc.Y+1,
		)
		messenger.Message("Moved down current line")
	}
	v.Buf.IsModified = true

	if usePlugin {
		return PostActionCall("MoveLinesDown", v)
	}
	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (v *View) Paste(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Paste", v) {
		return false
	}

	clip, _ := clipboard.ReadAll("clipboard")
	v.paste(clip)

	if usePlugin {
		return PostActionCall("Paste", v)
	}
	return true
}

// PastePrimary pastes from the primary clipboard (only use on linux)
func (v *View) PastePrimary(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Paste", v) {
		return false
	}

	clip, _ := clipboard.ReadAll("primary")
	v.paste(clip)

	if usePlugin {
		return PostActionCall("Paste", v)
	}
	return true
}

// JumpToMatchingBrace moves the cursor to the matching brace if it is
// currently on a brace
func (v *View) JumpToMatchingBrace(usePlugin bool) bool {
	if usePlugin && !PreActionCall("JumpToMatchingBrace", v) {
		return false
	}

	for _, bp := range bracePairs {
		r := v.Cursor.RuneUnder(v.Cursor.X)
		if r == bp[0] || r == bp[1] {
			matchingBrace := v.Buf.FindMatchingBrace(bp, v.Cursor.Loc)
			v.Cursor.GotoLoc(matchingBrace)
		}
	}

	if usePlugin {
		return PostActionCall("JumpToMatchingBrace", v)
	}
	return true
}

// SelectAll selects the entire buffer
func (v *View) SelectAll(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectAll", v) {
		return false
	}

	v.Cursor.SetSelectionStart(v.Buf.Start())
	v.Cursor.SetSelectionEnd(v.Buf.End())
	// Put the cursor at the beginning
	v.Cursor.X = 0
	v.Cursor.Y = 0

	if usePlugin {
		return PostActionCall("SelectAll", v)
	}
	return true
}

// OpenFile opens a new file in the buffer
func (v *View) OpenFile(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("OpenFile", v) {
			return false
		}

		if v.CanClose() {
			input, canceled := messenger.Prompt("> ", "open ", "Open", CommandCompletion)
			if !canceled {
				HandleCommand(input)
				if usePlugin {
					return PostActionCall("OpenFile", v)
				}
			}
		}
	}
	return false
}

// Start moves the viewport to the start of the buffer
func (v *View) Start(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("Start", v) {
			return false
		}

		v.Topline = 0

		if usePlugin {
			return PostActionCall("Start", v)
		}
	}
	return false
}

// End moves the viewport to the end of the buffer
func (v *View) End(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("End", v) {
			return false
		}

		if v.Height > v.Buf.NumLines {
			v.Topline = 0
		} else {
			v.Topline = v.Buf.NumLines - v.Height
		}

		if usePlugin {
			return PostActionCall("End", v)
		}
	}
	return false
}

// PageUp scrolls the view up a page
func (v *View) PageUp(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("PageUp", v) {
			return false
		}

		if v.Topline > v.Height {
			v.ScrollUp(v.Height)
		} else {
			v.Topline = 0
		}

		if usePlugin {
			return PostActionCall("PageUp", v)
		}
	}
	return false
}

// PageDown scrolls the view down a page
func (v *View) PageDown(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("PageDown", v) {
			return false
		}

		if v.Buf.NumLines-(v.Topline+v.Height) > v.Height {
			v.ScrollDown(v.Height)
		} else if v.Buf.NumLines >= v.Height {
			v.Topline = v.Buf.NumLines - v.Height
		}

		if usePlugin {
			return PostActionCall("PageDown", v)
		}
	}
	return false
}

// SelectPageUp selects up one page
func (v *View) SelectPageUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectPageUp", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.UpN(v.Height)
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectPageUp", v)
	}
	return true
}

// SelectPageDown selects down one page
func (v *View) SelectPageDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectPageDown", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.DownN(v.Height)
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectPageDown", v)
	}
	return true
}

// CursorPageUp places the cursor a page up
func (v *View) CursorPageUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorPageUp", v) {
		return false
	}

	v.deselect(0)

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
		v.Cursor.StoreVisualX()
	}
	v.Cursor.UpN(v.Height)

	if usePlugin {
		return PostActionCall("CursorPageUp", v)
	}
	return true
}

// CursorPageDown places the cursor a page up
func (v *View) CursorPageDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorPageDown", v) {
		return false
	}

	v.deselect(0)

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1]
		v.Cursor.ResetSelection()
		v.Cursor.StoreVisualX()
	}
	v.Cursor.DownN(v.Height)

	if usePlugin {
		return PostActionCall("CursorPageDown", v)
	}
	return true
}

// HalfPageUp scrolls the view up half a page
func (v *View) HalfPageUp(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("HalfPageUp", v) {
			return false
		}

		if v.Topline > v.Height/2 {
			v.ScrollUp(v.Height / 2)
		} else {
			v.Topline = 0
		}

		if usePlugin {
			return PostActionCall("HalfPageUp", v)
		}
	}
	return false
}

// HalfPageDown scrolls the view down half a page
func (v *View) HalfPageDown(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("HalfPageDown", v) {
			return false
		}

		if v.Buf.NumLines-(v.Topline+v.Height) > v.Height/2 {
			v.ScrollDown(v.Height / 2)
		} else {
			if v.Buf.NumLines >= v.Height {
				v.Topline = v.Buf.NumLines - v.Height
			}
		}

		if usePlugin {
			return PostActionCall("HalfPageDown", v)
		}
	}
	return false
}

// ToggleRuler turns line numbers off and on
func (v *View) ToggleRuler(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ToggleRuler", v) {
			return false
		}

		if v.Buf.Settings["ruler"] == false {
			v.Buf.Settings["ruler"] = true
			messenger.Message("Enabled ruler")
		} else {
			v.Buf.Settings["ruler"] = false
			messenger.Message("Disabled ruler")
		}

		if usePlugin {
			return PostActionCall("ToggleRuler", v)
		}
	}
	return false
}

// JumpLine jumps to a line and moves the view accordingly.
func (v *View) JumpLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("JumpLine", v) {
		return false
	}

	// Prompt for line number
	message := fmt.Sprintf("Jump to line:col (1 - %v) # ", v.Buf.NumLines)
	input, canceled := messenger.Prompt(message, "", "LineNumber", NoCompletion)
	if canceled {
		return false
	}
	var lineInt int
	var colInt int
	var err error
	if strings.Contains(input, ":") {
		split := strings.Split(input, ":")
		lineInt, err = strconv.Atoi(split[0])
		if err != nil {
			messenger.Message("Invalid line number")
			return false
		}
		colInt, err = strconv.Atoi(split[1])
		if err != nil {
			messenger.Message("Invalid column number")
			return false
		}
	} else {
		lineInt, err = strconv.Atoi(input)
		if err != nil {
			messenger.Message("Invalid line number")
			return false
		}
	}
	lineInt--
	// Move cursor and view if possible.
	if lineInt < v.Buf.NumLines && lineInt >= 0 {
		v.Cursor.X = colInt
		v.Cursor.Y = lineInt

		if usePlugin {
			return PostActionCall("JumpLine", v)
		}
		return true
	}
	messenger.Error("Only ", v.Buf.NumLines, " lines to jump")
	return false
}

// ClearStatus clears the messenger bar
func (v *View) ClearStatus(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ClearStatus", v) {
			return false
		}

		messenger.Message("")

		if usePlugin {
			return PostActionCall("ClearStatus", v)
		}
	}
	return false
}

// ToggleHelp toggles the help screen
func (v *View) ToggleHelp(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ToggleHelp", v) {
			return false
		}

		if v.Type != vtHelp {
			// Open the default help
			v.openHelp("help")
		} else {
			v.Quit(true)
		}

		if usePlugin {
			return PostActionCall("ToggleHelp", v)
		}
	}
	return true
}

// ToggleKeyMenu toggles the keymenu option and resizes all tabs
func (v *View) ToggleKeyMenu(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ToggleBindings", v) {
			return false
		}

		globalSettings["keymenu"] = !globalSettings["keymenu"].(bool)
		for _, tab := range tabs {
			tab.Resize()
		}

		if usePlugin {
			return PostActionCall("ToggleBindings", v)
		}
	}
	return true
}

// ShellMode opens a terminal to run a shell command
func (v *View) ShellMode(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ShellMode", v) {
			return false
		}

		input, canceled := messenger.Prompt("$ ", "", "Shell", NoCompletion)
		if !canceled {
			// The true here is for openTerm to make the command interactive
			HandleShellCommand(input, true, true)
			if usePlugin {
				return PostActionCall("ShellMode", v)
			}
		}
	}
	return false
}

// CommandMode lets the user enter a command
func (v *View) CommandMode(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("CommandMode", v) {
			return false
		}

		input, canceled := messenger.Prompt("> ", "", "Command", CommandCompletion)
		if !canceled {
			HandleCommand(input)
			if usePlugin {
				return PostActionCall("CommandMode", v)
			}
		}
	}

	return false
}

// ToggleOverwriteMode lets the user toggle the text overwrite mode
func (v *View) ToggleOverwriteMode(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ToggleOverwriteMode", v) {
			return false
		}

		v.isOverwriteMode = !v.isOverwriteMode

		if usePlugin {
			return PostActionCall("ToggleOverwriteMode", v)
		}
	}
	return false
}

// Escape leaves current mode
func (v *View) Escape(usePlugin bool) bool {
	if v.mainCursor() {
		// check if user is searching, or the last search is still active
		if searching || lastSearch != "" {
			ExitSearch(v)
			return true
		}
		// check if a prompt is shown, hide it and don't quit
		if messenger.hasPrompt {
			messenger.Reset() // FIXME
			return true
		}
	}

	return false
}

// Quit this will close the current tab or view that is open
func (v *View) Quit(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("Quit", v) {
			return false
		}

		// Make sure not to quit if there are unsaved changes
		if v.CanClose() {
			v.CloseBuffer()
			if len(tabs[curTab].Views) > 1 {
				v.splitNode.Delete()
				tabs[v.TabNum].Cleanup()
				tabs[v.TabNum].Resize()
			} else if len(tabs) > 1 {
				if len(tabs[v.TabNum].Views) == 1 {
					tabs = tabs[:v.TabNum+copy(tabs[v.TabNum:], tabs[v.TabNum+1:])]
					for i, t := range tabs {
						t.SetNum(i)
					}
					if curTab >= len(tabs) {
						curTab--
					}
					if curTab == 0 {
						CurView().ToggleTabbar()
					}
				}
			} else {
				if usePlugin {
					PostActionCall("Quit", v)
				}

				screen.Fini()
				messenger.SaveHistory()
				os.Exit(0)
			}
		}

		if usePlugin {
			return PostActionCall("Quit", v)
		}
	}
	return false
}

// QuitAll quits the whole editor; all splits and tabs
func (v *View) QuitAll(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("QuitAll", v) {
			return false
		}

		closeAll := true
		for _, tab := range tabs {
			for _, v := range tab.Views {
				if !v.CanClose() {
					closeAll = false
				}
			}
		}

		if closeAll {
			// only quit if all of the buffers can be closed and the user confirms that they actually want to quit everything
			shouldQuit, _ := messenger.YesNoPrompt("Do you want to quit micro (all open files will be closed)?")

			if shouldQuit {
				for _, tab := range tabs {
					for _, v := range tab.Views {
						v.CloseBuffer()
					}
				}

				if usePlugin {
					PostActionCall("QuitAll", v)
				}

				screen.Fini()
				messenger.SaveHistory()
				os.Exit(0)
			}
		}
	}

	return false
}

// AddTab adds a new tab with an empty buffer
func (v *View) AddTab(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("AddTab", v) {
			return false
		}

		tab := NewTabFromView(NewView(NewBufferFromString("", "")))
		tab.SetNum(len(tabs))
		tabs = append(tabs, tab)
		curTab = len(tabs) - 1
		if len(tabs) == 2 {
			for _, t := range tabs {
				for _, v := range t.Views {
					v.ToggleTabbar()
				}
			}
		}

		if usePlugin {
			return PostActionCall("AddTab", v)
		}
	}
	return true
}

// PreviousTab switches to the previous tab in the tab list
func (v *View) PreviousTab(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("PreviousTab", v) {
			return false
		}

		if curTab > 0 {
			curTab--
		} else if curTab == 0 {
			curTab = len(tabs) - 1
		}

		if usePlugin {
			return PostActionCall("PreviousTab", v)
		}
	}
	return false
}

// NextTab switches to the next tab in the tab list
func (v *View) NextTab(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("NextTab", v) {
			return false
		}

		if curTab < len(tabs)-1 {
			curTab++
		} else if curTab == len(tabs)-1 {
			curTab = 0
		}

		if usePlugin {
			return PostActionCall("NextTab", v)
		}
	}
	return false
}

// VSplitBinding opens an empty vertical split
func (v *View) VSplitBinding(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("VSplit", v) {
			return false
		}

		v.VSplit(NewBufferFromString("", ""))

		if usePlugin {
			return PostActionCall("VSplit", v)
		}
	}
	return false
}

// HSplitBinding opens an empty horizontal split
func (v *View) HSplitBinding(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("HSplit", v) {
			return false
		}

		v.HSplit(NewBufferFromString("", ""))

		if usePlugin {
			return PostActionCall("HSplit", v)
		}
	}
	return false
}

// Unsplit closes all splits in the current tab except the active one
func (v *View) Unsplit(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("Unsplit", v) {
			return false
		}

		curView := tabs[curTab].CurView
		for i := len(tabs[curTab].Views) - 1; i >= 0; i-- {
			view := tabs[curTab].Views[i]
			if view != nil && view.Num != curView {
				view.Quit(true)
				// messenger.Message("Quit ", view.Buf.Path)
			}
		}

		if usePlugin {
			return PostActionCall("Unsplit", v)
		}
	}
	return false
}

// NextSplit changes the view to the next split
func (v *View) NextSplit(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("NextSplit", v) {
			return false
		}

		tab := tabs[curTab]
		if tab.CurView < len(tab.Views)-1 {
			tab.CurView++
		} else {
			tab.CurView = 0
		}

		if usePlugin {
			return PostActionCall("NextSplit", v)
		}
	}
	return false
}

// PreviousSplit changes the view to the previous split
func (v *View) PreviousSplit(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("PreviousSplit", v) {
			return false
		}

		tab := tabs[curTab]
		if tab.CurView > 0 {
			tab.CurView--
		} else {
			tab.CurView = len(tab.Views) - 1
		}

		if usePlugin {
			return PostActionCall("PreviousSplit", v)
		}
	}
	return false
}

var curMacro []interface{}
var recordingMacro bool

// ToggleMacro toggles recording of a macro
func (v *View) ToggleMacro(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ToggleMacro", v) {
			return false
		}

		recordingMacro = !recordingMacro

		if recordingMacro {
			curMacro = []interface{}{}
			messenger.Message("Recording")
		} else {
			messenger.Message("Stopped recording")
		}

		if usePlugin {
			return PostActionCall("ToggleMacro", v)
		}
	}
	return true
}

// PlayMacro plays back the most recently recorded macro
func (v *View) PlayMacro(usePlugin bool) bool {
	if usePlugin && !PreActionCall("PlayMacro", v) {
		return false
	}

	for _, action := range curMacro {
		switch t := action.(type) {
		case rune:
			// Insert a character
			if v.Cursor.HasSelection() {
				v.Cursor.DeleteSelection()
				v.Cursor.ResetSelection()
			}
			v.Buf.Insert(v.Cursor.Loc, string(t))
			// v.Cursor.Right()

			for pl := range loadedPlugins {
				_, err := Call(pl+".onRune", string(t), v)
				if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
					TermMessage(err)
				}
			}
		case func(*View, bool) bool:
			t(v, true)
		}
	}

	if usePlugin {
		return PostActionCall("PlayMacro", v)
	}
	return true
}

// SpawnMultiCursor creates a new multiple cursor at the next occurrence of the current selection or current word
func (v *View) SpawnMultiCursor(usePlugin bool) bool {
	spawner := v.Buf.cursors[len(v.Buf.cursors)-1]
	// You can only spawn a cursor from the main cursor
	if v.Cursor == spawner {
		if usePlugin && !PreActionCall("SpawnMultiCursor", v) {
			return false
		}

		if !spawner.HasSelection() {
			spawner.SelectWord()
		} else {
			c := &Cursor{
				buf: v.Buf,
			}

			sel := spawner.GetSelection()

			searchStart = spawner.CurSelection[1]
			v.Cursor = c
			Search(regexp.QuoteMeta(sel), v, true)

			for _, cur := range v.Buf.cursors {
				if c.Loc == cur.Loc {
					return false
				}
			}
			v.Buf.cursors = append(v.Buf.cursors, c)
			v.Buf.UpdateCursors()
			v.Relocate()
			v.Cursor = spawner
		}

		if usePlugin {
			PostActionCall("SpawnMultiCursor", v)
		}
	}

	return false
}

// SpawnMultiCursorSelect adds a cursor at the beginning of each line of a selection
func (v *View) SpawnMultiCursorSelect(usePlugin bool) bool {
	if v.Cursor == &v.Buf.Cursor {
		if usePlugin && !PreActionCall("SpawnMultiCursorSelect", v) {
			return false
		}

		// Avoid cases where multiple cursors already exist, that would create problems
		if len(v.Buf.cursors) > 1 {
			return false
		}

		var startLine int
		var endLine int

		a, b := v.Cursor.CurSelection[0].Y, v.Cursor.CurSelection[1].Y
		if a > b {
			startLine, endLine = b, a
		} else {
			startLine, endLine = a, b
		}

		if v.Cursor.HasSelection() {
			v.Cursor.ResetSelection()
			v.Cursor.GotoLoc(Loc{0, startLine})

			for i := startLine; i <= endLine; i++ {
				c := &Cursor{
					buf: v.Buf,
				}
				c.GotoLoc(Loc{0, i})
				v.Buf.cursors = append(v.Buf.cursors, c)
			}
			v.Buf.MergeCursors()
			v.Buf.UpdateCursors()
		} else {
			return false
		}

		if usePlugin {
			PostActionCall("SpawnMultiCursorSelect", v)
		}

		messenger.Message("Added cursors from selection")
	}
	return false
}

// MouseMultiCursor is a mouse action which puts a new cursor at the mouse position
func (v *View) MouseMultiCursor(usePlugin bool, e *tcell.EventMouse) bool {
	if v.Cursor == &v.Buf.Cursor {
		if usePlugin && !PreActionCall("SpawnMultiCursorAtMouse", v, e) {
			return false
		}
		x, y := e.Position()
		x -= v.lineNumOffset - v.leftCol + v.x
		y += v.Topline - v.y

		c := &Cursor{
			buf: v.Buf,
		}
		v.Cursor = c
		v.MoveToMouseClick(x, y)
		v.Relocate()
		v.Cursor = &v.Buf.Cursor

		v.Buf.cursors = append(v.Buf.cursors, c)
		v.Buf.MergeCursors()
		v.Buf.UpdateCursors()

		if usePlugin {
			PostActionCall("SpawnMultiCursorAtMouse", v)
		}
	}
	return false
}

// SkipMultiCursor moves the current multiple cursor to the next available position
func (v *View) SkipMultiCursor(usePlugin bool) bool {
	cursor := v.Buf.cursors[len(v.Buf.cursors)-1]

	if v.mainCursor() {
		if usePlugin && !PreActionCall("SkipMultiCursor", v) {
			return false
		}
		sel := cursor.GetSelection()

		searchStart = cursor.CurSelection[1]
		v.Cursor = cursor
		Search(regexp.QuoteMeta(sel), v, true)
		v.Relocate()
		v.Cursor = cursor

		if usePlugin {
			PostActionCall("SkipMultiCursor", v)
		}
	}
	return false
}

// RemoveMultiCursor removes the latest multiple cursor
func (v *View) RemoveMultiCursor(usePlugin bool) bool {
	end := len(v.Buf.cursors)
	if end > 1 {
		if v.mainCursor() {
			if usePlugin && !PreActionCall("RemoveMultiCursor", v) {
				return false
			}

			v.Buf.cursors[end-1] = nil
			v.Buf.cursors = v.Buf.cursors[:end-1]
			v.Buf.UpdateCursors()
			v.Relocate()

			if usePlugin {
				return PostActionCall("RemoveMultiCursor", v)
			}
			return true
		}
	} else {
		v.RemoveAllMultiCursors(usePlugin)
	}
	return false
}

// RemoveAllMultiCursors removes all cursors except the base cursor
func (v *View) RemoveAllMultiCursors(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("RemoveAllMultiCursors", v) {
			return false
		}

		v.Buf.clearCursors()
		v.Relocate()

		if usePlugin {
			return PostActionCall("RemoveAllMultiCursors", v)
		}
		return true
	}
	return false
}
