package main

import (
	"github.com/zyedidia/tcell"
	"regexp"
)

var (
	// What was the last search
	lastSearch string

	// Where should we start the search down from (or up from)
	searchStart int

	// Is there currently a search in progress
	searching bool
)

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

	Search(messenger.response, v, true)

	return
}

// Search searches in the view for the given regex. The down bool
// specifies whether it should search down from the searchStart position
// or up from there
func Search(searchStr string, v *View, down bool) {
	if searchStr == "" {
		return
	}
	var str string
	var charPos int
	if down {
		str = v.buf.text[searchStart:]
		charPos = searchStart
	} else {
		str = v.buf.text[:searchStart]
	}
	r, err := regexp.Compile(searchStr)
	if err != nil {
		return
	}
	matches := r.FindAllStringIndex(str, -1)
	var match []int
	if matches == nil {
		// Search the entire buffer now
		matches = r.FindAllStringIndex(v.buf.text, -1)
		charPos = 0
		if matches == nil {
			v.cursor.ResetSelection()
			return
		}

		if !down {
			match = matches[len(matches)-1]
		} else {
			match = matches[0]
		}
	}

	if !down {
		match = matches[len(matches)-1]
	} else {
		match = matches[0]
	}

	v.cursor.curSelection[0] = charPos + match[0]
	v.cursor.curSelection[1] = charPos + match[1]
	v.cursor.x, v.cursor.y = FromCharPos(charPos+match[1]-1, v.buf)
	if v.Relocate() {
		v.matches = Match(v)
	}
	lastSearch = searchStr
}
