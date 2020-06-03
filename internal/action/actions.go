package action

import (
	"regexp"
	"runtime"
	"strings"
	"time"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/zyedidia/clipboard"
	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/micro/v2/internal/shell"
	"github.com/zyedidia/micro/v2/internal/util"
	"github.com/zyedidia/tcell"
)

// ScrollUp is not an action
func (h *BufPane) ScrollUp(n int) {
	v := h.GetView()
	if v.StartLine >= n {
		v.StartLine -= n
		h.SetView(v)
	} else {
		v.StartLine = 0
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
	mouseLoc := h.LocFromVisual(buffer.Loc{mx, my})
	h.Cursor.Loc = mouseLoc
	if h.mouseReleased {
		if b.NumCursors() > 1 {
			b.ClearCursors()
			h.Relocate()
			h.Cursor = h.Buf.GetActiveCursor()
			h.Cursor.Loc = mouseLoc
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
		}
	}

	h.Cursor.StoreVisualX()
	h.lastLoc = mouseLoc
	return true
}

// ScrollUpAction scrolls the view up
func (h *BufPane) ScrollUpAction() bool {
	h.ScrollUp(util.IntOpt(h.Buf.Settings["scrollspeed"]))
	return true
}

// ScrollDownAction scrolls the view up
func (h *BufPane) ScrollDownAction() bool {
	h.ScrollDown(util.IntOpt(h.Buf.Settings["scrollspeed"]))
	return true
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
	h.Relocate()
	return true
}

// CursorUp moves the cursor up
func (h *BufPane) CursorUp() bool {
	h.Cursor.Deselect(true)
	h.Cursor.Up()
	h.Relocate()
	return true
}

// CursorDown moves the cursor down
func (h *BufPane) CursorDown() bool {
	h.Cursor.Deselect(true)
	h.Cursor.Down()
	h.Relocate()
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
	h.Relocate()
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
			if h.Cursor.X+tabsize < util.CharacterCount(line) && util.IsSpaces(line[h.Cursor.X:h.Cursor.X+tabsize]) && util.IsBytesWhitespace(line[0:h.Cursor.X]) {
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

	h.Relocate()
	return true
}

// WordRight moves the cursor one word to the right
func (h *BufPane) WordRight() bool {
	h.Cursor.Deselect(false)
	h.Cursor.WordRight()
	h.Relocate()
	return true
}

// WordLeft moves the cursor one word to the left
func (h *BufPane) WordLeft() bool {
	h.Cursor.Deselect(true)
	h.Cursor.WordLeft()
	h.Relocate()
	return true
}

// SelectUp selects up one line
func (h *BufPane) SelectUp() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.Up()
	h.Cursor.SelectTo(h.Cursor.Loc)
	h.Relocate()
	return true
}

// SelectDown selects down one line
func (h *BufPane) SelectDown() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.Down()
	h.Cursor.SelectTo(h.Cursor.Loc)
	h.Relocate()
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
	h.Relocate()
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
	h.Relocate()
	return true
}

// SelectWordRight selects the word to the right of the cursor
func (h *BufPane) SelectWordRight() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.WordRight()
	h.Cursor.SelectTo(h.Cursor.Loc)
	h.Relocate()
	return true
}

// SelectWordLeft selects the word to the left of the cursor
func (h *BufPane) SelectWordLeft() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.WordLeft()
	h.Cursor.SelectTo(h.Cursor.Loc)
	h.Relocate()
	return true
}

// StartOfText moves the cursor to the start of the text of the line
func (h *BufPane) StartOfText() bool {
	h.Cursor.Deselect(true)
	h.Cursor.StartOfText()
	h.Relocate()
	return true
}

// StartOfTextToggle toggles the cursor between the start of the text of the line
// and the start of the line
func (h *BufPane) StartOfTextToggle() bool {
	h.Cursor.Deselect(true)
	if h.Cursor.IsStartOfText() {
		h.Cursor.Start()
	} else {
		h.Cursor.StartOfText()
	}
	h.Relocate()
	return true
}

