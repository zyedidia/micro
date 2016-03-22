package main

import (
	"fmt"
	"github.com/zyedidia/tcell"
	"io/ioutil"
	"os/user"
	"regexp"
	"strings"
)

const colorschemeName = "default"

// Colorscheme is a map from string to style -- it represents a colorscheme
type Colorscheme map[string]tcell.Style

// The current colorscheme
var colorscheme Colorscheme

// InitColorscheme picks and initializes the colorscheme when micro starts
func InitColorscheme() {
	LoadColorscheme()
	// LoadColorscheme may not have found any colorschemes
	if colorscheme == nil {
		colorscheme = DefaultColorscheme()
	}
}

// LoadColorscheme loads the colorscheme from ~/.micro/colorschemes
func LoadColorscheme() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	LoadColorschemeFromDir(dir + "/.micro/colorschemes")
}

// LoadColorschemeFromDir loads the colorscheme from a directory
func LoadColorschemeFromDir(dir string) {
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if f.Name() == colorschemeName+".micro" {
			text, err := ioutil.ReadFile(dir + "/" + f.Name())
			if err != nil {
				fmt.Println("Error loading colorscheme:", err)
				continue
			}
			colorscheme = ParseColorscheme(string(text))
		}
	}
}

// DefaultColorscheme returns the default colorscheme
func DefaultColorscheme() Colorscheme {
	c := make(Colorscheme)
	c["comment"] = StringToStyle("brightgreen")
	c["constant"] = StringToStyle("cyan")
	c["identifier"] = StringToStyle("blue")
	c["statement"] = StringToStyle("green")
	c["preproc"] = StringToStyle("brightred")
	c["type"] = StringToStyle("yellow")
	c["special"] = StringToStyle("red")
	c["underlined"] = StringToStyle("brightmagenta")
	c["ignore"] = StringToStyle("default")
	c["error"] = StringToStyle("bold brightred")
	c["todo"] = StringToStyle("bold brightmagenta")
	return c
}

// ParseColorscheme parses the text definition for a colorscheme and returns the corresponding object
func ParseColorscheme(text string) Colorscheme {
	parser := regexp.MustCompile(`color-link\s+(\S*)\s+"(.*)"`)

	lines := strings.Split(text, "\n")

	c := make(Colorscheme)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" ||
			strings.TrimSpace(line)[0] == '#' {
			// Ignore this line
			continue
		}

		matches := parser.FindSubmatch([]byte(line))
		if len(matches) == 3 {
			link := string(matches[1])
			colors := string(matches[2])

			c[link] = StringToStyle(colors)
		} else {
			fmt.Println("Color-link statement is not valid:", line)
		}
	}

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
	case "default":
		return tcell.ColorDefault
	default:
		return tcell.GetColor(str)
	}
}
