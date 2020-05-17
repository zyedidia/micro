package config

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/zyedidia/tcell"
)

// Micro's default style
var DefStyle tcell.Style = tcell.StyleDefault

// The current colorscheme
var Colorscheme map[string]tcell.Style

// GetColor takes in a syntax group and returns the colorscheme's style for that group
func GetColor(color string) tcell.Style {
	st := DefStyle
	if color == "" {
		return st
	}
	groups := strings.Split(color, ".")
	if len(groups) > 1 {
		curGroup := ""
		for i, g := range groups {
			if i != 0 {
				curGroup += "."
			}
			curGroup += g
			if style, ok := Colorscheme[curGroup]; ok {
				st = style
			}
		}
	} else if style, ok := Colorscheme[color]; ok {
		st = style
	} else {
		st = StringToStyle(color)
	}

	return st
}

// ColorschemeExists checks if a given colorscheme exists
func ColorschemeExists(colorschemeName string) bool {
	return FindRuntimeFile(RTColorscheme, colorschemeName) != nil
}

// InitColorscheme picks and initializes the colorscheme when micro starts
func InitColorscheme() error {
	Colorscheme = make(map[string]tcell.Style)
	DefStyle = tcell.StyleDefault

	return LoadDefaultColorscheme()
}

// LoadDefaultColorscheme loads the default colorscheme from $(ConfigDir)/colorschemes
func LoadDefaultColorscheme() error {
	return LoadColorscheme(GlobalSettings["colorscheme"].(string))
}

// LoadColorscheme loads the given colorscheme from a directory
func LoadColorscheme(colorschemeName string) error {
	file := FindRuntimeFile(RTColorscheme, colorschemeName)
	if file == nil {
		return errors.New(colorschemeName + " is not a valid colorscheme")
	}
	if data, err := file.Data(); err != nil {
		return errors.New("Error loading colorscheme: " + err.Error())
	} else {
		Colorscheme, err = ParseColorscheme(string(data))
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseColorscheme parses the text definition for a colorscheme and returns the corresponding object
// Colorschemes are made up of color-link statements linking a color group to a list of colors
// For example, color-link keyword (blue,red) makes all keywords have a blue foreground and
// red background
func ParseColorscheme(text string) (map[string]tcell.Style, error) {
	var err error
	parser := regexp.MustCompile(`color-link\s+(\S*)\s+"(.*)"`)

	lines := strings.Split(text, "\n")

	c := make(map[string]tcell.Style)

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

			style := StringToStyle(colors)
			c[link] = style

			if link == "default" {
				DefStyle = style
			}
		} else {
			err = errors.New("Color-link statement is not valid: " + line)
		}
	}

	return c, err
}

