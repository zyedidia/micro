package action

import (
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zyedidia/clipboard"
	"github.com/zyedidia/micro/internal/buffer"
	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/micro/internal/screen"
	"github.com/zyedidia/micro/internal/shell"
	"github.com/zyedidia/micro/internal/util"
	"github.com/zyedidia/micro/pkg/shellwords"
	"github.com/zyedidia/tcell"
)

// ScrollUp is not an action
func (h *BufPane) ScrollUp(n int) {
	v := h.GetView()
	if v.StartLine >= n {
		v.StartLine -= n
		h.SetView(v)
	}
}

// ScrollDown is not an action
func (h *BufPane) ScrollDown(n int) {
	v := h.GetView()
	if v.StartLine <= h.Buf.LinesNum()-1-n {
		v.StartLine += n
		h.SetView(v)
	}
}

// MousePress is the event that should happen when a normal click happens
// This is almost always bound to left click
func (h *BufPane) MousePress(e *tcell.EventMouse) bool {
	b := h.Buf
	mx, my := e.Position()
	mouseLoc := h.GetMouseLoc(buffer.Loc{mx, my})
	h.Cursor.Loc = mouseLoc
	if h.mouseReleased {
		if b.NumCursors() > 1 {
			b.ClearCursors()
			h.Relocate()
		}
		if time.Since(h.lastClickTime)/time.Millisecond < config.DoubleClickThreshold && (mouseLoc.X == h.lastLoc.X && mouseLoc.Y == h.lastLoc.Y) {
			if h.doubleClick {
				// Triple click
				h.lastClickTime = time.Now()

				h.tripleClick = true
				h.doubleClick = false

				h.Cursor.SelectLine()
				h.Cursor.CopySelection("primary")
			} else {
				// Double click
				h.lastClickTime = time.Now()

				h.doubleClick = true
				h.tripleClick = false

				h.Cursor.SelectWord()
				h.Cursor.CopySelection("primary")
			}
		} else {
			h.doubleClick = false
			h.tripleClick = false
			h.lastClickTime = time.Now()

			h.Cursor.OrigSelection[0] = h.Cursor.Loc
			h.Cursor.CurSelection[0] = h.Cursor.Loc
			h.Cursor.CurSelection[1] = h.Cursor.Loc
		}
		h.mouseReleased = false
	} else if !h.mouseReleased {
		if h.tripleClick {
			h.Cursor.AddLineToSelection()
		} else if h.doubleClick {
			h.Cursor.AddWordToSelection()
		} else {
			h.Cursor.SetSelectionEnd(h.Cursor.Loc)
			h.Cursor.CopySelection("primary")
		}
	}

	h.lastLoc = mouseLoc
	return false
}

// ScrollUpAction scrolls the view up
func (h *BufPane) ScrollUpAction() bool {
	h.ScrollUp(util.IntOpt(h.Buf.Settings["scrollspeed"]))
	return false
}

// ScrollDownAction scrolls the view up
func (h *BufPane) ScrollDownAction() bool {
	h.ScrollDown(util.IntOpt(h.Buf.Settings["scrollspeed"]))
	return false
}

// Center centers the view on the cursor
func (h *BufPane) Center() bool {
	v := h.GetView()
	v.StartLine = h.Cursor.Y - v.Height/2
	if v.StartLine+v.Height > h.Buf.LinesNum() {
		v.StartLine = h.Buf.LinesNum() - v.Height
	}
	if v.StartLine < 0 {
		v.StartLine = 0
	}
	h.SetView(v)
	return true
}

// CursorUp moves the cursor up
func (h *BufPane) CursorUp() bool {
	h.Cursor.Deselect(true)
	h.Cursor.Up()
	return true
}

// CursorDown moves the cursor down
func (h *BufPane) CursorDown() bool {
	h.Cursor.Deselect(true)
	h.Cursor.Down()
	return true
}

// CursorLeft moves the cursor left
func (h *BufPane) CursorLeft() bool {
	if h.Cursor.HasSelection() {
		h.Cursor.Deselect(true)
	} else {
		tabstospaces := h.Buf.Settings["tabstospaces"].(bool)
		tabmovement := h.Buf.Settings["tabmovement"].(bool)
		if tabstospaces && tabmovement {
			tabsize := int(h.Buf.Settings["tabsize"].(float64))
			line := h.Buf.LineBytes(h.Cursor.Y)
			if h.Cursor.X-tabsize >= 0 && util.IsSpaces(line[h.Cursor.X-tabsize:h.Cursor.X]) && util.IsBytesWhitespace(line[0:h.Cursor.X-tabsize]) {
				for i := 0; i < tabsize; i++ {
					h.Cursor.Left()
				}
			} else {
				h.Cursor.Left()
			}
		} else {
			h.Cursor.Left()
		}
	}
	return true
}

// CursorRight moves the cursor right
func (h *BufPane) CursorRight() bool {
	if h.Cursor.HasSelection() {
		h.Cursor.Deselect(false)
		h.Cursor.Loc = h.Cursor.Loc.Move(1, h.Buf)
	} else {
		tabstospaces := h.Buf.Settings["tabstospaces"].(bool)
		tabmovement := h.Buf.Settings["tabmovement"].(bool)
		if tabstospaces && tabmovement {
			tabsize := int(h.Buf.Settings["tabsize"].(float64))
			line := h.Buf.LineBytes(h.Cursor.Y)
			if h.Cursor.X+tabsize < utf8.RuneCount(line) && util.IsSpaces(line[h.Cursor.X:h.Cursor.X+tabsize]) && util.IsBytesWhitespace(line[0:h.Cursor.X]) {
				for i := 0; i < tabsize; i++ {
					h.Cursor.Right()
				}
			} else {
				h.Cursor.Right()
			}
		} else {
			h.Cursor.Right()
		}
	}

	return true
}

