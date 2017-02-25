package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/zyedidia/clipboard"
	"github.com/zyedidia/tcell"
)

// TermMessage sends a message to the user in the terminal. This usually occurs before
// micro has been fully initialized -- ie if there is an error in the syntax highlighting
// regular expressions
// The function must be called when the screen is not initialized
// This will write the message, and wait for the user
// to press and key to continue
func TermMessage(msg ...interface{}) {
	screenWasNil := screen == nil
	if !screenWasNil {
		screen.Fini()
		screen = nil
	}

	fmt.Println(msg...)
	fmt.Print("\nPress enter to continue")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	if !screenWasNil {
		InitScreen()
	}
}

// TermError sends an error to the user in the terminal. Like TermMessage except formatted
// as an error
func TermError(filename string, lineNum int, err string) {
	TermMessage(filename + ", " + strconv.Itoa(lineNum) + ": " + err)
}

// Messenger is an object that makes it easy to send messages to the user
// and get input from the user
type Messenger struct {
	log *Buffer
	// Are we currently prompting the user?
	hasPrompt bool
	// Is there a message to print
	hasMessage bool

	// Message to print
	message string
	// The user's response to a prompt
	response string
	// style to use when drawing the message
	style tcell.Style

	// We have to keep track of the cursor for prompting
	cursorx int

	// This map stores the history for all the different kinds of uses Prompt has
	// It's a map of history type -> history array
	history    map[string][]string
	historyNum int

	// Is the current message a message from the gutter
	gutterMessage bool
}

func (m *Messenger) AddLog(msg string) {
	buffer := m.getBuffer()
	buffer.insert(buffer.End(), []byte(msg+"\n"))
	buffer.Cursor.Loc = buffer.End()
	buffer.Cursor.Relocate()
}

func (m *Messenger) getBuffer() *Buffer {
	if m.log == nil {
		m.log = NewBuffer(strings.NewReader(""), "")
		m.log.name = "Log"
	}
	return m.log
}

// Message sends a message to the user
func (m *Messenger) Message(msg ...interface{}) {
	displayMessage := fmt.Sprint(msg...)
	// only display a new message if there isn't an active prompt
	// this is to prevent overwriting an existing prompt to the user
	if m.hasPrompt == false {
		// if there is no active prompt then style and display the message as normal
		m.message = displayMessage

		m.style = defStyle

		if _, ok := colorscheme["message"]; ok {
			m.style = colorscheme["message"]
		}

		m.hasMessage = true
	}
	// add the message to the log regardless of active prompts
	m.AddLog(displayMessage)
}

// Error sends an error message to the user
func (m *Messenger) Error(msg ...interface{}) {
	buf := new(bytes.Buffer)
	fmt.Fprint(buf, msg...)

	// only display a new message if there isn't an active prompt
	// this is to prevent overwriting an existing prompt to the user
	if m.hasPrompt == false {
		// if there is no active prompt then style and display the message as normal
		m.message = buf.String()
		m.style = defStyle.
			Foreground(tcell.ColorBlack).
			Background(tcell.ColorMaroon)

		if _, ok := colorscheme["error-message"]; ok {
			m.style = colorscheme["error-message"]
		}
		m.hasMessage = true
	}
	// add the message to the log regardless of active prompts
	m.AddLog(buf.String())
}

func (m *Messenger) PromptText(msg ...interface{}) {
	displayMessage := fmt.Sprint(msg...)
	// if there is no active prompt then style and display the message as normal
	m.message = displayMessage

	m.style = defStyle

	if _, ok := colorscheme["message"]; ok {
		m.style = colorscheme["message"]
	}

	m.hasMessage = true
	// add the message to the log regardless of active prompts
	m.AddLog(displayMessage)
}

// YesNoPrompt asks the user a yes or no question (waits for y or n) and returns the result
func (m *Messenger) YesNoPrompt(prompt string) (bool, bool) {
	m.hasPrompt = true
	m.PromptText(prompt)

	_, h := screen.Size()
	for {
		m.Clear()
		m.Display()
		screen.ShowCursor(Count(m.message), h-1)
		screen.Show()
		event := <-events

		switch e := event.(type) {
		case *tcell.EventKey:
			switch e.Key() {
			case tcell.KeyRune:
				if e.Rune() == 'y' {
					m.AddLog("\t--> y")
					m.hasPrompt = false
					return true, false
				} else if e.Rune() == 'n' {
					m.AddLog("\t--> n")
					m.hasPrompt = false
					return false, false
				}
			case tcell.KeyCtrlC, tcell.KeyCtrlQ, tcell.KeyEscape:
				m.AddLog("\t--> (cancel)")
				m.hasPrompt = false
				return false, true
			}
		}
	}
}

