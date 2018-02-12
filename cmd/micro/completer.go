package main

import (
	"fmt"
	"strings"

	"github.com/zyedidia/micro/cmd/micro/optionprovider"
	"github.com/zyedidia/tcell"
)

// OptionProvider is the signature of a function which returns all of the available options at the given offset.
type OptionProvider func(buffer []byte, offset int) (options []optionprovider.Option, err error)

// ContentSetter is the signature of a function which allows the content of a cell to be set.
type ContentSetter func(x int, y int, mainc rune, combc []rune, style tcell.Style)

// CompleterEnabledFlagFromView gets whether autocomplete is enabled from the buffer settings.
func CompleterEnabledFlagFromView(v *View) func() bool {
	return func() bool {
		setting, hasSetting := v.Buf.Settings["autocomplete"]
		if !hasSetting {
			return false
		}
		enabled, isBool := setting.(bool)
		if !isBool {
			return false
		}
		return enabled
	}
}

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
		buf.Cursor.GotoLoc(Loc{X: from.X + len(with), Y: to.Y})
	}
}

// ContentSetterForView sets the content of a cell for the x, y coordinate of a document.
func ContentSetterForView(v *View) ContentSetter {
	return func(x int, y int, mainc rune, combc []rune, style tcell.Style) {
		targetY := y - v.Topline
		targetX := x + v.leftCol
		LogToMessenger()("completer.ContentSetterForView: doc pos %v:%v drawing '%v' at %v:%v", y, x, string(mainc), targetY, targetX)
		screen.SetContent(targetX, targetY, mainc, combc, style)
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
	// Active is the state which determines whether the completer is active (visible) or not.
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
	// Setter is a function which draws to the console at a given location.
	Setter ContentSetter
	// OptionStyleInactive is the style for completer options which are not currently highlighted.
	OptionStyleInactive tcell.Style
	// OptionStyleActive is the style for completer options which are currently highlighted.
	OptionStyleActive tcell.Style
	// Enabled determines whether the view has the enabled flag set or not.
	Enabled func() bool
}

// NewCompleterForView creates a new autocompleter with defaults for writing to the console.
func NewCompleterForView(v *View) *Completer {
	// Default activators and deactivators.
	activators := []rune{'.', '('}
	deactivators := []rune{')', ',', ' ', '\n'}

	var provider OptionProvider

	// Load the provider based on filename.
	fileName := v.Buf.GetName()
	if strings.HasSuffix(fileName, ".go") {
		provider = optionprovider.GoCode
	} else {
		provider = optionprovider.Generic
	}

	// If no matching provider was found, we can't autocomplete.
	if provider == nil {
		provider = optionprovider.Noop
	}

	return NewCompleter(activators, deactivators,
		provider,
		LogToMessenger(),
		CurrentBytesAndOffsetFromView(v),
		CurrentLocationFromView(v),
		ReplaceFromBuffer(v.Buf),
		ContentSetterForView(v),
		colorscheme["default"].Reverse(true),
		colorscheme["default"],
		CompleterEnabledFlagFromView(v),
	)
}

// NewCompleter creates a new completer with all options exposed. See NewCompleterForView for more common usage.
func NewCompleter(activators []rune,
	deactivators []rune,
	provider OptionProvider,
	logger func(s string, values ...interface{}),
	currentBytesAndOffset func() (bytes []byte, offset int),
	currentLocation func() Loc,
	replacer func(from, to Loc, with string),
	setter ContentSetter,
	optionStyleInactive tcell.Style,
	optionStyleActive tcell.Style,
	enabled func() bool) *Completer {
	return &Completer{
		Activators:            activators,
		Deactivators:          deactivators,
		Provider:              provider,
		Logger:                logger,
		CurrentBytesAndOffset: currentBytesAndOffset,
		CurrentLocation:       currentLocation,
		Replacer:              replacer,
		Setter:                setter,
		OptionStyleInactive:   optionStyleInactive,
		OptionStyleActive:     optionStyleActive,
		Enabled:               enabled,
	}
}

// Process handles incoming events from the view and starts looking up via autocomplete.
func (c *Completer) Process(r rune) error {
	if !c.Enabled() {
		return nil
	}

	if c.Provider == nil {
		return nil
	}

	// Hide the autocomplete view if needed.
	if c.Active && containsRune(c.Deactivators, r) {
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
	//TODO: We only need the answer by the time Display is called, so we could let the rest of the
	// program continue until we're ready to receive the value by using a go routine or channel.
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
	if !c.Enabled() {
		c.Logger("completer.HandleEvent: not enabled")
		return false
	}
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
	if i > len(options)-1 {
		return "", false
	}
	if i < 0 {
		i = 0
	}
	return options[i].Text(), true
}

// Display the suggestion box.
func (c *Completer) Display() {
	if !c.Enabled() {
		c.Logger("completer.Display: not enabled")
		return
	}
	if !c.Active {
		c.Logger("completer.Display: not showing because inactive")
		return
	}

	c.Logger("completer.Display: showing %d options", len(c.Options))
	width := getWidth(c.Options)
	for iy, o := range c.Options {
		y := c.Y + iy + 1 // +1 to draw underneath the start position.

		// If it's active, show it differently.
		style := c.OptionStyleInactive
		if c.ActiveIndex == iy {
			style = c.OptionStyleActive
		}

		// Draw the runes.
		for ix, r := range padRight(o.Text(), width+1) {
			x := c.X + ix
			c.Setter(x, y, r, nil, style)
		}
	}
}

func getWidth(options []optionprovider.Option) (max int) {
	for _, o := range options {
		if l := len(o.Text()); l > max {
			max = l
		}
	}
	return
}

func padRight(s string, minSize int) string {
	extra := minSize - len(s)
	if extra > 0 {
		padding := make([]byte, extra)
		return s + string(padding)
	}
	return s
}

func containsRune(array []rune, r rune) bool {
	for _, r1 := range array {
		if r1 == r {
			return true
		}
	}
	return false
}