// WordRight moves the cursor one word to the right
func (h *BufPane) WordRight() bool {
	h.Cursor.Deselect(false)
	h.Cursor.WordRight()
	return true
}

// WordLeft moves the cursor one word to the left
func (h *BufPane) WordLeft() bool {
	h.Cursor.Deselect(true)
	h.Cursor.WordLeft()
	return true
}

// SelectUp selects up one line
func (h *BufPane) SelectUp() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.Up()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectDown selects down one line
func (h *BufPane) SelectDown() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.Down()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectLeft selects the character to the left of the cursor
func (h *BufPane) SelectLeft() bool {
	loc := h.Cursor.Loc
	count := h.Buf.End()
	if loc.GreaterThan(count) {
		loc = count
	}
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = loc
	}
	h.Cursor.Left()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectRight selects the character to the right of the cursor
func (h *BufPane) SelectRight() bool {
	loc := h.Cursor.Loc
	count := h.Buf.End()
	if loc.GreaterThan(count) {
		loc = count
	}
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = loc
	}
	h.Cursor.Right()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectWordRight selects the word to the right of the cursor
func (h *BufPane) SelectWordRight() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.WordRight()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectWordLeft selects the word to the left of the cursor
func (h *BufPane) SelectWordLeft() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.WordLeft()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// StartOfLine moves the cursor to the start of the line
func (h *BufPane) StartOfLine() bool {
	h.Cursor.Deselect(true)
	if h.Cursor.X != 0 {
		h.Cursor.Start()
	} else {
		h.Cursor.StartOfText()
	}
	return true
}

// EndOfLine moves the cursor to the end of the line
func (h *BufPane) EndOfLine() bool {
	h.Cursor.Deselect(true)
	h.Cursor.End()
	return true
}

// SelectLine selects the entire current line
func (h *BufPane) SelectLine() bool {
	h.Cursor.SelectLine()
	return true
}

// SelectToStartOfLine selects to the start of the current line
func (h *BufPane) SelectToStartOfLine() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.Start()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectToEndOfLine selects to the end of the current line
func (h *BufPane) SelectToEndOfLine() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.End()
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// ParagraphPrevious moves the cursor to the previous empty line, or beginning of the buffer if there's none
func (h *BufPane) ParagraphPrevious() bool {
	var line int
	for line = h.Cursor.Y; line > 0; line-- {
		if len(h.Buf.LineBytes(line)) == 0 && line != h.Cursor.Y {
			h.Cursor.X = 0
			h.Cursor.Y = line
			break
		}
	}
	// If no empty line found. move cursor to end of buffer
	if line == 0 {
		h.Cursor.Loc = h.Buf.Start()
	}
	return true
}

// ParagraphNext moves the cursor to the next empty line, or end of the buffer if there's none
func (h *BufPane) ParagraphNext() bool {
	var line int
	for line = h.Cursor.Y; line < h.Buf.LinesNum(); line++ {
		if len(h.Buf.LineBytes(line)) == 0 && line != h.Cursor.Y {
			h.Cursor.X = 0
			h.Cursor.Y = line
			break
		}
	}
	// If no empty line found. move cursor to end of buffer
	if line == h.Buf.LinesNum() {
		h.Cursor.Loc = h.Buf.End()
	}
	return true
}

