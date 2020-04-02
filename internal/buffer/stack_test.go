package buffer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	s := new(TEStack)
	e1 := &TextEvent{
		EventType: TextEventReplace,
		Time:      time.Now(),
	}
	e2 := &TextEvent{
		EventType: TextEventInsert,
		Time:      time.Now(),
	}
	s.Push(e1)
	s.Push(e2)

	p := s.Peek()
	assert.Equal(t, p.EventType, TextEventInsert)
	p = s.Pop()
	assert.Equal(t, p.EventType, TextEventInsert)
	p = s.Peek()
	assert.Equal(t, p.EventType, TextEventReplace)
	p = s.Pop()
	assert.Equal(t, p.EventType, TextEventReplace)
	p = s.Pop()
	assert.Nil(t, p)
	p = s.Peek()
	assert.Nil(t, p)
}
