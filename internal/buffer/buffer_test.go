package buffer

import (
	"strings"
	"testing"

	testifyAssert "github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"

	ulua "github.com/zyedidia/micro/internal/lua"
)

type operation struct {
	start Loc
	end   Loc
	text  []string
}

type asserter interface {
	Equal(interface{}, interface{}, ...interface{}) bool
	NotEqual(interface{}, interface{}, ...interface{}) bool
}

type noOpAsserter struct {
}

func (a *noOpAsserter) Equal(interface{}, interface{}, ...interface{}) bool {
	return true
}

func (a *noOpAsserter) NotEqual(interface{}, interface{}, ...interface{}) bool {
	return true
}

func init() {
	ulua.L = lua.NewState()
}

func check(t *testing.T, before []string, operations []operation, after []string) {
	var assert asserter
	if t == nil {
		// Benchmark mode; don't perform assertions
		assert = &noOpAsserter{}
	} else {
		assert = testifyAssert.New(t)
	}

	b := NewBufferFromString(strings.Join(before, "\n"), "", BTDefault)

	assert.NotEqual(b.GetName(), "")
	assert.Equal(b.ExternallyModified(), false)
	assert.Equal(b.Modified(), false)
	assert.Equal(b.NumCursors(), 1)

	checkText := func(lines []string) {
		assert.Equal(b.Bytes(), []byte(strings.Join(lines, "\n")))
		assert.Equal(b.LinesNum(), len(lines))
		for i, s := range lines {
			assert.Equal(b.Line(i), s)
			assert.Equal(b.LineBytes(i), []byte(s))
		}
	}

	checkText(before)

	var cursors []*Cursor

	for _, op := range operations {
		cursor := NewCursor(b, op.start)
		cursor.SetSelectionStart(op.start)
		cursor.SetSelectionEnd(op.end)
		b.AddCursor(cursor)
		cursors = append(cursors, cursor)
	}

	assert.Equal(b.NumCursors(), 1+len(operations))

	for i, op := range operations {
		cursor := cursors[i]
		cursor.DeleteSelection()
		b.Insert(cursor.Loc, strings.Join(op.text, "\n"))
	}

	checkText(after)

	for _ = range operations {
		b.UndoOneEvent()
		b.UndoOneEvent()
	}

	checkText(before)

	for i, op := range operations {
		cursor := cursors[i]
		assert.Equal(cursor.Loc, op.start)
		assert.Equal(cursor.CurSelection[0], op.start)
		assert.Equal(cursor.CurSelection[1], op.end)
	}

	for _ = range operations {
		b.RedoOneEvent()
		b.RedoOneEvent()
	}

	checkText(after)

	b.Close()
}
