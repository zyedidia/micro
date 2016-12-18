package main

import (
	"regexp"

	"github.com/zyedidia/tcell"
)

var (
	// What was the last search
	lastSearch string

	// Where should we start the search down from (or up from)
	searchStart int

	// Is there currently a search in progress
	searching bool

	// Stores the history for searching
	searchHistory []string
)

// BeginSearch starts a search
func BeginSearch(searchStr string) {
	searchHistory = append(searchHistory, "")
	messenger.historyNum = len(searchHistory) - 1
	searching = true
	messenger.response = searchStr
	messenger.cursorx = Count(searchStr)
	messenger.Message("Find: ")
	messenger.hasPrompt = true
}

// EndSearch stops the current search
func EndSearch() {
	searchHistory[len(searchHistory)-1] = messenger.response
	searching = false
	messenger.hasPrompt = false
	messenger.Clear()
	messenger.Reset()
	if lastSearch != "" {
		messenger.Message("^P Previous ^N Next")
	}
}

// exit the search mode, reset active search phrase, and clear status bar
func ExitSearch(v *View) {
	lastSearch = ""
	searching = false
	messenger.hasPrompt = false
	messenger.Clear()
	messenger.Reset()
	v.Cursor.ResetSelection()
}

// HandleSearchEvent takes an event and a view and will do a real time match from the messenger's output
// to the current buffer. It searches down the buffer.
func HandleSearchEvent(event tcell.Event, v *View) {
	switch e := event.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyEscape:
			// Exit the search mode
			ExitSearch(v)
			return
		case tcell.KeyCtrlQ, tcell.KeyCtrlC, tcell.KeyEnter:
			// Done
			EndSearch()
			return
		}
	}

	messenger.HandleEvent(event, searchHistory)

	if messenger.cursorx < 0 {
		// Done
		EndSearch()
		return
	}

	if messenger.response == "" {
		v.Cursor.ResetSelection()
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
	text := v.Buf.String()
	if down {
		str = string([]rune(text)[searchStart:])
		charPos = searchStart
	} else {
		str = string([]rune(text)[:searchStart])
	}
	r, err := regexp.Compile(searchStr)
	if v.Buf.Settings["ignorecase"].(bool) {
		r, err = regexp.Compile("(?i)" + searchStr)
	}
	if err != nil {
		return
	}
	matches := r.FindAllStringIndex(str, -1)
	var match []int
	if matches == nil {
		// Search the entire buffer now
		matches = r.FindAllStringIndex(text, -1)
		charPos = 0
		if matches == nil {
			v.Cursor.ResetSelection()
			return
		}

		if !down {
			match = matches[len(matches)-1]
		} else {
			match = matches[0]
		}
		str = text
	}

	if !down {
		match = matches[len(matches)-1]
	} else {
		match = matches[0]
	}

	if match[0] == match[1] {
		return
	}

	v.Cursor.SetSelectionStart(FromCharPos(charPos+runePos(match[0], str), v.Buf))
	v.Cursor.SetSelectionEnd(FromCharPos(charPos+runePos(match[1], str), v.Buf))
	v.Cursor.Loc = v.Cursor.CurSelection[1]
	if v.Relocate() {
		v.matches = Match(v)
	}
	lastSearch = searchStr
}
