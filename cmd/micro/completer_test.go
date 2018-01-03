package main

import "testing"
import "github.com/zyedidia/micro/cmd/micro/optionprovider"
import "reflect"
import "github.com/zyedidia/tcell"

var noopReplacer = func(from, to Loc, with string) {}

func TestThatTheCompleterIsDeactivatedByDeactivatorRunes(t *testing.T) {
	activators := []rune{'('}
	deactivators := []rune{')', ';'}
	currentBytesAndOffset := func() (bytes []byte, offset int) {
		return []byte("fmt.Println("), 3
	}
	currentLocation := func() Loc {
		return Loc{X: 1, Y: 2}
	}
	provider := func(buffer []byte, offset int) (options []optionprovider.Option, err error) {
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
		noopReplacer)

	c.Active = true

	err := c.Process(')')
	if err != nil {
		t.Fatalf("failed to process with error: %v", err)
	}
	if c.Active {
		t.Errorf("expected ')' to deactivate the completer, but it didn't")
	}
}

func TestThatTheCompleterIsActivatedByActivatorRunes(t *testing.T) {
	var providerReceivedBytes []byte
	var providerReceivedOffset int

	expectedOptions := []optionprovider.Option{
		optionprovider.New("text", "hint"),
		optionprovider.New("text1", "hint1"),
	}

	activators := []rune{'('}
	deactivators := []rune{')', ';'}
	currentBytesAndOffset := func() (bytes []byte, offset int) {
		return []byte("fmt.Println("), 3
	}
	currentLocation := func() Loc {
		return Loc{X: 1, Y: 2}
	}
	provider := func(buffer []byte, offset int) (options []optionprovider.Option, err error) {
		providerReceivedBytes = buffer
		providerReceivedOffset = offset
		options = expectedOptions
		return
	}
	c := NewCompleter(activators,
		deactivators,
		provider,
		t.Logf,
		currentBytesAndOffset,
		currentLocation,
		noopReplacer)

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

func TestThatAnInactiveCompleterIsNotTriggeredByOtherRunes(t *testing.T) {
	activators := []rune{'('}
	deactivators := []rune{')', ';'}
	currentBytesAndOffset := func() (bytes []byte, offset int) {
		return []byte("fmt.Println("), 3
	}
	currentLocation := func() Loc {
		return Loc{X: 1, Y: 2}
	}
	provider := func(buffer []byte, offset int) (options []optionprovider.Option, err error) {
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
		noopReplacer)

	err := c.Process('a')
	if err != nil {
		t.Fatalf("failed to process with error: %v", err)
	}
	if c.Active {
		t.Errorf("expected 'a' to do nothing to activate the completer, but the completer was activated")
	}
}

func TestHandleEventInactive(t *testing.T) {
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil)

	handled := c.HandleEvent(tcell.KeyRune)
	if handled {
		t.Error("when the completer is inactive, handling events should not take place")
	}
}

func TestHandleEventKeyUp(t *testing.T) {
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil)

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

func TestHandleEventKeyDown(t *testing.T) {
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil)

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

func TestHandleEventKeyEscape(t *testing.T) {
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil)

	c.Active = true

	handled := c.HandleEvent(tcell.KeyEscape)
	if !handled {
		t.Error("when the completer is active, KeyEscape should be handled")
	}
	if c.Active {
		t.Error("KeyEscape should stop the completer from being active")
	}
}

func TestHandleEventKeyTab(t *testing.T) {
	testHandleEventCompletion(tcell.KeyTab, t)
}

func TestHandleEventKeyEnter(t *testing.T) {
	testHandleEventCompletion(tcell.KeyEnter, t)
}

func testHandleEventCompletion(key tcell.Key, t *testing.T) {
	expectedFrom := Loc{X: 0, Y: 1}
	expectedTo := Loc{X: 1, Y: 2}

	currentLocation := func() Loc { return expectedTo }

	var receivedFrom, receivedTo Loc
	var receivedWith string

	replacer := func(from, to Loc, with string) {
		receivedFrom = from
		receivedTo = to
		receivedWith = with
	}
	c := NewCompleter(nil, nil, nil, t.Logf, nil, currentLocation, replacer)

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

func TestHandleEventKeyWhenActive(t *testing.T) {
	c := NewCompleter(nil, nil, nil, t.Logf, nil, nil, nil)

	c.Active = true

	handled := c.HandleEvent(tcell.KeyRune)
	if handled {
		t.Error("when the completer is active, KeyRune should have no effect")
	}
}

func TestGetOption(t *testing.T) {
	options := []optionprovider.Option{
		optionprovider.New("text", "hint"),
		optionprovider.New("text1", "hint1"),
	}

	tests := []struct {
		index        int
		expectedText string
		expectedOK   bool
	}{
		{
			index:        -1,
			expectedText: "",
			expectedOK:   false,
		},
		{
			index:        0,
			expectedText: "text",
			expectedOK:   true,
		},
		{
			index:        1,
			expectedText: "text1",
			expectedOK:   true,
		},
		{
			index:        2,
			expectedText: "",
			expectedOK:   false,
		},
	}

	for _, test := range tests {
		actualText, actualOK := getOption(test.index, options)
		if test.expectedText != actualText {
			t.Errorf("for index %v, expected '%v', but got '%v'", test.index, test.expectedText, actualText)
		}
		if test.expectedOK != actualOK {
			t.Errorf("for index %v, expected '%v', but got '%v'", test.index, test.expectedOK, actualOK)
		}
	}
}