// StartOfLine moves the cursor to the start of the line
func (h *BufPane) StartOfLine() bool {
	h.Cursor.Deselect(true)
	h.Cursor.Start()
	h.Relocate()
	return true
}

// EndOfLine moves the cursor to the end of the line
func (h *BufPane) EndOfLine() bool {
	h.Cursor.Deselect(true)
	h.Cursor.End()
	h.Relocate()
	return true
}

// SelectLine selects the entire current line
func (h *BufPane) SelectLine() bool {
	h.Cursor.SelectLine()
	h.Relocate()
	return true
}

// SelectToStartOfText selects to the start of the text on the current line
func (h *BufPane) SelectToStartOfText() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.StartOfText()
	h.Cursor.SelectTo(h.Cursor.Loc)
	h.Relocate()
	return true
}

// SelectToStartOfTextToggle toggles the selection between the start of the text
// on the current line and the start of the line
func (h *BufPane) SelectToStartOfTextToggle() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	if h.Cursor.IsStartOfText() {
		h.Cursor.Start()
	} else {
		h.Cursor.StartOfText()
	}
	h.Cursor.SelectTo(h.Cursor.Loc)
	h.Relocate()
	return true
}

// SelectToStartOfLine selects to the start of the current line
func (h *BufPane) SelectToStartOfLine() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.Start()
	h.Cursor.SelectTo(h.Cursor.Loc)
	h.Relocate()
	return true
}

// SelectToEndOfLine selects to the end of the current line
func (h *BufPane) SelectToEndOfLine() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.End()
	h.Cursor.SelectTo(h.Cursor.Loc)
	h.Relocate()
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
	h.Relocate()
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
	h.Relocate()
	return true
}

// Retab changes all tabs to spaces or all spaces to tabs depending
// on the user's settings
func (h *BufPane) Retab() bool {
	h.Buf.Retab()
	h.Relocate()
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (h *BufPane) CursorStart() bool {
	h.Cursor.Deselect(true)
	h.Cursor.X = 0
	h.Cursor.Y = 0
	h.Cursor.StoreVisualX()
	h.Relocate()
	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (h *BufPane) CursorEnd() bool {
	h.Cursor.Deselect(true)
	h.Cursor.Loc = h.Buf.End()
	h.Cursor.StoreVisualX()
	h.Relocate()
	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (h *BufPane) SelectToStart() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.CursorStart()
	h.Cursor.SelectTo(h.Buf.Start())
	h.Relocate()
	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (h *BufPane) SelectToEnd() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.CursorEnd()
	h.Cursor.SelectTo(h.Buf.End())
	h.Relocate()
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
			h.Buf.Remove(buffer.Loc{X: 0, Y: h.Cursor.Y - 1}, buffer.Loc{X: util.CharacterCount(line), Y: h.Cursor.Y - 1})
		}
	}
	h.Cursor.LastVisualX = h.Cursor.GetVisualX()
	h.Relocate()
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
		if h.Buf.Settings["tabstospaces"].(bool) && util.IsSpaces(lineStart) && len(lineStart) != 0 && util.CharacterCount(lineStart)%tabSize == 0 {
			loc := h.Cursor.Loc
			h.Buf.Remove(loc.Move(-tabSize, h.Buf), loc)
		} else {
			loc := h.Cursor.Loc
			h.Buf.Remove(loc.Move(-1, h.Buf), loc)
		}
	}
	h.Cursor.LastVisualX = h.Cursor.GetVisualX()
	h.Relocate()
	return true
}

// DeleteWordRight deletes the word to the right of the cursor
func (h *BufPane) DeleteWordRight() bool {
	h.SelectWordRight()
	if h.Cursor.HasSelection() {
		h.Cursor.DeleteSelection()
		h.Cursor.ResetSelection()
	}
	h.Relocate()
	return true
}

