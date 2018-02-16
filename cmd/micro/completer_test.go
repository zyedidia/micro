package main

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/zyedidia/micro/cmd/micro/optionprovider"
	"github.com/zyedidia/tcell"
)

var noopReplacer = func(from, to Loc, with string) {}
var noopContentSetter = func(x int, y int, mainc rune, combc []rune, style tcell.Style) {}
var enabledFlagSetToTrue = func() bool { return true }
var enabledFlagSetToFalse = func() bool { return false }

var optionStyleInactive = tcell.StyleDefault.Reverse(true)

const optionStyleActive = tcell.StyleDefault

func TestCompleterDoesNothingWhenNotEnabledOrProviderNotSet(t *testing.T) {
	var currentBytesAndOffsetCalled, currentLocationCalled, locationOffsetCalled, providerCalled bool

	currentBytesAndOffset := func() (bytes []byte, offset int) {
		currentBytesAndOffsetCalled = true
		return []byte("fmt.Println("), 3
	}
	currentLocation := func() Loc {
		currentLocationCalled = true
		return Loc{X: 1, Y: 2}
	}
	locationOffset := func(Loc) int {
		locationOffsetCalled = true
		return 0
	}
	provider := func(buffer []byte, startOffset, currentOffset int) (options []optionprovider.Option, delta int, err error) {
		providerCalled = true
		return
	}

	activators := map[rune]int{
		'(': 0,
	}
	deactivators := []rune{')', ';'}

	c := NewCompleter(activators,
		deactivators,
		provider,
		t.Logf,
		currentBytesAndOffset,
		currentLocation,
		locationOffset,
		noopReplacer,
		noopContentSetter,
		optionStyleInactive,
		optionStyleActive,
		enabledFlagSetToFalse)

	// It's not enabled, so nothing should be called.
	err := c.Process('(')
	if err != nil {
		t.Fatalf("not enabled: failed to process with error: %v", err)
	}
	if currentBytesAndOffsetCalled || currentLocationCalled || locationOffsetCalled {
		t.Errorf("not enabled: when disabled, no functions should be called")
	}

	// It's enabled, but the provider is nil.
	c.Enabled = enabledFlagSetToTrue
	c.Provider = nil

	err = c.Process('(')
	if err != nil {
		t.Fatalf("enabled: failed to process with error: %v", err)
	}
	if currentBytesAndOffsetCalled || currentLocationCalled || locationOffsetCalled || providerCalled {
		t.Errorf("enabled: when disabled, no functions should be called")
	}
}

func TestCompleterIsDeactivatedByDeactivatorRunes(t *testing.T) {
	activators := map[rune]int{
		'(': 0,
	}
	deactivators := []rune{')', ';'}
	currentBytesAndOffset := func() (bytes []byte, offset int) {
		return []byte("fmt.Println("), 3
	}
	currentLocation := func() Loc {
		return Loc{X: 1, Y: 2}
	}
	locationOffset := func(Loc) int {
		return 3
	}
	provider := func(buffer []byte, startOffset, currentOffset int) (options []optionprovider.Option, delta int, err error) {
		options = []optionprovider.Option{
			optionprovider.New("text", "hint"),
		}
		return
	}

	c := NewCompleter(activators,
		deactivators,
		provider,
		t.Logf,
		currentBytesAndOffset,
		currentLocation,
		locationOffset,
		noopReplacer,
		noopContentSetter,
		optionStyleInactive,
		optionStyleActive,
		enabledFlagSetToTrue)

	c.Active = true

	err := c.Process(')')
	if err != nil {
		t.Fatalf("failed to process with error: %v", err)
	}
	if c.Active {
		t.Errorf("expected ')' to deactivate the completer, but it didn't")
	}
}

