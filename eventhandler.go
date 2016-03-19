package main

import (
	"time"
)

const (
	// TextEventInsert repreasents an insertion event
	TextEventInsert = 1
	// TextEventRemove represents a deletion event
	TextEventRemove = -1
)

// TextEvent holds data for a manipulation on some text that can be undone
type TextEvent struct {
	c Cursor

	eventType int
	text      string
	start     int
	end       int
	buf       *Buffer
	time      time.Time
}

// ExecuteTextEvent runs a text event
func ExecuteTextEvent(t *TextEvent) {
	if t.eventType == TextEventInsert {
		t.buf.Insert(t.start, t.text)
	} else if t.eventType == TextEventRemove {
		t.text = t.buf.Remove(t.start, t.end)
	}
}

// UndoTextEvent undoes a text event
func UndoTextEvent(t *TextEvent) {
	t.eventType = -t.eventType
	ExecuteTextEvent(t)
}

// EventHandler executes text manipulations and allows undoing and redoing
type EventHandler struct {
	v    *View
	undo *Stack
	redo *Stack
}

// NewEventHandler returns a new EventHandler
func NewEventHandler(v *View) *EventHandler {
	eh := new(EventHandler)
	eh.undo = new(Stack)
	eh.redo = new(Stack)
	eh.v = v
	return eh
}

// Insert creates an insert text event and executes it
func (eh *EventHandler) Insert(start int, text string) {
	e := &TextEvent{
		c:         eh.v.cursor,
		eventType: TextEventInsert,
		text:      text,
		start:     start,
		end:       start + len(text),
		buf:       eh.v.buf,
		time:      time.Now(),
	}
	eh.Execute(e)
}

// Remove creates a remove text event and executes it
func (eh *EventHandler) Remove(start, end int) {
	e := &TextEvent{
		c:         eh.v.cursor,
		eventType: TextEventRemove,
		start:     start,
		end:       end,
		buf:       eh.v.buf,
		time:      time.Now(),
	}
	eh.Execute(e)
}

// Execute a textevent and add it to the undo stack
func (eh *EventHandler) Execute(t *TextEvent) {
	eh.undo.Push(t)
	ExecuteTextEvent(t)
}

// Undo the first event in the undo stack
func (eh *EventHandler) Undo() {
	t := eh.undo.Pop()
	if t == nil {
		return
	}

	te := t.(*TextEvent)
	// Modifies the text event
	UndoTextEvent(te)
	eh.redo.Push(t)

	eh.v.cursor = te.c
}

// Redo the first event in the redo stack
func (eh *EventHandler) Redo() {
	t := eh.redo.Pop()
	if t == nil {
		return
	}

	te := t.(*TextEvent)
	// Modifies the text event
	UndoTextEvent(te)
	eh.undo.Push(t)
	eh.v.cursor = te.c
}
