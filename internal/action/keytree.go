package action

import (
	"bytes"
	"log"
	"strings"
	"time"

	"github.com/micro-editor/tcell/v2"
)

type PaneKeyAction func(Pane) bool
type PaneMouseAction func(Pane, *tcell.EventMouse) bool
type PaneKeyAnyAction func(Pane, []KeyEvent) bool

// A KeyTreeNode stores a single node in the KeyTree (trie). The
// children are stored as a map, and any node may store a list of
// actions (the list will be nil if no actions correspond to a certain
// node)
type KeyTreeNode struct {
	children map[Event]*KeyTreeNode

	// Only one of these actions may be active in the current
	// mode, and only one will be returned. If multiple actions
	// are active, it is undefined which one will be the one
	// returned.
	actions []TreeAction
}

func NewKeyTreeNode() *KeyTreeNode {
	n := new(KeyTreeNode)
	n.children = make(map[Event]*KeyTreeNode)
	n.actions = []TreeAction{}
	return n
}

// A TreeAction stores an action, and a set of mode constraints for
// the action to be active.
type TreeAction struct {
	// only one of these can be non-nil
	action PaneKeyAction
	any    PaneKeyAnyAction
	mouse  PaneMouseAction

	modes []ModeConstraint
}

// A KeyTree is a data structure for storing keybindings. It maps
// key events to actions, and maintains a set of currently enabled
// modes, which affects the action that is returned for a key event.
// The tree acts like a Trie for Events to handle sequence events.
type KeyTree struct {
	root  *KeyTreeNode
	modes map[string]bool

	cursor KeyTreeCursor
}

// A KeyTreeCursor keeps track of the current location within the
// tree, and stores any information from previous events that may
// be needed to execute the action (values of wildcard events or
// mouse events)
const (
	// SequenceTimeout is the maximum time between key presses in a sequence
	SequenceTimeout = 1000 * time.Millisecond
)

type KeyTreeCursor struct {
	node *KeyTreeNode

	recordedEvents []Event
	wildcards      []KeyEvent
	mouseInfo      *tcell.EventMouse
	lastKeyTime    time.Time
	isInSequence   bool
}

// MakeClosure uses the information stored in a key tree cursor to construct
// a PaneKeyAction from a TreeAction (which may have a PaneKeyAction, PaneMouseAction,
// or AnyAction)
func (k *KeyTreeCursor) MakeClosure(a TreeAction) PaneKeyAction {
	if a.action != nil {
		return a.action
	} else if a.any != nil {
		return func(p Pane) bool {
			return a.any(p, k.wildcards)
		}
	} else if a.mouse != nil {
		return func(p Pane) bool {
			return a.mouse(p, k.mouseInfo)
		}
	}

	return nil
}

// NewKeyTree allocates and returns an empty key tree
func NewKeyTree() *KeyTree {
	root := NewKeyTreeNode()
	tree := new(KeyTree)

	tree.root = root
	tree.modes = make(map[string]bool)
	tree.cursor = KeyTreeCursor{
		node:          root,
		wildcards:     []KeyEvent{},
		mouseInfo:     nil,
		lastKeyTime:   time.Time{},
		isInSequence:  false,
	}

	return tree
}

// A ModeConstraint specifies that an action can only be executed
// while a certain mode is enabled or disabled.
type ModeConstraint struct {
	mode     string
	disabled bool
}

// RegisterKeyBinding registers a PaneKeyAction with an Event.
func (k *KeyTree) RegisterKeyBinding(e Event, a PaneKeyAction) {
	k.registerBinding(e, TreeAction{
		action: a,
		any:    nil,
		mouse:  nil,
		modes:  nil,
	})
}

// RegisterKeyAnyBinding registers a PaneKeyAnyAction with an Event.
// The event should contain an "any" event.
func (k *KeyTree) RegisterKeyAnyBinding(e Event, a PaneKeyAnyAction) {
	k.registerBinding(e, TreeAction{
		action: nil,
		any:    a,
		mouse:  nil,
		modes:  nil,
	})
}

// RegisterMouseBinding registers a PaneMouseAction with an Event.
// The event should contain a mouse event.
func (k *KeyTree) RegisterMouseBinding(e Event, a PaneMouseAction) {
	k.registerBinding(e, TreeAction{
		action: nil,
		any:    nil,
		mouse:  a,
		modes:  nil,
	})
}