// StringToStyle returns a style from a string
// The strings must be in the format "extra foregroundcolor,backgroundcolor"
// The 'extra' can be bold, reverse, or underline
func StringToStyle(str string) tcell.Style {
	var fg, bg string
	spaceSplit := strings.Split(str, " ")
	var split []string
	if len(spaceSplit) > 1 {
		split = strings.Split(spaceSplit[1], ",")
	} else {
		split = strings.Split(str, ",")
	}
	if len(split) > 1 {
		fg, bg = split[0], split[1]
	} else {
		fg = split[0]
	}
	fg = strings.TrimSpace(fg)
	bg = strings.TrimSpace(bg)

	var fgColor, bgColor tcell.Color
	if fg == "" {
		fgColor, _, _ = DefStyle.Decompose()
	} else {
		fgColor = StringToColor(fg)
	}
	if bg == "" {
		_, bgColor, _ = DefStyle.Decompose()
	} else {
		bgColor = StringToColor(bg)
	}

	style := DefStyle.Foreground(fgColor).Background(bgColor)
	if strings.Contains(str, "bold") {
		style = style.Bold(true)
	}
	if strings.Contains(str, "italic") {
		style = style.Italic(true)
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
	colors := []tcell.Color{tcell.ColorBlack, tcell.ColorMaroon, tcell.ColorGreen,
		tcell.ColorOlive, tcell.ColorNavy, tcell.ColorPurple,
		tcell.ColorTeal, tcell.ColorSilver, tcell.ColorGray,
		tcell.ColorRed, tcell.ColorLime, tcell.ColorYellow,
		tcell.ColorBlue, tcell.ColorFuchsia, tcell.ColorAqua,
		tcell.ColorWhite, tcell.Color16, tcell.Color17, tcell.Color18, tcell.Color19, tcell.Color20,
		tcell.Color21, tcell.Color22, tcell.Color23, tcell.Color24, tcell.Color25, tcell.Color26, tcell.Color27, tcell.Color28,
		tcell.Color29, tcell.Color30, tcell.Color31, tcell.Color32, tcell.Color33, tcell.Color34, tcell.Color35, tcell.Color36,
		tcell.Color37, tcell.Color38, tcell.Color39, tcell.Color40, tcell.Color41, tcell.Color42, tcell.Color43, tcell.Color44,
		tcell.Color45, tcell.Color46, tcell.Color47, tcell.Color48, tcell.Color49, tcell.Color50, tcell.Color51, tcell.Color52,
		tcell.Color53, tcell.Color54, tcell.Color55, tcell.Color56, tcell.Color57, tcell.Color58, tcell.Color59, tcell.Color60,
		tcell.Color61, tcell.Color62, tcell.Color63, tcell.Color64, tcell.Color65, tcell.Color66, tcell.Color67, tcell.Color68,
		tcell.Color69, tcell.Color70, tcell.Color71, tcell.Color72, tcell.Color73, tcell.Color74, tcell.Color75, tcell.Color76,
		tcell.Color77, tcell.Color78, tcell.Color79, tcell.Color80, tcell.Color81, tcell.Color82, tcell.Color83, tcell.Color84,
		tcell.Color85, tcell.Color86, tcell.Color87, tcell.Color88, tcell.Color89, tcell.Color90, tcell.Color91, tcell.Color92,
		tcell.Color93, tcell.Color94, tcell.Color95, tcell.Color96, tcell.Color97, tcell.Color98, tcell.Color99, tcell.Color100,
		tcell.Color101, tcell.Color102, tcell.Color103, tcell.Color104, tcell.Color105, tcell.Color106, tcell.Color107, tcell.Color108,
		tcell.Color109, tcell.Color110, tcell.Color111, tcell.Color112, tcell.Color113, tcell.Color114, tcell.Color115, tcell.Color116,
		tcell.Color117, tcell.Color118, tcell.Color119, tcell.Color120, tcell.Color121, tcell.Color122, tcell.Color123, tcell.Color124,
		tcell.Color125, tcell.Color126, tcell.Color127, tcell.Color128, tcell.Color129, tcell.Color130, tcell.Color131, tcell.Color132,
		tcell.Color133, tcell.Color134, tcell.Color135, tcell.Color136, tcell.Color137, tcell.Color138, tcell.Color139, tcell.Color140,
		tcell.Color141, tcell.Color142, tcell.Color143, tcell.Color144, tcell.Color145, tcell.Color146, tcell.Color147, tcell.Color148,
		tcell.Color149, tcell.Color150, tcell.Color151, tcell.Color152, tcell.Color153, tcell.Color154, tcell.Color155, tcell.Color156,
		tcell.Color157, tcell.Color158, tcell.Color159, tcell.Color160, tcell.Color161, tcell.Color162, tcell.Color163, tcell.Color164,
		tcell.Color165, tcell.Color166, tcell.Color167, tcell.Color168, tcell.Color169, tcell.Color170, tcell.Color171, tcell.Color172,
		tcell.Color173, tcell.Color174, tcell.Color175, tcell.Color176, tcell.Color177, tcell.Color178, tcell.Color179, tcell.Color180,
		tcell.Color181, tcell.Color182, tcell.Color183, tcell.Color184, tcell.Color185, tcell.Color186, tcell.Color187, tcell.Color188,
		tcell.Color189, tcell.Color190, tcell.Color191, tcell.Color192, tcell.Color193, tcell.Color194, tcell.Color195, tcell.Color196,
		tcell.Color197, tcell.Color198, tcell.Color199, tcell.Color200, tcell.Color201, tcell.Color202, tcell.Color203, tcell.Color204,
		tcell.Color205, tcell.Color206, tcell.Color207, tcell.Color208, tcell.Color209, tcell.Color210, tcell.Color211, tcell.Color212,
		tcell.Color213, tcell.Color214, tcell.Color215, tcell.Color216, tcell.Color217, tcell.Color218, tcell.Color219, tcell.Color220,
		tcell.Color221, tcell.Color222, tcell.Color223, tcell.Color224, tcell.Color225, tcell.Color226, tcell.Color227, tcell.Color228,
		tcell.Color229, tcell.Color230, tcell.Color231, tcell.Color232, tcell.Color233, tcell.Color234, tcell.Color235, tcell.Color236,
		tcell.Color237, tcell.Color238, tcell.Color239, tcell.Color240, tcell.Color241, tcell.Color242, tcell.Color243, tcell.Color244,
		tcell.Color245, tcell.Color246, tcell.Color247, tcell.Color248, tcell.Color249, tcell.Color250, tcell.Color251, tcell.Color252,
		tcell.Color253, tcell.Color254, tcell.Color255,
	}

	if color >= 0 && color < len(colors) {
		return colors[color]
	}

	return tcell.ColorDefault
}