func TestCompleterIsActivatedByActivatorRunes(t *testing.T) {
	var providerReceivedBytes []byte
	var providerReceivedOffset int

	expectedOptions := []optionprovider.Option{
		optionprovider.New("text", "hint"),
		optionprovider.New("text1", "hint1"),
	}

	activators := map[rune]int{
		'(': 0,
	}
	deactivators := []rune{')', ';'}
	currentBytesAndOffset := func() (bytes []byte, offset int) {
		return []byte("fmt.Println("), 3
	}
	currentLocation := func() Loc {
		return Loc{X: 1, Y: 2}
	}
	locationOffset := func(Loc) int {
		return 3
	}
	provider := func(buffer []byte, startOffset, currentOffset int) (options []optionprovider.Option, delta int, err error) {
		providerReceivedBytes = buffer
		providerReceivedOffset = currentOffset
		options = expectedOptions
		return
	}
	c := NewCompleter(activators,
		deactivators,
		provider,
		t.Logf,
		currentBytesAndOffset,
		currentLocation,
		locationOffset,
		noopReplacer,
		noopContentSetter,
		optionStyleInactive,
		optionStyleActive,
		enabledFlagSetToTrue)

	err := c.Process('(')
	if err != nil {
		t.Fatalf("failed to process with error: %v", err)
	}
	if !c.Active {
		t.Errorf("expected '(' to activate the completer, but it didn't")
	}
	if c.X != 1 && c.Y != 2 {
		t.Errorf("expected activating the completer to set the start position to {1, 2} but got {%v, %v}", c.X, c.Y)
	}
	if !reflect.DeepEqual([]byte("fmt.Println("), providerReceivedBytes) {
		t.Errorf("expected the provider to receive '%v', but got '%v'", "fmt.Println(", string(providerReceivedBytes))
	}
	if !reflect.DeepEqual(3, providerReceivedOffset) {
		t.Errorf("expected the provider to receive '%v', but got '%v'", "fmt.Println(", string(providerReceivedBytes))
	}
	if !reflect.DeepEqual(expectedOptions, c.Options) {
		t.Errorf("expected options %v, but got %v", expectedOptions, c.Options)
	}
	if c.ActiveIndex != -1 {
		t.Errorf("expected the active index to be reset to -1 after a refresh, but it was set to %v", c.ActiveIndex)
	}
}

func TestCompleterIsDeactivatedByNotReceivingAnyOptions(t *testing.T) {
	activators := map[rune]int{
		'(': 0,
	}
	deactivators := []rune{')', ';'}
	currentBytesAndOffset := func() (bytes []byte, offset int) {
		return []byte("fmt.Print"), len("fmt.Print")
	}
	currentLocation := func() Loc {
		return Loc{X: len("fmt.Print"), Y: 0}
	}
	locationOffset := func(Loc) int {
		return 0
	}
	provider := func(buffer []byte, startOffset, currentOffset int) (options []optionprovider.Option, delta int, err error) {
		options = []optionprovider.Option{} // No options.
		return
	}

	c := NewCompleter(activators,
		deactivators,
		provider,
		t.Logf,
		currentBytesAndOffset,
		currentLocation,
		locationOffset,
		noopReplacer,
		noopContentSetter,
		optionStyleInactive,
		optionStyleActive,
		enabledFlagSetToTrue)

	c.Active = true

	err := c.Process('l')
	if err != nil {
		t.Fatalf("failed to process with error: %v", err)
	}
	if c.Active {
		t.Errorf("expected receiving no autocomplete options to deactivate the completer, but it didn't")
	}
}

