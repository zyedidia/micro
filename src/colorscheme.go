package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"io/ioutil"
	"os/user"
	"regexp"
	"strconv"
	"strings"
)

const defaultColorscheme = "default"

// Colorscheme is a map from string to style -- it represents a colorscheme
type Colorscheme map[string]tcell.Style

// The current colorscheme
var colorscheme Colorscheme

// InitColorscheme picks and initializes the colorscheme when micro starts
func InitColorscheme() {
	LoadDefaultColorscheme()
}

// LoadDefaultColorscheme loads the default colorscheme from ~/.micro/colorschemes
func LoadDefaultColorscheme() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	LoadColorscheme(defaultColorscheme, dir+"/.micro/colorschemes")
}

// LoadColorscheme loads the given colorscheme from a directory
func LoadColorscheme(colorschemeName, dir string) {
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

// ParseColorscheme parses the text definition for a colorscheme and returns the corresponding object
// Colorschemes are made up of color-link statements linking a color group to a list of colors
// For example, color-link keyword (blue,red) makes all keywords have a blue foreground and
// red background
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
// The strings must be in the format "extra foregroundcolor,backgroundcolor"
// The 'extra' can be bold, reverse, or underline
func StringToStyle(str string) tcell.Style {
	var fg string
	var bg string
	split := strings.Split(str, ",")
	if len(split) > 1 {
		fg, bg = split[0], split[1]
	} else {
		fg = split[0]
	}
	fg = strings.TrimSpace(fg)
	bg = strings.TrimSpace(bg)

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
// We accept either bright... or light... to mean the brighter version of a color
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
	case "brightblack", "lightblack":
		return tcell.ColorGray
	case "brightred", "lightred":
		return tcell.ColorRed
	case "brightgreen", "lightgreen":
		return tcell.ColorLime
	case "brightyellow", "lightyellow":
		return tcell.ColorYellow
	case "brightblue", "lightblue":
		return tcell.ColorBlue
	case "brightmagenta", "lightmagenta":
		return tcell.ColorFuchsia
	case "brightcyan", "lightcyan":
		return tcell.ColorAqua
	case "brightwhite", "lightwhite":
		return tcell.ColorWhite
	case "default":
		return tcell.ColorDefault
	default:
		// Check if this is a 256 color
		if num, err := strconv.Atoi(str); err == nil {
			return GetColor256(num)
		}
		// Probably a truecolor hex value
		return tcell.GetColor(str)
	}
}

// GetColor256 returns the tcell color for a number between 0 and 255
func GetColor256(color int) tcell.Color {
	ansiColors := []tcell.Color{tcell.ColorBlack, tcell.ColorMaroon, tcell.ColorGreen,
		tcell.ColorOlive, tcell.ColorNavy, tcell.ColorPurple,
		tcell.ColorTeal, tcell.ColorSilver, tcell.ColorGray,
		tcell.ColorRed, tcell.ColorLime, tcell.ColorYellow,
		tcell.ColorBlue, tcell.ColorFuchsia, tcell.ColorAqua,
		tcell.ColorWhite}

	if color >= 0 && color <= 15 {
		return ansiColors[color]
	}

	return tcell.GetColor("Color" + strconv.Itoa(color))
}