// DeleteWordLeft deletes the word to the left of the cursor
func (h *BufPane) DeleteWordLeft() bool {
	h.SelectWordLeft()
	if h.Cursor.HasSelection() {
		h.Cursor.DeleteSelection()
		h.Cursor.ResetSelection()
	}
	h.Relocate()
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
	h.Relocate()
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
			if len(h.Buf.LineBytes(y)) > 0 {
				h.Buf.Insert(buffer.Loc{X: 0, Y: y}, h.Buf.IndentString(tabsize))
				if y == startY && start.X > 0 {
					h.Cursor.SetSelectionStart(start.Move(indentsize, h.Buf))
				}
				if y == endY {
					h.Cursor.SetSelectionEnd(buffer.Loc{X: endX + indentsize + 1, Y: endY})
				}
			}
		}
		h.Buf.RelocateCursors()

		h.Relocate()
		return true
	}
	return false
}

// IndentLine moves the current line forward one indentation
func (h *BufPane) IndentLine() bool {
	if h.Cursor.HasSelection() {
		return false
	}

	tabsize := int(h.Buf.Settings["tabsize"].(float64))
	indentstr := h.Buf.IndentString(tabsize)
	h.Buf.Insert(buffer.Loc{X: 0, Y: h.Cursor.Y}, indentstr)
	h.Buf.RelocateCursors()
	h.Relocate()
	return true
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
	h.Relocate()
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

		h.Relocate()
		return true
	}
	return false
}

// Autocomplete cycles the suggestions and performs autocompletion if there are suggestions
func (h *BufPane) Autocomplete() bool {
	b := h.Buf

	if h.Cursor.HasSelection() {
		return false
	}

	if !util.IsNonAlphaNumeric(h.Cursor.RuneUnder(h.Cursor.X)) {
		// don't autocomplete if cursor is on alpha numeric character (middle of a word)
		return false
	}

	if b.HasSuggestions {
		b.CycleAutocomplete(true)
		return true
	}
	return b.Autocomplete(buffer.BufferComplete)
}

// CycleAutocompleteBack cycles back in the autocomplete suggestion list
func (h *BufPane) CycleAutocompleteBack() bool {
	if h.Cursor.HasSelection() {
		return false
	}

	if h.Buf.HasSuggestions {
		h.Buf.CycleAutocomplete(false)
		return true
	}
	return false
}

// InsertTab inserts a tab or spaces
func (h *BufPane) InsertTab() bool {
	b := h.Buf
	indent := b.IndentString(util.IntOpt(b.Settings["tabsize"]))
	tabBytes := len(indent)
	bytesUntilIndent := tabBytes - (h.Cursor.GetVisualX() % tabBytes)
	b.Insert(h.Cursor.Loc, indent[:bytesUntilIndent])
	h.Relocate()
	return true
}

// SaveAll saves all open buffers
func (h *BufPane) SaveAll() bool {
	for _, b := range buffer.OpenBuffers {
		b.Save()
	}
	return true
}

// SaveCB performs a save and does a callback at the very end (after all prompts have been resolved)
func (h *BufPane) SaveCB(action string, callback func()) bool {
	// If this is an empty buffer, ask for a filename
	if h.Buf.Path == "" {
		h.SaveAsCB(action, callback)
	} else {
		noPrompt := h.saveBufToFile(h.Buf.Path, action, callback)
		if noPrompt {
			return true
		}
	}
	return false
}

// Save the buffer to disk
func (h *BufPane) Save() bool {
	return h.SaveCB("Save", nil)
}

