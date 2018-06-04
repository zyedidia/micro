package main

import (
	"testing"
)

func TestNumOccurences(t *testing.T) {
	var tests = []struct {
		inputStr  string
		inputChar byte
		want      int
	}{
		{"aaaa", 'a', 4},
		{"\trfd\ta", '\t', 2},
		{"∆ƒ\tø ® \t\t", '\t', 3},
	}
	for _, test := range tests {
		if got := NumOccurrences(test.inputStr, test.inputChar); got != test.want {
			t.Errorf("NumOccurences(%s, %c) = %d", test.inputStr, test.inputChar, got)
		}
	}
}

func TestSpaces(t *testing.T) {
	var tests = []struct {
		input int
		want  string
	}{
		{4, "    "},
		{0, ""},
	}
	for _, test := range tests {
		if got := Spaces(test.input); got != test.want {
			t.Errorf("Spaces(%d) = \"%s\"", test.input, got)
		}
	}
}

func TestIsWordChar(t *testing.T) {
	if IsWordChar("t") == false {
		t.Errorf("IsWordChar(t) = false")
	}
	if IsWordChar("T") == false {
		t.Errorf("IsWordChar(T) = false")
	}
	if IsWordChar("5") == false {
		t.Errorf("IsWordChar(5) = false")
	}
	if IsWordChar("_") == false {
		t.Errorf("IsWordChar(_) = false")
	}
	if IsWordChar("ß") == false {
		t.Errorf("IsWordChar(ß) = false")
	}
	if IsWordChar("~") == true {
		t.Errorf("IsWordChar(~) = true")
	}
	if IsWordChar(" ") == true {
		t.Errorf("IsWordChar( ) = true")
	}
	if IsWordChar(")") == true {
		t.Errorf("IsWordChar()) = true")
	}
	if IsWordChar("\n") == true {
		t.Errorf("IsWordChar(\n)) = true")
	}
}

func TestStringWidth(t *testing.T) {
	tabsize := 4
	if w := StringWidth("1\t2", tabsize); w != 5 {
		t.Error("StringWidth 1 Failed. Got", w)
	}
	if w := StringWidth("\t", tabsize); w != 4 {
		t.Error("StringWidth 2 Failed. Got", w)
	}
	if w := StringWidth("1\t", tabsize); w != 4 {
		t.Error("StringWidth 3 Failed. Got", w)
	}
	if w := StringWidth("\t\t", tabsize); w != 8 {
		t.Error("StringWidth 4 Failed. Got", w)
	}
	if w := StringWidth("12\t2\t", tabsize); w != 8 {
		t.Error("StringWidth 5 Failed. Got", w)
	}
}

func TestWidthOfLargeRunes(t *testing.T) {
	tabsize := 4
	if w := WidthOfLargeRunes("1\t2", tabsize); w != 2 {
		t.Error("WidthOfLargeRunes 1 Failed. Got", w)
	}
	if w := WidthOfLargeRunes("\t", tabsize); w != 3 {
		t.Error("WidthOfLargeRunes 2 Failed. Got", w)
	}
	if w := WidthOfLargeRunes("1\t", tabsize); w != 2 {
		t.Error("WidthOfLargeRunes 3 Failed. Got", w)
	}
	if w := WidthOfLargeRunes("\t\t", tabsize); w != 6 {
		t.Error("WidthOfLargeRunes 4 Failed. Got", w)
	}
	if w := WidthOfLargeRunes("12\t2\t", tabsize); w != 3 {
		t.Error("WidthOfLargeRunes 5 Failed. Got", w)
	}
}

func assertEqual(t *testing.T, expected interface{}, result interface{}) {
	if expected != result {
		t.Fatalf("Expected: %d != Got: %d", expected, result)
	}
}

func assertTrue(t *testing.T, condition bool) {
	if !condition {
		t.Fatalf("Condition was not true. Got false")
	}
}

func TestGetPathRelativeWithDot(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("./myfile:10:5")

	assertEqual(t, path, "./myfile")
	assertEqual(t, "10", cursorPosition[0])
	assertEqual(t, "5", cursorPosition[1])
}
func TestGetPathRelativeWithDotWindows(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition(".\\myfile:10:5")

	assertEqual(t, path, ".\\myfile")
	assertEqual(t, "10", cursorPosition[0])
	assertEqual(t, cursorPosition[1], "5")
}
func TestGetPathRelativeNoDot(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("myfile:10:5")

	assertEqual(t, path, "myfile")
	assertEqual(t, "10", cursorPosition[0])

	assertEqual(t, cursorPosition[1], "5")
}
func TestGetPathAbsoluteWindows(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("C:\\myfile:10:5")

	assertEqual(t, path, "C:\\myfile")
	assertEqual(t, "10", cursorPosition[0])

	assertEqual(t, cursorPosition[1], "5")

	path, cursorPosition = GetPathAndCursorPosition("C:/myfile:10:5")

	assertEqual(t, path, "C:/myfile")
	assertEqual(t, "10", cursorPosition[0])
	assertEqual(t, "5", cursorPosition[1])
}

func TestGetPathAbsoluteUnix(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("/home/user/myfile:10:5")

	assertEqual(t, path, "/home/user/myfile")
	assertEqual(t, "10", cursorPosition[0])
	assertEqual(t, "5", cursorPosition[1])
}

