package highlight

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const syntax = `
filetype: test
detect:
    filename: "test"
rules:
    - type: "\\b(one|two|three)\\b"
    - constant.string:
        start: "\""
        end: "\""
    - constant.string:
        start: "'"
        end: "'"
    - comment:
        start: "//"
        end: "$"
        rules:
            - todo: "(TODO):?"
    - comment:
        start: "/\\*"
        end: "\\*/"
        rules:
            - todo: "(TODO):?"
`

const content = `
one two three
// comment
one
two
three
/*
multi
line
// comment
with
TODO
*/
"string"
one two three
"
multi
line
str'ng
"
one two three
'string'
one two three
// '
one "string" two /*rule*/ three //all
/* " */
`

var expectation = []map[int]string{
	{},
	{0: "type", 3: "", 4: "type", 7: "", 8: "type"},
	{0: "comment"},
	{0: "type"},
	{0: "type"},
	{0: "type"},
	{0: "comment"},
	{0: "comment"},
	{0: "comment"},
	{0: "comment"},
	{0: "comment"},
	{0: "todo"},
	{0: "comment"},
	{0: "constant.string"},
	{0: "type", 3: "", 4: "type", 7: "", 8: "type"},
	{0: "constant.string"},
	{0: "constant.string"},
	{0: "constant.string"},
	{0: "constant.string"},
	{0: "constant.string"},
	{0: "type", 3: "", 4: "type", 7: "", 8: "type"},
	{0: "constant.string"},
	{0: "type", 3: "", 4: "type", 7: "", 8: "type"},
	{0: "comment"},
	{0: "type", 3: "", 4: "constant.string", 12: "", 13: "type", 16: "", 17: "comment", 25: "", 26: "type", 31: "", 32: "comment"},
	{0: "comment"},
	{},
}

func TestHighlightString(t *testing.T) {
	header, err := MakeHeaderYaml([]byte(syntax))
	if !assert.NoError(t, err) {
		return
	}

	file, err := ParseFile([]byte(syntax))
	if !assert.NoError(t, err) {
		return
	}

	def, err := ParseDef(file, header)
	if !assert.NoError(t, err) {
		return
	}

	highlighter := NewHighlighter(def)
	matches := highlighter.HighlightString(content)
	result := assert.Equal(t, len(expectation), len(matches))
	if !result {
		return
	}

	for i, m := range matches {
		result = assert.Equal(t, len(expectation[i]), len(m))
		if !result {
			return
		}
		actual := map[int]string{}
		for k, g := range m {
			actual[k] = g.String()
		}
		result := assert.Equal(t, expectation[i], actual, i)
		if !result {
			return
		}
	}
}
