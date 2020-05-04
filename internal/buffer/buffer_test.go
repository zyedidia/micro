package buffer

import (
	"strings"
	"testing"

	testifyAssert "github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"

	ulua "github.com/zyedidia/micro/v2/internal/lua"
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

	assert.NotEqual("", b.GetName())
	assert.Equal(false, b.ExternallyModified())
	assert.Equal(false, b.Modified())
	assert.Equal(1, b.NumCursors())

	checkText := func(lines []string) {
		assert.Equal([]byte(strings.Join(lines, "\n")), b.Bytes())
		assert.Equal(len(lines), b.LinesNum())
		for i, s := range lines {
			assert.Equal(s, b.Line(i))
			assert.Equal([]byte(s), b.LineBytes(i))
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

	assert.Equal(1+len(operations), b.NumCursors())

	for i, op := range operations {
		cursor := cursors[i]
		b.SetCurCursor(cursor.Num)
		cursor.DeleteSelection()
		b.Insert(cursor.Loc, strings.Join(op.text, "\n"))
	}

	checkText(after)

	// must have exactly two events per operation (delete and insert)
	for range operations {
		b.UndoOneEvent()
		b.UndoOneEvent()
	}

	checkText(before)

	for i, op := range operations {
		cursor := cursors[i]
		if op.start == op.end {
			assert.Equal(op.start, cursor.Loc)
		} else {
			assert.Equal(op.start, cursor.CurSelection[0])
			assert.Equal(op.end, cursor.CurSelection[1])
		}
	}

	for range operations {
		b.RedoOneEvent()
		b.RedoOneEvent()
	}

	checkText(after)

	b.Close()
}
