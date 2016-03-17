package main

import (
	// "fmt"
	"math"
	"unicode/utf8"
)

const (
	ropeSplitLength    = 1000
	ropeJoinLength     = 500
	ropeRebalanceRatio = 1.2
)

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type Rope struct {
	left     *Rope
	right    *Rope
	value    string
	valueNil bool

	len int
}

func newRope(str string) *Rope {
	r := new(Rope)
	r.value = str
	r.valueNil = false
	r.len = utf8.RuneCountInString(r.value)

	r.adjust()

	return r
}

func (r *Rope) adjust() {
	if !r.valueNil {
		if r.len > ropeSplitLength {
			divide := int(math.Floor(float64(r.len) / 2))
			r.left = newRope(r.value[:divide])
			r.right = newRope(r.value[divide:])
			r.valueNil = true
		}
	} else {
		if r.len < ropeJoinLength {
			r.value = r.left.toString() + r.right.toString()
			r.valueNil = false
			r.left = nil
			r.right = nil
		}
	}
}

func (r *Rope) toString() string {
	if !r.valueNil {
		return r.value
	}
	return r.left.toString() + r.right.toString()
}

func (r *Rope) remove(start, end int) {
	if !r.valueNil {
		r.value = string(append([]rune(r.value)[:start], []rune(r.value)[end:]...))
		r.valueNil = false
		r.len = utf8.RuneCountInString(r.value)
	} else {
		leftStart := min(start, r.left.len)
		leftEnd := min(end, r.left.len)
		rightStart := max(0, min(start-r.left.len, r.right.len))
		rightEnd := max(0, min(end-r.left.len, r.right.len))
		if leftStart < r.left.len {
			r.left.remove(leftStart, leftEnd)
		}
		if rightEnd > 0 {
			r.right.remove(rightStart, rightEnd)
		}
		r.len = r.left.len + r.right.len
	}

	r.adjust()
}

func (r *Rope) insert(pos int, value string) {
	if !r.valueNil {
		first := append([]rune(r.value)[:pos], []rune(value)...)
		r.value = string(append(first, []rune(r.value)[pos:]...))
		r.valueNil = false
		r.len = utf8.RuneCountInString(r.value)
	} else {
		if pos < r.left.len {
			r.left.insert(pos, value)
			r.len = r.left.len + r.right.len
		} else {
			r.right.insert(pos-r.left.len, value)
		}
	}

	r.adjust()
}
