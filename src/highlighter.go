package main

import (
	"fmt"
	"github.com/zyedidia/tcell"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

type FileTypeRules struct {
	filetype string
	rules    []SyntaxRule
}

type SyntaxRule struct {
	regex *regexp.Regexp
	style tcell.Style
}

var syntaxFiles map[[2]*regexp.Regexp]FileTypeRules

// LoadSyntaxFiles loads the syntax files from the default directory ~/.micro
func LoadSyntaxFiles() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	LoadSyntaxFilesFromDir(dir + "/.micro/syntax")
}

// JoinRule takes a syntax rule (which can be multiple regular expressions)
// and joins it into one regular expression by ORing everything together
func JoinRule(rule string) string {
	split := strings.Split(rule, `" "`)
	joined := strings.Join(split, ")|(")
	joined = "(" + joined + ")"
	return joined
}

// LoadSyntaxFilesFromDir loads the syntax files from a specified directory
// To load the syntax files, we must fill the `syntaxFiles` map
// This involves finding the regex for syntax and if it exists, the regex
// for the header. Then we must get the text for the file and the filetype.
func LoadSyntaxFilesFromDir(dir string) {
	InitColorscheme()

	syntaxFiles = make(map[[2]*regexp.Regexp]FileTypeRules)
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".micro" {
			text, err := ioutil.ReadFile(dir + "/" + f.Name())

			if err != nil {
				fmt.Println("Error loading syntax files:", err)
				continue
			}
			lines := strings.Split(string(text), "\n")

			syntaxParser := regexp.MustCompile(`syntax "(.*?)"\s+"(.*)"+`)
			headerParser := regexp.MustCompile(`header "(.*)"`)
			ruleParser := regexp.MustCompile(`color (.*?)\s+(?:\((.*?)\)\s+)?"(.*)"`)

			var syntaxRegex *regexp.Regexp
			var headerRegex *regexp.Regexp
			var filetype string
			var rules []SyntaxRule
			for lineNum, line := range lines {
				if strings.TrimSpace(line) == "" ||
					strings.TrimSpace(line)[0] == '#' {
					// Ignore this line
					continue
				}

				if strings.HasPrefix(line, "syntax") {
					syntaxMatches := syntaxParser.FindSubmatch([]byte(line))
					if len(syntaxMatches) == 3 {
						if syntaxRegex != nil {
							regexes := [2]*regexp.Regexp{syntaxRegex, headerRegex}
							syntaxFiles[regexes] = FileTypeRules{filetype, rules}
						}
						rules = rules[:0]

						filetype = string(syntaxMatches[1])
						extensions := JoinRule(string(syntaxMatches[2]))

						syntaxRegex, err = regexp.Compile(extensions)
						if err != nil {
							fmt.Println("Regex error:", err)
							continue
						}
					} else {
						fmt.Println("Syntax statement is not valid:", line)
						continue
					}
				} else if strings.HasPrefix(line, "header") {
					headerMatches := headerParser.FindSubmatch([]byte(line))
					if len(headerMatches) == 2 {
						header := JoinRule(string(headerMatches[1]))

						headerRegex, err = regexp.Compile(header)
						if err != nil {
							fmt.Println("Regex error:", err)
							continue
						}
					} else {
						fmt.Println("Header statement is not valid:", line)
						continue
					}
				} else {
					if ruleParser.MatchString(line) {
						submatch := ruleParser.FindSubmatch([]byte(line))
						color := string(submatch[1])
						var regexStr string
						if len(submatch) == 4 {
							regexStr = "(?m" + string(submatch[2]) + ")" + JoinRule(string(submatch[3]))
						} else if len(submatch) == 3 {
							regexStr = "(?m)" + JoinRule(string(submatch[2]))
						}
						regex, err := regexp.Compile(regexStr)
						if err != nil {
							fmt.Println(f.Name(), lineNum, err)
							continue
						}

						st := tcell.StyleDefault
						if _, ok := colorscheme[color]; ok {
							st = colorscheme[color]
						} else {
							st = StringToStyle(color)
						}
						rules = append(rules, SyntaxRule{regex, st})
					}
				}
			}
			if syntaxRegex != nil {
				regexes := [2]*regexp.Regexp{syntaxRegex, headerRegex}
				syntaxFiles[regexes] = FileTypeRules{filetype, rules}
			}
		}
	}
}

// GetRules finds the syntax rules that should be used for the buffer
// and returns them. It also returns the filetype of the file
func GetRules(buf *Buffer) ([]SyntaxRule, string) {
	for r := range syntaxFiles {
		if r[0] != nil && r[0].MatchString(buf.path) {
			return syntaxFiles[r].rules, syntaxFiles[r].filetype
		} else if r[1] != nil && r[1].MatchString(buf.lines[0]) {
			return syntaxFiles[r].rules, syntaxFiles[r].filetype
		}
	}
	return nil, "Unknown"
}

// Match takes a buffer and returns a map specifying how it should be syntax highlighted
// The map is from character numbers to styles, so map[3] represents the style change
// at the third character in the buffer
// Note that this map only stores changes in styles, not each character's style
func Match(rules []SyntaxRule, buf *Buffer, v *View) map[int]tcell.Style {
	start := v.topline - synLinesUp
	end := v.topline + v.height + synLinesDown
	if start < 0 {
		start = 0
	}
	if end > len(buf.lines) {
		end = len(buf.lines)
	}
	str := strings.Join(buf.lines[start:end], "\n")
	startNum := v.cursor.loc + v.cursor.Distance(0, start)
	toplineNum := v.cursor.loc + v.cursor.Distance(0, v.topline)

	m := make(map[int]tcell.Style)
	for _, rule := range rules {
		if rule.regex.MatchString(str) {
			indicies := rule.regex.FindAllStringIndex(str, -1)
			for _, value := range indicies {
				value[0] += startNum
				value[1] += startNum
				for i := value[0]; i < value[1]; i++ {
					if i >= toplineNum {
						m[i] = rule.style
					}
				}
			}
		}
	}

	return m
}
