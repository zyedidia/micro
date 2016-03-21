package main

import (
	"github.com/zyedidia/tcell"
	"strings"
)

// The current colorscheme
var colorscheme map[string]tcell.Style

func InitColorscheme() {
	colorscheme = DefaultColorscheme()
}

// DefaultColorscheme returns the default colorscheme
func DefaultColorscheme() map[string]tcell.Style {
	c := make(map[string]tcell.Style)
	c["comment"] = StringToStyle("brightgreen")
	c["constant"] = StringToStyle("cyan")
	c["identifier"] = StringToStyle("blue")
	c["statement"] = StringToStyle("green")
	c["preproc"] = StringToStyle("brightred")
	c["type"] = StringToStyle("yellow")
	c["special"] = StringToStyle("red")
	c["underlined"] = StringToStyle("brightmagenta")
	c["ignore"] = StringToStyle("")
	c["error"] = StringToStyle("bold brightred")
	c["todo"] = StringToStyle("bold brightmagenta")
	return c
}

// StringToStyle returns a style from a string
func StringToStyle(str string) tcell.Style {
	var fg string
	var bg string
	split := strings.Split(str, ",")
	if len(split) > 1 {
		fg, bg = split[0], split[1]
	} else {
		fg = split[0]
	}

	style := tcell.StyleDefault.Foreground(StringToColor(fg)).Background(StringToColor(bg))
	if strings.Contains(str, "bold") {
		style = style.Bold(true)
	}
	if strings.Contains(str, "blink") {
		style = style.Blink(true)
	}
	if strings.Contains(str, "reverse") {
		style = style.Reverse(true)
	}
	if strings.Contains(str, "underline") {
		style = style.Underline(true)
	}
	return style
}

// StringToColor returns a tcell color from a string representation of a color
func StringToColor(str string) tcell.Color {
	switch str {
	case "black":
		return tcell.ColorBlack
	case "red":
		return tcell.ColorMaroon
	case "green":
		return tcell.ColorGreen
	case "yellow":
		return tcell.ColorOlive
	case "blue":
		return tcell.ColorNavy
	case "magenta":
		return tcell.ColorPurple
	case "cyan":
		return tcell.ColorTeal
	case "white":
		return tcell.ColorSilver
	case "brightblack":
		return tcell.ColorGray
	case "brightred":
		return tcell.ColorRed
	case "brightgreen":
		return tcell.ColorLime
	case "brightyellow":
		return tcell.ColorYellow
	case "brightblue":
		return tcell.ColorBlue
	case "brightmagenta":
		return tcell.ColorFuchsia
	case "brightcyan":
		return tcell.ColorAqua
	case "brightwhite":
		return tcell.ColorWhite
	default:
		return tcell.ColorDefault
	}
}
