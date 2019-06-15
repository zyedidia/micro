package display

import (
	"unicode/utf8"

	runewidth "github.com/mattn/go-runewidth"
	"github.com/zyedidia/micro/internal/buffer"
	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/micro/internal/info"
	"github.com/zyedidia/micro/internal/screen"
	"github.com/zyedidia/micro/internal/util"
	"github.com/zyedidia/tcell"
)

type InfoWindow struct {
	*info.InfoBuf
	*View

	defStyle tcell.Style
	errStyle tcell.Style
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

	iw.Width, iw.Y = screen.Screen.Size()
	iw.Y--

	return iw
}

func (i *InfoWindow) Resize(w, h int) {
	i.Width = w
	i.Y = h
}

func (i *InfoWindow) SetBuffer(b *buffer.Buffer) {
	i.InfoBuf.Buffer = b
}

func (i *InfoWindow) Relocate() bool   { return false }
func (i *InfoWindow) GetView() *View   { return i.View }
func (i *InfoWindow) SetView(v *View)  {}
func (i *InfoWindow) SetActive(b bool) {}

func (i *InfoWindow) GetMouseLoc(vloc buffer.Loc) buffer.Loc {
	c := i.Buffer.GetActiveCursor()
	l := i.Buffer.LineBytes(0)
	n := utf8.RuneCountInString(i.Msg)
	return buffer.Loc{c.GetCharPosInLine(l, vloc.X-n), 0}
}

func (i *InfoWindow) Clear() {
	for x := 0; x < i.Width; x++ {
		screen.Screen.SetContent(x, i.Y, ' ', nil, config.DefStyle)
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

			rw := runewidth.RuneWidth(r)
			for j := 0; j < rw; j++ {
				c := r
				if j > 0 {
					c = ' '
				}
				screen.Screen.SetContent(vlocX, i.Y, c, nil, style)
			}
			vlocX++
		}
		nColsBeforeStart--
	}

	totalwidth := blocX - nColsBeforeStart
	for len(line) > 0 {
		if activeC.X == blocX {
			screen.Screen.ShowCursor(vlocX, i.Y)
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
		if vlocX >= i.Width {
			break
		}
	}
	if activeC.X == blocX {
		screen.Screen.ShowCursor(vlocX, i.Y)
	}
}

var keydisplay = []string{"^Q Quit, ^S Save, ^O Open, ^G Help, ^E Command Bar, ^K Cut Line", "^F Find, ^Z Undo, ^Y Redo, ^A Select All, ^D Duplicate Line, ^T New Tab"}

func (i *InfoWindow) displayKeyMenu() {
	// TODO: maybe make this based on the actual keybindings

	for y := 0; y < len(keydisplay); y++ {
		for x := 0; x < i.Width; x++ {
			if x < len(keydisplay[y]) {
				screen.Screen.SetContent(x, i.Y-len(keydisplay)+y, rune(keydisplay[y][x]), nil, config.DefStyle)
			} else {
				screen.Screen.SetContent(x, i.Y-len(keydisplay)+y, ' ', nil, config.DefStyle)
			}
		}
	}
}

func (i *InfoWindow) Display() {
	x := 0
	if config.GetGlobalOption("keymenu").(bool) {
		i.displayKeyMenu()
	}

	if i.HasPrompt || config.GlobalSettings["infobar"].(bool) {
		if !i.HasPrompt && !i.HasMessage && !i.HasError {
			return
		}
		i.Clear()
		style := i.defStyle

		if i.HasError {
			style = i.errStyle
		}

		display := i.Msg
		for _, c := range display {
			screen.Screen.SetContent(x, i.Y, c, nil, style)
			x += runewidth.RuneWidth(c)
		}

		if i.HasPrompt {
			i.displayBuffer()
		}
	}

	if i.HasSuggestions && len(i.Suggestions) > 1 {
		statusLineStyle := config.DefStyle.Reverse(true)
		if style, ok := config.Colorscheme["statusline"]; ok {
			statusLineStyle = style
		}
		keymenuOffset := 0
		if config.GetGlobalOption("keymenu").(bool) {
			keymenuOffset = len(keydisplay)
		}
		x := 0
		for j, s := range i.Suggestions {
			style := statusLineStyle
			if i.CurSuggestion == j {
				style = style.Reverse(true)
			}
			for _, r := range s {
				screen.Screen.SetContent(x, i.Y-keymenuOffset-1, r, nil, style)
				x++
				if x >= i.Width {
					return
				}
			}
			screen.Screen.SetContent(x, i.Y-keymenuOffset-1, ' ', nil, statusLineStyle)
			x++
			if x >= i.Width {
				return
			}
		}

		for x < i.Width {
			screen.Screen.SetContent(x, i.Y-keymenuOffset-1, ' ', nil, statusLineStyle)
			x++
		}
	}
}