// SaveAsCB performs a save as and does a callback at the very end (after all prompts have been resolved)
// The callback is only called if the save was successful
func (h *BufPane) SaveAsCB(action string, callback func()) bool {
	InfoBar.Prompt("Filename: ", "", "Save", nil, func(resp string, canceled bool) {
		if !canceled {
			// the filename might or might not be quoted, so unquote first then join the strings.
			args, err := shellquote.Split(resp)
			if err != nil {
				InfoBar.Error("Error parsing arguments: ", err)
				return
			}
			if len(args) == 0 {
				InfoBar.Error("No filename given")
				return
			}
			filename := strings.Join(args, " ")
			noPrompt := h.saveBufToFile(filename, action, callback)
			if noPrompt {
				h.completeAction(action)
			}
		}
	})
	return false
}

// SaveAs saves the buffer to disk with the given name
func (h *BufPane) SaveAs() bool {
	return h.SaveAsCB("SaveAs", nil)
}

// This function saves the buffer to `filename` and changes the buffer's path and name
// to `filename` if the save is successful
// The callback is only called if the save was successful
func (h *BufPane) saveBufToFile(filename string, action string, callback func()) bool {
	err := h.Buf.SaveAs(filename)
	if err != nil {
		if strings.HasSuffix(err.Error(), "permission denied") {
			saveWithSudo := func() {
				err = h.Buf.SaveAsWithSudo(filename)
				if err != nil {
					InfoBar.Error(err)
				} else {
					h.Buf.Path = filename
					h.Buf.SetName(filename)
					InfoBar.Message("Saved " + filename)
					if callback != nil {
						callback()
					}
				}
			}
			if h.Buf.Settings["autosu"].(bool) {
				saveWithSudo()
			} else {
				InfoBar.YNPrompt("Permission denied. Do you want to save this file using sudo? (y,n)", func(yes, canceled bool) {
					if yes && !canceled {
						saveWithSudo()
						h.completeAction(action)
					}
				})
				return false
			}
		} else {
			InfoBar.Error(err)
		}
	} else {
		h.Buf.Path = filename
		h.Buf.SetName(filename)
		InfoBar.Message("Saved " + filename)
		if callback != nil {
			callback()
		}
	}
	return true
}

// Find opens a prompt and searches forward for the input
func (h *BufPane) Find() bool {
	return h.find(true)
}

// FindLiteral is the same as Find() but does not support regular expressions
func (h *BufPane) FindLiteral() bool {
	return h.find(false)
}

func (h *BufPane) find(useRegex bool) bool {
	h.searchOrig = h.Cursor.Loc
	prompt := "Find: "
	if useRegex {
		prompt = "Find (regex): "
	}
	InfoBar.Prompt(prompt, "", "Find", func(resp string) {
		// Event callback
		match, found, _ := h.Buf.FindNext(resp, h.Buf.Start(), h.Buf.End(), h.searchOrig, true, useRegex)
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
			match, found, err := h.Buf.FindNext(resp, h.Buf.Start(), h.Buf.End(), h.searchOrig, true, useRegex)
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
				h.lastSearchRegex = useRegex
			} else {
				h.Cursor.ResetSelection()
				InfoBar.Message("No matches found")
			}
		} else {
			h.Cursor.ResetSelection()
		}
		h.Relocate()
	})

	return true
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
	match, found, err := h.Buf.FindNext(h.lastSearch, h.Buf.Start(), h.Buf.End(), searchLoc, true, h.lastSearchRegex)
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
	h.Relocate()
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
	match, found, err := h.Buf.FindNext(h.lastSearch, h.Buf.Start(), h.Buf.End(), searchLoc, false, h.lastSearchRegex)
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
	h.Relocate()
	return true
}

// Undo undoes the last action
func (h *BufPane) Undo() bool {
	h.Buf.Undo()
	InfoBar.Message("Undid action")
	h.Relocate()
	return true
}

// Redo redoes the last action
func (h *BufPane) Redo() bool {
	h.Buf.Redo()
	InfoBar.Message("Redid action")
	h.Relocate()
	return true
}

