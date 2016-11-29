package main

import (
	"github.com/zyedidia/clipboard"
	"github.com/zyedidia/tcell"
	"index/suffixarray"
	"sort"
	"strings"
)

// Display autocompletions in a box above or below the cursor
type AutocompletionBox struct {
	open       bool
	showPrompt bool
	width      int
	// Message to print
	message string
	// The user's response to a prompt
	response string
	cursorx  int
	// style to use when drawing the message
	style tcell.Style

	// We have to keep track of the cursor for selecting
	cursory int

	messages       Messages
	messagesToshow Messages

	selected int

	Index       suffixarray.Index
	AcceptEnter AcceptFcn
	AcceptTab   AcceptFcn
	Pop         PopulateFcn
}
type AcceptFcn func(message Message)
type PopulateFcn func(v *View) (messages Messages)
type Message struct {
	// Searchable is the target of search
	Searchable string
	// MessageToDisplay is the string inside the box
	MessageToDisplay string
	// Value2 is used as a return type for accept
	Value2 []byte
}
type Messages []Message

func (s Messages) Len() int {
	return len(s)
}
func (s Messages) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Messages) Less(i, j int) bool {
	return s[i].MessageToDisplay < s[j].MessageToDisplay
}

//Open opens a box with prompt
func (a *AutocompletionBox) Open(pop PopulateFcn, acceptEnter, acceptTab AcceptFcn, v *View) {
	a.Pop = pop
	a.generateAutocomplete(v)
	a.open = true
	a.showPrompt = true
	a.AcceptEnter = acceptEnter
	a.AcceptTab = acceptTab
}

//OpenNoPrompt opens a box with no prompt. Typing will cause the box to move.
func (a *AutocompletionBox) OpenNoPrompt(pop PopulateFcn, acceptEnter, acceptTab AcceptFcn, v *View) {
	a.Pop = pop
	a.generateAutocomplete(v)
	a.open = true
	a.showPrompt = false
	a.AcceptEnter = acceptEnter
	a.AcceptTab = acceptTab
}

func (a *AutocompletionBox) generateAutocomplete(v *View) {
	a.messages = a.Pop(v)
	for _, message := range a.messages {
		if message.Searchable == "" {
			message.Searchable = message.MessageToDisplay
		}
		a.width = Max(a.width, Count(message.MessageToDisplay))
	}
	sort.Sort(a.messages)
	a.filterAutocomplete()
}

// Display autocompletionbox
func (a *AutocompletionBox) Display(v *View) {
	if !a.open {
		return
	}
	h := 0
	if a.showPrompt {
		screen.ShowCursor(a.cursorx+cursorGX, cursorGY)
		a.style = tcell.StyleDefault
		a.style = a.style.Foreground(tcell.ColorYellow).Background(tcell.ColorBlue)
		runes := []rune(a.response)
		for x := 0; x < a.width; x++ {
			i := rune(' ')
			if x < len(runes) {
				i = runes[x]
			}
			screen.SetContent(cursorGX+x, cursorGY, i, nil, a.style)
		}
	}
	a.selected = Min(a.selected, len(a.messagesToshow)-1)

	skipped := Max(0, a.selected-5)
	messages := a.messagesToshow[skipped:]

	for i, message := range messages[:Min(len(messages), 11)] {
		runes := []rune(message.MessageToDisplay)
		var j int
		var indexes []int
		for _, r := range a.response {
			j = j + strings.IndexRune(message.MessageToDisplay[j:], r)
			indexes = append(indexes, j)
		}
		for x := 0; x < a.width; x++ {
			if i == a.selected-skipped {
				a.style = defStyle
			} else {
				a.style = defStyle.Reverse(true)
			}
			for _, value := range indexes {
				if value == x {
					a.style = a.style.Foreground(tcell.ColorYellow)
					break
				}
			}

			i := rune(' ')
			if x < len(runes) {
				i = runes[x]
			}
			screen.SetContent(cursorGX+x, cursorGY+h+1, i, nil, a.style)
		}
		h++
	}
}

