package display

import (
	"fmt"
	"strings"

	runewidth "github.com/mattn/go-runewidth"
	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/tcell"
)

type InfoBar struct {
	*buffer.Buffer

	hasPrompt  bool
	hasMessage bool

	message string
	// style to use when drawing the message
	style tcell.Style

	width int
	y     int

	// This map stores the history for all the different kinds of uses Prompt has
	// It's a map of history type -> history array
	history    map[string][]string
	historyNum int

	// Is the current message a message from the gutter
	gutterMessage bool
}

func NewInfoBar() *InfoBar {
	ib := new(InfoBar)
	ib.style = config.DefStyle
	ib.history = make(map[string][]string)

	ib.Buffer = buffer.NewBufferFromString("", "infobar")
	ib.Type = buffer.BTScratch

	ib.width, ib.y = screen.Screen.Size()

	return ib
}

func (i *InfoBar) Clear() {
	for x := 0; x < i.width; x++ {
		screen.Screen.SetContent(x, i.y, ' ', nil, config.DefStyle)
	}
}

func (i *InfoBar) Display() {
	x := 0
	if i.hasPrompt || config.GlobalSettings["infobar"].(bool) {
		display := i.message + strings.TrimSpace(string(i.Bytes()))
		for _, c := range display {
			screen.Screen.SetContent(x, i.y, c, nil, i.style)
			x += runewidth.RuneWidth(c)
		}
	}
}

// Message sends a message to the user
func (i *InfoBar) Message(msg ...interface{}) {
	displayMessage := fmt.Sprint(msg...)
	// only display a new message if there isn't an active prompt
	// this is to prevent overwriting an existing prompt to the user
	if i.hasPrompt == false {
		// if there is no active prompt then style and display the message as normal
		i.message = displayMessage
		i.style = config.DefStyle

		if _, ok := config.Colorscheme["message"]; ok {
			i.style = config.Colorscheme["message"]
		}

		i.hasMessage = true
	}
}
