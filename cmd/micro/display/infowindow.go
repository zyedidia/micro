package display

import (
	"unicode/utf8"

	runewidth "github.com/mattn/go-runewidth"
	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/info"
	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/micro/cmd/micro/util"
	"github.com/zyedidia/tcell"
)

type InfoWindow struct {
	*info.InfoBuf
	*View

	defStyle tcell.Style
	errStyle tcell.Style

	width int
	y     int
}

func NewInfoWindow(b *info.InfoBuf) *InfoWindow {
	iw := new(InfoWindow)
	iw.InfoBuf = b
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

// func (i *InfoWindow) YesNoPrompt() (bool, bool) {
// 	for {
// 		i.Clear()
// 		i.Display()
// 		screen.Screen.ShowCursor(utf8.RuneCountInString(i.Msg), i.y)
// 		screen.Show()
// 		event := <-events
//
// 		switch e := event.(type) {
// 		case *tcell.EventKey:
// 			switch e.Key() {
// 			case tcell.KeyRune:
// 				if e.Rune() == 'y' || e.Rune() == 'Y' {
// 					i.HasPrompt = false
// 					return true, false
// 				} else if e.Rune() == 'n' || e.Rune() == 'N' {
// 					i.HasPrompt = false
// 					return false, false
// 				}
// 			case tcell.KeyCtrlC, tcell.KeyCtrlQ, tcell.KeyEscape:
// 				i.Clear()
// 				i.Reset()
// 				i.HasPrompt = false
// 				return false, true
// 			}
// 		}
// 	}
// }

func (i *InfoWindow) Relocate() bool  { return false }
func (i *InfoWindow) GetView() *View  { return i.View }
func (i *InfoWindow) SetView(v *View) {}

func (i *InfoWindow) GetMouseLoc(vloc buffer.Loc) buffer.Loc {
	c := i.Buffer.GetActiveCursor()
	l := i.Buffer.LineBytes(0)
	n := utf8.RuneCountInString(i.Msg)
	return buffer.Loc{c.GetCharPosInLine(l, vloc.X-n), 0}
}

func (i *InfoWindow) Clear() {
	for x := 0; x < i.width; x++ {
		screen.Screen.SetContent(x, i.y, ' ', nil, config.DefStyle)
	}
}

func (i *InfoWindow) displayBuffer() {
	b := i.Buffer
	line := b.LineBytes(0)
	activeC := b.GetActiveCursor()

	blocX := 0
	vlocX := utf8.RuneCountInString(i.Msg)

	tabsize := 4
	line, nColsBeforeStart, bslice := util.SliceVisualEnd(line, blocX, tabsize)
	blocX = bslice

	draw := func(r rune, style tcell.Style) {
		if nColsBeforeStart <= 0 {
			bloc := buffer.Loc{X: blocX, Y: 0}
			if activeC.HasSelection() &&
				(bloc.GreaterEqual(activeC.CurSelection[0]) && bloc.LessThan(activeC.CurSelection[1]) ||
					bloc.LessThan(activeC.CurSelection[0]) && bloc.GreaterEqual(activeC.CurSelection[1])) {
				// The current character is selected
				style = config.DefStyle.Reverse(true)

				if s, ok := config.Colorscheme["selection"]; ok {
					style = s
				}

			}

			screen.Screen.SetContent(vlocX, i.y, r, nil, style)
			vlocX++
		}
		nColsBeforeStart--
	}

	totalwidth := blocX - nColsBeforeStart
	for len(line) > 0 {
		if activeC.X == blocX {
			screen.Screen.ShowCursor(vlocX, i.y)
		}

		r, size := utf8.DecodeRune(line)

		draw(r, i.defStyle)

		width := 0

		char := ' '
		switch r {
		case '\t':
			ts := tabsize - (totalwidth % tabsize)
			width = ts
		default:
			width = runewidth.RuneWidth(r)
			char = '@'
		}

		blocX++
		line = line[size:]

		// Draw any extra characters either spaces for tabs or @ for incomplete wide runes
		if width > 1 {
			for j := 1; j < width; j++ {
				draw(char, i.defStyle)
			}
		}
		totalwidth += width
		if vlocX >= i.width {
			break
		}
	}
	if activeC.X == blocX {
		screen.Screen.ShowCursor(vlocX, i.y)
	}
}

func (i *InfoWindow) Display() {
	x := 0
	if i.HasPrompt || config.GlobalSettings["infobar"].(bool) {
		if !i.HasPrompt && !i.HasMessage && !i.HasError {
			return
		}
		style := i.defStyle

		if i.HasError {
			style = i.errStyle
		}

		display := i.Msg
		for _, c := range display {
			screen.Screen.SetContent(x, i.y, c, nil, style)
			x += runewidth.RuneWidth(c)
		}

		if i.HasPrompt {
			i.displayBuffer()
		}
	}
}
