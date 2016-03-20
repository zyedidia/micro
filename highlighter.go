package main

import (
	"github.com/gdamore/tcell"
	"io/ioutil"
	"regexp"
	"strings"
)

var syntaxRules string

func GetRules() string {
	file, err := ioutil.ReadFile("syntax.micro")
	if err != nil {
		return ""
	}
	return string(file)
}

// Match ...
func Match(str string) map[int]tcell.Style {
	rules := strings.TrimSpace(GetRules())

	lines := strings.Split(rules, "\n")
	m := make(map[int]tcell.Style)
	parser, _ := regexp.Compile(`color (.*?)\s+"(.*)"`)
	for _, line := range lines {
		submatch := parser.FindSubmatch([]byte(line))
		color := string(submatch[1])
		regex, err := regexp.Compile(string(submatch[2]))
		if err != nil {
			continue
			// Error with the regex!
		}
		st := StringToStyle(color)

		if regex.MatchString(str) {
			indicies := regex.FindAllStringIndex(str, -1)
			for _, value := range indicies {
				for i := value[0] + 1; i < value[1]; i++ {
					if _, exists := m[i]; exists {
						delete(m, i)
					}
				}
				m[value[0]] = st
				if _, exists := m[value[1]]; !exists {
					m[value[1]] = tcell.StyleDefault
				}
			}
		}
	}

	return m
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

	return tcell.StyleDefault.Foreground(StringToColor(fg)).Background(StringToColor(bg))
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
