package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

var syntaxFiles map[[2]*regexp.Regexp][2]string

// LoadSyntaxFiles loads the syntax files from the default directory ~/.micro
func LoadSyntaxFiles() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	LoadSyntaxFilesFromDir(dir + "/.micro")
}

// LoadSyntaxFilesFromDir loads the syntax files from a specified directory
// To load the syntax files, we must fill the `syntaxFiles` map
// This involves finding the regex for syntax and if it exists, the regex
// for the header. Then we must get the text for the file and the filetype.
func LoadSyntaxFilesFromDir(dir string) {
	syntaxFiles = make(map[[2]*regexp.Regexp][2]string)
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".micro" {
			text, err := ioutil.ReadFile(dir + "/" + f.Name())

			if err != nil {
				fmt.Println("Error loading syntax files:", err)
				continue
			}
			lines := strings.Split(string(text), "\n")
			syntaxParser := regexp.MustCompile(`syntax "(.*?)"\s+"(.*)"`)
			headerParser := regexp.MustCompile(`header "(.*)"`)
			syntaxMatches := syntaxParser.FindSubmatch([]byte(lines[0]))

			var headerRegex *regexp.Regexp
			if strings.HasPrefix(lines[1], "header") {
				headerMatches := headerParser.FindSubmatch([]byte(lines[1]))
				headerRegex, err = regexp.Compile(string(headerMatches[1]))

				if err != nil {
					// Error with the regex!
					continue
				}
			}

			syntaxRegex, err := regexp.Compile(string(syntaxMatches[2]))
			if err != nil {
				// Error with the regex!
				continue
			}

			regexes := [2]*regexp.Regexp{syntaxRegex, headerRegex}
			syntaxFiles[regexes] = [2]string{string(text), string(syntaxMatches[1])}
		}
	}
}

// GetRules finds the syntax rules that should be used for the buffer
// and returns them. It also returns the filetype of the file
func GetRules(buf *Buffer) (string, string) {
	for r := range syntaxFiles {
		if r[0].MatchString(buf.path) {
			return syntaxFiles[r][0], syntaxFiles[r][1]
		} else if r[1] != nil && r[1].MatchString(buf.lines[0]) {
			return syntaxFiles[r][0], syntaxFiles[r][1]
		}
	}
	return "", "Unknown"
}

// Match takes a buffer and returns a map specifying how it should be syntax highlighted
// The map is from character numbers to styles, so map[3] represents the style change
// at the third character in the buffer
func Match(rules string, buf *Buffer) map[int]tcell.Style {
	// rules := strings.TrimSpace(GetRules(buf))
	str := buf.text

	lines := strings.Split(rules, "\n")
	m := make(map[int]tcell.Style)
	parser := regexp.MustCompile(`color (.*?)\s+"(.*)"`)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" ||
			strings.TrimSpace(line)[0] == '#' ||
			strings.HasPrefix(line, "syntax") ||
			strings.HasPrefix(line, "header") {
			// Ignore this line
			continue
		}
		submatch := parser.FindSubmatch([]byte(line))
		color := string(submatch[1])
		regex, err := regexp.Compile(string(submatch[2]))
		if err != nil {
			// Error with the regex!
			continue
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
