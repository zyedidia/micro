package action

import (
	"github.com/zyedidia/tcell"
)

type Event interface{}

// RawEvent is simply an escape code
// We allow users to directly bind escape codes
// to get around some of a limitations of terminals
type RawEvent struct {
	esc string
}

// KeyEvent is a key event containing a key code,
// some possible modifiers (alt, ctrl, etc...) and
// a rune if it was simply a character press
// Note: to be compatible with tcell events,
// for ctrl keys r=code
type KeyEvent struct {
	code tcell.Key
	mod  tcell.ModMask
	r    rune
}

// MouseEvent is a mouse event with a mouse button and
// any possible key modifiers
type MouseEvent struct {
	btn tcell.ButtonMask
	mod tcell.ModMask
}

type KeyAction func(Handler) bool
type MouseAction func(Handler, tcell.EventMouse) bool

// A Handler will take a tcell event and execute it
// appropriately
type Handler interface {
	HandleEvent(tcell.Event)
	HandleCommand(string)
}
