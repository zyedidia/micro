package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/mattn/go-runewidth"
	"github.com/zyedidia/clipboard"
	"github.com/zyedidia/micro/cmd/micro/shellwords"
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

// Password a password
type Password struct {
	Secret   string
	Prompted bool
}

// TermPasswords gets the passwords for the encrypted files
func TermPasswords(filenames []string) []Password {
	passwords := make([]Password, len(filenames))

	for i, filename := range filenames {
		if !Encrypted(filename) {
			continue
		}

		if _, e := os.Stat(filename); e != nil {
			continue
		}

		for {
			fmt.Printf("Password for %v: ", filename)
			password, err := terminal.ReadPassword(syscall.Stdin)
			fmt.Printf("\n")
			if err != nil {
				fmt.Println(err)
				continue
			}
			passwords[i].Secret = string(password)
			passwords[i].Prompted = true
			break
		}
	}

	return passwords
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

// AddLog sends a message to the log view
func (m *Messenger) AddLog(msg ...interface{}) {
	logMessage := fmt.Sprint(msg...)
	buffer := m.getBuffer()
	buffer.insert(buffer.End(), []byte(logMessage+"\n"))
	buffer.Cursor.Loc = buffer.End()
	buffer.Cursor.Relocate()
}

func (m *Messenger) getBuffer() *Buffer {
	if m.log == nil {
		m.log = NewBufferFromString("", "")
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
				if e.Rune() == 'y' || e.Rune() == 'Y' {
					m.AddLog("\t--> y")
					m.hasPrompt = false
					return true, false
				} else if e.Rune() == 'n' || e.Rune() == 'N' {
					m.AddLog("\t--> n")
					m.hasPrompt = false
					return false, false
				}
			case tcell.KeyCtrlC, tcell.KeyCtrlQ, tcell.KeyEscape:
				m.AddLog("\t--> (cancel)")
				m.Clear()
				m.Reset()
				m.hasPrompt = false
				return false, true
			}
		}
	}
}