// Copy the selection to the system clipboard
func (h *BufPane) Copy() bool {
	if h.Cursor.HasSelection() {
		h.Cursor.CopySelection("clipboard")
		h.freshClip = true
		if clipboard.Unsupported {
			InfoBar.Message("Copied selection (install xclip for external clipboard)")
		} else {
			InfoBar.Message("Copied selection")
		}
	}
	h.Relocate()
	return true
}

// Copy the current line to the clipboard
func (h *BufPane) CopyLine() bool {
	if h.Cursor.HasSelection() {
		return false
	} else {
		h.Cursor.SelectLine()
		h.Cursor.CopySelection("clipboard")
		h.freshClip = true
		if clipboard.Unsupported {
			InfoBar.Message("Copied line (install xclip for external clipboard)")
		} else {
			InfoBar.Message("Copied line")
		}
	}
	h.Cursor.Deselect(true)
	h.Relocate()
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
	h.Relocate()
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

		h.Relocate()
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
	h.Relocate()
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
	h.Relocate()
	return true
}

// MoveLinesUp moves up the current line or selected lines if any
func (h *BufPane) MoveLinesUp() bool {
	if h.Cursor.HasSelection() {
		if h.Cursor.CurSelection[0].Y == 0 {
			InfoBar.Message("Cannot move further up")
			return false
		}
		start := h.Cursor.CurSelection[0].Y
		end := h.Cursor.CurSelection[1].Y
		sel := 1
		if start > end {
			end, start = start, end
			sel = 0
		}

		compensate := false
		if h.Cursor.CurSelection[sel].X != 0 {
			end++
		} else {
			compensate = true
		}

		h.Buf.MoveLinesUp(
			start,
			end,
		)
		if compensate {
			h.Cursor.CurSelection[sel].Y -= 1
		}
	} else {
		if h.Cursor.Loc.Y == 0 {
			InfoBar.Message("Cannot move further up")
			return false
		}
		h.Buf.MoveLinesUp(
			h.Cursor.Loc.Y,
			h.Cursor.Loc.Y+1,
		)
	}

	h.Relocate()
	return true
}