func TestCompleterIsRestartedIfARuneIsAnActivatorAndDeactivator(t *testing.T) {
	activators := map[rune]int{
		'.': 0,
	}
	deactivators := []rune{'.'}
	currentBytesAndOffset := func() (bytes []byte, offset int) {
		return []byte("test test"), 9
	}
	currentLocation := func() Loc {
		return Loc{X: 9, Y: 0}
	}
	locationOffset := func(Loc) int {
		return 9
	}
	provider := func(buffer []byte, startOffset, currentOffset int) (options []optionprovider.Option, delta int, err error) {
		options = []optionprovider.Option{
			optionprovider.New("test", "test"),
		}
		return
	}

	c := NewCompleter(activators,
		deactivators,
		provider,
		t.Logf,
		currentBytesAndOffset,
		currentLocation,
		locationOffset,
		noopReplacer,
		noopContentSetter,
		optionStyleInactive,
		optionStyleActive,
		enabledFlagSetToTrue)

	c.Active = true

	err := c.Process('.')
	if err != nil {
		t.Fatalf("failed to process with error: %v", err)
	}
	if !c.Active {
		t.Errorf("expected '.' to deactivate, then reactivate the completer, but it didn't")
	}
	if c.X != 9 {
		t.Errorf("expected the start position to be reset to x:9, but was %v", c.X)
	}
}

func TestCompleterIsNotTriggeredByOtherRunesWhenInactive(t *testing.T) {
	activators := map[rune]int{
		'(': 0,
	}
	deactivators := []rune{')', ';'}
	currentBytesAndOffset := func() (bytes []byte, offset int) {
		return []byte("fmt.Println("), 3
	}
	currentLocation := func() Loc {
		return Loc{X: 1, Y: 2}
	}
	locationOffset := func(Loc) int {
		return 0
	}
	provider := func(buffer []byte, startOffset, currentOffset int) (options []optionprovider.Option, delta int, err error) {
		options = []optionprovider.Option{
			optionprovider.New("text", "hint"),
		}
		return
	}
	c := NewCompleter(activators,
		deactivators,
		provider,
		t.Logf,
		currentBytesAndOffset,
		currentLocation,
		locationOffset,
		noopReplacer,
		noopContentSetter,
		optionStyleInactive,
		optionStyleActive,
		enabledFlagSetToTrue)

	err := c.Process('a')
	if err != nil {
		t.Fatalf("failed to process with error: %v", err)
	}
	if c.Active {
		t.Errorf("expected 'a' to do nothing to activate the completer, but the completer was activated")
	}
}

func TestCompleterAdjustsStartPositionIfOptionProviderMovesIt(t *testing.T) {
	activators := map[rune]int{
		'.': 0,
	}
	deactivators := []rune{'.'}
	currentBytesAndOffset := func() (bytes []byte, offset int) {
		return []byte("test test"), 9
	}
	currentLocation := func() Loc {
		return Loc{X: 9, Y: 0}
	}
	locationOffset := func(Loc) int {
		return 9
	}
	provider := func(buffer []byte, startOffset, currentOffset int) (options []optionprovider.Option, startPositionDelta int, err error) {
		options = []optionprovider.Option{
			optionprovider.New("test", "test"),
		}
		startPositionDelta = -1
		return
	}

	c := NewCompleter(activators,
		deactivators,
		provider,
		t.Logf,
		currentBytesAndOffset,
		currentLocation,
		locationOffset,
		noopReplacer,
		noopContentSetter,
		optionStyleInactive,
		optionStyleActive,
		enabledFlagSetToTrue)

	c.Active = true
	c.X = 7
	err := c.Process('e')
	if err != nil {
		t.Fatalf("failed to process with error: %v", err)
	}
	if c.X != 6 {
		t.Errorf("expected the start position to shift by the -1 delta returned by the options, but was %v", c.X)
	}
}

func TestCompleterHandleEventNotEnabled(t *testing.T) {
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil, nil, nil, optionStyleInactive, optionStyleActive, enabledFlagSetToFalse)

	handled := c.HandleEvent(tcell.KeyRune)
	if handled {
		t.Error("when the completer is not enabled, handling events should not take place")
	}
}

func TestCompleterHandleEventInactive(t *testing.T) {
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil, nil, nil, optionStyleInactive, optionStyleActive, enabledFlagSetToTrue)

	handled := c.HandleEvent(tcell.KeyRune)
	if handled {
		t.Error("when the completer is inactive, handling events should not take place")
	}
}

