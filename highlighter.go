package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"regexp"
	"strings"
)

// Match ...
func Match(str string) map[int]tcell.Style {
	rules := `color blue "[A-Za-z_][A-Za-z0-9_]*[[:space:]]*[()]"
color blue "\b(append|cap|close|complex|copy|delete|imag|len)\b"
color blue "\b(make|new|panic|print|println|protect|real|recover)\b"
color green     "\b(u?int(8|16|32|64)?|float(32|64)|complex(64|128))\b"
color green     "\b(uintptr|byte|rune|string|interface|bool|map|chan|error)\b"
color cyan  "\b(package|import|const|var|type|struct|func|go|defer|nil|iota)\b"
color cyan  "\b(for|range|if|else|case|default|switch|return)\b"
color red     "\b(go|goto|break|continue)\b"
color cyan "\b(true|false)\b"
color red "[-+/*=<>!~%&|^]|:="
color blue   "\b([0-9]+|0x[0-9a-fA-F]*)\b|'.'"
color magenta   "\\([0-7]{3}|x[A-Fa-f0-9]{2}|u[A-Fa-f0-9]{4}|U[A-Fa-f0-9]{8})"
color yellow   "` + "`" + `[^` + "`" + `]*` + "`" + `"
color green "(^|[[:space:]])//.*"
color brightwhite,cyan "TODO:?"
color ,green "[[:space:]]+$"
color ,red "	+ +| +	+"`

	lines := strings.Split(rules, "\n")
	m := make(map[int]tcell.Style)
	for _, line := range lines {
		split := strings.Split(line, "\"")
		color := strings.Split(split[0], " ")[1]
		regex, err := regexp.Compile(split[1])
		if err != nil {
			fmt.Println("\a")
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
				m[value[1]] = tcell.StyleDefault
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
