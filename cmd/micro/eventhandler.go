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
	Cursor Cursor

	EventType int
	Text      string
	Start     int
	End       int
	Time      time.Time
}

// ExecuteTextEvent runs a text event
func ExecuteTextEvent(t *TextEvent, buf *Buffer) {
	if t.EventType == TextEventInsert {
		buf.Insert(t.Start, t.Text)
	} else if t.EventType == TextEventRemove {
		t.Text = buf.Remove(t.Start, t.End)
	}
}

// UndoTextEvent undoes a text event
func UndoTextEvent(t *TextEvent, buf *Buffer) {
	t.EventType = -t.EventType
	ExecuteTextEvent(t, buf)
}

// EventHandler executes text manipulations and allows undoing and redoing
type EventHandler struct {
	cursor    *Cursor
	buf       *Buffer
	UndoStack *Stack
	RedoStack *Stack
}

// NewEventHandler returns a new EventHandler
func NewEventHandler(buf *Buffer) *EventHandler {
	eh := new(EventHandler)
	eh.UndoStack = new(Stack)
	eh.RedoStack = new(Stack)
	eh.buf = buf
	return eh
}

// Insert creates an insert text event and executes it
func (eh *EventHandler) Insert(start int, text string) {
	e := &TextEvent{
		Cursor:    *eh.cursor,
		EventType: TextEventInsert,
		Text:      text,
		Start:     start,
		End:       start + Count(text),
		Time:      time.Now(),
	}
	eh.Execute(e)
}

// Remove creates a remove text event and executes it
func (eh *EventHandler) Remove(start, end int) {
	e := &TextEvent{
		Cursor:    *eh.cursor,
		EventType: TextEventRemove,
		Start:     start,
		End:       end,
		Time:      time.Now(),
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
	if eh.RedoStack.Len() > 0 {
		eh.RedoStack = new(Stack)
	}
	eh.UndoStack.Push(t)
	ExecuteTextEvent(t, eh.buf)
}

// Undo the first event in the undo stack
func (eh *EventHandler) Undo() {
	t := eh.UndoStack.Peek()
	if t == nil {
		return
	}

	startTime := t.Time.UnixNano() / int64(time.Millisecond)

	eh.UndoOneEvent()

	for {
		t = eh.UndoStack.Peek()
		if t == nil {
			return
		}

		if startTime-(t.Time.UnixNano()/int64(time.Millisecond)) > undoThreshold {
			return
		}

		eh.UndoOneEvent()
	}
}

// UndoOneEvent undoes one event
func (eh *EventHandler) UndoOneEvent() {
	// This event should be undone
	// Pop it off the stack
	t := eh.UndoStack.Pop()
	if t == nil {
		return
	}

	// Undo it
	// Modifies the text event
	UndoTextEvent(t, eh.buf)

	eh.cursor.X = t.Cursor.X
	eh.cursor.Y = t.Cursor.Y
	eh.cursor.LastVisualX = t.Cursor.LastVisualX
	eh.cursor.CurSelection = t.Cursor.CurSelection
	eh.cursor.OrigSelection = t.Cursor.OrigSelection

	// Push it to the redo stack
	eh.RedoStack.Push(t)
}

// Redo the first event in the redo stack
func (eh *EventHandler) Redo() {
	t := eh.RedoStack.Peek()
	if t == nil {
		return
	}

	startTime := t.Time.UnixNano() / int64(time.Millisecond)

	eh.RedoOneEvent()

	for {
		t = eh.RedoStack.Peek()
		if t == nil {
			return
		}

		if (t.Time.UnixNano()/int64(time.Millisecond))-startTime > undoThreshold {
			return
		}

		eh.RedoOneEvent()
	}
}

// RedoOneEvent redoes one event
func (eh *EventHandler) RedoOneEvent() {
	t := eh.RedoStack.Pop()
	if t == nil {
		return
	}

	// Modifies the text event
	UndoTextEvent(t, eh.buf)

	eh.cursor.X = t.Cursor.X + 1
	eh.cursor.Y = t.Cursor.Y
	eh.cursor.LastVisualX = t.Cursor.LastVisualX
	eh.cursor.CurSelection = t.Cursor.CurSelection
	eh.cursor.OrigSelection = t.Cursor.OrigSelection

	eh.UndoStack.Push(t)
}
