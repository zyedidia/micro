package main

import (
	"io/ioutil"
	"strings"
)

type Buffer struct {
	r *Rope

	// Path to the file on disk
	path string
	// Name of the buffer on the status line
	name string

	// This is the text stored everytime the buffer is saved to check if the buffer is modified
	savedText string

	text  string
	lines []string
}

func newBuffer(txt, path string) *Buffer {
	b := new(Buffer)
	b.r = newRope(txt)
	b.path = path
	b.name = path
	b.savedText = txt

	b.update()

	return b
}

func (b *Buffer) update() {
	b.text = b.r.toString()
	b.lines = strings.Split(b.text, "\n")
}

func (b *Buffer) save() error {
	return b.saveAs(b.path)
}

func (b *Buffer) saveAs(filename string) error {
	err := ioutil.WriteFile(filename, []byte(b.text), 0644)
	return err
}

func (b *Buffer) insert(idx int, value string) {
	b.r.insert(idx, value)
	b.update()
}

func (b *Buffer) remove(start, end int) {
	b.r.remove(start, end)
	b.update()
}

func (b *Buffer) length() int {
	return b.r.len
}
