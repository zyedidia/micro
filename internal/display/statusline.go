package display

import (
	"bytes"
	"fmt"
	"log"
	"path"
	"regexp"
	"strconv"
	"unicode/utf8"

	runewidth "github.com/mattn/go-runewidth"
	"github.com/zyedidia/micro/internal/buffer"
	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/micro/internal/screen"
	"github.com/zyedidia/micro/internal/util"
)

// StatusLine represents the information line at the bottom
// of each window
// It gives information such as filename, whether the file has been
// modified, filetype, cursor location
type StatusLine struct {
	Info map[string]func(*buffer.Buffer) string

	win *BufWindow
}

// TODO: plugin modify status line formatter

// NewStatusLine returns a statusline bound to a window
func NewStatusLine(win *BufWindow) *StatusLine {
	s := new(StatusLine)
	s.Info = map[string]func(*buffer.Buffer) string{
		"filename": func(b *buffer.Buffer) string {
			if b.Settings["basename"].(bool) {
				return path.Base(b.GetName())
			}
			return b.GetName()
		},
		"line": func(b *buffer.Buffer) string {
			return strconv.Itoa(b.GetActiveCursor().Y + 1)
		},
		"col": func(b *buffer.Buffer) string {
			return strconv.Itoa(b.GetActiveCursor().X + 1)
		},
		"modified": func(b *buffer.Buffer) string {
			if b.Modified() {
				return "+ "
			}
			return ""
		},
	}
	s.win = win
	return s
}

// FindOpt finds a given option in the current buffer's settings
func (s *StatusLine) FindOpt(opt string) interface{} {
	if val, ok := s.win.Buf.Settings[opt]; ok {
		return val
	}
	return "null"
}

var formatParser = regexp.MustCompile(`\$\(.+?\)`)

// Display draws the statusline to the screen
func (s *StatusLine) Display() {
	// We'll draw the line at the lowest line in the window
	y := s.win.Height + s.win.Y - 1

	formatter := func(match []byte) []byte {
		name := match[2 : len(match)-1]
		if bytes.HasPrefix(name, []byte("opt")) {
			option := name[4:]
			return []byte(fmt.Sprint(s.FindOpt(string(option))))
		} else if bytes.HasPrefix(name, []byte("bind")) {
			binding := string(name[5:])
			for k, v := range config.Bindings {
				if v == binding {
					return []byte(k)
				}
			}
			return []byte("null")
		} else {
			return []byte(s.Info[string(name)](s.win.Buf))
		}
	}

	leftText := []byte(s.win.Buf.StatusFormatLeft)
	leftText = formatParser.ReplaceAllFunc(leftText, formatter)
	rightText := []byte(s.win.Buf.StatusFormatRight)
	rightText = formatParser.ReplaceAllFunc(rightText, formatter)

	statusLineStyle := config.DefStyle.Reverse(true)
	if style, ok := config.Colorscheme["statusline"]; ok {
		statusLineStyle = style
	}

	leftLen := util.StringWidth(leftText, utf8.RuneCount(leftText), 1)
	rightLen := util.StringWidth(rightText, utf8.RuneCount(rightText), 1)

	winX := s.win.X
	for x := 0; x < s.win.Width; x++ {
		if x < leftLen {
			r, size := utf8.DecodeRune(leftText)
			leftText = leftText[size:]
			rw := runewidth.RuneWidth(r)
			for j := 0; j < rw; j++ {
				c := r
				if j > 0 {
					c = ' '
					x++
				}
				log.Println(x, string(c))
				screen.Screen.SetContent(winX+x, y, c, nil, statusLineStyle)
			}
		} else if x >= s.win.Width-rightLen && x < rightLen+s.win.Width-rightLen {
			r, size := utf8.DecodeRune(rightText)
			rightText = rightText[size:]
			rw := runewidth.RuneWidth(r)
			for j := 0; j < rw; j++ {
				c := r
				if j > 0 {
					c = ' '
					x++
				}
				screen.Screen.SetContent(winX+x, y, c, nil, statusLineStyle)
			}
		} else {
			screen.Screen.SetContent(winX+x, y, ' ', nil, statusLineStyle)
		}
	}
}
