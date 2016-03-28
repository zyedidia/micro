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

// LoadSyntaxFile loads the specified syntax file
// A syntax file is a list of syntax rules, explaining how to color certain
// regular expressions
// Example: color comment "//.*"
// This would color all strings that match the regex "//.*" in the comment color defined
// by the colorscheme
func LoadSyntaxFile(filename string) {
	text, err := ioutil.ReadFile(filename)

	if err != nil {
		TermMessage("Error loading syntax file " + filename + ": " + err.Error())
		return
	}
	lines := strings.Split(string(text), "\n")

	// Regex for parsing syntax statements
	syntaxParser := regexp.MustCompile(`syntax "(.*?)"\s+"(.*)"+`)
	// Regex for parsing header statements
	headerParser := regexp.MustCompile(`header "(.*)"`)

	// Regex for parsing standard syntax rules
	ruleParser := regexp.MustCompile(`color (.*?)\s+(?:\((.*?)\)\s+)?"(.*)"`)
	// Regex for parsing syntax rules with start="..." end="..."
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
			// Syntax statement
			syntaxMatches := syntaxParser.FindSubmatch([]byte(line))
			if len(syntaxMatches) == 3 {
				if syntaxRegex != nil {
					// Add the current rules to the syntaxFiles variable
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
			// Header statement
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
			// Syntax rule, but it could be standard or start-end
			if ruleParser.MatchString(line) {
				// Standard syntax rule
				// Parse the line
				submatch := ruleParser.FindSubmatch([]byte(line))
				var color string
				var regexStr string
				var flags string
				if len(submatch) == 4 {
					// If len is 4 then the user specified some additional flags to use
					color = string(submatch[1])
					flags = string(submatch[2])
					regexStr = "(?" + flags + ")" + JoinRule(string(submatch[3]))
				} else if len(submatch) == 3 {
					// If len is 3, no additional flags were given
					color = string(submatch[1])
					regexStr = JoinRule(string(submatch[2]))
				} else {
					// If len is not 3 or 4 there is a problem
					TermError(filename, lineNum, "Invalid statement: "+line)
					continue
				}
				// Compile the regex
				regex, err := regexp.Compile(regexStr)
				if err != nil {
					TermError(filename, lineNum, err.Error())
					continue
				}

				// Get the style
				// The user could give us a "color" that is really a part of the colorscheme
				// in which case we should look that up in the colorscheme
				// They can also just give us a straight up color
				st := tcell.StyleDefault
				if _, ok := colorscheme[color]; ok {
					st = colorscheme[color]
				} else {
					st = StringToStyle(color)
				}
				// Add the regex, flags, and style
				// False because this is not start-end
				rules = append(rules, SyntaxRule{regex, flags, false, st})
			} else if ruleStartEndParser.MatchString(line) {
				// Start-end syntax rule
				submatch := ruleStartEndParser.FindSubmatch([]byte(line))
				var color string
				var start string
				var end string
				// Use m and s flags by default
				flags := "ms"
				if len(submatch) == 5 {
					// If len is 5 the user provided some additional flags
					color = string(submatch[1])
					flags += string(submatch[2])
					start = string(submatch[3])
					end = string(submatch[4])
				} else if len(submatch) == 4 {
					// If len is 4 the user did not provide additional flags
					color = string(submatch[1])
					start = string(submatch[2])
					end = string(submatch[3])
				} else {
					// If len is not 4 or 5 there is a problem
					TermError(filename, lineNum, "Invalid statement: "+line)
					continue
				}

				// Compile the regex
				regex, err := regexp.Compile("(?" + flags + ")" + "(" + start + ").*?(" + end + ")")
				if err != nil {
					TermError(filename, lineNum, err.Error())
					continue
				}

				// Get the style
				// The user could give us a "color" that is really a part of the colorscheme
				// in which case we should look that up in the colorscheme
				// They can also just give us a straight up color
				st := tcell.StyleDefault
				if _, ok := colorscheme[color]; ok {
					st = colorscheme[color]
				} else {
					st = StringToStyle(color)
				}
				// Add the regex, flags, and style
				// True because this is start-end
				rules = append(rules, SyntaxRule{regex, flags, true, st})
			}
		}
	}
	if syntaxRegex != nil {
		// Add the current rules to the syntaxFiles variable
		regexes := [2]*regexp.Regexp{syntaxRegex, headerRegex}
		syntaxFiles[regexes] = FileTypeRules{filetype, rules}
	}
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
			LoadSyntaxFile(dir + "/" + f.Name())
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
type SyntaxMatches [][]tcell.Style

// Match takes a buffer and returns the syntax matches a map specifying how it should be syntax highlighted
// We need to check the start-end regexes for the entire buffer every time Match is called, but for the
// non start-end rules, we only have to update the updateLines provided by the view
func Match(v *View) SyntaxMatches {
	buf := v.buf
	rules := v.buf.rules

	viewStart := v.topline
	viewEnd := v.topline + v.height
	if viewEnd > len(buf.lines) {
		viewEnd = len(buf.lines)
	}

	// updateStart := v.updateLines[0]
	// updateEnd := v.updateLines[1]
	//
	// if updateEnd > len(buf.lines) {
	// 	updateEnd = len(buf.lines)
	// }
	// if updateStart < 0 {
	// 	updateStart = 0
	// }
	lines := buf.lines[viewStart:viewEnd]
	// updateLines := buf.lines[updateStart:updateEnd]
	matches := make(SyntaxMatches, len(lines))

	for i, line := range lines {
		matches[i] = make([]tcell.Style, len(line))
	}

	// We don't actually check the entire buffer, just from synLinesUp to synLinesDown
	totalStart := v.topline - synLinesUp
	totalEnd := v.topline + v.height + synLinesDown
	if totalStart < 0 {
		totalStart = 0
	}
	if totalEnd > len(buf.lines) {
		totalEnd = len(buf.lines)
	}

	str := strings.Join(buf.lines[totalStart:totalEnd], "\n")
	startNum := ToCharPos(0, totalStart, v.buf)

	for _, rule := range rules {
		if rule.startend {
			if rule.regex.MatchString(str) {
				indicies := rule.regex.FindAllStringIndex(str, -1)
				for _, value := range indicies {
					value[0] += startNum
					value[1] += startNum
					for i := value[0]; i < value[1]; i++ {
						colNum, lineNum := FromCharPos(i, buf)
						if lineNum == -1 || colNum == -1 {
							continue
						}
						lineNum -= viewStart
						if lineNum >= 0 && lineNum < v.height {
							matches[lineNum][colNum] = rule.style
						}
					}
				}
			}
		} else {
			for lineN, line := range lines {
				if rule.regex.MatchString(line) {
					indicies := rule.regex.FindAllStringIndex(line, -1)
					for _, value := range indicies {
						for i := value[0]; i < value[1]; i++ {
							// matches[lineN+updateStart][i] = rule.style
							matches[lineN][i] = rule.style
						}
					}
				}
			}
		}
	}

	return matches
}