// Retab changes all tabs to spaces or all spaces to tabs depending
// on the user's settings
func (h *BufPane) Retab() bool {
	h.Buf.Retab()
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (h *BufPane) CursorStart() bool {
	h.Cursor.Deselect(true)
	h.Cursor.X = 0
	h.Cursor.Y = 0
	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (h *BufPane) CursorEnd() bool {
	h.Cursor.Deselect(true)
	h.Cursor.Loc = h.Buf.End()
	h.Cursor.StoreVisualX()
	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (h *BufPane) SelectToStart() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.CursorStart()
	h.Cursor.SelectTo(h.Buf.Start())
	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (h *BufPane) SelectToEnd() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.CursorEnd()
	h.Cursor.SelectTo(h.Buf.End())
	return true
}

// InsertNewline inserts a newline plus possible some whitespace if autoindent is on
func (h *BufPane) InsertNewline() bool {
	// Insert a newline
	if h.Cursor.HasSelection() {
		h.Cursor.DeleteSelection()
		h.Cursor.ResetSelection()
	}

	ws := util.GetLeadingWhitespace(h.Buf.LineBytes(h.Cursor.Y))
	cx := h.Cursor.X
	h.Buf.Insert(h.Cursor.Loc, "\n")
	// h.Cursor.Right()

	if h.Buf.Settings["autoindent"].(bool) {
		if cx < len(ws) {
			ws = ws[0:cx]
		}
		h.Buf.Insert(h.Cursor.Loc, string(ws))
		// for i := 0; i < len(ws); i++ {
		// 	h.Cursor.Right()
		// }

		// Remove the whitespaces if keepautoindent setting is off
		if util.IsSpacesOrTabs(h.Buf.LineBytes(h.Cursor.Y-1)) && !h.Buf.Settings["keepautoindent"].(bool) {
			line := h.Buf.LineBytes(h.Cursor.Y - 1)
			h.Buf.Remove(buffer.Loc{X: 0, Y: h.Cursor.Y - 1}, buffer.Loc{X: utf8.RuneCount(line), Y: h.Cursor.Y - 1})
		}
	}
	h.Cursor.LastVisualX = h.Cursor.GetVisualX()
	return true
}

// Backspace deletes the previous character
func (h *BufPane) Backspace() bool {
	if h.Cursor.HasSelection() {
		h.Cursor.DeleteSelection()
		h.Cursor.ResetSelection()
	} else if h.Cursor.Loc.GreaterThan(h.Buf.Start()) {
		// We have to do something a bit hacky here because we want to
		// delete the line by first moving left and then deleting backwards
		// but the undo redo would place the cursor in the wrong place
		// So instead we move left, save the position, move back, delete
		// and restore the position

		// If the user is using spaces instead of tabs and they are deleting
		// whitespace at the start of the line, we should delete as if it's a
		// tab (tabSize number of spaces)
		lineStart := util.SliceStart(h.Buf.LineBytes(h.Cursor.Y), h.Cursor.X)
		tabSize := int(h.Buf.Settings["tabsize"].(float64))
		if h.Buf.Settings["tabstospaces"].(bool) && util.IsSpaces(lineStart) && len(lineStart) != 0 && utf8.RuneCount(lineStart)%tabSize == 0 {
			loc := h.Cursor.Loc
			h.Buf.Remove(loc.Move(-tabSize, h.Buf), loc)
		} else {
			loc := h.Cursor.Loc
			h.Buf.Remove(loc.Move(-1, h.Buf), loc)
		}
	}
	h.Cursor.LastVisualX = h.Cursor.GetVisualX()
	return true
}

// DeleteWordRight deletes the word to the right of the cursor
func (h *BufPane) DeleteWordRight() bool {
	h.SelectWordRight()
	if h.Cursor.HasSelection() {
		h.Cursor.DeleteSelection()
		h.Cursor.ResetSelection()
	}
	return true
}

// DeleteWordLeft deletes the word to the left of the cursor
func (h *BufPane) DeleteWordLeft() bool {
	h.SelectWordLeft()
	if h.Cursor.HasSelection() {
		h.Cursor.DeleteSelection()
		h.Cursor.ResetSelection()
	}
	return true
}

// Delete deletes the next character
func (h *BufPane) Delete() bool {
	if h.Cursor.HasSelection() {
		h.Cursor.DeleteSelection()
		h.Cursor.ResetSelection()
	} else {
		loc := h.Cursor.Loc
		if loc.LessThan(h.Buf.End()) {
			h.Buf.Remove(loc, loc.Move(1, h.Buf))
		}
	}
	return true
}

// IndentSelection indents the current selection
func (h *BufPane) IndentSelection() bool {
	if h.Cursor.HasSelection() {
		start := h.Cursor.CurSelection[0]
		end := h.Cursor.CurSelection[1]
		if end.Y < start.Y {
			start, end = end, start
			h.Cursor.SetSelectionStart(start)
			h.Cursor.SetSelectionEnd(end)
		}

		startY := start.Y
		endY := end.Move(-1, h.Buf).Y
		endX := end.Move(-1, h.Buf).X
		tabsize := int(h.Buf.Settings["tabsize"].(float64))
		indentsize := len(h.Buf.IndentString(tabsize))
		for y := startY; y <= endY; y++ {
			h.Buf.Insert(buffer.Loc{X: 0, Y: y}, h.Buf.IndentString(tabsize))
			if y == startY && start.X > 0 {
				h.Cursor.SetSelectionStart(start.Move(indentsize, h.Buf))
			}
			if y == endY {
				h.Cursor.SetSelectionEnd(buffer.Loc{X: endX + indentsize + 1, Y: endY})
			}
		}
		h.Buf.RelocateCursors()

		return true
	}
	return false
}

// OutdentLine moves the current line back one indentation
func (h *BufPane) OutdentLine() bool {
	if h.Cursor.HasSelection() {
		return false
	}

	for x := 0; x < len(h.Buf.IndentString(util.IntOpt(h.Buf.Settings["tabsize"]))); x++ {
		if len(util.GetLeadingWhitespace(h.Buf.LineBytes(h.Cursor.Y))) == 0 {
			break
		}
		h.Buf.Remove(buffer.Loc{X: 0, Y: h.Cursor.Y}, buffer.Loc{X: 1, Y: h.Cursor.Y})
	}
	h.Buf.RelocateCursors()
	return true
}

// OutdentSelection takes the current selection and moves it back one indent level
func (h *BufPane) OutdentSelection() bool {
	if h.Cursor.HasSelection() {
		start := h.Cursor.CurSelection[0]
		end := h.Cursor.CurSelection[1]
		if end.Y < start.Y {
			start, end = end, start
			h.Cursor.SetSelectionStart(start)
			h.Cursor.SetSelectionEnd(end)
		}

		startY := start.Y
		endY := end.Move(-1, h.Buf).Y
		for y := startY; y <= endY; y++ {
			for x := 0; x < len(h.Buf.IndentString(util.IntOpt(h.Buf.Settings["tabsize"]))); x++ {
				if len(util.GetLeadingWhitespace(h.Buf.LineBytes(y))) == 0 {
					break
				}
				h.Buf.Remove(buffer.Loc{X: 0, Y: y}, buffer.Loc{X: 1, Y: y})
			}
		}
		h.Buf.RelocateCursors()

		return true
	}
	return false
}

// InsertTab inserts a tab or spaces
func (h *BufPane) InsertTab() bool {
	b := h.Buf
	if b.HasSuggestions {
		b.CycleAutocomplete(true)
		return true
	}

	l := b.LineBytes(h.Cursor.Y)
	l = util.SliceStart(l, h.Cursor.X)
	hasComplete := b.Autocomplete(buffer.BufferComplete)
	if !hasComplete {
		indent := b.IndentString(util.IntOpt(b.Settings["tabsize"]))
		tabBytes := len(indent)
		bytesUntilIndent := tabBytes - (h.Cursor.GetVisualX() % tabBytes)
		b.Insert(h.Cursor.Loc, indent[:bytesUntilIndent])
		return true
	}
	return true
}

// SaveAll saves all open buffers
func (h *BufPane) SaveAll() bool {
	for _, b := range buffer.OpenBuffers {
		b.Save()
	}
	return false
}

// Save the buffer to disk
func (h *BufPane) Save() bool {
	// If this is an empty buffer, ask for a filename
	if h.Buf.Path == "" {
		h.SaveAs()
	} else {
		h.saveBufToFile(h.Buf.Path)
	}

	return false
}

// SaveAs saves the buffer to disk with the given name
func (h *BufPane) SaveAs() bool {
	InfoBar.Prompt("Filename: ", "", "Save", nil, func(resp string, canceled bool) {
		if !canceled {
			// the filename might or might not be quoted, so unquote first then join the strings.
			args, err := shellwords.Split(resp)
			filename := strings.Join(args, " ")
			if err != nil {
				InfoBar.Error("Error parsing arguments: ", err)
				return
			}
			h.saveBufToFile(filename)

		}
	})
	return false
}

// This function saves the buffer to `filename` and changes the buffer's path and name
// to `filename` if the save is successful
func (h *BufPane) saveBufToFile(filename string) {
	err := h.Buf.SaveAs(filename)
	if err != nil {
		if strings.HasSuffix(err.Error(), "permission denied") {
			InfoBar.YNPrompt("Permission denied. Do you want to save this file using sudo? (y,n)", func(yes, canceled bool) {
				if yes && !canceled {
					err = h.Buf.SaveAsWithSudo(filename)
					if err != nil {
						InfoBar.Error(err)
					} else {
						h.Buf.Path = filename
						h.Buf.SetName(filename)
						InfoBar.Message("Saved " + filename)
					}
				}
			})
		} else {
			InfoBar.Error(err)
		}
	} else {
		h.Buf.Path = filename
		h.Buf.SetName(filename)
		InfoBar.Message("Saved " + filename)
	}
}

// Find opens a prompt and searches forward for the input
func (h *BufPane) Find() bool {
	h.searchOrig = h.Cursor.Loc
	InfoBar.Prompt("Find: ", "", "Find", func(resp string) {
		// Event callback
		match, found, _ := h.Buf.FindNext(resp, h.Buf.Start(), h.Buf.End(), h.searchOrig, true, true)
		if found {
			h.Cursor.SetSelectionStart(match[0])
			h.Cursor.SetSelectionEnd(match[1])
			h.Cursor.OrigSelection[0] = h.Cursor.CurSelection[0]
			h.Cursor.OrigSelection[1] = h.Cursor.CurSelection[1]
			h.Cursor.GotoLoc(match[1])
		} else {
			h.Cursor.GotoLoc(h.searchOrig)
			h.Cursor.ResetSelection()
		}
		h.Relocate()
	}, func(resp string, canceled bool) {
		// Finished callback
		if !canceled {
			match, found, err := h.Buf.FindNext(resp, h.Buf.Start(), h.Buf.End(), h.searchOrig, true, true)
			if err != nil {
				InfoBar.Error(err)
			}
			if found {
				h.Cursor.SetSelectionStart(match[0])
				h.Cursor.SetSelectionEnd(match[1])
				h.Cursor.OrigSelection[0] = h.Cursor.CurSelection[0]
				h.Cursor.OrigSelection[1] = h.Cursor.CurSelection[1]
				h.Cursor.GotoLoc(h.Cursor.CurSelection[1])
				h.lastSearch = resp
			} else {
				h.Cursor.ResetSelection()
				InfoBar.Message("No matches found")
			}
		} else {
			h.Cursor.ResetSelection()
		}
		h.Relocate()
	})

	return false
}

// FindNext searches forwards for the last used search term
func (h *BufPane) FindNext() bool {
	// If the cursor is at the start of a selection and we search we want
	// to search from the end of the selection in the case that
	// the selection is a search result in which case we wouldn't move at
	// at all which would be bad
	searchLoc := h.Cursor.Loc
	if h.Cursor.HasSelection() {
		searchLoc = h.Cursor.CurSelection[1]
	}
	match, found, err := h.Buf.FindNext(h.lastSearch, h.Buf.Start(), h.Buf.End(), searchLoc, true, true)
	if err != nil {
		InfoBar.Error(err)
	}
	if found {
		h.Cursor.SetSelectionStart(match[0])
		h.Cursor.SetSelectionEnd(match[1])
		h.Cursor.OrigSelection[0] = h.Cursor.CurSelection[0]
		h.Cursor.OrigSelection[1] = h.Cursor.CurSelection[1]
		h.Cursor.Loc = h.Cursor.CurSelection[1]
	} else {
		h.Cursor.ResetSelection()
	}
	return true
}

// FindPrevious searches backwards for the last used search term
func (h *BufPane) FindPrevious() bool {
	// If the cursor is at the end of a selection and we search we want
	// to search from the beginning of the selection in the case that
	// the selection is a search result in which case we wouldn't move at
	// at all which would be bad
	searchLoc := h.Cursor.Loc
	if h.Cursor.HasSelection() {
		searchLoc = h.Cursor.CurSelection[0]
	}
	match, found, err := h.Buf.FindNext(h.lastSearch, h.Buf.Start(), h.Buf.End(), searchLoc, false, true)
	if err != nil {
		InfoBar.Error(err)
	}
	if found {
		h.Cursor.SetSelectionStart(match[0])
		h.Cursor.SetSelectionEnd(match[1])
		h.Cursor.OrigSelection[0] = h.Cursor.CurSelection[0]
		h.Cursor.OrigSelection[1] = h.Cursor.CurSelection[1]
		h.Cursor.Loc = h.Cursor.CurSelection[1]
	} else {
		h.Cursor.ResetSelection()
	}
	return true
}

// Undo undoes the last action
func (h *BufPane) Undo() bool {
	h.Buf.Undo()
	InfoBar.Message("Undid action")
	return true
}

// Redo redoes the last action
func (h *BufPane) Redo() bool {
	h.Buf.Redo()
	InfoBar.Message("Redid action")
	return true
}

// Copy the selection to the system clipboard
func (h *BufPane) Copy() bool {
	if h.Cursor.HasSelection() {
		h.Cursor.CopySelection("clipboard")
		h.freshClip = true
		InfoBar.Message("Copied selection")
	}
	return true
}

// CutLine cuts the current line to the clipboard
func (h *BufPane) CutLine() bool {
	h.Cursor.SelectLine()
	if !h.Cursor.HasSelection() {
		return false
	}
	if h.freshClip == true {
		if h.Cursor.HasSelection() {
			if clip, err := clipboard.ReadAll("clipboard"); err != nil {
				// messenger.Error(err)
			} else {
				clipboard.WriteAll(clip+string(h.Cursor.GetSelection()), "clipboard")
			}
		}
	} else if time.Since(h.lastCutTime)/time.Second > 10*time.Second || h.freshClip == false {
		h.Copy()
	}
	h.freshClip = true
	h.lastCutTime = time.Now()
	h.Cursor.DeleteSelection()
	h.Cursor.ResetSelection()
	InfoBar.Message("Cut line")
	return true
}

// Cut the selection to the system clipboard
func (h *BufPane) Cut() bool {
	if h.Cursor.HasSelection() {
		h.Cursor.CopySelection("clipboard")
		h.Cursor.DeleteSelection()
		h.Cursor.ResetSelection()
		h.freshClip = true
		InfoBar.Message("Cut selection")

		return true
	} else {
		return h.CutLine()
	}
}

// DuplicateLine duplicates the current line or selection
func (h *BufPane) DuplicateLine() bool {
	if h.Cursor.HasSelection() {
		h.Buf.Insert(h.Cursor.CurSelection[1], string(h.Cursor.GetSelection()))
	} else {
		h.Cursor.End()
		h.Buf.Insert(h.Cursor.Loc, "\n"+string(h.Buf.LineBytes(h.Cursor.Y)))
		// h.Cursor.Right()
	}

	InfoBar.Message("Duplicated line")
	return true
}

// DeleteLine deletes the current line
func (h *BufPane) DeleteLine() bool {
	h.Cursor.SelectLine()
	if !h.Cursor.HasSelection() {
		return false
	}
	h.Cursor.DeleteSelection()
	h.Cursor.ResetSelection()
	InfoBar.Message("Deleted line")
	return true
}

// MoveLinesUp moves up the current line or selected lines if any
func (h *BufPane) MoveLinesUp() bool {
	if h.Cursor.HasSelection() {
		if h.Cursor.CurSelection[0].Y == 0 {
			InfoBar.Message("Can not move further up")
			return true
		}
		start := h.Cursor.CurSelection[0].Y
		end := h.Cursor.CurSelection[1].Y
		if start > end {
			end, start = start, end
		}

		h.Buf.MoveLinesUp(
			start,
			end,
		)
		h.Cursor.CurSelection[1].Y -= 1
	} else {
		if h.Cursor.Loc.Y == 0 {
			InfoBar.Message("Can not move further up")
			return true
		}
		h.Buf.MoveLinesUp(
			h.Cursor.Loc.Y,
			h.Cursor.Loc.Y+1,
		)
	}

	return true
}

// MoveLinesDown moves down the current line or selected lines if any
func (h *BufPane) MoveLinesDown() bool {
	if h.Cursor.HasSelection() {
		if h.Cursor.CurSelection[1].Y >= h.Buf.LinesNum() {
			InfoBar.Message("Can not move further down")
			return true
		}
		start := h.Cursor.CurSelection[0].Y
		end := h.Cursor.CurSelection[1].Y
		if start > end {
			end, start = start, end
		}

		h.Buf.MoveLinesDown(
			start,
			end,
		)
	} else {
		if h.Cursor.Loc.Y >= h.Buf.LinesNum()-1 {
			InfoBar.Message("Can not move further down")
			return true
		}
		h.Buf.MoveLinesDown(
			h.Cursor.Loc.Y,
			h.Cursor.Loc.Y+1,
		)
	}

	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (h *BufPane) Paste() bool {
	clip, _ := clipboard.ReadAll("clipboard")
	h.paste(clip)
	return true
}

// PastePrimary pastes from the primary clipboard (only use on linux)
func (h *BufPane) PastePrimary() bool {
	clip, _ := clipboard.ReadAll("primary")
	h.paste(clip)
	return true
}

func (h *BufPane) paste(clip string) {
	if h.Buf.Settings["smartpaste"].(bool) {
		if h.Cursor.X > 0 && len(util.GetLeadingWhitespace([]byte(strings.TrimLeft(clip, "\r\n")))) == 0 {
			leadingWS := util.GetLeadingWhitespace(h.Buf.LineBytes(h.Cursor.Y))
			clip = strings.Replace(clip, "\n", "\n"+string(leadingWS), -1)
		}
	}

	if h.Cursor.HasSelection() {
		h.Cursor.DeleteSelection()
		h.Cursor.ResetSelection()
	}

	h.Buf.Insert(h.Cursor.Loc, clip)
	// h.Cursor.Loc = h.Cursor.Loc.Move(Count(clip), h.Buf)
	h.freshClip = false
	InfoBar.Message("Pasted clipboard")
}

// JumpToMatchingBrace moves the cursor to the matching brace if it is
// currently on a brace
func (h *BufPane) JumpToMatchingBrace() bool {
	for _, bp := range buffer.BracePairs {
		r := h.Cursor.RuneUnder(h.Cursor.X)
		if r == bp[0] || r == bp[1] {
			matchingBrace := h.Buf.FindMatchingBrace(bp, h.Cursor.Loc)
			h.Cursor.GotoLoc(matchingBrace)
		}
	}

	return true
}

// SelectAll selects the entire buffer
func (h *BufPane) SelectAll() bool {
	h.Cursor.SetSelectionStart(h.Buf.Start())
	h.Cursor.SetSelectionEnd(h.Buf.End())
	// Put the cursor at the beginning
	h.Cursor.X = 0
	h.Cursor.Y = 0
	return true
}

// OpenFile opens a new file in the buffer
func (h *BufPane) OpenFile() bool {
	InfoBar.Prompt("> ", "open ", "Open", nil, func(resp string, canceled bool) {
		if !canceled {
			h.HandleCommand(resp)
		}
	})
	return false
}

// Start moves the viewport to the start of the buffer
func (h *BufPane) Start() bool {
	v := h.GetView()
	v.StartLine = 0
	h.SetView(v)
	return false
}

// End moves the viewport to the end of the buffer
func (h *BufPane) End() bool {
	// TODO: softwrap problems?
	v := h.GetView()
	if v.Height > h.Buf.LinesNum() {
		v.StartLine = 0
		h.SetView(v)
	} else {
		v.StartLine = h.Buf.LinesNum() - v.Height
		h.SetView(v)
	}
	return false
}

// PageUp scrolls the view up a page
func (h *BufPane) PageUp() bool {
	v := h.GetView()
	if v.StartLine > v.Height {
		h.ScrollUp(v.Height)
	} else {
		v.StartLine = 0
	}
	h.SetView(v)
	return false
}

// PageDown scrolls the view down a page
func (h *BufPane) PageDown() bool {
	v := h.GetView()
	if h.Buf.LinesNum()-(v.StartLine+v.Height) > v.Height {
		h.ScrollDown(v.Height)
	} else if h.Buf.LinesNum() >= v.Height {
		v.StartLine = h.Buf.LinesNum() - v.Height
	}
	return false
}

// SelectPageUp selects up one page
func (h *BufPane) SelectPageUp() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.UpN(h.GetView().Height)
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// SelectPageDown selects down one page
func (h *BufPane) SelectPageDown() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.DownN(h.GetView().Height)
	h.Cursor.SelectTo(h.Cursor.Loc)
	return true
}

// CursorPageUp places the cursor a page up
func (h *BufPane) CursorPageUp() bool {
	h.Cursor.Deselect(true)

	if h.Cursor.HasSelection() {
		h.Cursor.Loc = h.Cursor.CurSelection[0]
		h.Cursor.ResetSelection()
		h.Cursor.StoreVisualX()
	}
	h.Cursor.UpN(h.GetView().Height)
	return true
}

// CursorPageDown places the cursor a page up
func (h *BufPane) CursorPageDown() bool {
	h.Cursor.Deselect(false)

	if h.Cursor.HasSelection() {
		h.Cursor.Loc = h.Cursor.CurSelection[1]
		h.Cursor.ResetSelection()
		h.Cursor.StoreVisualX()
	}
	h.Cursor.DownN(h.GetView().Height)
	return true
}

// HalfPageUp scrolls the view up half a page
func (h *BufPane) HalfPageUp() bool {
	v := h.GetView()
	if v.StartLine > v.Height/2 {
		h.ScrollUp(v.Height / 2)
	} else {
		v.StartLine = 0
	}
	h.SetView(v)
	return false
}

// HalfPageDown scrolls the view down half a page
func (h *BufPane) HalfPageDown() bool {
	v := h.GetView()
	if h.Buf.LinesNum()-(v.StartLine+v.Height) > v.Height/2 {
		h.ScrollDown(v.Height / 2)
	} else {
		if h.Buf.LinesNum() >= v.Height {
			v.StartLine = h.Buf.LinesNum() - v.Height
		}
	}
	h.SetView(v)
	return false
}

// ToggleRuler turns line numbers off and on
func (h *BufPane) ToggleRuler() bool {
	if !h.Buf.Settings["ruler"].(bool) {
		h.Buf.Settings["ruler"] = true
		InfoBar.Message("Enabled ruler")
	} else {
		h.Buf.Settings["ruler"] = false
		InfoBar.Message("Disabled ruler")
	}
	return false
}

// JumpLine jumps to a line and moves the view accordingly.
func (h *BufPane) JumpLine() bool {
	return false
}

// ClearStatus clears the messenger bar
func (h *BufPane) ClearStatus() bool {
	InfoBar.Message("")
	return false
}

// ToggleHelp toggles the help screen
func (h *BufPane) ToggleHelp() bool {
	if h.Buf.Type == buffer.BTHelp {
		h.Quit()
	} else {
		h.openHelp("help")
	}
	return false
}

// ToggleKeyMenu toggles the keymenu option and resizes all tabs
func (h *BufPane) ToggleKeyMenu() bool {
	config.GlobalSettings["keymenu"] = !config.GetGlobalOption("keymenu").(bool)
	Tabs.Resize()
	return false
}

// ShellMode opens a terminal to run a shell command
func (h *BufPane) ShellMode() bool {
	InfoBar.Prompt("$ ", "", "Shell", nil, func(resp string, canceled bool) {
		if !canceled {
			// The true here is for openTerm to make the command interactive
			shell.RunInteractiveShell(resp, true, false)
		}
	})

	return false
}

// CommandMode lets the user enter a command
func (h *BufPane) CommandMode() bool {
	InfoBar.Prompt("> ", "", "Command", nil, func(resp string, canceled bool) {
		if !canceled {
			h.HandleCommand(resp)
		}
	})
	return false
}

// ToggleOverwriteMode lets the user toggle the text overwrite mode
func (h *BufPane) ToggleOverwriteMode() bool {
	h.isOverwriteMode = !h.isOverwriteMode
	return false
}

// Escape leaves current mode
func (h *BufPane) Escape() bool {
	return false
}

// Quit this will close the current tab or view that is open
func (h *BufPane) Quit() bool {
	quit := func() {
		h.Buf.Close()
		if len(MainTab().Panes) > 1 {
			h.Unsplit()
		} else if len(Tabs.List) > 1 {
			Tabs.RemoveTab(h.splitID)
		} else {
			screen.Screen.Fini()
			InfoBar.Close()
			os.Exit(0)
		}
	}
	if h.Buf.Modified() {
		InfoBar.YNPrompt("Save changes to "+h.Buf.GetName()+" before closing? (y,n,esc)", func(yes, canceled bool) {
			if !canceled && !yes {
				quit()
			} else if !canceled && yes {
				h.Save()
				quit()
			}
		})
	} else {
		quit()
	}
	return false
}

// QuitAll quits the whole editor; all splits and tabs
func (h *BufPane) QuitAll() bool {
	return false
}

// AddTab adds a new tab with an empty buffer
func (h *BufPane) AddTab() bool {
	width, height := screen.Screen.Size()
	iOffset := config.GetInfoBarOffset()
	b := buffer.NewBufferFromString("", "", buffer.BTDefault)
	tp := NewTabFromBuffer(0, 0, width, height-iOffset, b)
	Tabs.AddTab(tp)
	Tabs.SetActive(len(Tabs.List) - 1)

	return false
}

// PreviousTab switches to the previous tab in the tab list
func (h *BufPane) PreviousTab() bool {
	a := Tabs.Active()
	Tabs.SetActive(util.Clamp(a-1, 0, len(Tabs.List)-1))

	return false
}

// NextTab switches to the next tab in the tab list
func (h *BufPane) NextTab() bool {
	a := Tabs.Active()
	Tabs.SetActive(util.Clamp(a+1, 0, len(Tabs.List)-1))
	return false
}

// VSplitAction opens an empty vertical split
func (h *BufPane) VSplitAction() bool {
	h.VSplitBuf(buffer.NewBufferFromString("", "", buffer.BTDefault))

	return false
}

// HSplitAction opens an empty horizontal split
func (h *BufPane) HSplitAction() bool {
	h.HSplitBuf(buffer.NewBufferFromString("", "", buffer.BTDefault))

	return false
}

// Unsplit closes all splits in the current tab except the active one
func (h *BufPane) Unsplit() bool {
	n := MainTab().GetNode(h.splitID)
	n.Unsplit()

	MainTab().RemovePane(MainTab().GetPane(h.splitID))
	MainTab().Resize()
	MainTab().SetActive(len(MainTab().Panes) - 1)
	return false
}

// NextSplit changes the view to the next split
func (h *BufPane) NextSplit() bool {
	a := MainTab().active
	if a < len(MainTab().Panes)-1 {
		a++
	} else {
		a = 0
	}

	MainTab().SetActive(a)

	return false
}

// PreviousSplit changes the view to the previous split
func (h *BufPane) PreviousSplit() bool {
	a := MainTab().active
	if a > 0 {
		a--
	} else {
		a = len(MainTab().Panes) - 1
	}
	MainTab().SetActive(a)

	return false
}

var curMacro []interface{}
var recordingMacro bool

// ToggleMacro toggles recording of a macro
func (h *BufPane) ToggleMacro() bool {
	return true
}

// PlayMacro plays back the most recently recorded macro
func (h *BufPane) PlayMacro() bool {
	return true
}

// SpawnMultiCursor creates a new multiple cursor at the next occurrence of the current selection or current word
func (h *BufPane) SpawnMultiCursor() bool {
	spawner := h.Buf.GetCursor(h.Buf.NumCursors() - 1)
	if !spawner.HasSelection() {
		spawner.SelectWord()
		h.multiWord = true
		return true
	}

	sel := spawner.GetSelection()
	searchStart := spawner.CurSelection[1]

	search := string(sel)
	search = regexp.QuoteMeta(search)
	if h.multiWord {
		search = "\\b" + search + "\\b"
	}
	match, found, err := h.Buf.FindNext(search, h.Buf.Start(), h.Buf.End(), searchStart, true, true)
	if err != nil {
		InfoBar.Error(err)
	}
	if found {
		c := buffer.NewCursor(h.Buf, buffer.Loc{})
		c.SetSelectionStart(match[0])
		c.SetSelectionEnd(match[1])
		c.OrigSelection[0] = c.CurSelection[0]
		c.OrigSelection[1] = c.CurSelection[1]
		c.Loc = c.CurSelection[1]

		h.Buf.AddCursor(c)
		h.Buf.SetCurCursor(h.Buf.NumCursors() - 1)
		h.Buf.MergeCursors()
	} else {
		InfoBar.Message("No matches found")
	}

	return true
}

// SpawnMultiCursorSelect adds a cursor at the beginning of each line of a selection
func (h *BufPane) SpawnMultiCursorSelect() bool {
	// Avoid cases where multiple cursors already exist, that would create problems
	if h.Buf.NumCursors() > 1 {
		return false
	}

	var startLine int
	var endLine int

	a, b := h.Cursor.CurSelection[0].Y, h.Cursor.CurSelection[1].Y
	if a > b {
		startLine, endLine = b, a
	} else {
		startLine, endLine = a, b
	}

	if h.Cursor.HasSelection() {
		h.Cursor.ResetSelection()
		h.Cursor.GotoLoc(buffer.Loc{0, startLine})

		for i := startLine; i <= endLine; i++ {
			c := buffer.NewCursor(h.Buf, buffer.Loc{0, i})
			c.StoreVisualX()
			h.Buf.AddCursor(c)
		}
		h.Buf.MergeCursors()
	} else {
		return false
	}
	InfoBar.Message("Added cursors from selection")
	return false
}

// MouseMultiCursor is a mouse action which puts a new cursor at the mouse position
func (h *BufPane) MouseMultiCursor(e *tcell.EventMouse) bool {
	b := h.Buf
	mx, my := e.Position()
	mouseLoc := h.GetMouseLoc(buffer.Loc{X: mx, Y: my})
	c := buffer.NewCursor(b, mouseLoc)
	b.AddCursor(c)
	b.MergeCursors()

	return false
}

// SkipMultiCursor moves the current multiple cursor to the next available position
func (h *BufPane) SkipMultiCursor() bool {
	lastC := h.Buf.GetCursor(h.Buf.NumCursors() - 1)
	sel := lastC.GetSelection()
	searchStart := lastC.CurSelection[1]

	search := string(sel)
	search = regexp.QuoteMeta(search)
	if h.multiWord {
		search = "\\b" + search + "\\b"
	}

	match, found, err := h.Buf.FindNext(search, h.Buf.Start(), h.Buf.End(), searchStart, true, true)
	if err != nil {
		InfoBar.Error(err)
	}
	if found {
		lastC.SetSelectionStart(match[0])
		lastC.SetSelectionEnd(match[1])
		lastC.OrigSelection[0] = lastC.CurSelection[0]
		lastC.OrigSelection[1] = lastC.CurSelection[1]
		lastC.Loc = lastC.CurSelection[1]

		h.Buf.MergeCursors()
		h.Buf.SetCurCursor(h.Buf.NumCursors() - 1)
	} else {
		InfoBar.Message("No matches found")
	}
	return true
}

// RemoveMultiCursor removes the latest multiple cursor
func (h *BufPane) RemoveMultiCursor() bool {
	if h.Buf.NumCursors() > 1 {
		h.Buf.RemoveCursor(h.Buf.NumCursors() - 1)
		h.Buf.SetCurCursor(h.Buf.NumCursors() - 1)
		h.Buf.UpdateCursors()
	} else {
		h.multiWord = false
	}
	return true
}

// RemoveAllMultiCursors removes all cursors except the base cursor
func (h *BufPane) RemoveAllMultiCursors() bool {
	h.Buf.ClearCursors()
	h.multiWord = false
	return true
}
