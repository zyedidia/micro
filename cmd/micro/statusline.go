package main

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"unicode/utf8"

	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/screen"
)

// StatusLine represents the information line at the bottom
// of each window
// It gives information such as filename, whether the file has been
// modified, filetype, cursor location
type StatusLine struct {
	FormatLeft  string
	FormatRight string
	Info        map[string]func(*buffer.Buffer) string

	win *Window
}

// TODO: plugin modify status line formatter

// NewStatusLine returns a statusline bound to a window
func NewStatusLine(win *Window) *StatusLine {
	s := new(StatusLine)
	// s.FormatLeft = "$(filename) $(modified)($(line),$(col)) $(opt:filetype) $(opt:fileformat)"
	s.FormatLeft = "$(filename) $(modified)(line,col) $(opt:filetype) $(opt:fileformat)"
	s.FormatRight = "$(bind:ToggleKeyMenu): show bindings, $(bind:ToggleHelp): open help"
	s.Info = map[string]func(*buffer.Buffer) string{
		"filename": func(b *buffer.Buffer) string {
			if b.Settings["basename"].(bool) {
				return path.Base(b.GetName())
			}
			return b.GetName()
		},
		"line": func(b *buffer.Buffer) string {
			return strconv.Itoa(b.GetActiveCursor().Y)
		},
		"col": func(b *buffer.Buffer) string {
			return strconv.Itoa(b.GetActiveCursor().X)
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
	// TODO: don't display if infobar off and has message
	// if !GetGlobalOption("infobar").(bool) {
	// 	return
	// }

	// We'll draw the line at the lowest line in the window
	y := s.win.Height + s.win.Y - 1

	formatter := func(match []byte) []byte {
		name := match[2 : len(match)-1]
		if bytes.HasPrefix(name, []byte("opt")) {
			option := name[4:]
			return []byte(fmt.Sprint(s.FindOpt(string(option))))
		} else if bytes.HasPrefix(name, []byte("bind")) {
			binding := string(name[5:])
			for k, v := range bindings {
				if v == binding {
					return []byte(k)
				}
			}
			return []byte("null")
		} else {
			return []byte(s.Info[string(name)](s.win.Buf))
		}
	}

	leftText := []byte(s.FormatLeft)
	leftText = formatParser.ReplaceAllFunc([]byte(s.FormatLeft), formatter)
	rightText := []byte(s.FormatRight)
	rightText = formatParser.ReplaceAllFunc([]byte(s.FormatRight), formatter)

	statusLineStyle := config.DefStyle.Reverse(true)
	if style, ok := config.Colorscheme["statusline"]; ok {
		statusLineStyle = style
	}

	leftLen := utf8.RuneCount(leftText)
	rightLen := utf8.RuneCount(rightText)

	winX := s.win.X
	for x := 0; x < s.win.Width; x++ {
		if x < leftLen {
			r, size := utf8.DecodeRune(leftText)
			leftText = leftText[size:]
			screen.Screen.SetContent(winX+x, y, r, nil, statusLineStyle)
		} else if x >= s.win.Width-rightLen && x < rightLen+s.win.Width-rightLen {
			r, size := utf8.DecodeRune(rightText)
			rightText = rightText[size:]
			screen.Screen.SetContent(winX+x, y, r, nil, statusLineStyle)
		} else {
			screen.Screen.SetContent(winX+x, y, ' ', nil, statusLineStyle)
		}
	}
}