func TestCompleterHandleEventKeyUp(t *testing.T) {
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil, nil, nil, optionStyleInactive, optionStyleActive, enabledFlagSetToTrue)

	c.Active = true
	c.ActiveIndex = 10

	handled := c.HandleEvent(tcell.KeyUp)
	if !handled {
		t.Error("when the completer is active, KeyUp should be handled")
	}
	if c.ActiveIndex != 9 {
		t.Errorf("KeyUp should decrease the active index from 10 to 9, but the result was %v", c.ActiveIndex)
	}

	// Check that it's not possible to go before option index zero.
	c.ActiveIndex = 0
	c.HandleEvent(tcell.KeyUp)
	if c.ActiveIndex != 0 {
		t.Errorf("Once the top of the selections are reached, it shouldn't be possible to go any further, but the result was %v", c.ActiveIndex)
	}
}

func TestCompleterHandleEventKeyDown(t *testing.T) {
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil, nil, nil, optionStyleInactive, optionStyleActive, enabledFlagSetToTrue)

	c.Active = true
	c.Options = []optionprovider.Option{
		{H: "hint0", T: "text0"},
		{H: "hint1", T: "text1"},
	}
	c.ActiveIndex = 0

	handled := c.HandleEvent(tcell.KeyDown)
	if !handled {
		t.Error("when the completer is active, KeyDown should be handled")
	}
	if c.ActiveIndex != 1 {
		t.Errorf("KeyDown should increase the active index from 0 to 1, but the result was %v", c.ActiveIndex)
	}

	// Check that it's not possible to exceed the number of options.
	c.HandleEvent(tcell.KeyDown)
	if c.ActiveIndex != 1 {
		t.Errorf("Once the bottom of the selections are reached, it shouldn't be possible to go any further, but the result was %v", c.ActiveIndex)
	}
}

func TestCompleterHandleEventKeyEscape(t *testing.T) {
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil, nil, nil, optionStyleInactive, optionStyleActive, enabledFlagSetToTrue)

	c.Active = true

	handled := c.HandleEvent(tcell.KeyEscape)
	if !handled {
		t.Error("when the completer is active, KeyEscape should be handled")
	}
	if c.Active {
		t.Error("KeyEscape should stop the completer from being active")
	}
}

func TestCompleterHandleEventKeyTab(t *testing.T) {
	testCompleterHandleEventCompletion(tcell.KeyTab, t)
}

func TestCompleterHandleEventKeyEnter(t *testing.T) {
	testCompleterHandleEventCompletion(tcell.KeyEnter, t)
}

func testCompleterHandleEventCompletion(key tcell.Key, t *testing.T) {
	expectedFrom := Loc{X: 0, Y: 1}
	expectedTo := Loc{X: 1, Y: 2}

	currentLocation := func() Loc { return expectedTo }
	locationOffset := func(Loc) int {
		return 0
	}

	var receivedFrom, receivedTo Loc
	var receivedWith string

	replacer := func(from, to Loc, with string) {
		receivedFrom = from
		receivedTo = to
		receivedWith = with
	}
	c := NewCompleter(nil, nil, nil, t.Logf, nil, currentLocation, locationOffset, replacer, nil, optionStyleInactive, optionStyleActive, enabledFlagSetToTrue)

	c.X = expectedFrom.X
	c.Y = expectedFrom.Y
	c.Active = true
	c.Options = []optionprovider.Option{
		{H: "hint0", T: "text0"},
		{H: "hint1", T: "text1"},
	}
	c.ActiveIndex = 1

	handled := c.HandleEvent(key)
	if !handled {
		t.Error("when the completer is active, the completion should be handled")
	}
	if expectedFrom != receivedFrom {
		t.Errorf("expected from location %v but got %v", expectedFrom, receivedFrom)
	}
	if expectedTo != receivedTo {
		t.Errorf("expected to location %v but got %v", expectedTo, receivedTo)
	}
	if "text1" != receivedWith {
		t.Errorf("expected to receive a replacement of 'text1', but got %v", receivedWith)
	}
}

