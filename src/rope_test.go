package main

import "testing"

func TestInsert(t *testing.T) {
	var tests = []struct {
		origStr   string
		insertStr string
		insertPos int
		want      string
	}{
		{"foo", " bar", 3, "foo bar"},
		{"üñîç", "ø∂é", 4, "üñîçø∂é"},
		{"test", "3", 2, "te3st"},
		{"", "test", 0, "test"},
	}
	for _, test := range tests {
		r := NewRope(test.origStr)
		r.Insert(test.insertPos, test.insertStr)
		got := r.String()
		if got != test.want {
			t.Errorf("Insert(%d, %s) = %s", test.insertPos, test.insertStr, got)
		}
	}
}

func TestRemove(t *testing.T) {
	var tests = []struct {
		inputStr    string
		removeStart int
		removeEnd   int
		want        string
	}{
		{"foo bar", 3, 7, "foo"},
		{"üñîçø∂é", 0, 3, "çø∂é"},
		{"test", 0, 4, ""},
	}
	for _, test := range tests {
		r := NewRope(test.inputStr)
		r.Remove(test.removeStart, test.removeEnd)
		got := r.String()
		if got != test.want {
			t.Errorf("Remove(%d, %d) = %s", test.removeStart, test.removeEnd, got)
		}
	}
}