// LetterPrompt gives the user a prompt and waits for a one letter response
func (m *Messenger) LetterPrompt(prompt string, responses ...rune) (rune, bool) {
	m.hasPrompt = true
	m.PromptText(prompt)

	_, h := screen.Size()
	for {
		m.Clear()
		m.Display()
		screen.ShowCursor(Count(m.message), h-1)
		screen.Show()
		event := <-events

		switch e := event.(type) {
		case *tcell.EventKey:
			switch e.Key() {
			case tcell.KeyRune:
				for _, r := range responses {
					if e.Rune() == r {
						m.AddLog("\t--> " + string(r))
						m.Clear()
						m.Reset()
						m.hasPrompt = false
						return r, false
					}
				}
			case tcell.KeyCtrlC, tcell.KeyCtrlQ, tcell.KeyEscape:
				m.AddLog("\t--> (cancel)")
				m.Clear()
				m.Reset()
				m.hasPrompt = false
				return ' ', true
			}
		}
	}
}

type Completion int

const (
	NoCompletion Completion = iota
	FileCompletion
	CommandCompletion
	HelpCompletion
	OptionCompletion
	PluginCmdCompletion
	PluginNameCompletion
)

// Prompt sends the user a message and waits for a response to be typed in
// This function blocks the main loop while waiting for input
func (m *Messenger) Prompt(prompt, placeholder, historyType string, completionTypes ...Completion) (string, bool) {
	m.hasPrompt = true
	m.PromptText(prompt)
	if _, ok := m.history[historyType]; !ok {
		m.history[historyType] = []string{""}
	} else {
		m.history[historyType] = append(m.history[historyType], "")
	}
	m.historyNum = len(m.history[historyType]) - 1

	response, canceled := placeholder, true
	m.response = response
	m.cursorx = Count(placeholder)

	RedrawAll()
	for m.hasPrompt {
		var suggestions []string
		m.Clear()

		event := <-events

		switch e := event.(type) {
		case *tcell.EventKey:
			switch e.Key() {
			case tcell.KeyCtrlQ, tcell.KeyCtrlC, tcell.KeyEscape:
				// Cancel
				m.AddLog("\t--> (cancel)")
				m.hasPrompt = false
			case tcell.KeyEnter:
				// User is done entering their response
				m.AddLog("\t--> " + m.response)
				m.hasPrompt = false
				response, canceled = m.response, false
				m.history[historyType][len(m.history[historyType])-1] = response
			case tcell.KeyTab:
				args := SplitCommandArgs(m.response)
				currentArgNum := len(args) - 1
				currentArg := args[currentArgNum]
				var completionType Completion

				if completionTypes[0] == CommandCompletion && currentArgNum > 0 {
					if command, ok := commands[args[0]]; ok {
						completionTypes = append([]Completion{CommandCompletion}, command.completions...)
					}
				}

				if currentArgNum >= len(completionTypes) {
					completionType = completionTypes[len(completionTypes)-1]
				} else {
					completionType = completionTypes[currentArgNum]
				}

				var chosen string
				if completionType == FileCompletion {
					chosen, suggestions = FileComplete(currentArg)
				} else if completionType == CommandCompletion {
					chosen, suggestions = CommandComplete(currentArg)
				} else if completionType == HelpCompletion {
					chosen, suggestions = HelpComplete(currentArg)
				} else if completionType == OptionCompletion {
					chosen, suggestions = OptionComplete(currentArg)
				} else if completionType == PluginCmdCompletion {
					chosen, suggestions = PluginCmdComplete(currentArg)
				} else if completionType == PluginNameCompletion {
					chosen, suggestions = PluginNameComplete(currentArg)
				} else if completionType < NoCompletion {
					chosen, suggestions = PluginComplete(completionType, currentArg)
				}

				if len(suggestions) > 1 {
					chosen = chosen + CommonSubstring(suggestions...)
				}

				if chosen != "" {
					m.response = JoinCommandArgs(append(args[:len(args)-1], chosen)...)
					m.cursorx = Count(m.response)
				}
			}
		}

		m.HandleEvent(event, m.history[historyType])

		m.Clear()
		for _, v := range tabs[curTab].views {
			v.Display()
		}
		DisplayTabs()
		m.Display()
		if len(suggestions) > 1 {
			m.DisplaySuggestions(suggestions)
		}
		screen.Show()
	}

	m.Clear()
	m.Reset()
	return response, canceled
}

