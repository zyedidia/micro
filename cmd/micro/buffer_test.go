package main

import (
	"testing"
)

func TestGetBufferCursorLocationEmptyArgs(t *testing.T) {
	buf := NewBufferFromString("this is my\nbuffer\nfile\nhello", "")

	location, err := GetBufferCursorLocation(nil, buf)

	assertEqual(t, 0, location.Y)
	assertEqual(t, 0, location.X)

	// an error is present due to the cursorLocation being nil
	assertTrue(t, err != nil)

}

func TestGetBufferCursorLocationStartposFlag(t *testing.T) {
	buf := NewBufferFromString("this is my\nbuffer\nfile\nhello", "")

	*flagStartPos = "1,2"

	location, err := GetBufferCursorLocation(nil, buf)

	assertTrue(t, 1 == location.Y)
	assertTrue(t, 2 == location.X)

	// an error is present due to the cursorLocation being nil
	assertTrue(t, err != nil)
}

func TestGetBufferCursorLocationInvalidStartposFlag(t *testing.T) {
	buf := NewBufferFromString("this is my\nbuffer\nfile\nhello", "")

	*flagStartPos = "apples,2"

	location, err := GetBufferCursorLocation(nil, buf)
	// expect to default to the start of the file, which is 0,0
	assertEqual(t, 0, location.Y)
	assertEqual(t, 0, location.X)

	// an error is present due to the cursorLocation being nil
	assertTrue(t, err != nil)
}
func TestGetBufferCursorLocationStartposFlagAndCursorPosition(t *testing.T) {
	text := "this is my\nbuffer\nfile\nhello"
	cursorPosition := []string{"3", "1"}

	buf := NewBufferFromString(text, "")

	*flagStartPos = "1,2"

	location, err := GetBufferCursorLocation(cursorPosition, buf)
	// expect to have the flag positions, not the cursor position
	assertEqual(t, 1, location.Y)
	assertEqual(t, 2, location.X)

	assertTrue(t, err == nil)
}
func TestGetBufferCursorLocationCursorPositionAndInvalidStartposFlag(t *testing.T) {
	text := "this is my\nbuffer\nfile\nhello"
	cursorPosition := []string{"3", "1"}

	buf := NewBufferFromString(text, "")

	*flagStartPos = "apples,2"

	location, err := GetBufferCursorLocation(cursorPosition, buf)
	// expect to have the flag positions, not the cursor position
	assertEqual(t, 3, location.Y)
	assertEqual(t, 1, location.X)

	// no errors this time as cursorPosition is not nil
	assertTrue(t, err == nil)
}

func TestGetBufferCursorLocationNoErrorWhenOverflowWithStartpos(t *testing.T) {
	text := "this is my\nbuffer\nfile\nhello"

	buf := NewBufferFromString(text, "")

	*flagStartPos = "50,50"

	location, err := GetBufferCursorLocation(nil, buf)
	// expect to have the flag positions, not the cursor position
	assertEqual(t, buf.NumLines-1, location.Y)
	assertEqual(t, 5, location.X)

	// error is expected as cursorPosition is nil
	assertTrue(t, err != nil)
}
func TestGetBufferCursorLocationNoErrorWhenOverflowWithCursorPosition(t *testing.T) {
	text := "this is my\nbuffer\nfile\nhello"
	cursorPosition := []string{"50", "2"}

	*flagStartPos = ""

	buf := NewBufferFromString(text, "")

	location, err := GetBufferCursorLocation(cursorPosition, buf)
	// expect to have the flag positions, not the cursor position
	assertEqual(t, buf.NumLines-1, location.Y)
	assertEqual(t, 2, location.X)

	// error is expected as cursorPosition is nil
	assertTrue(t, err == nil)
}

//func TestGetBufferCursorLocationColonArgs(t *testing.T) {
//	buf := new(Buffer)

//}
