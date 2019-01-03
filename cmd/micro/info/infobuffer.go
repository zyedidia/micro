package info

import (
	"fmt"
	"strings"

	"github.com/zyedidia/micro/cmd/micro/buffer"
)

// The InfoBuf displays messages and other info at the bottom of the screen.
// It is respresented as a buffer and a message with a style.
type InfoBuf struct {
	*buffer.Buffer

	HasPrompt  bool
	HasMessage bool
	HasError   bool

	Msg string

	// This map stores the history for all the different kinds of uses Prompt has
	// It's a map of history type -> history array
	History    map[string][]string
	HistoryNum int

	// Is the current message a message from the gutter
	GutterMessage bool

	PromptCallback func(resp string, canceled bool)
}

func NewBuffer() *InfoBuf {
	ib := new(InfoBuf)
	ib.History = make(map[string][]string)

	ib.Buffer = buffer.NewBufferFromString("", "infobar", buffer.BTInfo)

	return ib
}

// Message sends a message to the user
func (i *InfoBuf) Message(msg ...interface{}) {
	// only display a new message if there isn't an active prompt
	// this is to prevent overwriting an existing prompt to the user
	if i.HasPrompt == false {
		displayMessage := fmt.Sprint(msg...)
		// if there is no active prompt then style and display the message as normal
		i.Msg = displayMessage
		i.HasMessage = true
	}
}

// Error sends an error message to the user
func (i *InfoBuf) Error(msg ...interface{}) {
	// only display a new message if there isn't an active prompt
	// this is to prevent overwriting an existing prompt to the user
	if i.HasPrompt == false {
		// if there is no active prompt then style and display the message as normal
		i.Msg = fmt.Sprint(msg...)
		i.HasMessage, i.HasError = false, true
	}
	// TODO: add to log?
}

func (i *InfoBuf) Prompt(prompt string, msg string, callback func(string, bool)) {
	// If we get another prompt mid-prompt we cancel the one getting overwritten
	if i.HasPrompt {
		i.DonePrompt(true)
	}

	i.Msg = prompt
	i.HasPrompt = true
	i.HasMessage, i.HasError = false, false
	i.PromptCallback = callback
	i.Buffer.Insert(i.Buffer.Start(), msg)
}

func (i *InfoBuf) DonePrompt(canceled bool) {
	i.HasPrompt = false
	if canceled {
		i.PromptCallback("", true)
	} else {
		i.PromptCallback(strings.TrimSpace(string(i.LineBytes(0))), false)
	}
	i.Replace(i.Start(), i.End(), "")
}

// Reset resets the messenger's cursor, message and response
func (i *InfoBuf) Reset() {
	i.Msg = ""
	i.HasPrompt, i.HasMessage, i.HasError = false, false, false
}
