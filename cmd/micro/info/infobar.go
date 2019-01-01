package info

import (
	"fmt"

	"github.com/zyedidia/micro/cmd/micro/buffer"
)

var MainBar *Bar

func InitInfoBar() {
	MainBar = NewBar()
}

// The Bar displays messages and other info at the bottom of the screen.
// It is respresented as a buffer and a message with a style.
type Bar struct {
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
}

func NewBar() *Bar {
	ib := new(Bar)
	ib.History = make(map[string][]string)

	ib.Buffer = buffer.NewBufferFromString("", "infobar", buffer.BTScratch)

	return ib
}

// Message sends a message to the user
func (i *Bar) Message(msg ...interface{}) {
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
func (i *Bar) Error(msg ...interface{}) {
	// only display a new message if there isn't an active prompt
	// this is to prevent overwriting an existing prompt to the user
	if i.HasPrompt == false {
		// if there is no active prompt then style and display the message as normal
		i.Msg = fmt.Sprint(msg...)
		i.HasError = true
	}
	// TODO: add to log?
}
