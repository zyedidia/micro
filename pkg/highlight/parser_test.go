package highlight

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseTestDef(t *testing.T, src string) (*Def, *File) {
	t.Helper()
	f, err := ParseFile([]byte(src))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	d, err := ParseDef(f, nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return d, f
}

func TestResolveIncludesSkipsInvalidFile(t *testing.T) {
	def, _ := parseTestDef(t, `
filetype: outer
rules:
    - include: "broken"
    - default:
        start: "<"
        end: ">"
        rules:
            - include: "broken"
`)

	broken, err := ParseFile([]byte(`
filetype: broken
rules:
    - identifier: "("
`))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.NotPanics(t, func() {
		ResolveIncludes(def, []*File{broken})
	})
	assert.Empty(t, def.rules.patterns)
	assert.Empty(t, def.rules.regions[0].rules.patterns)
}
