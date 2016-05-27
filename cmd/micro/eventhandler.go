package main

import (
	"time"
)

const (
	// Opposite and undoing events must have opposite values

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
	time      time.Time
}

// ExecuteTextEvent runs a text event
func ExecuteTextEvent(t *TextEvent, buf *Buffer) {
	if t.eventType == TextEventInsert {
		buf.insert(t.start, t.text)
	} else if t.eventType == TextEventRemove {
		t.text = buf.remove(t.start, t.end)
	}
}

// UndoTextEvent undoes a text event
func UndoTextEvent(t *TextEvent, buf *Buffer) {
	t.eventType = -t.eventType
	ExecuteTextEvent(t, buf)
}

// EventHandler executes text manipulations and allows undoing and redoing
type EventHandler struct {
	buf  *Buffer
	undo *Stack
	redo *Stack
}

// NewEventHandler returns a new EventHandler
func NewEventHandler(buf *Buffer) *EventHandler {
	eh := new(EventHandler)
	eh.undo = new(Stack)
	eh.redo = new(Stack)
	eh.buf = buf
	return eh
}

// Insert creates an insert text event and executes it
func (eh *EventHandler) Insert(start int, text string) {
	e := &TextEvent{
		c:         eh.buf.Cursor,
		eventType: TextEventInsert,
		text:      text,
		start:     start,
		end:       start + Count(text),
		time:      time.Now(),
	}
	eh.Execute(e)
}

// Remove creates a remove text event and executes it
func (eh *EventHandler) Remove(start, end int) {
	e := &TextEvent{
		c:         eh.buf.Cursor,
		eventType: TextEventRemove,
		start:     start,
		end:       end,
		time:      time.Now(),
	}
	eh.Execute(e)
}

// Replace deletes from start to end and replaces it with the given string
func (eh *EventHandler) Replace(start, end int, replace string) {
	eh.Remove(start, end)
	eh.Insert(start, replace)
}

// Execute a textevent and add it to the undo stack
func (eh *EventHandler) Execute(t *TextEvent) {
	if eh.redo.Len() > 0 {
		eh.redo = new(Stack)
	}
	eh.undo.Push(t)
	ExecuteTextEvent(t, eh.buf)
}

// Undo the first event in the undo stack
func (eh *EventHandler) Undo() {
	t := eh.undo.Peek()
	if t == nil {
		return
	}

	te := t.(*TextEvent)
	startTime := t.(*TextEvent).time.UnixNano() / int64(time.Millisecond)

	eh.UndoOneEvent()

	for {
		t = eh.undo.Peek()
		if t == nil {
			return
		}

		te = t.(*TextEvent)

		if startTime-(te.time.UnixNano()/int64(time.Millisecond)) > undoThreshold {
			return
		} else {
			startTime = t.(*TextEvent).time.UnixNano() / int64(time.Millisecond)
		}

		eh.UndoOneEvent()
	}
}

// UndoOneEvent undoes one event
func (eh *EventHandler) UndoOneEvent() {
	// This event should be undone
	// Pop it off the stack
	t := eh.undo.Pop()
	if t == nil {
		return
	}

	te := t.(*TextEvent)
	// Undo it
	// Modifies the text event
	UndoTextEvent(te, eh.buf)

	// Set the cursor in the right place
	teCursor := te.c
	te.c = eh.buf.Cursor
	eh.buf.Cursor = teCursor

	// Push it to the redo stack
	eh.redo.Push(te)
}

// Redo the first event in the redo stack
func (eh *EventHandler) Redo() {
	t := eh.redo.Peek()
	if t == nil {
		return
	}

	te := t.(*TextEvent)
	startTime := t.(*TextEvent).time.UnixNano() / int64(time.Millisecond)

	eh.RedoOneEvent()

	for {
		t = eh.redo.Peek()
		if t == nil {
			return
		}

		te = t.(*TextEvent)

		if (te.time.UnixNano()/int64(time.Millisecond))-startTime > undoThreshold {
			return
		}

		eh.RedoOneEvent()
	}
}

// RedoOneEvent redoes one event
func (eh *EventHandler) RedoOneEvent() {
	t := eh.redo.Pop()
	if t == nil {
		return
	}

	te := t.(*TextEvent)
	// Modifies the text event
	UndoTextEvent(te, eh.buf)

	teCursor := te.c
	te.c = eh.buf.Cursor
	eh.buf.Cursor = teCursor

	eh.undo.Push(te)
}
