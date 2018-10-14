package main

import (
	"regexp"
	"strings"

	"github.com/zyedidia/tcell"
)

var (
	// What was the last search
	lastSearch string

	// Where should we start the search down from (or up from)
	searchStart Loc

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

// ExitSearch exits the search mode, reset active search phrase, and clear status bar
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
		case tcell.KeyEnter:
			// If the user has pressed Enter, they want this to be the lastSearch
			lastSearch = messenger.response
			EndSearch()
			return
		case tcell.KeyCtrlQ, tcell.KeyCtrlC:
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

	v.Relocate()

	return
}

func searchDown(r *regexp.Regexp, v *View, start, end Loc) bool {
	if start.Y >= v.Buf.NumLines {
		start.Y = v.Buf.NumLines - 1
	}
	if start.Y < 0 {
		start.Y = 0
	}
	for i := start.Y; i <= end.Y; i++ {
		var l []byte
		var charPos int
		if i == start.Y {
			runes := []rune(string(v.Buf.lines[i].data))
			if start.X >= len(runes) {
				start.X = len(runes) - 1
			}
			if start.X < 0 {
				start.X = 0
			}
			l = []byte(string(runes[start.X:]))
			charPos = start.X

			if strings.Contains(r.String(), "^") && start.X != 0 {
				continue
			}
		} else {
			l = v.Buf.lines[i].data
		}

		match := r.FindIndex(l)

		if match != nil {
			v.Cursor.SetSelectionStart(Loc{charPos + runePos(match[0], string(l)), i})
			v.Cursor.SetSelectionEnd(Loc{charPos + runePos(match[1], string(l)), i})
			v.Cursor.OrigSelection[0] = v.Cursor.CurSelection[0]
			v.Cursor.OrigSelection[1] = v.Cursor.CurSelection[1]
			v.Cursor.Loc = v.Cursor.CurSelection[1]

			return true
		}
	}
	return false
}

func searchUp(r *regexp.Regexp, v *View, start, end Loc) bool {
	if start.Y >= v.Buf.NumLines {
		start.Y = v.Buf.NumLines - 1
	}
	if start.Y < 0 {
		start.Y = 0
	}
	for i := start.Y; i >= end.Y; i-- {
		var l []byte
		if i == start.Y {
			runes := []rune(string(v.Buf.lines[i].data))
			if start.X >= len(runes) {
				start.X = len(runes) - 1
			}
			if start.X < 0 {
				start.X = 0
			}
			l = []byte(string(runes[:start.X]))

			if strings.Contains(r.String(), "$") && start.X != Count(string(l)) {
				continue
			}
		} else {
			l = v.Buf.lines[i].data
		}

		match := r.FindIndex(l)

		if match != nil {
			v.Cursor.SetSelectionStart(Loc{runePos(match[0], string(l)), i})
			v.Cursor.SetSelectionEnd(Loc{runePos(match[1], string(l)), i})
			v.Cursor.OrigSelection[0] = v.Cursor.CurSelection[0]
			v.Cursor.OrigSelection[1] = v.Cursor.CurSelection[1]
			v.Cursor.Loc = v.Cursor.CurSelection[1]

			return true
		}
	}
	return false
}

// Search searches in the view for the given regex. The down bool
// specifies whether it should search down from the searchStart position
// or up from there
func Search(searchStr string, v *View, down bool) {
	if searchStr == "" {
		return
	}
	r, err := regexp.Compile(searchStr)
	if v.Buf.Settings["ignorecase"].(bool) {
		r, err = regexp.Compile("(?i)" + searchStr)
	}
	if err != nil {
		return
	}

	var found bool
	if down {
		found = searchDown(r, v, searchStart, v.Buf.End())
		if !found {
			found = searchDown(r, v, v.Buf.Start(), searchStart)
		}
	} else {
		found = searchUp(r, v, searchStart, v.Buf.Start())
		if !found {
			found = searchUp(r, v, v.Buf.End(), searchStart)
		}
	}
	if !found {
		v.Cursor.ResetSelection()
	}
}
