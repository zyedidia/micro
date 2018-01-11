package optionprovider

import (
	"reflect"
	"testing"
)

func TestGeneric(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		leading  string
		expected []Option
	}{
		{
			name:     "words are sorted alphabetically if they appear the same number of times",
			text:     "fish dog cat",
			leading:  "",
			expected: []Option{New("cat", ""), New("dog", ""), New("fish", "")},
		},
		{
			name:    "capitals preceed lowercase",
			text:    "A a B b C c",
			leading: "",
			expected: []Option{
				New("A", ""), New("B", ""), New("C", ""),
				New("a", ""), New("b", ""), New("c", ""),
			},
		},
		{
			name:     "bare numbers are not included",
			text:     "1 2 3 1.23 a",
			leading:  "",
			expected: []Option{New("a", "")},
		},
		{
			name:     "words that include a number are included",
			text:     "a1",
			leading:  "",
			expected: []Option{New("a1", "")},
		},
		{
			name:     "words are ordered by their frequency descending (most common words are first in the list)",
			text:     "ccc ccc ccc bb bb a",
			leading:  "",
			expected: []Option{New("ccc", ""), New("bb", ""), New("a", "")},
		},
		{
			name:     "the autocomplete looks for the previous word boundary to see if you're partway through a word to limit results",
			text:     `A AB ABC ABCD`,
			leading:  `A AB AB`,
			expected: []Option{New("ABC", ""), New("ABCD", ""), New("A", "")},
		},
		{
			name:     "realistic example",
			text:     `fmt.Println("hello") fmt.P`,
			leading:  `fmt.Println("hello") fmt.P`,
			expected: []Option{New("Println", ""), New("fmt", ""), New("hello", "")},
		},
	}

	for _, test := range tests {
		options, err := Generic([]byte(test.text), len(test.leading))
		if err != nil {
			t.Fatalf("%s: generic complete failed with error %v", test.name, err)
			continue
		}
		if !reflect.DeepEqual(options, test.expected) {
			t.Errorf("%s: expected '%v', got '%v'", test.name, test.expected, options)
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