func (k *KeyTree) registerBinding(e Event, a TreeAction) {
	switch ev := e.(type) {
	case KeyEvent, MouseEvent, RawEvent:
		newNode, ok := k.root.children[e]
		if !ok {
			newNode = NewKeyTreeNode()
			k.root.children[e] = newNode
		}
		// newNode.actions = append(newNode.actions, a)
		newNode.actions = []TreeAction{a}
	case KeySequenceEvent:
		n := k.root
		for _, key := range ev.keys {
			newNode, ok := n.children[key]
			if !ok {
				newNode = NewKeyTreeNode()
				n.children[key] = newNode
			}

			n = newNode
		}
		// n.actions = append(n.actions, a)
		n.actions = []TreeAction{a}
	}
}

// NextEvent returns the action for the current sequence where e is the next
// event. Even if the action was registered as a PaneKeyAnyAction or PaneMouseAction,
// it will be returned as a PaneKeyAction closure where the appropriate arguments
// have been provided.
// If no action is associated with the given Event, or mode constraints are not
// met for that action, nil is returned.
// A boolean is returned to indicate if there is a conflict with this action. A
// conflict occurs when there is an active action for this event but there are
// bindings associated with further sequences starting with this event. The
// calling function can decide what to do about the conflict (e.g. use a
// timeout).
func (k *KeyTree) NextEvent(e Event, mouse *tcell.EventMouse) (PaneKeyAction, bool) {
	now := time.Now()
	
	// Log the incoming event with more context
	log.Printf("NextEvent: %v (type: %T), isInSequence: %v, time since last: %v, current node: %+v, recorded events: %d", 
		e.Name(), e, k.cursor.isInSequence, now.Sub(k.cursor.lastKeyTime), k.cursor.node != nil, len(k.cursor.recordedEvents))
	
	// Log the current sequence of events
	if len(k.cursor.recordedEvents) > 0 {
		events := make([]string, 0, len(k.cursor.recordedEvents))
		for _, evt := range k.cursor.recordedEvents {
			events = append(events, evt.Name())
		}
		log.Printf("Current sequence: %v", strings.Join(events, " "))
	}
	
	// Log the children of the current node for debugging
	if k.cursor.node != nil {
		node := k.cursor.node
		actionInfo := "<none>"
		if len(node.actions) > 0 {
			actionInfo = "<has actions>"
		}
		log.Printf("Current node has %d children, actions: %v", len(node.children), actionInfo)
		for child := range node.children {
			log.Printf("  Child: %v (type: %T)", child.Name(), child)
		}
	} else {
		log.Printf("Current node is nil, resetting sequence state")
	}

	// Reset sequence if timeout occurred or if a mouse event is received during a sequence
	if k.cursor.isInSequence && (now.Sub(k.cursor.lastKeyTime) > SequenceTimeout || e.Name() == "MouseEvent") {
		reason := "mouse event"
		if now.Sub(k.cursor.lastKeyTime) > SequenceTimeout {
			reason = "timeout"
		}
		log.Printf("Resetting sequence due to %s", reason)
		k.ResetEvents()
		// If this was a mouse event, let it be handled normally
		if e.Name() == "MouseEvent" {
			return nil, false
		}
	}

	n := k.cursor.node
	c, hasChild := n.children[e]

	// Check if this node has any children that could form a valid sequence
	hasSequence := false
	for range n.children {
		// If we find any child that could be part of a sequence, set hasSequence to true
		hasSequence = true
		break
	}

	log.Printf("Current node has %d children, has actions: %v, has sequence: %v", 
		len(n.children), len(n.actions) > 0, hasSequence)

	// If this node has no children, execute the action if it exists
	if !hasChild {
		if len(n.actions) > 0 {
			action := k.cursor.MakeClosure(n.actions[0])
			k.ResetEvents()
			log.Printf("Executing action for %v (no children)", e.Name())
			return action, false
		}
		log.Printf("No matching child for %v and no action to execute", e.Name())
		k.ResetEvents()
		return nil, false
	}

	// If we have a child node that could form a sequence
	if hasChild {
		hasAction := len(c.actions) > 0  // Check actions on the child node, not current node
		hasChildren := len(c.children) > 0
		
		log.Printf("Found matching child for %v, has action: %v, has children: %v", 
			e.Name(), hasAction, hasChildren)

		// If we're in a sequence, always try to continue it if possible
		if k.cursor.isInSequence {
			// If this node has an action and no children, execute it
			if hasAction && !hasChildren {
				action := k.cursor.MakeClosure(c.actions[0])
				k.ResetEvents()
				log.Printf("Executing sequence action for %v", e.Name())
				return action, false
			}
			
			// Otherwise, continue the sequence
			k.cursor.node = c
			k.cursor.recordedEvents = append(k.cursor.recordedEvents, e)
			k.cursor.lastKeyTime = now
			k.cursor.isInSequence = true
			log.Printf("Continuing sequence with %v (in sequence)", e.Name())
			return nil, true
		}

		// If we're not in a sequence yet, but this node has children, start a new sequence
		if hasChildren {
			k.cursor.node = c
			k.cursor.recordedEvents = append(k.cursor.recordedEvents, e)
			k.cursor.lastKeyTime = now
			k.cursor.isInSequence = true
			log.Printf("Starting new sequence with %v", e.Name())
			return nil, true
		}
		
		// If we have an action and no children, execute it immediately
		if hasAction && !hasChildren {
			action := k.cursor.MakeClosure(c.actions[0])
			k.ResetEvents()
			log.Printf("Executing action for %v (no children)", e.Name())
			return action, false
		}
	}

	more := len(c.children) > 0

	// Update cursor state
	k.cursor.node = c
	k.cursor.lastKeyTime = now
	k.cursor.isInSequence = true

	k.cursor.recordedEvents = append(k.cursor.recordedEvents, e)

	switch ev := e.(type) {
	case KeyEvent:
		if ev.any {
			k.cursor.wildcards = append(k.cursor.wildcards, ev)
		}
	case MouseEvent:
		k.cursor.mouseInfo = mouse
	}

	// If we have an action at this node, return it
	if len(c.actions) > 0 {
		// check if actions are active
		for _, a := range c.actions {
			active := true
			for _, mc := range a.modes {
				// if any mode constraint is not met, the action is not active
				hasMode := k.modes[mc.mode]
				if hasMode != mc.disabled {
					active = false
				}
			}

			if active {
				// Reset sequence when we find a matching action
				defer k.ResetEvents()
				return k.cursor.MakeClosure(a), more
			}
		}
	}

	// If no more children, reset sequence after this key
	if !more {
		defer k.ResetEvents()
	}

	return nil, more
}

