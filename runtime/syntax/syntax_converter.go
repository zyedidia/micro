//+build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type SingleRule struct {
	color string
	regex string
}

type MultiRule struct {
	color string
	start string
	end   string
}

// JoinRule takes a syntax rule (which can be multiple regular expressions)
// and joins it into one regular expression by ORing everything together
func JoinRule(rule string) string {
	split := strings.Split(rule, `" "`)
	joined := strings.Join(split, "|")
	return joined
}

func parseFile(text, filename string) (filetype, syntax, header string, rules []interface{}) {
	lines := strings.Split(text, "\n")

	// Regex for parsing syntax statements
	syntaxParser := regexp.MustCompile(`syntax "(.*?)"\s+"(.*)"+`)
	// Regex for parsing header statements
	headerParser := regexp.MustCompile(`header "(.*)"`)

	// Regex for parsing standard syntax rules
	ruleParser := regexp.MustCompile(`color (.*?)\s+(?:\((.+?)?\)\s+)?"(.*)"`)
	// Regex for parsing syntax rules with start="..." end="..."
	ruleStartEndParser := regexp.MustCompile(`color (.*?)\s+(?:\((.+?)?\)\s+)?start="(.*)"\s+end="(.*)"`)

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "syntax") {
			syntaxMatches := syntaxParser.FindSubmatch([]byte(line))
			if len(syntaxMatches) == 3 {
				filetype = string(syntaxMatches[1])
				syntax = JoinRule(string(syntaxMatches[2]))
			} else {
				fmt.Println(filename, lineNum, "Syntax statement is not valid: "+line)
				continue
			}
		}
		if strings.HasPrefix(line, "header") {
			// Header statement
			headerMatches := headerParser.FindSubmatch([]byte(line))
			if len(headerMatches) == 2 {
				header = JoinRule(string(headerMatches[1]))
			} else {
				fmt.Println(filename, lineNum, "Header statement is not valid: "+line)
				continue
			}
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
				if flags != "" {
					regexStr = "(?" + flags + ")" + JoinRule(string(submatch[3]))
				} else {
					regexStr = JoinRule(string(submatch[3]))
				}
			} else if len(submatch) == 3 {
				// If len is 3, no additional flags were given
				color = string(submatch[1])
				regexStr = JoinRule(string(submatch[2]))
			} else {
				// If len is not 3 or 4 there is a problem
				fmt.Println(filename, lineNum, "Invalid statement: "+line)
				continue
			}

			rules = append(rules, SingleRule{color, regexStr})
		} else if ruleStartEndParser.MatchString(line) {
			// Start-end syntax rule
			submatch := ruleStartEndParser.FindSubmatch([]byte(line))
			var color string
			var start string
			var end string
			// Use m and s flags by default
			if len(submatch) == 5 {
				// If len is 5 the user provided some additional flags
				color = string(submatch[1])
				start = string(submatch[3])
				end = string(submatch[4])
			} else if len(submatch) == 4 {
				// If len is 4 the user did not provide additional flags
				color = string(submatch[1])
				start = string(submatch[2])
				end = string(submatch[3])
			} else {
				// If len is not 4 or 5 there is a problem
				fmt.Println(filename, lineNum, "Invalid statement: "+line)
				continue
			}

			// rules[color] = "(?" + flags + ")" + "(" + start + ").*?(" + end + ")"
			rules = append(rules, MultiRule{color, start, end})
		}
	}

	return
}

func generateFile(filetype, syntax, header string, rules []interface{}) string {
	output := ""

	output += fmt.Sprintf("filetype: %s\n\n", filetype)
	output += fmt.Sprintf("detect: \n    filename: \"%s\"\n", strings.Replace(strings.Replace(syntax, "\\", "\\\\", -1), "\"", "\\\"", -1))

	if header != "" {
		output += fmt.Sprintf("    signature: \"%s\"\n", strings.Replace(strings.Replace(header, "\\", "\\\\", -1), "\"", "\\\"", -1))
	}

	output += "\nrules:\n"

	for _, r := range rules {
		if rule, ok := r.(SingleRule); ok {
			output += fmt.Sprintf("    - %s: \"%s\"\n", rule.color, strings.Replace(strings.Replace(rule.regex, "\\", "\\\\", -1), "\"", "\\\"", -1))
		} else if rule, ok := r.(MultiRule); ok {
			output += fmt.Sprintf("    - %s:\n", rule.color)
			output += fmt.Sprintf("        start: \"%s\"\n", strings.Replace(strings.Replace(rule.start, "\\", "\\\\", -1), "\"", "\\\"", -1))
			output += fmt.Sprintf("        end: \"%s\"\n", strings.Replace(strings.Replace(rule.end, "\\", "\\\\", -1), "\"", "\\\"", -1))
		}
	}

	return output
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("no args")
		return
	}

	data, _ := ioutil.ReadFile(os.Args[1])
	fmt.Print(generateFile(parseFile(string(data), os.Args[1])))
}
