package main

import (
	"regexp"
	"strings"

	"github.com/zyedidia/tcell"
)

// FileTypeRules represents a complete set of syntax rules for a filetype
type FileTypeRules struct {
	filetype string
	filename string
	text     string
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

var syntaxKeys [][2]*regexp.Regexp
var syntaxFiles map[[2]*regexp.Regexp]FileTypeRules

// LoadSyntaxFiles loads the syntax files from the default directory (configDir)
func LoadSyntaxFiles() {
	InitColorscheme()
	syntaxFiles = make(map[[2]*regexp.Regexp]FileTypeRules)
	for _, f := range ListRuntimeFiles(RTSyntax) {
		data, err := f.Data()
		if err != nil {
			TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
		} else {
			LoadSyntaxFile(string(data), f.Name())
		}
	}
}

// JoinRule takes a syntax rule (which can be multiple regular expressions)
// and joins it into one regular expression by ORing everything together
func JoinRule(rule string) string {
	split := strings.Split(rule, `" "`)
	joined := strings.Join(split, ")|(")
	joined = "(" + joined + ")"
	return joined
}

// LoadSyntaxFile simply gets the filetype of a the syntax file and the source for the
// file and creates FileTypeRules out of it. If this filetype is the one opened by the user
// the rules will be loaded and compiled later
// In this function we are only concerned with loading the syntax and header regexes
func LoadSyntaxFile(text, filename string) {
	var err error
	lines := strings.Split(string(text), "\n")

	// Regex for parsing syntax statements
	syntaxParser := regexp.MustCompile(`syntax "(.*?)"\s+"(.*)"+`)
	// Regex for parsing header statements
	headerParser := regexp.MustCompile(`header "(.*)"`)

	// Is there a syntax definition in this file?
	hasSyntax := syntaxParser.MatchString(text)
	// Is there a header definition in this file?
	hasHeader := headerParser.MatchString(text)

	var syntaxRegex *regexp.Regexp
	var headerRegex *regexp.Regexp
	var filetype string
	for lineNum, line := range lines {
		if (hasSyntax == (syntaxRegex != nil)) && (hasHeader == (headerRegex != nil)) {
			// We found what we we're supposed to find
			break
		}

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
					TermError(filename, lineNum, "Syntax statement redeclaration")
				}

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
		}
	}
	if syntaxRegex != nil {
		// Add the current rules to the syntaxFiles variable
		regexes := [2]*regexp.Regexp{syntaxRegex, headerRegex}
		syntaxKeys = append(syntaxKeys, regexes)
		syntaxFiles[regexes] = FileTypeRules{filetype, filename, text}
	}
}

