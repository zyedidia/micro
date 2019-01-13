package display

import (
	"unicode/utf8"

	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/micro/cmd/micro/shell"
	"github.com/zyedidia/tcell"
	"github.com/zyedidia/terminal"
)

type TermWindow struct {
	*View
	*shell.Terminal

	active bool
}

func NewTermWindow(x, y, w, h int, term *shell.Terminal) *TermWindow {
	tw := new(TermWindow)
	tw.View = new(View)
	tw.Terminal = term
	tw.X, tw.Y = x, y
	tw.Width, tw.Height = w, h-1
	tw.Resize(tw.Width, tw.Height)
	return tw
}

// Resize informs the terminal of a resize event
func (w *TermWindow) Resize(width, height int) {
	w.Term.Resize(width, height)
}

func (w *TermWindow) SetActive(b bool) {
	w.active = b
}

func (w *TermWindow) GetMouseLoc(vloc buffer.Loc) buffer.Loc {
	return vloc
}

func (w *TermWindow) Clear() {
	for y := 0; y < w.Height; y++ {
		for x := 0; x < w.Width; x++ {
			screen.Screen.SetContent(w.X+x, w.Y+y, ' ', nil, config.DefStyle)
		}
	}
}

func (w *TermWindow) Relocate() bool { return true }
func (w *TermWindow) GetView() *View {
	return w.View
}
func (w *TermWindow) SetView(v *View) {
	w.View = v
}

// Display displays this terminal in a view
func (w *TermWindow) Display() {
	w.State.Lock()
	defer w.State.Unlock()

	var l buffer.Loc
	for y := 0; y < w.Height; y++ {
		for x := 0; x < w.Width; x++ {
			l.X, l.Y = x, y
			c, f, b := w.State.Cell(x, y)

			fg, bg := int(f), int(b)
			if f == terminal.DefaultFG {
				fg = int(tcell.ColorDefault)
			}
			if b == terminal.DefaultBG {
				bg = int(tcell.ColorDefault)
			}
			st := tcell.StyleDefault.Foreground(config.GetColor256(int(fg))).Background(config.GetColor256(int(bg)))

			if l.LessThan(w.Selection[1]) && l.GreaterEqual(w.Selection[0]) || l.LessThan(w.Selection[0]) && l.GreaterEqual(w.Selection[1]) {
				st = st.Reverse(true)
			}

			screen.Screen.SetContent(w.X+x, w.Y+y, c, nil, st)
		}
	}
	if config.GetGlobalOption("statusline").(bool) {
		statusLineStyle := config.DefStyle.Reverse(true)
		if style, ok := config.Colorscheme["statusline"]; ok {
			statusLineStyle = style
		}

		text := []byte(w.Name())
		textLen := utf8.RuneCount(text)
		for x := 0; x < w.Width; x++ {
			if x < textLen {
				r, size := utf8.DecodeRune(text)
				text = text[size:]
				screen.Screen.SetContent(w.X+x, w.Y+w.Height, r, nil, statusLineStyle)
			} else {
				screen.Screen.SetContent(w.X+x, w.Y+w.Height, ' ', nil, statusLineStyle)
			}
		}
	}
	if w.State.CursorVisible() && w.active {
		curx, cury := w.State.Cursor()
		screen.Screen.ShowCursor(curx+w.X, cury+w.Y)
	}
}
