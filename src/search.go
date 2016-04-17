package main

import (
	"github.com/gdamore/tcell"
	"regexp"
	"strings"
)

var lastSearch string

// BeginSearch starts a search
func BeginSearch() {
	searching = true
	messenger.hasPrompt = true
	messenger.Message("Find: ")
}

// EndSearch stops the current search
func EndSearch() {
	searching = false
	messenger.hasPrompt = false
	messenger.Clear()
	messenger.Reset()
}

// HandleSearchEvent takes an event and a view and will do a real time match from the messenger's output
// to the current buffer. It searches down the buffer.
func HandleSearchEvent(event tcell.Event, v *View) {
	switch e := event.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyCtrlQ, tcell.KeyCtrlC, tcell.KeyEscape, tcell.KeyEnter:
			// Done
			EndSearch()
			return
		}
	}

	messenger.HandleEvent(event)

	if messenger.cursorx < 0 {
		// Done
		EndSearch()
		return
	}

	if messenger.response == "" {
		v.cursor.ResetSelection()
		// We don't end the search though
		return
	}

	Search(messenger.response, v)

	return
}

// Search searches for a given regular expression inside a view
// It also moves the cursor and highlights the search result
func Search(searchStr string, v *View) {
	lines := v.buf.lines[v.cursor.y:]
	var charPos int
	if v.cursor.HasSelection() {
		x, y := FromCharPos(v.cursor.curSelection[1]-1, v.buf)
		lines = v.buf.lines[y:]
		lines[0] = lines[0][x:]
		charPos = ToCharPos(x, y, v.buf)
	} else {
		lines[0] = lines[0][v.cursor.x:]
		charPos = ToCharPos(v.cursor.x, v.cursor.y, v.buf)
	}
	str := strings.Join(lines, "\n")
	r, err := regexp.Compile(searchStr)
	if err != nil {
		return
	}
	match := r.FindStringIndex(str)
	if match == nil {
		// FIXME
		str = strings.Join(v.buf.lines[:v.cursor.y], "\n")
		match = r.FindStringIndex(str)
		charPos = 0
		if match == nil {
			v.cursor.ResetSelection()
			return
		}
	}
	v.cursor.curSelection[0] = charPos + match[0]
	v.cursor.curSelection[1] = charPos + match[1]
	v.cursor.x, v.cursor.y = FromCharPos(charPos+match[1]-1, v.buf)
	if v.Relocate() {
		v.matches = Match(v)
	}
	lastSearch = searchStr
}
