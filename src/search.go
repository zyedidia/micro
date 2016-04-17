package main

import (
	"github.com/gdamore/tcell"
	"regexp"
)

var lastSearch string
var searchStart int

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

func Search(searchStr string, v *View) {
	str := v.buf.text[searchStart:]
	r, err := regexp.Compile(searchStr)
	if err != nil {
		return
	}
	match := r.FindStringIndex(str)
	if match == nil {
		// Search the entire buffer now
		match = r.FindStringIndex(v.buf.text)
		searchStart = 0
		if match == nil {
			v.cursor.ResetSelection()
			return
		}
	}
	v.cursor.curSelection[0] = searchStart + match[0]
	v.cursor.curSelection[1] = searchStart + match[1]
	v.cursor.x, v.cursor.y = FromCharPos(searchStart+match[1]-1, v.buf)
	if v.Relocate() {
		v.matches = Match(v)
	}
	lastSearch = searchStr
}