func (a *AutocompletionBox) Reset() {
	a.selected = 0
	a.response = ""
	a.cursorx = 0
	a.cursory = 0
	a.open = false
	a.messages = a.messages[:0]
	a.messagesToshow = a.messagesToshow[:0]
	a.AcceptTab = nil
	a.AcceptEnter = nil
	a.Pop = nil
}

func (a *AutocompletionBox) filterAutocomplete() {
	mess := Messages{}
	for _, message := range a.messages {

		var j int
		var notFound bool
		for _, r := range a.response {
			i := strings.IndexRune(message.Searchable[j:], r)
			if i == -1 {
				notFound = true
				break
			}
			j += i
		}
		if !notFound {
			mess = append(mess, message)
		}

	}
	a.messagesToshow = mess
	if a.selected < 0 && len(a.messages) > 0 {
		a.selected = 0
	}
	if a.selected-1 > len(a.messages) {
		a.selected = len(a.message) - 1
	}
}

// HandleEvent handles an event for the prompter
func (a *AutocompletionBox) HandleEvent(event tcell.Event, v *View) (swallow bool) {
	switch e := event.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyEnter:
			if a.AcceptEnter != nil {
				if len(a.messagesToshow) > a.selected && len(a.messagesToshow) > 0 {
					a.AcceptEnter(a.messagesToshow[a.selected])
				}
				a.Reset()
			}
			return true
		case tcell.KeyTAB:
			if a.AcceptTab != nil {
				if len(a.messagesToshow) > a.selected {
					a.AcceptTab(a.messagesToshow[a.selected])
				}
				a.Reset()
			}
			return true
		case tcell.KeyESC:
			a.Reset()
			return true
		case tcell.KeyUp:
			if autocomplete.selected > 0 {
				autocomplete.selected--
			}
			return true
		case tcell.KeyDown:
			if len(autocomplete.messagesToshow)-1 > autocomplete.selected {
				autocomplete.selected++
			}
			return true
		}
	}
	if !a.showPrompt {
		return false
	}
	switch e := event.(type) {
	case *tcell.EventKey:
		if e.Key() != tcell.KeyRune || e.Modifiers() != 0 {
			for key, actions := range bindings {
				if e.Key() == key.keyCode {
					if e.Key() == tcell.KeyRune {
						if e.Rune() != key.r {
							continue
						}
					}
					if e.Modifiers() == key.modifiers {
						for _, action := range actions {
							funcName := FuncName(action)
							switch funcName {
							case "main.(*View).CursorLeft":
								if a.cursorx > 0 {
									a.cursorx--
								}
							case "main.(*View).CursorRight":
								if a.cursorx < Count(a.response) {
									a.cursorx++
								}
							case "main.(*View).CursorStart", "main.(*View).StartOfLine":
								a.cursorx = 0
							case "main.(*View).CursorEnd", "main.(*View).EndOfLine":
								a.cursorx = Count(a.response)
							case "main.(*View).Backspace":
								if a.cursorx > 0 {
									a.response = string([]rune(a.response)[:a.cursorx-1]) + string([]rune(a.response)[a.cursorx:])
									a.cursorx--
								}
							case "main.(*View).Paste":
								clip, _ := clipboard.ReadAll("clipboard")
								a.response = Insert(a.response, a.cursorx, clip)
								a.cursorx += Count(clip)
							}
						}
					}
				}
			}
		}
		switch e.Key() {
		case tcell.KeyRune:
			a.response = Insert(a.response, a.cursorx, string(e.Rune()))
			a.cursorx++
		}

	case *tcell.EventPaste:
		clip := e.Text()
		a.response = Insert(a.response, a.cursorx, clip)
		a.cursorx += Count(clip)
	case *tcell.EventMouse:
		x, y := e.Position()
		button := e.Buttons()
		if y == cursorGY {
			switch button {
			case tcell.Button1:
				a.cursorx = x
				if a.cursorx < 0 {
					a.cursorx = 0
				} else if a.cursorx > Count(a.response) {
					a.cursorx = Count(a.response)
				}
			}
		}
	}
	a.filterAutocomplete()
	return true
}
