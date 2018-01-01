package main

import "github.com/zyedidia/micro/cmd/micro/optionprovider"
import "github.com/zyedidia/tcell"

// OptionProvider returns all of the available options at the given offset.
type OptionProvider func(buffer []byte, offset int) (options []optionprovider.Option, err error)

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
	// Activators are insertions that start autocomplete, e.g. a "." or an opening bracket "(".
	Activators []rune
	// Deactivators are insertions that stop autocomplete, e.g. a closing bracket, or a semicolon.
	Deactivators []rune
	// Provider is the provider of completion options, e.g. gocode, or another provider such as a language server.
	Provider OptionProvider
	// Logger is where log messages are written via fmt.Sprintf.
	Logger func(s string, values ...interface{})
}

// Process handles incoming events from the view and starts looking up via autocomplete.
func (c *Completer) Process(v *View, r rune) error {
	// Hide the autocomplete view if needed.
	if containsRune(c.Deactivators, r) {
		c.Logger("completer.Process: deactivating, because received %v", string(r))
		c.Active = false
		return c.Hide(v)
	}

	// Check to work out whether we should activate the autocomplete.
	if containsRune(c.Activators, r) {
		c.Logger("completer.Process: activating, because received %v", string(r))
		c.Active = true
		c.SetPosition(v.Cursor.Loc)
	}

	if !c.Active {
		// We're not active.
		return nil
	}

	// Get options.
	//TODO: Make this run as a goroutine. We only need the answer by the time Display is called, so we can let the rest of the program continue until we're ready.
	options, err := c.Provider(v.Buf.Buffer(false).Bytes(), ByteOffset(v.Cursor.Loc, CurView().Buf))
	if err != nil {
		return err
	}
	c.Options = options
	return err
}

func (c *Completer) SetPosition(l Loc) {
	// Draw the suggestions directly under the current cursor position.
	c.X, c.Y = l.X, l.Y+1
	c.Logger("completer.SetPosition: %d, %d", c.X, c.Y)
}

type ContentSetter func(x int, y int, mainc rune, combc []rune, style tcell.Style)

// Display the suggestion box.
func (c *Completer) Display(setter ContentSetter) {
	c.Logger("completer.Show: showing %d options", len(c.Options))
	for iy, o := range c.Options {
		y := c.Y + iy
		// Draw the runes.
		// TODO: Only draw up to n characters?
		// TODO: Limit the number of options displayed?
		// TODO: Draw a box around the options, or restyle.
		for ix, r := range o.Text() {
			x := c.X + ix
			setter(x, y, r, nil, defStyle.Reverse(true))
		}
	}
}

// Hide the suggestion box.
func (c *Completer) Hide(v *View) error {
	//TODO: Hide the box
	return nil
}

func containsRune(array []rune, r rune) bool {
	for _, r1 := range array {
		if r1 == r {
			return true
		}
	}
	return false
}
