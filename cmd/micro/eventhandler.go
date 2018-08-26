package main

import (
	"time"
	"unicode/utf8"

	dmp "github.com/sergi/go-diff/diffmatchpatch"
)

const (
	// Opposite and undoing events must have opposite values

	// TextEventInsert represents an insertion event
	TextEventInsert = 1
	// TextEventRemove represents a deletion event
	TextEventRemove = -1
	// TextEventReplace represents a replace event
	TextEventReplace = 0
)

// TextEvent holds data for a manipulation on some text that can be undone
type TextEvent struct {
	C Cursor

	EventType int
	Deltas    []Delta
	Time      time.Time
}

// A Delta is a change to the buffer
type Delta struct {
	Text  []byte
	Start Loc
	End   Loc
}

// ExecuteTextEvent runs a text event
func ExecuteTextEvent(t *TextEvent, buf *Buffer) {
	if t.EventType == TextEventInsert {
		for _, d := range t.Deltas {
			buf.insert(d.Start, d.Text)
		}
	} else if t.EventType == TextEventRemove {
		for i, d := range t.Deltas {
			t.Deltas[i].Text = buf.remove(d.Start, d.End)
		}
	} else if t.EventType == TextEventReplace {
		for i, d := range t.Deltas {
			t.Deltas[i].Text = buf.remove(d.Start, d.End)
			buf.insert(d.Start, d.Text)
			t.Deltas[i].Start = d.Start
			t.Deltas[i].End = Loc{d.Start.X + utf8.RuneCount(d.Text), d.Start.Y}
		}
		for i, j := 0, len(t.Deltas)-1; i < j; i, j = i+1, j-1 {
			t.Deltas[i], t.Deltas[j] = t.Deltas[j], t.Deltas[i]
		}
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
	UndoStack *TEStack
	RedoStack *TEStack
}

// NewEventHandler returns a new EventHandler
func NewEventHandler(buf *Buffer) *EventHandler {
	eh := new(EventHandler)
	eh.UndoStack = new(TEStack)
	eh.RedoStack = new(TEStack)
	eh.buf = buf
	return eh
}

// ApplyDiff takes a string and runs the necessary insertion and deletion events to make
// the buffer equal to that string
// This means that we can transform the buffer into any string and still preserve undo/redo
// through insert and delete events
func (eh *EventHandler) ApplyDiff(new string) {
	differ := dmp.New()
	diff := differ.DiffMain(string(eh.buf.Bytes()), new, false)
	loc := eh.buf.Start()
	for _, d := range diff {
		if d.Type == dmp.DiffDelete {
			eh.Remove(loc, loc.Move(utf8.RuneCountInString(d.Text), eh.buf))
		} else {
			if d.Type == dmp.DiffInsert {
				eh.Insert(loc, d.Text)
			}
			loc = loc.Move(utf8.RuneCountInString(d.Text), eh.buf)
		}
	}
}

// Insert creates an insert text event and executes it
func (eh *EventHandler) Insert(start Loc, textStr string) {
	text := []byte(textStr)
	e := &TextEvent{
		C:         *eh.buf.GetActiveCursor(),
		EventType: TextEventInsert,
		Deltas:    []Delta{{text, start, Loc{0, 0}}},
		Time:      time.Now(),
	}
	eh.Execute(e)
	e.Deltas[0].End = start.Move(utf8.RuneCount(text), eh.buf)
	end := e.Deltas[0].End

	for _, c := range eh.buf.GetCursors() {
		move := func(loc Loc) Loc {
			if start.Y != end.Y && loc.GreaterThan(start) {
				loc.Y += end.Y - start.Y
			} else if loc.Y == start.Y && loc.GreaterEqual(start) {
				loc = loc.Move(utf8.RuneCount(text), eh.buf)
			}
			return loc
		}
		c.Loc = move(c.Loc)
		c.CurSelection[0] = move(c.CurSelection[0])
		c.CurSelection[1] = move(c.CurSelection[1])
		c.OrigSelection[0] = move(c.OrigSelection[0])
		c.OrigSelection[1] = move(c.OrigSelection[1])
		c.LastVisualX = c.GetVisualX()
	}
}

// Remove creates a remove text event and executes it
func (eh *EventHandler) Remove(start, end Loc) {
	e := &TextEvent{
		C:         *eh.buf.GetActiveCursor(),
		EventType: TextEventRemove,
		Deltas:    []Delta{{[]byte{}, start, end}},
		Time:      time.Now(),
	}
	eh.Execute(e)

	for _, c := range eh.buf.GetCursors() {
		move := func(loc Loc) Loc {
			if start.Y != end.Y && loc.GreaterThan(end) {
				loc.Y -= end.Y - start.Y
			} else if loc.Y == end.Y && loc.GreaterEqual(end) {
				loc = loc.Move(-Diff(start, end, eh.buf), eh.buf)
			}
			return loc
		}
		c.Loc = move(c.Loc)
		c.CurSelection[0] = move(c.CurSelection[0])
		c.CurSelection[1] = move(c.CurSelection[1])
		c.OrigSelection[0] = move(c.OrigSelection[0])
		c.OrigSelection[1] = move(c.OrigSelection[1])
		c.LastVisualX = c.GetVisualX()
	}
}

// MultipleReplace creates an multiple insertions executes them
func (eh *EventHandler) MultipleReplace(deltas []Delta) {
	e := &TextEvent{
		C:         *eh.buf.GetActiveCursor(),
		EventType: TextEventReplace,
		Deltas:    deltas,
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
		eh.RedoStack = new(TEStack)
	}
	eh.UndoStack.Push(t)

	// TODO: Call plugins on text events
	// for pl := range loadedPlugins {
	// 	ret, err := Call(pl+".onBeforeTextEvent", t)
	// 	if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
	// 		TermMessage(err)
	// 	}
	// 	if val, ok := ret.(lua.LBool); ok && val == lua.LFalse {
	// 		return
	// 	}
	// }

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
	if teCursor.Num >= 0 && teCursor.Num < eh.buf.NumCursors() {
		t.C = *eh.buf.GetCursor(teCursor.Num)
		eh.buf.GetCursor(teCursor.Num).Goto(teCursor)
	} else {
		teCursor.Num = -1
	}

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
	if teCursor.Num >= 0 && teCursor.Num < eh.buf.NumCursors() {
		t.C = *eh.buf.GetCursor(teCursor.Num)
		eh.buf.GetCursor(teCursor.Num).Goto(teCursor)
	} else {
		teCursor.Num = -1
	}

	eh.UndoStack.Push(t)
}
