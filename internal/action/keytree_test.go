package action

import (
	"testing"
	"time"

	"github.com/micro-editor/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestKeySequence(t *testing.T) {
	// Create a new key tree
	kt := NewKeyTree()
	
	// Track if our action was called
	sequenceActionCalled := false
	action := func(p Pane) bool {
		sequenceActionCalled = true
		return true
	}

	// Register a sequence: Ctrl-x Ctrl-s
	e1 := KeyEvent{code: tcell.KeyCtrlX, mod: tcell.ModCtrl}
	e2 := KeyEvent{code: tcell.KeyCtrlS, mod: tcell.ModCtrl}
	seq := KeySequenceEvent{keys: []Event{e1, e2}}
	
	kt.RegisterKeyBinding(seq, action)

	// Test the sequence
	// First key (Ctrl-x)
	a, more := kt.NextEvent(e1, nil)
	assert.Nil(t, a, "Action should be nil after first key")
	assert.True(t, more, "More keys expected in sequence")
	assert.False(t, sequenceActionCalled, "Action should not be called yet")

	// Second key (Ctrl-s) - should complete the sequence
	a, more = kt.NextEvent(e2, nil)
	assert.NotNil(t, a, "Action should not be nil after sequence completion")
	assert.False(t, more, "No more keys expected in sequence")
	
	// Execute the action
	a(nil)
	assert.True(t, sequenceActionCalled, "Action should be called after sequence completion")

	// Reset and test partial sequence with timeout
	kt.ResetEvents()
	sequenceActionCalled = false

	// First key (Ctrl-x)
	a, more = kt.NextEvent(e1, nil)
	assert.Nil(t, a, "Action should be nil after first key")
	assert.True(t, more, "More keys expected in sequence")

	// Wait for timeout
	time.Sleep(SequenceTimeout + 10*time.Millisecond)

	// Next key should start a new sequence
	a, more = kt.NextEvent(e2, nil)
	assert.Nil(t, a, "Action should be nil for invalid sequence")
	assert.False(t, more, "No more keys expected for invalid sequence")
}

func TestKeySequenceWithRunes(t *testing.T) {
	kt := NewKeyTree()
	
	// Track if our action was called
	actionCalled := false
	action := func(p Pane) bool {
		actionCalled = true
		return true
	}

	// Register a sequence: g g (like in vim)
	e1 := KeyEvent{code: tcell.KeyRune, r: 'g'}
	e2 := KeyEvent{code: tcell.KeyRune, r: 'g'}
	seq := KeySequenceEvent{keys: []Event{e1, e2}}
	
	kt.RegisterKeyBinding(seq, action)

	// Test the sequence
	// First 'g'
	a, more := kt.NextEvent(e1, nil)
	assert.Nil(t, a, "Action should be nil after first 'g'")
	assert.True(t, more, "More keys expected in sequence")

	// Second 'g' - should complete the sequence
	a, more = kt.NextEvent(e2, nil)
	assert.NotNil(t, a, "Action should not be nil after sequence completion")
	assert.False(t, more, "No more keys expected in sequence")
	
	// Execute the action
	a(nil)
	assert.True(t, actionCalled, "Action should be called after sequence completion")
}

func TestKeySequenceWithPartialMatch(t *testing.T) {
	kt := NewKeyTree()
	
	// Track which action was called
	action1Called := false
	action2Called := false

	action1 := func(p Pane) bool {
		action1Called = true
		return true
	}

	action2 := func(p Pane) bool {
		action2Called = true
		return true
	}

	// Register two sequences with common prefix: Ctrl-x Ctrl-s and Ctrl-x Ctrl-c
	e1 := KeyEvent{code: tcell.KeyCtrlX, mod: tcell.ModCtrl}
	e2s := KeyEvent{code: tcell.KeyCtrlS, mod: tcell.ModCtrl}
	e2c := KeyEvent{code: tcell.KeyCtrlC, mod: tcell.ModCtrl}
	
	seqSave := KeySequenceEvent{keys: []Event{e1, e2s}}
	seqCancel := KeySequenceEvent{keys: []Event{e1, e2c}}
	
	kt.RegisterKeyBinding(seqSave, action1)
	kt.RegisterKeyBinding(seqCancel, action2)

	// Test the common prefix (Ctrl-x)
	a, more := kt.NextEvent(e1, nil)
	assert.Nil(t, a, "Action should be nil after first key")
	assert.True(t, more, "More keys expected in sequence")

	// Test the first sequence (Ctrl-x Ctrl-s)
	a, more = kt.NextEvent(e2s, nil)
	assert.NotNil(t, a, "Action should not be nil after sequence completion")
	assert.False(t, more, "No more keys expected in sequence")
	
	a(nil)
	assert.True(t, action1Called, "First action should be called")
	assert.False(t, action2Called, "Second action should not be called yet")

	// Reset and test the second sequence
	kt.ResetEvents()
	action1Called = false
	action2Called = false

	// First key (Ctrl-x)
	a, more = kt.NextEvent(e1, nil)

	// Second key (Ctrl-c) - should complete the second sequence
	a, more = kt.NextEvent(e2c, nil)
	a(nil)
	assert.False(t, action1Called, "First action should not be called")
	assert.True(t, action2Called, "Second action should be called")
}

func TestKeySequenceWithInvalidKey(t *testing.T) {
	kt := NewKeyTree()
	
	// Register a sequence: Ctrl-x Ctrl-s
	e1 := KeyEvent{code: tcell.KeyCtrlX, mod: tcell.ModCtrl}
	e2 := KeyEvent{code: tcell.KeyCtrlS, mod: tcell.ModCtrl}
	seq := KeySequenceEvent{keys: []Event{e1, e2}}
	
	kt.RegisterKeyBinding(seq, func(p Pane) bool { return true })

	// First key (Ctrl-x)
	a, more := kt.NextEvent(e1, nil)
	assert.Nil(t, a)
	assert.True(t, more)

	// Invalid key - should reset the sequence
	invalidKey := KeyEvent{code: tcell.KeyRune, r: 'a'}
	a, more = kt.NextEvent(invalidKey, nil)
	assert.Nil(t, a)
	assert.False(t, more)
}

func TestKeySequenceWithTimeout(t *testing.T) {
	kt := NewKeyTree()
	
	// Register a sequence: Ctrl-x Ctrl-s
	e1 := KeyEvent{code: tcell.KeyCtrlX, mod: tcell.ModCtrl}
	e2 := KeyEvent{code: tcell.KeyCtrlS, mod: tcell.ModCtrl}
	seq := KeySequenceEvent{keys: []Event{e1, e2}}
	
	actionCalled := false
	kt.RegisterKeyBinding(seq, func(p Pane) bool {
		actionCalled = true
		return true
	})

	// First key (Ctrl-x)
	a, more := kt.NextEvent(e1, nil)
	assert.Nil(t, a)
	assert.True(t, more)

	// Wait for timeout
	time.Sleep(SequenceTimeout + 10*time.Millisecond)

	// Next key should start a new sequence
	a, more = kt.NextEvent(e2, nil)
	assert.Nil(t, a)
	assert.False(t, more)
	assert.False(t, actionCalled, "Action should not be called after timeout")
}