// HandleEvent handles an event for the prompter
func (m *Messenger) HandleEvent(event tcell.Event, history []string) {
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
							case "main.(*View).CursorUp":
								if m.historyNum > 0 {
									m.historyNum--
									m.response = history[m.historyNum]
									m.cursorx = Count(m.response)
								}
							case "main.(*View).CursorDown":
								if m.historyNum < len(history)-1 {
									m.historyNum++
									m.response = history[m.historyNum]
									m.cursorx = Count(m.response)
								}
							case "main.(*View).CursorLeft":
								if m.cursorx > 0 {
									m.cursorx--
								}
							case "main.(*View).CursorRight":
								if m.cursorx < Count(m.response) {
									m.cursorx++
								}
							case "main.(*View).CursorStart", "main.(*View).StartOfLine":
								m.cursorx = 0
							case "main.(*View).CursorEnd", "main.(*View).EndOfLine":
								m.cursorx = Count(m.response)
							case "main.(*View).Backspace":
								if m.cursorx > 0 {
									m.response = string([]rune(m.response)[:m.cursorx-1]) + string([]rune(m.response)[m.cursorx:])
									m.cursorx--
								}
							case "main.(*View).Paste":
								clip, _ := clipboard.ReadAll("clipboard")
								m.response = Insert(m.response, m.cursorx, clip)
								m.cursorx += Count(clip)
							}
						}
					}
				}
			}
		}
		switch e.Key() {
		case tcell.KeyRune:
			m.response = Insert(m.response, m.cursorx, string(e.Rune()))
			m.cursorx++
		}
		history[m.historyNum] = m.response

	case *tcell.EventPaste:
		clip := e.Text()
		m.response = Insert(m.response, m.cursorx, clip)
		m.cursorx += Count(clip)
	case *tcell.EventMouse:
		x, y := e.Position()
		x -= Count(m.message)
		button := e.Buttons()
		_, screenH := screen.Size()

		if y == screenH-1 {
			switch button {
			case tcell.Button1:
				m.cursorx = x
				if m.cursorx < 0 {
					m.cursorx = 0
				} else if m.cursorx > Count(m.response) {
					m.cursorx = Count(m.response)
				}
			}
		}
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
	w, h := screen.Size()
	for x := 0; x < w; x++ {
		screen.SetContent(x, h-1, ' ', nil, defStyle)
	}
}

func (m *Messenger) DisplaySuggestions(suggestions []string) {
	w, screenH := screen.Size()

	y := screenH - 2

	statusLineStyle := defStyle.Reverse(true)
	if style, ok := colorscheme["statusline"]; ok {
		statusLineStyle = style
	}

	for x := 0; x < w; x++ {
		screen.SetContent(x, y, ' ', nil, statusLineStyle)
	}

	x := 0
	for _, suggestion := range suggestions {
		for _, c := range suggestion {
			screen.SetContent(x, y, c, nil, statusLineStyle)
			x++
		}
		screen.SetContent(x, y, ' ', nil, statusLineStyle)
		x++
	}
}

// Display displays messages or prompts
func (m *Messenger) Display() {
	_, h := screen.Size()
	if m.hasMessage {
		if m.hasPrompt || globalSettings["infobar"].(bool) {
			runes := []rune(m.message + m.response)
			for x := 0; x < len(runes); x++ {
				screen.SetContent(x, h-1, runes[x], nil, m.style)
			}
		}
	}

	if m.hasPrompt {
		screen.ShowCursor(Count(m.message)+m.cursorx, h-1)
		screen.Show()
	}
}

// A GutterMessage is a message displayed on the side of the editor
type GutterMessage struct {
	lineNum int
	msg     string
	kind    int
}

// These are the different types of messages
const (
	// GutterInfo represents a simple info message
	GutterInfo = iota
	// GutterWarning represents a compiler warning
	GutterWarning
	// GutterError represents a compiler error
	GutterError
)