// MoveLinesDown moves down the current line or selected lines if any
func (h *BufPane) MoveLinesDown() bool {
	if h.Cursor.HasSelection() {
		if h.Cursor.CurSelection[1].Y >= h.Buf.LinesNum() {
			InfoBar.Message("Cannot move further down")
			return false
		}
		start := h.Cursor.CurSelection[0].Y
		end := h.Cursor.CurSelection[1].Y
		sel := 1
		if start > end {
			end, start = start, end
			sel = 0
		}

		if h.Cursor.CurSelection[sel].X != 0 {
			end++
		}

		h.Buf.MoveLinesDown(
			start,
			end,
		)
	} else {
		if h.Cursor.Loc.Y >= h.Buf.LinesNum()-1 {
			InfoBar.Message("Cannot move further down")
			return false
		}
		h.Buf.MoveLinesDown(
			h.Cursor.Loc.Y,
			h.Cursor.Loc.Y+1,
		)
	}

	h.Relocate()
	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (h *BufPane) Paste() bool {
	clip, _ := clipboard.ReadAll("clipboard")
	h.paste(clip)
	h.Relocate()
	return true
}

// PastePrimary pastes from the primary clipboard (only use on linux)
func (h *BufPane) PastePrimary() bool {
	clip, _ := clipboard.ReadAll("primary")
	h.paste(clip)
	h.Relocate()
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
	if clipboard.Unsupported {
		InfoBar.Message("Pasted clipboard (install xclip for external clipboard)")
	} else {
		InfoBar.Message("Pasted clipboard")
	}
}

// JumpToMatchingBrace moves the cursor to the matching brace if it is
// currently on a brace
func (h *BufPane) JumpToMatchingBrace() bool {
	for _, bp := range buffer.BracePairs {
		r := h.Cursor.RuneUnder(h.Cursor.X)
		rl := h.Cursor.RuneUnder(h.Cursor.X - 1)
		if r == bp[0] || r == bp[1] || rl == bp[0] || rl == bp[1] {
			matchingBrace, left, found := h.Buf.FindMatchingBrace(bp, h.Cursor.Loc)
			if found {
				if left {
					h.Cursor.GotoLoc(matchingBrace)
				} else {
					h.Cursor.GotoLoc(matchingBrace.Move(1, h.Buf))
				}
			} else {
				return false
			}
		}
	}

	h.Relocate()
	return true
}

// SelectAll selects the entire buffer
func (h *BufPane) SelectAll() bool {
	h.Cursor.SetSelectionStart(h.Buf.Start())
	h.Cursor.SetSelectionEnd(h.Buf.End())
	// Put the cursor at the beginning
	h.Cursor.X = 0
	h.Cursor.Y = 0
	h.Relocate()
	return true
}

// OpenFile opens a new file in the buffer
func (h *BufPane) OpenFile() bool {
	InfoBar.Prompt("> ", "open ", "Open", nil, func(resp string, canceled bool) {
		if !canceled {
			h.HandleCommand(resp)
		}
	})
	return true
}

// OpenFile opens a new file in the buffer
func (h *BufPane) JumpLine() bool {
	InfoBar.Prompt("> ", "goto ", "Command", nil, func(resp string, canceled bool) {
		if !canceled {
			h.HandleCommand(resp)
		}
	})
	return true
}

// Start moves the viewport to the start of the buffer
func (h *BufPane) Start() bool {
	v := h.GetView()
	v.StartLine = 0
	h.SetView(v)
	return true
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
	return true
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
	return true
}

// PageDown scrolls the view down a page
func (h *BufPane) PageDown() bool {
	v := h.GetView()
	if h.Buf.LinesNum()-(v.StartLine+v.Height) > v.Height {
		h.ScrollDown(v.Height)
	} else if h.Buf.LinesNum() >= v.Height {
		v.StartLine = h.Buf.LinesNum() - v.Height
	}
	return true
}

// SelectPageUp selects up one page
func (h *BufPane) SelectPageUp() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.UpN(h.GetView().Height)
	h.Cursor.SelectTo(h.Cursor.Loc)
	h.Relocate()
	return true
}

// SelectPageDown selects down one page
func (h *BufPane) SelectPageDown() bool {
	if !h.Cursor.HasSelection() {
		h.Cursor.OrigSelection[0] = h.Cursor.Loc
	}
	h.Cursor.DownN(h.GetView().Height)
	h.Cursor.SelectTo(h.Cursor.Loc)
	h.Relocate()
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
	h.Relocate()
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
	h.Relocate()
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
	return true
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
	return true
}

// ToggleDiffGutter turns the diff gutter off and on
func (h *BufPane) ToggleDiffGutter() bool {
	if !h.Buf.Settings["diffgutter"].(bool) {
		h.Buf.Settings["diffgutter"] = true
		h.Buf.UpdateDiff(func(synchronous bool) {
			screen.Redraw()
		})
		InfoBar.Message("Enabled diff gutter")
	} else {
		h.Buf.Settings["diffgutter"] = false
		InfoBar.Message("Disabled diff gutter")
	}
	return true
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
	return true
}

// ClearStatus clears the messenger bar
func (h *BufPane) ClearStatus() bool {
	InfoBar.Message("")
	return true
}

// ToggleHelp toggles the help screen
func (h *BufPane) ToggleHelp() bool {
	if h.Buf.Type == buffer.BTHelp {
		h.Quit()
	} else {
		h.openHelp("help")
	}
	return true
}

// ToggleKeyMenu toggles the keymenu option and resizes all tabs
func (h *BufPane) ToggleKeyMenu() bool {
	config.GlobalSettings["keymenu"] = !config.GetGlobalOption("keymenu").(bool)
	Tabs.Resize()
	return true
}

