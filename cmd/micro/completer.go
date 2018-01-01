package main

import "github.com/zyedidia/micro/cmd/micro/optionprovider"
import "fmt"

// OptionProvider returns all of the available options at the given offset.
type OptionProvider func(buffer []byte, offset int) (options []optionprovider.Option, err error)

// Completer completes code as you type.
type Completer struct {
	// Active is the state which determines whether the completer is active or not.
	Active bool
	// Activators are insertions that start autocomplete, e.g. a "." or an opening bracket "(".
	Activators []rune
	// Deactivators are insertions that stop autocomplete, e.g. a closing bracket, or a semicolon.
	Deactivators []rune
	// Provider is the provider of completion options, e.g. gocode, or another provider such as a language server.
	Provider OptionProvider
}

// Process handles incoming events from the view and determines whether to render the completer options.
func (c *Completer) Process(v *View, r rune) error {
	// Hide the autocomplete view if needed.
	if containsRune(c.Deactivators, r) {
		c.Active = false
		return c.Hide(v)
	}

	// Check to work out whether we should activate the autocomplete.
	if containsRune(c.Activators, r) {
		c.Active = true
		TermMessage("active")
	}

	if !c.Active {
		// We're not active.
		return nil
	}

	// Get options.
	options, err := c.Provider(v.Buf.Buffer(false).Bytes(), ByteOffset(v.Cursor.Loc, CurView().Buf))
	if err != nil {
		return err
	}

	if len(options) > 0 {
		TermMessage(fmt.Sprintf("got %v options", len(options)))
		err = c.Show(v, options)
	}
	return err
}

// Show the suggestion box.
func (c Completer) Show(v *View, options []optionprovider.Option) error {
	//TODO: Show the box
	return nil
}

// Hide the suggestion box.
func (c Completer) Hide(v *View) error {
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