func TestCompleterHandleEventKeyWhenActive(t *testing.T) {
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil, nil, nil, optionStyleInactive, optionStyleActive, enabledFlagSetToTrue)

	c.Active = true

	handled := c.HandleEvent(tcell.KeyRune)
	if handled {
		t.Error("when the completer is active, KeyRune should have no effect")
	}
}

func TestCompleterGetOption(t *testing.T) {
	defaultOptions := []optionprovider.Option{
		optionprovider.New("text", "hint"),
		optionprovider.New("text1", "hint1"),
	}

	tests := []struct {
		index        int
		options      []optionprovider.Option
		expectedText string
		expectedOK   bool
	}{
		{
			index:        -1,
			options:      defaultOptions,
			expectedText: "text", // Use the first entry by default.
			expectedOK:   true,
		},
		{
			index:        -1,
			options:      []optionprovider.Option{},
			expectedText: "", // If there are no options, don't use anything.
			expectedOK:   false,
		},
		{
			index:        0,
			options:      defaultOptions,
			expectedText: "text",
			expectedOK:   true,
		},
		{
			index:        1,
			options:      defaultOptions,
			expectedText: "text1",
			expectedOK:   true,
		},
		{
			index:        2,
			options:      defaultOptions,
			expectedText: "",
			expectedOK:   false,
		},
	}

	for _, test := range tests {
		actualText, actualOK := getOption(test.index, test.options)
		if test.expectedText != actualText {
			t.Errorf("for index %v, expected '%v', but got '%v'", test.index, test.expectedText, actualText)
		}
		if test.expectedOK != actualOK {
			t.Errorf("for index %v, expected '%v', but got '%v'", test.index, test.expectedOK, actualOK)
		}
	}
}

func TestCompleterDisplayDoesNotWriteToConsoleWhenNotEnabled(t *testing.T) {
	var setterCalled bool
	setter := func(x int, y int, mainc rune, combc []rune, style tcell.Style) {
		setterCalled = true
	}
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil, nil, setter, optionStyleInactive, optionStyleActive, enabledFlagSetToFalse)

	c.Display()
	if setterCalled {
		t.Error("when the completer is not enabled, expected no content to be written to the screen")
	}
}

func TestCompleterDisplayDoesNotWriteToConsoleWhenInactive(t *testing.T) {
	var setterCalled bool
	setter := func(x int, y int, mainc rune, combc []rune, style tcell.Style) {
		setterCalled = true
	}
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil, nil, setter, optionStyleInactive, optionStyleActive, enabledFlagSetToTrue)

	c.Display()
	if setterCalled {
		t.Error("when the completer is inactive, expected no content to be written to the screen")
	}
}

type rs struct {
	Rune  rune
	Style tcell.Style
}

type displayMap map[Loc]rs

func (a displayMap) Eq(b displayMap) bool {
	if len(a) != len(b) {
		return false
	}
	for locA, rsA := range a {
		rsB, ok := b[locA]
		if !ok {
			return false
		}
		if rsA.Rune != rsB.Rune {
			return false
		}
		if rsA.Style != rsB.Style {
			return false
		}
	}
	return true
}

func (a displayMap) String() string {
	buf := bytes.NewBuffer([]byte{})

	if len(a) == 0 {
		return ""
	}

	minX, maxX := a.X()
	minY, maxY := a.Y()

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			l := Loc{X: x, Y: y}
			// Get the value of the rune.
			v, ok := a[l]
			if ok {
				buf.WriteRune(v.Rune)
			} else {
				buf.WriteRune(' ')
			}
		}
		if y < maxY {
			buf.WriteRune('\n')
		}
	}
	return buf.String()
}

func (a displayMap) X() (min, max int) {
	first := true
	for k := range a {
		if first {
			min = k.X
			max = k.X
			first = false
		}
		if k.X < min {
			min = k.X
		}
		if k.X > max {
			max = k.X
		}
	}
	return
}

