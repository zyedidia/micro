package main

import "testing"

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
		if got := NumOccurences(test.inputStr, test.inputChar); got != test.want {
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
	if IsWordChar("~") == true {
		t.Errorf("IsWordChar(~) = true")
	}
	if IsWordChar(" ") == true {
		t.Errorf("IsWordChar( ) = true")
	}
	if IsWordChar("ß") == true {
		t.Errorf("IsWordChar(ß) = true")
	}
	if IsWordChar(")") == true {
		t.Errorf("IsWordChar()) = true")
	}
	if IsWordChar("\n") == true {
		t.Errorf("IsWordChar(\n)) = true")
	}
}