// LoadRulesFromFile loads just the syntax rules from a given file
// Only the necessary rules are loaded when the buffer is opened.
// If we load all the rules for every filetype when micro starts, there's a bit of lag
// A rule just explains how to color certain regular expressions
// Example: color comment "//.*"
// This would color all strings that match the regex "//.*" in the comment color defined
// by the colorscheme
func LoadRulesFromFile(text, filename string) []SyntaxRule {
	lines := strings.Split(string(text), "\n")

	// Regex for parsing standard syntax rules
	ruleParser := regexp.MustCompile(`color (.*?)\s+(?:\((.*?)\)\s+)?"(.*)"`)
	// Regex for parsing syntax rules with start="..." end="..."
	ruleStartEndParser := regexp.MustCompile(`color (.*?)\s+(?:\((.*?)\)\s+)?start="(.*)"\s+end="(.*)"`)

	var rules []SyntaxRule
	for lineNum, line := range lines {
		if strings.TrimSpace(line) == "" ||
			strings.TrimSpace(line)[0] == '#' ||
			strings.HasPrefix(line, "syntax") ||
			strings.HasPrefix(line, "header") {
			// Ignore this line
			continue
		}

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
			st := defStyle
			groups := strings.Split(color, ".")
			if len(groups) > 1 {
				curGroup := ""
				for i, g := range groups {
					if i != 0 {
						curGroup += "."
					}
					curGroup += g
					if style, ok := colorscheme[curGroup]; ok {
						st = style
					}
				}
			} else if style, ok := colorscheme[color]; ok {
				st = style
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
			st := defStyle
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
	return rules
}

// FindFileType finds the filetype for the given buffer
func FindFileType(buf *Buffer) string {
	for _, r := range syntaxKeys {
		if r[1] != nil && r[1].MatchString(buf.Line(0)) {
			// The header statement matches the first line
			return syntaxFiles[r].filetype
		}
	}
	for _, r := range syntaxKeys {
		if r[0] != nil && r[0].MatchString(buf.Path) {
			// The syntax statement matches the extension
			return syntaxFiles[r].filetype
		}
	}
	return "Unknown"
}

// GetRules finds the syntax rules that should be used for the buffer
// and returns them. It also returns the filetype of the file
func GetRules(buf *Buffer) []SyntaxRule {
	for _, r := range syntaxKeys {
		if syntaxFiles[r].filetype == buf.FileType() {
			return LoadRulesFromFile(syntaxFiles[r].text, syntaxFiles[r].filename)
		}
	}
	return nil
}

// SyntaxMatches is an alias to a map from character numbers to styles,
// so map[3] represents the style of the third character
type SyntaxMatches [][]tcell.Style

// Match takes a buffer and returns the syntax matches: a 2d array specifying how it should be syntax highlighted
// We match the rules from up `synLinesUp` lines and down `synLinesDown` lines
func Match(v *View) SyntaxMatches {
	buf := v.Buf
	rules := v.Buf.rules

	viewStart := v.Topline
	viewEnd := v.Topline + v.Height
	if viewEnd > buf.NumLines {
		viewEnd = buf.NumLines
	}

	lines := buf.Lines(viewStart, viewEnd)
	matches := make(SyntaxMatches, len(lines))

	for i, line := range lines {
		matches[i] = make([]tcell.Style, len(line)+1)
		for j := range matches[i] {
			matches[i][j] = defStyle
		}
	}

	// We don't actually check the entire buffer, just from synLinesUp to synLinesDown
	totalStart := v.Topline - synLinesUp
	totalEnd := v.Topline + v.Height + synLinesDown
	if totalStart < 0 {
		totalStart = 0
	}
	if totalEnd > buf.NumLines {
		totalEnd = buf.NumLines
	}

	str := strings.Join(buf.Lines(totalStart, totalEnd), "\n")
	startNum := ToCharPos(Loc{0, totalStart}, v.Buf)

	for _, rule := range rules {
		if rule.startend {
			if indicies := rule.regex.FindAllStringIndex(str, -1); indicies != nil {
				for _, value := range indicies {
					value[0] = runePos(value[0], str) + startNum
					value[1] = runePos(value[1], str) + startNum
					startLoc := FromCharPos(value[0], buf)
					endLoc := FromCharPos(value[1], buf)
					for curLoc := startLoc; curLoc.LessThan(endLoc); curLoc = curLoc.Move(1, buf) {
						if curLoc.Y < v.Topline {
							continue
						}
						colNum, lineNum := curLoc.X, curLoc.Y
						if lineNum == -1 || colNum == -1 {
							continue
						}
						lineNum -= viewStart
						if lineNum >= 0 && lineNum < v.Height {
							matches[lineNum][colNum] = rule.style
						}
					}
				}
			}
		} else {
			for lineN, line := range lines {
				if indicies := rule.regex.FindAllStringIndex(line, -1); indicies != nil {
					for _, value := range indicies {
						start := runePos(value[0], line)
						end := runePos(value[1], line)
						for i := start; i < end; i++ {
							matches[lineN][i] = rule.style
						}
					}
				}
			}
		}
	}

	return matches
}