func (a displayMap) Y() (min, max int) {
	first := true
	for k := range a {
		if first {
			min = k.Y
			max = k.Y
			first = false
		}
		if k.Y < min {
			min = k.Y
		}
		if k.Y > max {
			max = k.Y
		}
	}
	return
}

func TestDisplayMapString(t *testing.T) {
	m := make(displayMap)

	if m.String() != "" {
		t.Errorf("an empty display should be an empty string, but got '%v'", m.String())
	}

	m[Loc{Y: 0, X: 0}] = rs{Rune: 'a'}
	if m.String() != "a" {
		t.Errorf("expected 'a', got '%v'", m.String())
	}
	m[Loc{Y: 0, X: 2}] = rs{Rune: 'c'}
	if m.String() != "a c" {
		t.Errorf("expected 'a c', got '%v'", m.String())
	}
}

func TestCompleterDisplayRendersOptionsWhenActive(t *testing.T) {
	acs := optionStyleInactive
	act := optionStyleActive

	tests := []struct {
		name                string
		options             []optionprovider.Option
		selectedOptionIndex int
		expected            displayMap
	}{
		{
			name:                "no options",
			options:             []optionprovider.Option{},
			selectedOptionIndex: -1,
			expected:            displayMap{},
		},
		{
			name: "single option",
			options: []optionprovider.Option{
				optionprovider.New("Text", "Hint"),
			},
			selectedOptionIndex: -1,
			expected: displayMap{
				Loc{Y: 1, X: 0}: rs{'T', acs}, Loc{Y: 1, X: 1}: rs{'e', acs}, Loc{Y: 1, X: 2}: rs{'x', acs}, Loc{Y: 1, X: 3}: rs{'t', acs}, Loc{Y: 1, X: 4}: rs{rune(0), acs},
			},
		},
		{
			name: "multiple options",
			options: []optionprovider.Option{
				optionprovider.New("Text", "Hint"),
				optionprovider.New("Text2", "Hint2"),
			},
			selectedOptionIndex: -1,
			expected: displayMap{
				Loc{Y: 1, X: 0}: rs{'T', acs}, Loc{Y: 1, X: 1}: rs{'e', acs}, Loc{Y: 1, X: 2}: rs{'x', acs}, Loc{Y: 1, X: 3}: rs{'t', acs}, Loc{Y: 1, X: 4}: rs{0, acs}, Loc{Y: 1, X: 5}: rs{0, acs},
				Loc{Y: 2, X: 0}: rs{'T', acs}, Loc{Y: 2, X: 1}: rs{'e', acs}, Loc{Y: 2, X: 2}: rs{'x', acs}, Loc{Y: 2, X: 3}: rs{'t', acs}, Loc{Y: 2, X: 4}: rs{'2', acs}, Loc{Y: 2, X: 5}: rs{0, acs},
			},
		},
		{
			name: "multiple options, last selected",
			options: []optionprovider.Option{
				optionprovider.New("Text", "Hint"),
				optionprovider.New("Text2", "Hint2"),
			},
			selectedOptionIndex: 1,
			expected: displayMap{
				Loc{Y: 1, X: 0}: rs{'T', acs}, Loc{Y: 1, X: 1}: rs{'e', acs}, Loc{Y: 1, X: 2}: rs{'x', acs}, Loc{Y: 1, X: 3}: rs{'t', acs}, Loc{Y: 1, X: 4}: rs{0, acs}, Loc{Y: 1, X: 5}: rs{0, acs},
				Loc{Y: 2, X: 0}: rs{'T', act}, Loc{Y: 2, X: 1}: rs{'e', act}, Loc{Y: 2, X: 2}: rs{'x', act}, Loc{Y: 2, X: 3}: rs{'t', act}, Loc{Y: 2, X: 4}: rs{'2', act}, Loc{Y: 2, X: 5}: rs{0, act},
			},
		},
	}

	for _, test := range tests {
		testCompleterDisplayRendersOptionsWhenActive(test.name, test.options, test.selectedOptionIndex, test.expected, t)
	}
}