// ResetEvents sets the current sequence back to the initial value.
func (k *KeyTree) ResetEvents() {
	k.cursor = KeyTreeCursor{
		node:          k.root,
		wildcards:     []KeyEvent{},
		recordedEvents: []Event{},
		mouseInfo:     nil,
	}
}

// RecordedEventsStr returns the list of recorded events as a string
func (k *KeyTree) RecordedEventsStr() string {
	buf := &bytes.Buffer{}
	for i, e := range k.cursor.recordedEvents {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(e.Name())
	}
	return buf.String()
}

// IsInSequence returns true if we're in the middle of a key sequence
func (k *KeyTree) IsInSequence() bool {
	return k.cursor.isInSequence
}

// GetCurrentSequence returns the current key sequence as a slice of events
func (k *KeyTree) GetCurrentSequence() []Event {
	return k.cursor.recordedEvents
}

// GetPossibleCompletions returns a list of possible completions for the current sequence
func (k *KeyTree) GetPossibleCompletions() []string {
	var completions []string
	
	n := k.cursor.node
	for e := range n.children {
		completions = append(completions, e.Name())
	}
	
	return completions
}

// DeleteBinding removes any currently active actions associated with the
// given event.
func (k *KeyTree) DeleteBinding(e Event) {

}

// DeleteAllBindings removes all actions associated with the given event,
// regardless of whether they are active or not.
func (k *KeyTree) DeleteAllBindings(e Event) {

}

// SetMode enables or disabled a given mode
func (k *KeyTree) SetMode(mode string, en bool) {
	k.modes[mode] = en
}

// HasMode returns if the given mode is currently active
func (k *KeyTree) HasMode(mode string) bool {
	return k.modes[mode]
}
