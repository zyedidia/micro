package main

import (
	"strings"
	"time"

	dmp "github.com/sergi/go-diff/diffmatchpatch"
	"github.com/yuin/gopher-lua"
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
	C Cursor

	EventType int
	Text      string
	Start     Loc
	End       Loc
	Time      time.Time
}

// ExecuteTextEvent runs a text event
func ExecuteTextEvent(t *TextEvent, buf *Buffer) {
	if t.EventType == TextEventInsert {
		buf.insert(t.Start, []byte(t.Text))
	} else if t.EventType == TextEventRemove {
		t.Text = buf.remove(t.Start, t.End)
	}
}

// UndoTextEvent undoes a text event
func UndoTextEvent(t *TextEvent, buf *Buffer) {
	t.EventType = -t.EventType
	ExecuteTextEvent(t, buf)
}

// EventHandler executes text manipulations and allows undoing and redoing
type EventHandler struct {
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

// ApplyDiff takes a string and runs the necessary insertion and deletion events to make
// the buffer equal to that string
// This means that we can transform the buffer into any string and still preserve undo/redo
// through insert and delete events
func (eh *EventHandler) ApplyDiff(new string) {
	differ := dmp.New()
	diff := differ.DiffMain(eh.buf.String(), new, false)
	loc := eh.buf.Start()
	for _, d := range diff {
		if d.Type == dmp.DiffDelete {
			eh.Remove(loc, loc.Move(Count(d.Text), eh.buf))
		} else {
			if d.Type == dmp.DiffInsert {
				eh.Insert(loc, d.Text)
			}
			loc = loc.Move(Count(d.Text), eh.buf)
		}
	}
}

// Insert creates an insert text event and executes it
func (eh *EventHandler) Insert(start Loc, text string) {
	e := &TextEvent{
		C:         eh.buf.Cursor,
		EventType: TextEventInsert,
		Text:      text,
		Start:     start,
		Time:      time.Now(),
	}
	eh.Execute(e)
	e.End = start.Move(Count(text), eh.buf)
}

// Remove creates a remove text event and executes it
func (eh *EventHandler) Remove(start, end Loc) {
	e := &TextEvent{
		C:         eh.buf.Cursor,
		EventType: TextEventRemove,
		Start:     start,
		End:       end,
		Time:      time.Now(),
	}
	eh.Execute(e)
}

// Replace deletes from start to end and replaces it with the given string
func (eh *EventHandler) Replace(start, end Loc, replace string) {
	eh.Remove(start, end)
	eh.Insert(start, replace)
}

// Execute a textevent and add it to the undo stack
func (eh *EventHandler) Execute(t *TextEvent) {
	if eh.RedoStack.Len() > 0 {
		eh.RedoStack = new(Stack)
	}
	eh.UndoStack.Push(t)

	for pl := range loadedPlugins {
		ret, err := Call(pl+".onBeforeTextEvent", t)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
		}
		if val, ok := ret.(lua.LBool); ok && val == lua.LFalse {
			return
		}
	}

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
		startTime = t.Time.UnixNano() / int64(time.Millisecond)

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

	// Set the cursor in the right place
	teCursor := t.C
	t.C = eh.buf.Cursor
	eh.buf.Cursor.Goto(teCursor)

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

	teCursor := t.C
	t.C = eh.buf.Cursor
	eh.buf.Cursor.Goto(teCursor)

	eh.UndoStack.Push(t)
}
