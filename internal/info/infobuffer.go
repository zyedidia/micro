package info

import (
	"fmt"

	"github.com/zyedidia/micro/v2/internal/buffer"
)

// The InfoBuf displays messages and other info at the bottom of the screen.
// It is represented as a buffer and a message with a style.
type InfoBuf struct {
	*buffer.Buffer

	HasPrompt  bool
	HasMessage bool
	HasError   bool
	HasYN      bool

	PromptType string

	Msg    string
	YNResp bool

	// This map stores the history for all the different kinds of uses Prompt has
	// It's a map of history type -> history array
	History    map[string][]string
	HistoryNum int
	// HistorySearch indicates whether we are searching for history items
	// beginning with HistorySearchPrefix
	HistorySearch       bool
	HistorySearchPrefix string

	// Is the current message a message from the gutter
	HasGutter bool

	PromptCallback func(resp string, canceled bool)
	EventCallback  func(resp string)
	YNCallback     func(yes bool, canceled bool)
}

// NewBuffer returns a new infobuffer
func NewBuffer() *InfoBuf {
	ib := new(InfoBuf)
	ib.History = make(map[string][]string)

	ib.Buffer = buffer.NewBufferFromString("", "", buffer.BTInfo)
	ib.LoadHistory()

	return ib
}

// Close performs any cleanup necessary when shutting down the infobuffer
func (i *InfoBuf) Close() {
	i.SaveHistory()
}

// Message sends a message to the user
func (i *InfoBuf) Message(msg ...interface{}) {
	// only display a new message if there isn't an active prompt
	// this is to prevent overwriting an existing prompt to the user
	if !i.HasPrompt {
		displayMessage := fmt.Sprint(msg...)
		// if there is no active prompt then style and display the message as normal
		i.Msg = displayMessage
		i.HasMessage, i.HasError = true, false
	}
}

// GutterMessage displays a message and marks it as a gutter message
func (i *InfoBuf) GutterMessage(msg ...interface{}) {
	i.Message(msg...)
	i.HasGutter = true
}

// ClearGutter clears the info bar and unmarks the message
func (i *InfoBuf) ClearGutter() {
	i.HasGutter = false
	i.Message("")
}

// Error sends an error message to the user
func (i *InfoBuf) Error(msg ...interface{}) {
	// only display a new message if there isn't an active prompt
	// this is to prevent overwriting an existing prompt to the user
	if !i.HasPrompt {
		// if there is no active prompt then style and display the message as normal
		i.Msg = fmt.Sprint(msg...)
		i.HasMessage, i.HasError = false, true
	}
	// TODO: add to log?
}

// Prompt starts a prompt for the user, it takes a prompt, a possibly partially filled in msg
// and callbacks executed when the user executes an event and when the user finishes the prompt
// The eventcb passes the current user response as the argument and donecb passes the user's message
// and a boolean indicating if the prompt was canceled
func (i *InfoBuf) Prompt(prompt string, msg string, ptype string, eventcb func(string), donecb func(string, bool)) {
	// If we get another prompt mid-prompt we cancel the one getting overwritten
	if i.HasPrompt {
		i.DonePrompt(true)
	}

	if _, ok := i.History[ptype]; !ok {
		i.History[ptype] = []string{""}
	} else {
		i.History[ptype] = append(i.History[ptype], "")
	}
	i.HistoryNum = len(i.History[ptype]) - 1
	i.HistorySearch = false

	i.PromptType = ptype
	i.Msg = prompt
	i.HasPrompt = true
	i.HasMessage, i.HasError, i.HasYN = false, false, false
	i.HasGutter = false
	i.PromptCallback = donecb
	i.EventCallback = eventcb
	i.Buffer.Insert(i.Buffer.Start(), msg)
}

// YNPrompt creates a yes or no prompt, and the callback returns the yes/no result and whether
// the prompt was canceled
func (i *InfoBuf) YNPrompt(prompt string, donecb func(bool, bool)) {
	if i.HasPrompt {
		i.DonePrompt(true)
	}

	i.Msg = prompt
	i.HasPrompt = true
	i.HasYN = true
	i.HasMessage, i.HasError = false, false
	i.HasGutter = false
	i.YNCallback = donecb
}

// DonePrompt finishes the current prompt and indicates whether or not it was canceled
func (i *InfoBuf) DonePrompt(canceled bool) {
	hadYN := i.HasYN
	i.HasPrompt = false
	i.HasYN = false
	i.HasGutter = false
	if !hadYN {
		if i.PromptCallback != nil {
			if canceled {
				i.Replace(i.Start(), i.End(), "")
				h := i.History[i.PromptType]
				i.History[i.PromptType] = h[:len(h)-1]
				i.PromptCallback("", true)
			} else {
				resp := string(i.LineBytes(0))
				i.Replace(i.Start(), i.End(), "")
				h := i.History[i.PromptType]
				h[len(h)-1] = resp

				// avoid duplicates
				for j := len(h) - 2; j >= 0; j-- {
					if h[j] == h[len(h)-1] {
						i.History[i.PromptType] = append(h[:j], h[j+1:]...)
						break
					}
				}

				i.PromptCallback(resp, false)
			}
			// i.PromptCallback = nil
		}
	}
	if i.YNCallback != nil && hadYN {
		i.YNCallback(i.YNResp, canceled)
	}
}

// Reset resets the infobuffer's msg and info
func (i *InfoBuf) Reset() {
	i.Msg = ""
	i.HasPrompt, i.HasMessage, i.HasError = false, false, false
	i.HasGutter = false
}
