package main

import (
	"github.com/gdamore/tcell"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

// FileTypeRules represents a complete set of syntax rules for a filetype
type FileTypeRules struct {
	filetype string
	rules    []SyntaxRule
}

// SyntaxRule represents a regex to highlight in a certain style
type SyntaxRule struct {
	// What to highlight
	regex *regexp.Regexp
	// Any flags
	flags string
	// Whether this regex is a start=... end=... regex
	startend bool
	// How to highlight it
	style tcell.Style
}

var syntaxFiles map[[2]*regexp.Regexp]FileTypeRules

// LoadSyntaxFiles loads the syntax files from the default directory ~/.micro
func LoadSyntaxFiles() {
	dir, err := homedir.Dir()
	if err != nil {
		TermMessage("Error finding your home directory\nCan't load runtime files")
		return
	}
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
			filename := dir + "/" + f.Name()

			if err != nil {
				TermMessage("Error loading syntax files: " + err.Error())
				continue
			}
			lines := strings.Split(string(text), "\n")

			syntaxParser := regexp.MustCompile(`syntax "(.*?)"\s+"(.*)"+`)
			headerParser := regexp.MustCompile(`header "(.*)"`)

			ruleParser := regexp.MustCompile(`color (.*?)\s+(?:\((.*?)\)\s+)?"(.*)"`)
			ruleStartEndParser := regexp.MustCompile(`color (.*?)\s+(?:\((.*?)\)\s+)?start="(.*)"\s+end="(.*)"`)

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
							TermError(filename, lineNum, err.Error())
							continue
						}
					} else {
						TermError(filename, lineNum, "Syntax statement is not valid: "+line)
						continue
					}
				} else if strings.HasPrefix(line, "header") {
					headerMatches := headerParser.FindSubmatch([]byte(line))
					if len(headerMatches) == 2 {
						header := JoinRule(string(headerMatches[1]))

						headerRegex, err = regexp.Compile(header)
						if err != nil {
							TermError(filename, lineNum, "Regex error: "+err.Error())
							continue
						}
					} else {
						TermError(filename, lineNum, "Header statement is not valid: "+line)
						continue
					}
				} else {
					if ruleParser.MatchString(line) {
						submatch := ruleParser.FindSubmatch([]byte(line))
						var color string
						var regexStr string
						var flags string
						if len(submatch) == 4 {
							color = string(submatch[1])
							flags = string(submatch[2])
							regexStr = "(?" + flags + ")" + JoinRule(string(submatch[3]))
						} else if len(submatch) == 3 {
							color = string(submatch[1])
							regexStr = JoinRule(string(submatch[2]))
						} else {
							TermError(filename, lineNum, "Invalid statement: "+line)
						}
						regex, err := regexp.Compile(regexStr)
						if err != nil {
							TermError(filename, lineNum, err.Error())
							continue
						}

						st := tcell.StyleDefault
						if _, ok := colorscheme[color]; ok {
							st = colorscheme[color]
						} else {
							st = StringToStyle(color)
						}
						rules = append(rules, SyntaxRule{regex, flags, false, st})
					} else if ruleStartEndParser.MatchString(line) {
						submatch := ruleStartEndParser.FindSubmatch([]byte(line))
						var color string
						var start string
						var end string
						// Use m and s flags by default
						flags := "ms"
						if len(submatch) == 5 {
							color = string(submatch[1])
							flags += string(submatch[2])
							start = string(submatch[3])
							end = string(submatch[4])
						} else if len(submatch) == 4 {
							color = string(submatch[1])
							start = string(submatch[2])
							end = string(submatch[3])
						} else {
							TermError(filename, lineNum, "Invalid statement: "+line)
						}

						regex, err := regexp.Compile("(?" + flags + ")" + "(" + start + ").*?(" + end + ")")
						if err != nil {
							TermError(filename, lineNum, err.Error())
							continue
						}

						st := tcell.StyleDefault
						if _, ok := colorscheme[color]; ok {
							st = colorscheme[color]
						} else {
							st = StringToStyle(color)
						}
						rules = append(rules, SyntaxRule{regex, flags, true, st})
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

// SyntaxMatches is an alias to a map from character numbers to styles,
// so map[3] represents the style of the third character
type SyntaxMatches map[int]tcell.Style

// Match takes a buffer and returns the syntax matches a map specifying how it should be syntax highlighted
func Match(rules []SyntaxRule, buf *Buffer, v *View) SyntaxMatches {
	m := make(SyntaxMatches)

	lineStart := v.updateLines[0]
	lineEnd := v.updateLines[1] + 1
	if lineStart < 0 {
		// Don't need to update syntax highlighting
		return m
	}

	totalStart := v.topline - synLinesUp
	totalEnd := v.topline + v.height + synLinesDown
	if totalStart < 0 {
		totalStart = 0
	}
	if totalEnd > len(buf.lines) {
		totalEnd = len(buf.lines)
	}

	if lineEnd > len(buf.lines) {
		lineEnd = len(buf.lines)
	}

	lines := buf.lines[lineStart:lineEnd]
	str := strings.Join(buf.lines[totalStart:totalEnd], "\n")
	startNum := v.cursor.loc + v.cursor.Distance(0, totalStart)
	toplineNum := v.cursor.loc + v.cursor.Distance(0, v.topline)
	for _, rule := range rules {
		if rule.startend && rule.regex.MatchString(str) {
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
		} else {
			for _, line := range lines {
				if rule.regex.MatchString(line) {
					indicies := rule.regex.FindAllStringIndex(str, -1)
					for _, value := range indicies {
						value[0] += toplineNum
						value[1] += toplineNum
						for i := value[0]; i < value[1]; i++ {
							m[i] = rule.style
						}
					}
				}
			}
		}
	}

	return m
}