// ShellMode opens a terminal to run a shell command
func (h *BufPane) ShellMode() bool {
	InfoBar.Prompt("$ ", "", "Shell", nil, func(resp string, canceled bool) {
		if !canceled {
			// The true here is for openTerm to make the command interactive
			shell.RunInteractiveShell(resp, true, false)
		}
	})

	return true
}

// CommandMode lets the user enter a command
func (h *BufPane) CommandMode() bool {
	InfoBar.Prompt("> ", "", "Command", nil, func(resp string, canceled bool) {
		if !canceled {
			h.HandleCommand(resp)
		}
	})
	return true
}

// ToggleOverwriteMode lets the user toggle the text overwrite mode
func (h *BufPane) ToggleOverwriteMode() bool {
	h.isOverwriteMode = !h.isOverwriteMode
	return true
}

// Escape leaves current mode
func (h *BufPane) Escape() bool {
	return true
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
			runtime.Goexit()
		}
	}
	if h.Buf.Modified() {
		if config.GlobalSettings["autosave"].(float64) > 0 {
			// autosave on means we automatically save when quitting
			h.SaveCB("Quit", func() {
				quit()
			})
		} else {
			InfoBar.YNPrompt("Save changes to "+h.Buf.GetName()+" before closing? (y,n,esc)", func(yes, canceled bool) {
				if !canceled && !yes {
					quit()
				} else if !canceled && yes {
					h.SaveCB("Quit", func() {
						quit()
					})
				}
			})
		}
	} else {
		quit()
	}
	return true
}

// QuitAll quits the whole editor; all splits and tabs
func (h *BufPane) QuitAll() bool {
	anyModified := false
	for _, b := range buffer.OpenBuffers {
		if b.Modified() {
			anyModified = true
			break
		}
	}

	quit := func() {
		for _, b := range buffer.OpenBuffers {
			b.Close()
		}
		screen.Screen.Fini()
		InfoBar.Close()
		runtime.Goexit()
	}

	if anyModified {
		InfoBar.YNPrompt("Quit micro? (all open buffers will be closed without saving)", func(yes, canceled bool) {
			if !canceled && yes {
				quit()
			}
		})
	} else {
		quit()
	}

	return true
}

// AddTab adds a new tab with an empty buffer
func (h *BufPane) AddTab() bool {
	width, height := screen.Screen.Size()
	iOffset := config.GetInfoBarOffset()
	b := buffer.NewBufferFromString("", "", buffer.BTDefault)
	tp := NewTabFromBuffer(0, 0, width, height-iOffset, b)
	Tabs.AddTab(tp)
	Tabs.SetActive(len(Tabs.List) - 1)

	return true
}

// PreviousTab switches to the previous tab in the tab list
func (h *BufPane) PreviousTab() bool {
	tabsLen := len(Tabs.List)
	a := Tabs.Active() + tabsLen
	Tabs.SetActive((a - 1) % tabsLen)

	return true
}

// NextTab switches to the next tab in the tab list
func (h *BufPane) NextTab() bool {
	a := Tabs.Active()
	Tabs.SetActive((a + 1) % len(Tabs.List))

	return true
}

// VSplitAction opens an empty vertical split
func (h *BufPane) VSplitAction() bool {
	h.VSplitBuf(buffer.NewBufferFromString("", "", buffer.BTDefault))

	return true
}

// HSplitAction opens an empty horizontal split
func (h *BufPane) HSplitAction() bool {
	h.HSplitBuf(buffer.NewBufferFromString("", "", buffer.BTDefault))

	return true
}

// Unsplit closes all splits in the current tab except the active one
func (h *BufPane) Unsplit() bool {
	tab := h.tab
	n := tab.GetNode(h.splitID)
	ok := n.Unsplit()
	if ok {
		tab.RemovePane(tab.GetPane(h.splitID))
		tab.Resize()
		tab.SetActive(len(tab.Panes) - 1)

		return true
	}
	return false
}

