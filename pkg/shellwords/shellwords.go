package shellwords

import (
	"bytes"
	"errors"
	"os"
	"regexp"
)

var envRe = regexp.MustCompile(`\$({[a-zA-Z0-9_]+}|[a-zA-Z0-9_]+)`)

func isSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\r', '\n':
		return true
	}
	return false
}

func replaceEnv(s string) string {
	return envRe.ReplaceAllStringFunc(s, func(s string) string {
		s = s[1:]
		if s[0] == '{' {
			s = s[1 : len(s)-1]
		}
		return os.Getenv(s)
	})
}

type Parser struct {
	Position int
}

func NewParser() *Parser {
	return &Parser{0}
}

func (p *Parser) Parse(line string) ([]string, error) {
	args := []string{}
	buf := ""
	var escaped, doubleQuoted, singleQuoted, backQuote, dollarQuote bool
	backtick := ""

	pos := -1
	got := false

loop:
	for i, r := range line {
		if escaped {
			buf += string(r)
			escaped = false
			continue
		}

		if r == '\\' {
			if singleQuoted {
				buf += string(r)
			} else {
				escaped = true
			}
			continue
		}

		if isSpace(r) {
			if singleQuoted || doubleQuoted || backQuote || dollarQuote {
				buf += string(r)
				backtick += string(r)
			} else if got {
				buf = replaceEnv(buf)
				args = append(args, buf)
				buf = ""
				got = false
			}
			continue
		}

		switch r {
		case '`':
			if !singleQuoted && !doubleQuoted && !dollarQuote {
				if backQuote {
					out, err := shellRun(backtick)
					if err != nil {
						return nil, err
					}
					buf = out
				}
				backtick = ""
				backQuote = !backQuote
				continue
				backtick = ""
				backQuote = !backQuote
			}
		case ')':
			if !singleQuoted && !doubleQuoted && !backQuote {
				if dollarQuote {
					out, err := shellRun(backtick)
					if err != nil {
						return nil, err
					}
					buf = out
				}
				backtick = ""
				dollarQuote = !dollarQuote
				continue
				backtick = ""
				dollarQuote = !dollarQuote
			}
		case '(':
			if !singleQuoted && !doubleQuoted && !backQuote {
				if !dollarQuote && len(buf) > 0 && buf == "$" {
					dollarQuote = true
					buf += "("
					continue
				} else {
					return nil, errors.New("invalid command line string")
				}
			}
		case '"':
			if !singleQuoted && !dollarQuote {
				doubleQuoted = !doubleQuoted
				continue
			}
		case '\'':
			if !doubleQuoted && !dollarQuote {
				singleQuoted = !singleQuoted
				continue
			}
		case ';', '&', '|', '<', '>':
			if !(escaped || singleQuoted || doubleQuoted || backQuote) {
				pos = i
				break loop
			}
		}

		got = true
		buf += string(r)
		if backQuote || dollarQuote {
			backtick += string(r)
		}
	}

	buf = replaceEnv(buf)
	args = append(args, buf)

	if escaped || singleQuoted || doubleQuoted || backQuote || dollarQuote {
		return nil, errors.New("invalid command line string")
	}

	p.Position = pos

	return args, nil
}

func Split(line string) ([]string, error) {
	return NewParser().Parse(line)
}

func Join(args ...string) string {
	var buf bytes.Buffer
	for i, w := range args {
		if i != 0 {
			buf.WriteByte(' ')
		}
		if w == "" {
			buf.WriteString("''")
			continue
		}

		for _, b := range w {
			switch b {
			case ' ', '\t', '\r', '\n':
				buf.WriteByte('\\')
				buf.WriteString(string(b))
			default:
				buf.WriteString(string(b))
			}
		}
	}
	return buf.String()
}
