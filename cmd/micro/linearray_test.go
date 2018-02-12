package main

import (
	"strings"
	"testing"
)

func TestLineArraySubstr(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		start, end Loc
		expected   string
	}{
		{
			name:     "first line",
			input:    "abc\ndef",
			start:    Loc{Y: 0, X: 0},
			end:      Loc{Y: 0, X: 3},
			expected: "abc",
		},
		{
			name:     "second line, first char",
			input:    "abc\ndef",
			start:    Loc{Y: 1, X: 0},
			end:      Loc{Y: 1, X: 1},
			expected: "d",
		},
		{
			name:     "middle lines lines",
			input:    "abc\ndef\nghi",
			start:    Loc{Y: 0, X: 0},
			end:      Loc{Y: 1, X: 2},
			expected: "abc\nde",
		},
		{
			name:     "all lines",
			input:    "abc\ndef",
			start:    Loc{Y: 0, X: 0},
			end:      Loc{Y: 1, X: 3},
			expected: "abc\ndef",
		},
	}

	for _, test := range tests {
		r := strings.NewReader(test.input)
		la := NewLineArray(10, r)
		actual := la.Substr(test.start, test.end)
		if test.expected != actual {
			t.Errorf("%s: expected '%v', got '%v'", test.name, test.expected, actual)
		}
	}
}

func TestLineArrayDeleteLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		line     int
		expected string
	}{
		{
			name:     "remove first line",
			input:    "abc",
			line:     0,
			expected: "",
		},
		{
			name:     "remove second line",
			input:    "abc\ndef",
			line:     1,
			expected: "abc",
		},
		{
			name:     "remove second line of 3",
			input:    "abc\ndef\nhij",
			line:     1,
			expected: "abc\nhij",
		},
	}

	for _, test := range tests {
		r := strings.NewReader(test.input)
		la := NewLineArray(10, r)
		la.DeleteLine(test.line)
		if la.String() != test.expected {
			t.Errorf("%s: expected '%v', got '%v'", test.name, test.expected, la.String())
		}
	}
}

func TestLineArrayRemove(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		start, end Loc
		expected   string
	}{
		{
			name:     "remove first character",
			input:    "abc",
			start:    Loc{Y: 0, X: 0},
			end:      Loc{Y: 0, X: 1},
			expected: "bc",
		},
	}

	for _, test := range tests {
		r := strings.NewReader(test.input)
		la := NewLineArray(10, r)
		la.remove(test.start, test.end)
		if la.String() != test.expected {
			t.Errorf("%s: expected '%v', got '%v'", test.name, test.expected, la.String())
		}
	}
}