// PasswordPrompt asks the user for a password and returns the result
func (m *Messenger) PasswordPrompt(verify bool) (string, bool) {
	passwordPrompt := func(prompt string) (string, bool) {
		password := ""
		m.hasPrompt = true
		m.PromptText(prompt)
		m.Clear()
		m.Display()
		defer func() {
			m.Clear()
			m.Reset()
			m.hasPrompt = false
		}()
		for {
			event := <-events
			switch e := event.(type) {
			case *tcell.EventKey:
				switch e.Key() {
				case tcell.KeyRune:
					password += string(e.Rune())
				case tcell.KeyEnter:
					m.AddLog("\t--> (enter)")
					return password, false
				case tcell.KeyCtrlC, tcell.KeyCtrlQ, tcell.KeyEscape:
					m.AddLog("\t--> (cancel)")
					return "", true
				}
			}
		}
	}

	for verify {
		password, canceled := passwordPrompt("Password: ")
		if canceled {
			return "", true
		}
		verifyPassword, canceled := passwordPrompt("Verify Password: ")
		if canceled {
			return "", true
		}
		if password == verifyPassword {
			return password, false
		}
	}

	return passwordPrompt("Password: ")
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

// Completion represents a type of completion
type Completion int

const (
	NoCompletion Completion = iota
	FileCompletion
	CommandCompletion
	HelpCompletion
	OptionCompletion
	PluginCmdCompletion
	PluginNameCompletion
	OptionValueCompletion
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
		case *tcell.EventResize:
			for _, t := range tabs {
				t.Resize()
			}
			RedrawAll()
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
				args, err := shellwords.Split(m.response)
				if err != nil {
					break
				}
				currentArg := ""
				currentArgNum := 0
				if len(args) > 0 {
					currentArgNum = len(args) - 1
					currentArg = args[currentArgNum]
				}
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
				} else if completionType == OptionValueCompletion {
					if currentArgNum-1 > 0 {
						chosen, suggestions = OptionValueComplete(args[currentArgNum-1], currentArg)
					}
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

				if len(suggestions) != 0 && chosen != "" {
					m.response = shellwords.Join(append(args[:len(args)-1], chosen)...)
					m.cursorx = Count(m.response)
				}
			}
		}

		m.HandleEvent(event, m.history[historyType])

		m.Clear()
		for _, v := range tabs[curTab].Views {
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

// UpHistory fetches the previous item in the history
func (m *Messenger) UpHistory(history []string) {
	if m.historyNum > 0 {
		m.historyNum--
		m.response = history[m.historyNum]
		m.cursorx = Count(m.response)
	}
}

// DownHistory fetches the next item in the history
func (m *Messenger) DownHistory(history []string) {
	if m.historyNum < len(history)-1 {
		m.historyNum++
		m.response = history[m.historyNum]
		m.cursorx = Count(m.response)
	}
}

// CursorLeft moves the cursor one character left
func (m *Messenger) CursorLeft() {
	if m.cursorx > 0 {
		m.cursorx--
	}
}

// CursorRight moves the cursor one character right
func (m *Messenger) CursorRight() {
	if m.cursorx < Count(m.response) {
		m.cursorx++
	}
}

// Start moves the cursor to the start of the line
func (m *Messenger) Start() {
	m.cursorx = 0
}

// End moves the cursor to the end of the line
func (m *Messenger) End() {
	m.cursorx = Count(m.response)
}

// Backspace deletes one character
func (m *Messenger) Backspace() {
	if m.cursorx > 0 {
		m.response = string([]rune(m.response)[:m.cursorx-1]) + string([]rune(m.response)[m.cursorx:])
		m.cursorx--
	}
}

// Paste pastes the clipboard
func (m *Messenger) Paste() {
	clip, _ := clipboard.ReadAll("clipboard")
	m.response = Insert(m.response, m.cursorx, clip)
	m.cursorx += Count(clip)
}

// WordLeft moves the cursor one word to the left
func (m *Messenger) WordLeft() {
	response := []rune(m.response)
	m.CursorLeft()
	if m.cursorx <= 0 {
		return
	}
	for IsWhitespace(response[m.cursorx]) {
		if m.cursorx <= 0 {
			return
		}
		m.CursorLeft()
	}
	m.CursorLeft()
	for IsWordChar(string(response[m.cursorx])) {
		if m.cursorx <= 0 {
			return
		}
		m.CursorLeft()
	}
	m.CursorRight()
}

// WordRight moves the cursor one word to the right
func (m *Messenger) WordRight() {
	response := []rune(m.response)
	if m.cursorx >= len(response) {
		return
	}
	for IsWhitespace(response[m.cursorx]) {
		m.CursorRight()
		if m.cursorx >= len(response) {
			m.CursorRight()
			return
		}
	}
	m.CursorRight()
	if m.cursorx >= len(response) {
		return
	}
	for IsWordChar(string(response[m.cursorx])) {
		m.CursorRight()
		if m.cursorx >= len(response) {
			return
		}
	}
}

// DeleteWordLeft deletes one word to the left
func (m *Messenger) DeleteWordLeft() {
	m.WordLeft()
	m.response = string([]rune(m.response)[:m.cursorx])
}

// HandleEvent handles an event for the prompter
func (m *Messenger) HandleEvent(event tcell.Event, history []string) {
	switch e := event.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyCtrlA:
			m.Start()
		case tcell.KeyCtrlE:
			m.End()
		case tcell.KeyUp:
			m.UpHistory(history)
		case tcell.KeyDown:
			m.DownHistory(history)
		case tcell.KeyLeft:
			if e.Modifiers() == tcell.ModCtrl {
				m.Start()
			} else if e.Modifiers() == tcell.ModAlt || e.Modifiers() == tcell.ModMeta {
				m.WordLeft()
			} else {
				m.CursorLeft()
			}
		case tcell.KeyRight:
			if e.Modifiers() == tcell.ModCtrl {
				m.End()
			} else if e.Modifiers() == tcell.ModAlt || e.Modifiers() == tcell.ModMeta {
				m.WordRight()
			} else {
				m.CursorRight()
			}
		case tcell.KeyBackspace2, tcell.KeyBackspace:
			if e.Modifiers() == tcell.ModCtrl || e.Modifiers() == tcell.ModAlt || e.Modifiers() == tcell.ModMeta {
				m.DeleteWordLeft()
			} else {
				m.Backspace()
			}
		case tcell.KeyCtrlW:
			m.DeleteWordLeft()
		case tcell.KeyCtrlV:
			m.Paste()
		case tcell.KeyCtrlF:
			m.WordRight()
		case tcell.KeyCtrlB:
			m.WordLeft()
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
			posx := 0
			for x := 0; x < len(runes); x++ {
				screen.SetContent(posx, h-1, runes[x], nil, m.style)
				posx += runewidth.RuneWidth(runes[x])
			}
		}
	}

	if m.hasPrompt {
		screen.ShowCursor(Count(m.message)+m.cursorx, h-1)
		screen.Show()
	}
}

// LoadHistory attempts to load user history from configDir/buffers/history
// into the history map
// The savehistory option must be on
func (m *Messenger) LoadHistory() {
	if GetGlobalOption("savehistory").(bool) {
		file, err := os.Open(configDir + "/buffers/history")
		defer file.Close()
		var decodedMap map[string][]string
		if err == nil {
			decoder := gob.NewDecoder(file)
			err = decoder.Decode(&decodedMap)

			if err != nil {
				m.Error("Error loading history:", err)
				return
			}
		}

		if decodedMap != nil {
			m.history = decodedMap
		} else {
			m.history = make(map[string][]string)
		}
	} else {
		m.history = make(map[string][]string)
	}
}

// SaveHistory saves the user's command history to configDir/buffers/history
// only if the savehistory option is on
func (m *Messenger) SaveHistory() {
	if GetGlobalOption("savehistory").(bool) {
		// Don't save history past 100
		for k, v := range m.history {
			if len(v) > 100 {
				m.history[k] = v[len(m.history[k])-100:]
			}
		}

		file, err := os.Create(configDir + "/buffers/history")
		defer file.Close()
		if err == nil {
			encoder := gob.NewEncoder(file)

			err = encoder.Encode(m.history)
			if err != nil {
				m.Error("Error saving history:", err)
				return
			}
		}
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
