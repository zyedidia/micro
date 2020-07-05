package clipboard

import "github.com/zyedidia/micro/v2/internal/screen"

type terminalClipboard struct{}

var terminal terminalClipboard

func (t terminalClipboard) read(reg string) error {
	return screen.Screen.GetClipboard(reg)
}

func (t terminalClipboard) write(text, reg string) error {
	return screen.Screen.SetClipboard(text, reg)
}
