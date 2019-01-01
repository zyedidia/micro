package display

import (
	"strings"

	runewidth "github.com/mattn/go-runewidth"
	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/info"
	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/tcell"
)

type InfoWindow struct {
	*info.Bar
	*View

	defStyle tcell.Style
	errStyle tcell.Style

	width int
	y     int
}

func NewInfoWindow(b *info.Bar) *InfoWindow {
	iw := new(InfoWindow)
	iw.Bar = b
	iw.View = new(View)

	iw.defStyle = config.DefStyle

	if _, ok := config.Colorscheme["message"]; ok {
		iw.defStyle = config.Colorscheme["message"]
	}

	iw.errStyle = config.DefStyle.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorMaroon)

	if _, ok := config.Colorscheme["error-message"]; ok {
		iw.errStyle = config.Colorscheme["error-message"]
	}

	iw.width, iw.y = screen.Screen.Size()
	iw.y--

	return iw
}

func (i *InfoWindow) Relocate() bool  { return false }
func (i *InfoWindow) GetView() *View  { return i.View }
func (i *InfoWindow) SetView(v *View) {}

func (i *InfoWindow) Clear() {
	for x := 0; x < i.width; x++ {
		screen.Screen.SetContent(x, i.y, ' ', nil, config.DefStyle)
	}
}

func (i *InfoWindow) Display() {
	x := 0
	if i.HasPrompt || config.GlobalSettings["infobar"].(bool) {
		style := i.defStyle

		if i.HasError {
			style = i.errStyle
		}

		display := i.Msg + strings.TrimSpace(string(i.Bytes()))
		for _, c := range display {
			screen.Screen.SetContent(x, i.y, c, nil, style)
			x += runewidth.RuneWidth(c)
		}
	}
}
