package main

import (
	"fmt"

	"github.com/zyedidia/micro/cmd/micro/optionprovider"
	"github.com/zyedidia/tcell"
)

// OptionProvider returns all of the available options at the given offset.
type OptionProvider func(buffer []byte, offset int) (options []optionprovider.Option, err error)

// CurrentBytesAndOffsetFromView gets bytes from a view.
func CurrentBytesAndOffsetFromView(v *View) func() (bytes []byte, offset int) {
	return func() (bytes []byte, offset int) {
		bytes = v.Buf.Buffer(false).Bytes()
		offset = ByteOffset(v.Cursor.Loc, v.Buf)
		return
	}
}

// CurrentLocationFromView gets the current location from a view.
func CurrentLocationFromView(v *View) func() Loc {
	return func() Loc {
		return v.Cursor.Loc
	}
}

// ReplaceFromBuffer replaces text in a buffer.
func ReplaceFromBuffer(buf *Buffer) func(from, to Loc, with string) {
	return func(from, to Loc, with string) {
		buf.Replace(from, to, with)
	}
}

// LogToMessenger logs to the global messenger.
func LogToMessenger() func(s string, values ...interface{}) {
	return func(s string, values ...interface{}) {
		messenger.AddLog(fmt.Sprintf(s, values...))
	}
}

// Completer completes code as you type.
type Completer struct {
	// Active is the state which determines whether the completer is active or not.
	Active bool
	// X stores the X position of the suggestion box
	X int
	// Y stores the Y position of the suggestion box
	Y int
	// Options stores the current list of suggestions.
	Options []optionprovider.Option
	// ActiveIndex store the index of the active option (the one that will be selected).
	ActiveIndex int
	// Activators are insertions that start autocomplete, e.g. a "." or an opening bracket "(".
	Activators []rune
	// Deactivators are insertions that stop autocomplete, e.g. a closing bracket, or a semicolon.
	Deactivators []rune
	// Provider is the provider of completion options, e.g. gocode, or another provider such as a language server.
	Provider OptionProvider
	// Logger is where log messages are written via fmt.Sprintf.
	Logger func(s string, values ...interface{})
	// CurrentBytesAndOffset is a function which returns the bytes and the current offset position from the current view.
	CurrentBytesAndOffset func() (bytes []byte, offset int)
	// CurrentLocation is a function which returns the current location of the cursor.
	CurrentLocation func() Loc
	// Replacer is a function which replaces text.
	Replacer func(from, to Loc, with string)
}

// NewCompleter creates a new autocompleter.
func NewCompleter(activators []rune,
	deactivators []rune,
	provider OptionProvider,
	logger func(s string, values ...interface{}),
	currentBytesAndOffset func() (bytes []byte, offset int),
	currentLocation func() Loc,
	replacer func(from, to Loc, with string)) *Completer {
	return &Completer{
		Activators:            activators,
		Deactivators:          deactivators,
		Provider:              provider,
		Logger:                logger,
		CurrentBytesAndOffset: currentBytesAndOffset,
		CurrentLocation:       currentLocation,
		Replacer:              replacer,
	}
}

// Process handles incoming events from the view and starts looking up via autocomplete.
func (c *Completer) Process(r rune) error {
	// Hide the autocomplete view if needed.
	if containsRune(c.Deactivators, r) {
		c.Logger("completer.Process: deactivating, because received %v", string(r))
		c.Active = false
		return nil
	}

	// Check to work out whether we should activate the autocomplete.
	if containsRune(c.Activators, r) {
		c.Logger("completer.Process: activating, because received %v", string(r))
		c.Active = true
		currentLocation := c.CurrentLocation()
		c.X, c.Y = currentLocation.X, currentLocation.Y
		c.Logger("completer.Process: SetStartPosition to %d, %d", c.X, c.Y)
	}

	if !c.Active {
		// We're not active.
		return nil
	}

	// Get options.
	//TODO: We only need the answer by the time Display is called, so we can let the rest of the
	// program continue until we're ready to receive the value.
	bytes, offset := c.CurrentBytesAndOffset()
	options, err := c.Provider(bytes, offset)
	if err != nil {
		return err
	}
	c.Options = options
	c.ActiveIndex = -1
	return err
}

// HandleEvent handles incoming key presses if the completer is active.
// It returns true if it took over the key action, or false if it didn't.
func (c *Completer) HandleEvent(key tcell.Key) bool {
	if !c.Active {
		c.Logger("completer.HandleEvent: not active")
		return false
	}

	// Handle selecting various options in the list.
	switch key {
	case tcell.KeyUp:
		if c.ActiveIndex > 0 {
			c.ActiveIndex--
		}
		break
	case tcell.KeyDown:
		if c.ActiveIndex < len(c.Options)-1 {
			c.ActiveIndex++
		}
		break
	case tcell.KeyEsc:
		c.Active = false
		break
	case tcell.KeyTab, tcell.KeyEnter:
		// Complete the text.
		if toUse, ok := getOption(c.ActiveIndex, c.Options); ok {
			c.Replacer(Loc{X: c.X, Y: c.Y}, c.CurrentLocation(), toUse)
		}
		c.Active = false
		break
	default:
		// Not part of the keys that the autocomplete menu handles.
		return false
	}

	// The completer handled the key.
	return true
}

func getOption(i int, options []optionprovider.Option) (toUse string, ok bool) {
	if i < 0 || i > len(options)-1 {
		return "", false
	}
	return options[i].Text(), true
}

// ContentSetter is the signature of a function which allows the content of a cell to be set.
type ContentSetter func(x int, y int, mainc rune, combc []rune, style tcell.Style)

// Display the suggestion box.
func (c *Completer) Display(setter ContentSetter) {
	if !c.Active {
		c.Logger("completer.Display: not showing because inactive")
		return
	}

	c.Logger("completer.Display: showing %d options", len(c.Options))
	for iy, o := range c.Options {
		y := c.Y + iy + 1 // +1 to draw underneath the start position.
		// Draw the runes.
		// TODO: Only draw up to n characters?
		// TODO: Limit the number of options displayed?
		// TODO: Draw a box around the options, or restyle.
		for ix, r := range o.Text() {
			x := c.X + ix

			// If it's active, show it differently
			style := defStyle.Reverse(true)
			if c.ActiveIndex == iy {
				style = defStyle
			}
			setter(x, y, r, nil, style)
		}
	}
}

func containsRune(array []rune, r rune) bool {
	for _, r1 := range array {
		if r1 == r {
			return true
		}
	}
	return false
}