// NextSplit changes the view to the next split
func (h *BufPane) NextSplit() bool {
	a := h.tab.active
	if a < len(h.tab.Panes)-1 {
		a++
	} else {
		a = 0
	}

	h.tab.SetActive(a)

	return true
}

// PreviousSplit changes the view to the previous split
func (h *BufPane) PreviousSplit() bool {
	a := h.tab.active
	if a > 0 {
		a--
	} else {
		a = len(h.tab.Panes) - 1
	}
	h.tab.SetActive(a)

	return true
}

var curmacro []interface{}
var recording_macro bool

// ToggleMacro toggles recording of a macro
func (h *BufPane) ToggleMacro() bool {
	recording_macro = !recording_macro
	if recording_macro {
		curmacro = []interface{}{}
		InfoBar.Message("Recording")
	} else {
		InfoBar.Message("Stopped recording")
	}
	h.Relocate()
	return true
}

// PlayMacro plays back the most recently recorded macro
func (h *BufPane) PlayMacro() bool {
	if recording_macro {
		return false
	}
	for _, action := range curmacro {
		switch t := action.(type) {
		case rune:
			h.DoRuneInsert(t)
		case func(*BufPane) bool:
			t(h)
		}
	}
	h.Relocate()
	return true
}

// SpawnMultiCursor creates a new multiple cursor at the next occurrence of the current selection or current word
func (h *BufPane) SpawnMultiCursor() bool {
	spawner := h.Buf.GetCursor(h.Buf.NumCursors() - 1)
	if !spawner.HasSelection() {
		spawner.SelectWord()
		h.multiWord = true
		h.Relocate()
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

	h.Relocate()
	return true
}

// SpawnMultiCursorUp creates additional cursor, at the same X (if possible), one Y less.
func (h *BufPane) SpawnMultiCursorUp() bool {
	if h.Cursor.Y == 0 {
		return false
	} else {
		h.Cursor.GotoLoc(buffer.Loc{h.Cursor.X, h.Cursor.Y - 1})
		h.Cursor.Relocate()
	}

	c := buffer.NewCursor(h.Buf, buffer.Loc{h.Cursor.X, h.Cursor.Y + 1})
	h.Buf.AddCursor(c)
	h.Buf.SetCurCursor(h.Buf.NumCursors() - 1)
	h.Buf.MergeCursors()

	h.Relocate()
	return true
}

// SpawnMultiCursorDown creates additional cursor, at the same X (if possible), one Y more.
func (h *BufPane) SpawnMultiCursorDown() bool {
	if h.Cursor.Y+1 == h.Buf.LinesNum() {
		return false
	} else {
		h.Cursor.GotoLoc(buffer.Loc{h.Cursor.X, h.Cursor.Y + 1})
		h.Cursor.Relocate()
	}

	c := buffer.NewCursor(h.Buf, buffer.Loc{h.Cursor.X, h.Cursor.Y - 1})
	h.Buf.AddCursor(c)
	h.Buf.SetCurCursor(h.Buf.NumCursors() - 1)
	h.Buf.MergeCursors()
	h.Relocate()
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
	return true
}

// MouseMultiCursor is a mouse action which puts a new cursor at the mouse position
func (h *BufPane) MouseMultiCursor(e *tcell.EventMouse) bool {
	b := h.Buf
	mx, my := e.Position()
	mouseLoc := h.LocFromVisual(buffer.Loc{X: mx, Y: my})
	c := buffer.NewCursor(b, mouseLoc)
	b.AddCursor(c)
	b.MergeCursors()

	return true
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
	h.Relocate()
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
	h.Relocate()
	return true
}

// RemoveAllMultiCursors removes all cursors except the base cursor
func (h *BufPane) RemoveAllMultiCursors() bool {
	h.Buf.ClearCursors()
	h.multiWord = false
	h.Relocate()
	return true
}

// None is an action that does nothing
func (h *BufPane) None() bool {
	return true
}
