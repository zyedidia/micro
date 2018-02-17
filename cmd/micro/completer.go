package main

import (
	"fmt"
	"strings"

	"github.com/zyedidia/micro/cmd/micro/optionprovider"
	"github.com/zyedidia/tcell"
)

// OptionProvider is the signature of a function which returns all of the available options, potentially using the prefix
// data. For example, given input "abc\nab", start offset 4 and end offset 5, then the prefix is "ab", and the result
// should be the option "abc".
// Logger provides logging. Can be satisfied with t.Logf for tests, or LogToMessenger.
type OptionProvider func(logger func(s string, values ...interface{}), buffer []byte, startOffset, endOffset int) (options []optionprovider.Option, startOffsetDelta int, err error)

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

// LocationOffsetFromView provides the offset of a given location.
func LocationOffsetFromView(v *View) func(Loc) (offset int) {
	return func(l Loc) (offset int) {
		return ByteOffset(l, v.Buf)
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
		LogToMessenger()("replacing from %v to %v with %s", from, to, with)
		buf.Replace(from, to, with)
		buf.Cursor.GotoLoc(Loc{X: from.X + len(with), Y: from.Y})
	}
}

// ContentSetterForView sets the content of a cell for the x, y coordinate of a document.
func ContentSetterForView(v *View) ContentSetter {
	return func(x int, y int, mainc rune, combc []rune, style tcell.Style) {
		targetY := y - v.Topline
		targetX := x + v.leftCol + v.lineNumOffset
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
	Activators map[rune]int
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
	// LocationOffset is a function which returns the offset of a given location.
	LocationOffset func(Loc) int
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
	// PreviousLocation stores the last known location of the cursor.
	PreviousLocation Loc
}

// defaultActivators sets whether the character should start autocompletion. The value of zero means that
// the character itself is not included in the replacement, -1 means that it is.
var defaultActivators = map[rune]int{
	'.': 0,
	'(': 0,
	'a': -1, 'b': -1, 'c': -1, 'd': -1, 'e': -1, 'f': -1, 'g': -1, 'h': -1, 'i': -1, 'j': -1, 'k': -1, 'l': -1, 'm': -1, 'n': -1, 'o': -1, 'p': -1, 'q': -1, 'r': -1, 's': -1, 't': -1, 'u': -1, 'v': -1, 'w': -1, 'x': -1, 'y': -1, 'z': -1,
	'A': -1, 'B': -1, 'C': -1, 'D': -1, 'E': -1, 'F': -1, 'G': -1, 'H': -1, 'I': -1, 'J': -1, 'K': -1, 'L': -1, 'M': -1, 'N': -1, 'O': -1, 'P': -1, 'Q': -1, 'R': -1, 'S': -1, 'T': -1, 'U': -1, 'V': -1, 'W': -1, 'X': -1, 'Y': -1, 'Z': -1,
}

const defaultDeactivators = "), \n."

// NewCompleterForView creates a new autocompleter with defaults for writing to the console.
func NewCompleterForView(v *View) *Completer {
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

	return NewCompleter(defaultActivators, []rune(defaultDeactivators),
		provider,
		LogToMessenger(),
		CurrentBytesAndOffsetFromView(v),
		CurrentLocationFromView(v),
		LocationOffsetFromView(v),
		ReplaceFromBuffer(v.Buf),
		ContentSetterForView(v),
		colorscheme["default"].Reverse(true),
		colorscheme["default"],
		CompleterEnabledFlagFromView(v),
	)
}

// NewCompleter creates a new completer with all options exposed. See NewCompleterForView for more common usage.
func NewCompleter(activators map[rune]int,
	deactivators []rune,
	provider OptionProvider,
	logger func(s string, values ...interface{}),
	currentBytesAndOffset func() (bytes []byte, offset int),
	currentLocation func() Loc,
	locationOffset func(Loc) int,
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
		LocationOffset:        locationOffset,
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
	}

	if !c.Active {
		// Check to work out whether we should activate the autocomplete.
		if indexAdjustment, ok := c.Activators[r]; ok {
			c.Logger("completer.Process: activating, because received %v", string(r))
			c.Active = true
			currentLocation := c.CurrentLocation()
			c.PreviousLocation = currentLocation
			c.X, c.Y = currentLocation.X+indexAdjustment, currentLocation.Y
			c.Logger("completer.Process: SetStartPosition to %d, %d", c.X, c.Y)
		}
	}

	if !c.Active {
		// We're not active.
		return nil
	}

	// Get options.
	//TODO: We only need the answer by the time Display is called, so we could let the rest of the
	// program continue until we're ready to receive the value by using a go routine or channel.
	bytes, currentOffset := c.CurrentBytesAndOffset()
	startOffset := c.LocationOffset(Loc{X: c.X, Y: c.Y})
	options, delta, err := c.Provider(c.Logger, bytes, startOffset, currentOffset)
	if err != nil {
		return err
	}
	c.X += delta
	c.Options = options
	c.ActiveIndex = -1
	// If there are no options, just deactivate.
	if len(options) == 0 {
		c.Logger("completer.Process: Deactivating because there are no options")
		c.Active = false
	}
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
	if len(options) == 0 {
		return "", false
	}
	if i > len(options)-1 {
		return "", false
	}
	if i < 0 {
		i = 0
	}
	return options[i].Text(), true
}

// DeactivateIfOutOfBounds for example, if duplicating lines or backspacing past the start of the completion.
func (c *Completer) DeactivateIfOutOfBounds() {
	// Disable autocomplete if we've switched lines (e.g. by duplicating a line, or moving the cursor away)
	// of if the X position is equal to or less than current.
	if !c.Active {
		return
	}
	cur := c.CurrentLocation()
	beforeStart := cur.X < c.X
	movedMoreThanOneXSinceLastCheck := distance(c.PreviousLocation.X, cur.X) > 1
	c.Logger("completed.DeactivateIfOutOfBounds: Previous loc %v, current loc %v, distance: %v", cur, c.PreviousLocation, distance(c.PreviousLocation.X, cur.X))
	movedLine := cur.Y != c.Y
	if beforeStart || movedMoreThanOneXSinceLastCheck || movedLine {
		c.Logger("completer.DeactivateIfOutOfBounds: deactivating")
		c.Active = false
	}
	c.PreviousLocation = cur
}

func distance(a, b int) int {
	if a == b {
		return 0
	}
	if a > b {
		return a - b
	}
	return b - a
}

// Display the suggestion box.
func (c *Completer) Display() {
	if !c.Enabled() {
		c.Logger("completer.Display: not enabled")
		return
	}
	if !c.Active {
		return
	}

	c.Logger("completer.Display: showing %d options", len(c.Options))
	width := getWidth(c.Options)
	start := c.CurrentLocation()
	for iy, o := range c.Options {
		y := start.Y + iy + 1 // +1 to draw a line below the cursor.

		// If it's active, show it differently.
		style := c.OptionStyleInactive
		if c.ActiveIndex == iy {
			style = c.OptionStyleActive
		}

		// Draw the runes.
		for ix, r := range padRight(o.Text(), width+1) {
			x := start.X + ix
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
