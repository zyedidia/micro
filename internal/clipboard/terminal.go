package clipboard

import (
	"errors"
	"time"

	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/tcell/v2"
)

type terminalClipboard struct{}

var terminal terminalClipboard

func (t terminalClipboard) read(reg string) (string, error) {
	screen.Screen.GetClipboard(reg)
	// wait at most 200ms for response
	for {
		select {
		case event := <-screen.Events:
			e, ok := event.(*tcell.EventPaste)
			if ok {
				return e.Text(), nil
			}
		case <-time.After(200 * time.Millisecond):
			return "", errors.New("No clipboard received from terminal")
		}
	}
}

func (t terminalClipboard) write(text, reg string) error {
	return screen.Screen.SetClipboard(text, reg)
}
