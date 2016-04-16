package main

import (
	"github.com/gdamore/tcell"
	"regexp"
	"strings"
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

	str := strings.Join(v.buf.lines[v.cursor.y:], "\n")
	charPos := ToCharPos(0, v.cursor.y, v.buf)
	r, err := regexp.Compile(messenger.response)
	if err != nil {
		return
	}
	match := r.FindStringIndex(str)
	if match == nil {
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
	return
}
