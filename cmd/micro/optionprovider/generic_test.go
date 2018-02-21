package optionprovider

import (
	"reflect"
	"testing"
)

func TestGeneric(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		from          string
		to            string
		expected      []Option
		expectedDelta int
	}{
		{
			name:     "words are sorted alphabetically if they appear the same number of times",
			text:     "fish dog cat ",
			expected: []Option{New("cat", ""), New("dog", ""), New("fish", "")},
		},
		{
			name: "capitals preceed lowercase",
			text: "A a B b C c",
			expected: []Option{
				New("A", ""), New("B", ""), New("C", ""),
				New("a", ""), New("b", ""), New("c", ""),
			},
		},
		{
			name:     "bare numbers are not included",
			text:     "1 2 3 1.23 a",
			expected: []Option{New("a", "")},
		},
		{
			name:     "words that include a number are included",
			text:     "a1",
			expected: []Option{New("a1", "")},
		},
		{
			name:     "words are ordered by their frequency descending (most common words are first in the list)",
			text:     "ccc ccc ccc bb bb a",
			expected: []Option{New("ccc", ""), New("bb", ""), New("a", "")},
		},
		{
			name:     "prefix matches preceed other matches",
			text:     "common common common something",
			from:     "common common common ",
			to:       "common common common s",
			expected: []Option{New("something", "")},
		},
		{
			name:     "the autocomplete looks for the previous word boundary to see if you're partway through a word to limit results",
			text:     "A AB ABC ABCD",
			from:     "A AB ",
			to:       "A AB AB",
			expected: []Option{New("ABC", ""), New("ABCD", "")},
		},
		{
			name:     "realistic example",
			text:     `fmt.Println("hello") fmt.P`,
			from:     `fmt.Println("hello") fmt.`,
			to:       `fmt.Println("hello") fmt.P`,
			expected: []Option{New("Println", "")},
		},
		{
			name:          "go back further than the start position",
			text:          `testing`,
			from:          `test`,
			to:            `testi`,
			expected:      []Option{New("testing", "")},
			expectedDelta: -4,
		},
	}

	for _, test := range tests {
		options, delta, err := Generic(t.Logf, []byte(test.text), len(test.from), len(test.to))
		if err != nil {
			t.Fatalf("%s: generic complete failed with error %v", test.name, err)
			continue
		}
		if !reflect.DeepEqual(options, test.expected) {
			t.Errorf("%s: expected '%v', got '%v'", test.name, test.expected, options)
		}
		if delta != test.expectedDelta {
			t.Errorf("%s: expected delta %v, got %v", test.name, test.expectedDelta, delta)
		}
	}
}

func TestLastCharacters(t *testing.T) {
	tests := []struct {
		input    string
		end      int
		length   int
		expected string
	}{
		{
			input:    "ABC",
			end:      3,
			length:   1,
			expected: "C",
		},
		{
			input:    "ABC",
			end:      3,
			length:   2,
			expected: "BC",
		},
		{
			input:    "ABC",
			end:      3,
			length:   3,
			expected: "ABC",
		},
		{
			input:    "ABC",
			end:      3,
			length:   4,
			expected: "ABC",
		},
		{
			input:    "ABCD",
			end:      3,
			length:   3,
			expected: "ABC",
		},
		{
			input:    "ABCD",
			end:      2,
			length:   6,
			expected: "AB",
		},
		{
			input:    "ABCD",
			end:      0,
			length:   6,
			expected: "",
		},
	}

	for _, test := range tests {
		actual := lastCharacters(test.input, test.end, test.length)
		if actual != test.expected {
			t.Errorf("for input '%v', the %v characters before index %v, expected: '%v', but got '%v'",
				test.input, test.length, test.end, test.expected, actual)
		}
	}
}

func TestPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "fmt.Prin",
			expected: "Prin",
		},
		{
			input:    "Test t",
			expected: "t",
		},
		{
			input:    "1234 123",
			expected: "123",
		},
		{
			input:    "word-connection",
			expected: "connection",
		},
		{
			input:    `"quote`,
			expected: "quote",
		},
	}

	for _, test := range tests {
		actual := prefix(test.input)
		if actual != test.expected {
			t.Errorf("for input '%v', expected: '%v', but got '%v'",
				test.input, test.expected, actual)
		}
	}
}
