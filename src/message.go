package main

import (
	"github.com/gdamore/tcell"
)

// Messenger is an object that can send messages to the user and get input from the user (with a prompt)
type Messenger struct {
	hasPrompt  bool
	hasMessage bool

	message  string
	response string
	style    tcell.Style

	cursorx int

	s tcell.Screen
}

// NewMessenger returns a new Messenger struct
func NewMessenger(s tcell.Screen) *Messenger {
	m := new(Messenger)
	m.s = s
	return m
}

// Message sends a message to the user
func (m *Messenger) Message(msg string) {
	m.message = msg
	m.style = tcell.StyleDefault

	if _, ok := colorscheme["message"]; ok {
		m.style = colorscheme["message"]
	}
	m.hasMessage = true
}

// Error sends an error message to the user
func (m *Messenger) Error(msg string) {
	m.message = msg
	m.style = tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorMaroon)

	if _, ok := colorscheme["error-message"]; ok {
		m.style = colorscheme["error-message"]
	}
	m.hasMessage = true
}

// Prompt sends the user a message and waits for a response to be typed in
// This function blocks the main loop while waiting for input
func (m *Messenger) Prompt(prompt string) (string, bool) {
	m.hasPrompt = true
	m.Message(prompt)

	response, canceled := "", true

	for m.hasPrompt {
		m.Clear()
		m.Display()

		event := m.s.PollEvent()

		switch e := event.(type) {
		case *tcell.EventKey:
			if e.Key() == tcell.KeyEscape {
				// Cancel
				m.hasPrompt = false
			} else if e.Key() == tcell.KeyCtrlC {
				// Cancel
				m.hasPrompt = false
			} else if e.Key() == tcell.KeyCtrlQ {
				// Cancel
				m.hasPrompt = false
			} else if e.Key() == tcell.KeyEnter {
				// User is done entering their response
				m.hasPrompt = false
				response, canceled = m.response, false
			}
		}

		m.HandleEvent(event)
	}

	m.Reset()
	return response, canceled
}

// HandleEvent handles an event for the prompter
func (m *Messenger) HandleEvent(event tcell.Event) {
	switch e := event.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyLeft:
			m.cursorx--
		case tcell.KeyRight:
			m.cursorx++
		case tcell.KeyBackspace2:
			if m.cursorx > 0 {
				m.response = string([]rune(m.response)[:Count(m.response)-1])
				m.cursorx--
			}
		case tcell.KeySpace:
			m.response += " "
			m.cursorx++
		case tcell.KeyRune:
			m.response += string(e.Rune())
			m.cursorx++
		}
	}
	if m.cursorx < 0 {
		m.cursorx = 0
	}
}

// Reset resets the messenger's cursor, message and response
func (m *Messenger) Reset() {
	m.cursorx = 0
	m.message = ""
	m.response = ""
}

// Clear clears the line at the bottom of the editor
func (m *Messenger) Clear() {
	w, h := m.s.Size()
	for x := 0; x < w; x++ {
		m.s.SetContent(x, h-1, ' ', nil, tcell.StyleDefault)
	}
}

// Display displays and messages or prompts
func (m *Messenger) Display() {
	_, h := m.s.Size()
	if m.hasMessage {
		runes := []rune(m.message + m.response)
		for x := 0; x < len(runes); x++ {
			m.s.SetContent(x, h-1, runes[x], nil, m.style)
		}
	}
	if m.hasPrompt {
		m.s.ShowCursor(Count(m.message)+m.cursorx, h-1)
		m.s.Show()
	}
}
