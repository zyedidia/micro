package main

import (
	"io/ioutil"
	"strings"
)

// Buffer stores the text for files that are loaded into the text editor
// It uses a rope to efficiently store the string and contains some
// simple functions for saving and wrapper functions for modifying the rope
type Buffer struct {
	// Stores the text of the buffer
	r *Rope

	// Path to the file on disk
	path string
	// Name of the buffer on the status line
	name string

	// This is the text stored every time the buffer is saved to check if the buffer is modified
	savedText string

	// Provide efficient and easy access to text and lines so the rope String does not
	// need to be constantly recalculated
	// These variables are updated in the update() function
	text  string
	lines []string

	// Syntax highlighting rules
	rules string
	// File type of the buffer
	filetype string
}

// NewBuffer creates a new buffer from `txt` with path and name `path`
func NewBuffer(txt, path string) *Buffer {
	b := new(Buffer)
	b.r = NewRope(txt)
	b.path = path
	b.name = path
	b.savedText = txt

	b.Update()

	b.rules, b.filetype = GetRules(b)

	return b
}

// Update fetches the string from the rope and updates the `text` and `lines` in the buffer
func (b *Buffer) Update() {
	b.text = b.r.String()
	b.lines = strings.Split(b.text, "\n")
}

// Save saves the buffer to its default path
func (b *Buffer) Save() error {
	return b.SaveAs(b.path)
}

// SaveAs saves the buffer to a specified path (filename), creating the file if it does not exist
func (b *Buffer) SaveAs(filename string) error {
	err := ioutil.WriteFile(filename, []byte(b.text), 0644)
	if err == nil {
		b.savedText = b.text
	}
	return err
}

// Insert a string into the rope
func (b *Buffer) Insert(idx int, value string) {
	b.r.Insert(idx, value)
	b.Update()
}

// Remove a slice of the rope from start to end (exclusive)
// Returns the string that was removed
func (b *Buffer) Remove(start, end int) string {
	removed := b.text[start:end]
	b.r.Remove(start, end)
	b.Update()
	return removed
}

// Len gives the length of the buffer
func (b *Buffer) Len() int {
	return b.r.len
}