func testCompleterDisplayRendersOptionsWhenActive(name string,
	options []optionprovider.Option,
	activeIndex int,
	expected displayMap,
	t *testing.T) {
	actual := make(displayMap)

	currentLocation := func() Loc {
		return Loc{X: 0, Y: 0}
	}
	locationOffset := func(Loc) int {
		return 0
	}

	var setterCalled bool
	setter := func(x int, y int, mainc rune, combc []rune, style tcell.Style) {
		setterCalled = true
		actual[Loc{X: x, Y: y}] = rs{Rune: mainc, Style: style}
	}
	c := NewCompleter(nil, nil, nil, t.Logf, nil, currentLocation, locationOffset, nil, setter, optionStyleInactive, optionStyleActive, enabledFlagSetToTrue)
	c.Active = true
	c.ActiveIndex = activeIndex
	c.Options = options

	c.Display()
	if !expected.Eq(actual) {
		t.Errorf("%s: expected characters '%v', got '%v'", name, expected.String(), actual.String())
		t.Errorf("%s: expected '%v', got '%v'", name, map[Loc]rs(expected), map[Loc]rs(actual))
	}
}

func TestDeactivateIfOutOfBounds(t *testing.T) {
	tests := []struct {
		name             string
		previousLocation Loc
		currentLocation  Loc
		currentlyActive  bool
		expectedActive   bool
	}{
		{
			name:             "No change results in cancellation",
			previousLocation: Loc{X: 30, Y: 0},
			currentLocation:  Loc{X: 30, Y: 0},
			currentlyActive:  true,
			expectedActive:   false,
		},
		{
			name:             "Moving forward retains active",
			previousLocation: Loc{X: 30, Y: 0},
			currentLocation:  Loc{X: 31, Y: 0},
			currentlyActive:  true,
			expectedActive:   true,
		},
		{
			name:             "Moving back to the start cancels",
			previousLocation: Loc{X: 31, Y: 0},
			currentLocation:  Loc{X: 30, Y: 0},
			currentlyActive:  true,
			expectedActive:   false,
		},
		{
			name:             "Moving before the start cancels",
			previousLocation: Loc{X: 31, Y: 0},
			currentLocation:  Loc{X: 25, Y: 0},
			currentlyActive:  true,
			expectedActive:   false,
		},
		{
			name:             "Moving to the line start cancels",
			previousLocation: Loc{X: 31, Y: 0},
			currentLocation:  Loc{X: 0, Y: 0},
			currentlyActive:  true,
			expectedActive:   false,
		},
		{
			name:             "Moving more than one character to the right cancels (could be a paste, or mouse click)",
			previousLocation: Loc{X: 31, Y: 0},
			currentLocation:  Loc{X: 35, Y: 0},
			currentlyActive:  true,
			expectedActive:   false,
		},
		{
			name:             "Switching lines cancels",
			previousLocation: Loc{X: 30, Y: 0},
			currentLocation:  Loc{X: 31, Y: 1},
			currentlyActive:  true,
			expectedActive:   false,
		},
	}

	for _, test := range tests {
		currentLocation := func() Loc {
			return test.currentLocation
		}
		c := NewCompleter(nil, nil, nil, t.Logf, nil, currentLocation, nil, nil, nil, optionStyleInactive, optionStyleActive, enabledFlagSetToTrue)
		c.X = test.previousLocation.X
		c.Y = test.previousLocation.Y
		c.PreviousLocation = test.previousLocation
		c.Active = test.currentlyActive

		c.DeactivateIfOutOfBounds()

		if c.Active != test.expectedActive {
			t.Errorf("%s: expected active to now be %v, but was %v", test.name, test.expectedActive, c.Active)
		}
	}
}