func TestGetPathRelativeWithDotWithoutLineAndColumn(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("./myfile")

	assertEqual(t, path, "./myfile")
	// no cursor position in filename, nil should be returned
	assertTrue(t, cursorPosition == nil)
}
func TestGetPathRelativeWithDotWindowsWithoutLineAndColumn(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition(".\\myfile")

	assertEqual(t, path, ".\\myfile")
	assertTrue(t, cursorPosition == nil)

}
func TestGetPathRelativeNoDotWithoutLineAndColumn(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("myfile")

	assertEqual(t, path, "myfile")
	assertTrue(t, cursorPosition == nil)

}
func TestGetPathAbsoluteWindowsWithoutLineAndColumn(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("C:\\myfile")

	assertEqual(t, path, "C:\\myfile")
	assertTrue(t, cursorPosition == nil)

	path, cursorPosition = GetPathAndCursorPosition("C:/myfile")

	assertEqual(t, path, "C:/myfile")
	assertTrue(t, cursorPosition == nil)

}
func TestGetPathAbsoluteUnixWithoutLineAndColumn(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("/home/user/myfile")

	assertEqual(t, path, "/home/user/myfile")
	assertTrue(t, cursorPosition == nil)

}
func TestGetPathSingleLetterFileRelativePath(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("a:5:6")

	assertEqual(t, path, "a")
	assertEqual(t, "5", cursorPosition[0])
	assertEqual(t, "6", cursorPosition[1])
}
func TestGetPathSingleLetterFileAbsolutePathWindows(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("C:\\a:5:6")

	assertEqual(t, path, "C:\\a")
	assertEqual(t, "5", cursorPosition[0])
	assertEqual(t, "6", cursorPosition[1])

	path, cursorPosition = GetPathAndCursorPosition("C:/a:5:6")

	assertEqual(t, path, "C:/a")
	assertEqual(t, "5", cursorPosition[0])
	assertEqual(t, "6", cursorPosition[1])
}
func TestGetPathSingleLetterFileAbsolutePathUnix(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("/home/user/a:5:6")

	assertEqual(t, path, "/home/user/a")
	assertEqual(t, "5", cursorPosition[0])
	assertEqual(t, "6", cursorPosition[1])
}
func TestGetPathSingleLetterFileAbsolutePathWindowsWithoutLineAndColumn(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("C:\\a")

	assertEqual(t, path, "C:\\a")
	assertTrue(t, cursorPosition == nil)

	path, cursorPosition = GetPathAndCursorPosition("C:/a")

	assertEqual(t, path, "C:/a")
	assertTrue(t, cursorPosition == nil)

}
func TestGetPathSingleLetterFileAbsolutePathUnixWithoutLineAndColumn(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("/home/user/a")

	assertEqual(t, path, "/home/user/a")
	assertTrue(t, cursorPosition == nil)

}

func TestGetPathRelativeWithDotOnlyLine(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("./myfile:10")

	assertEqual(t, path, "./myfile")
	assertEqual(t, "10", cursorPosition[0])
	assertEqual(t, "0", cursorPosition[1])
}
func TestGetPathRelativeWithDotWindowsOnlyLine(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition(".\\myfile:10")

	assertEqual(t, path, ".\\myfile")
	assertEqual(t, "10", cursorPosition[0])
	assertEqual(t, "0", cursorPosition[1])
}
func TestGetPathRelativeNoDotOnlyLine(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("myfile:10")

	assertEqual(t, path, "myfile")
	assertEqual(t, "10", cursorPosition[0])
	assertEqual(t, "0", cursorPosition[1])
}
func TestGetPathAbsoluteWindowsOnlyLine(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("C:\\myfile:10")

	assertEqual(t, path, "C:\\myfile")
	assertEqual(t, "10", cursorPosition[0])
	assertEqual(t, "0", cursorPosition[1])

	path, cursorPosition = GetPathAndCursorPosition("C:/myfile:10")

	assertEqual(t, path, "C:/myfile")
	assertEqual(t, "10", cursorPosition[0])
	assertEqual(t, "0", cursorPosition[1])
}
func TestGetPathAbsoluteUnixOnlyLine(t *testing.T) {
	path, cursorPosition := GetPathAndCursorPosition("/home/user/myfile:10")

	assertEqual(t, path, "/home/user/myfile")
	assertEqual(t, "10", cursorPosition[0])
	assertEqual(t, "0", cursorPosition[1])
}
func TestParseCursorLocationOneArg(t *testing.T) {
	location, err := ParseCursorLocation([]string{"3"})

	assertEqual(t, 3, location.Y)
	assertEqual(t, 0, location.X)
	assertEqual(t, nil, err)
}
func TestParseCursorLocationTwoArgs(t *testing.T) {
	location, err := ParseCursorLocation([]string{"3", "15"})

	assertEqual(t, 3, location.Y)
	assertEqual(t, 15, location.X)
	assertEqual(t, nil, err)
}
func TestParseCursorLocationNoArgs(t *testing.T) {
	location, err := ParseCursorLocation(nil)
	// the expected result is the start position - 0, 0
	assertEqual(t, 0, location.Y)
	assertEqual(t, 0, location.X)
	// an error will be present here as the positions we're parsing are a nil
	assertTrue(t, err != nil)
}
func TestParseCursorLocationFirstArgNotValidNumber(t *testing.T) {
	// the messenger is necessary as ParseCursorLocation
	// puts a message in it on error
	messenger = new(Messenger)
	_, err := ParseCursorLocation([]string{"apples", "1"})
	// the expected result is the start position - 0, 0
	assertTrue(t, messenger.hasMessage)
	assertTrue(t, err != nil)
}
func TestParseCursorLocationSecondArgNotValidNumber(t *testing.T) {
	// the messenger is necessary as ParseCursorLocation
	// puts a message in it on error
	messenger = new(Messenger)
	_, err := ParseCursorLocation([]string{"1", "apples"})
	// the expected result is the start position - 0, 0
	assertTrue(t, messenger.hasMessage)
	assertTrue(t, err != nil)
}
