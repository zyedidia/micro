package main

import "testing"

func TestFromCharPos(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		charpos  int
		expected Loc
	}{
		{
			name:     "first line",
			input:    "abc\ndef",
			charpos:  0,
			expected: Loc{Y: 0, X: 0},
		},
		{
			name:     "secnod line",
			input:    "abc\ndef",
			charpos:  5,
			expected: Loc{Y: 1, X: 1},
		},
	}

	for _, test := range tests {
		b := NewBufferFromString(test.input, "/dev/null")
		actual := FromCharPos(test.charpos, b)
		if test.expected != actual {
			t.Errorf("%s: expected '%v', got '%v'", test.name, test.expected, actual)
		}
	}
}
