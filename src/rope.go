package main

import (
	"math"
)

const (
	// RopeSplitLength is the threshold used to split a leaf node into two child nodes.
	RopeSplitLength = 1000000000
	// RopeJoinLength is the threshold used to join two child nodes into one leaf node.
	RopeJoinLength = 500
	// RopeRebalanceRatio = 1.2
)

// A Rope is a data structure for efficiently manipulating large strings
type Rope struct {
	left     *Rope
	right    *Rope
	value    string
	valueNil bool

	len int
}

// NewRope returns a new rope from a given string
func NewRope(str string) *Rope {
	r := new(Rope)
	r.value = str
	r.valueNil = false
	r.len = Count(r.value)

	r.Adjust()

	return r
}

// Adjust modifies the rope so it is more balanced
func (r *Rope) Adjust() {
	if !r.valueNil {
		if r.len > RopeSplitLength {
			divide := int(math.Floor(float64(r.len) / 2))
			runes := []rune(r.value)
			r.left = NewRope(string(runes[:divide]))
			r.right = NewRope(string(runes[divide:]))
			r.valueNil = true
		}
	} else {
		if r.len < RopeJoinLength {
			r.value = r.left.String() + r.right.String()
			r.valueNil = false
			r.left = nil
			r.right = nil
		}
	}
}

// String returns the string representation of the rope
func (r *Rope) String() string {
	if !r.valueNil {
		return r.value
	}
	return r.left.String() + r.right.String()
}

// Remove deletes a slice of the rope from start the to end (exclusive)
func (r *Rope) Remove(start, end int) {
	if !r.valueNil {
		r.value = string(append([]rune(r.value)[:start], []rune(r.value)[end:]...))
		r.valueNil = false
		r.len = Count(r.value)
	} else {
		leftStart := Min(start, r.left.len)
		leftEnd := Min(end, r.left.len)
		rightStart := Max(0, Min(start-r.left.len, r.right.len))
		rightEnd := Max(0, Min(end-r.left.len, r.right.len))
		if leftStart < r.left.len {
			r.left.Remove(leftStart, leftEnd)
		}
		if rightEnd > 0 {
			r.right.Remove(rightStart, rightEnd)
		}
		r.len = r.left.len + r.right.len
	}

	r.Adjust()
}

// Insert inserts a string into the rope at a specified position
func (r *Rope) Insert(pos int, value string) {
	if !r.valueNil {
		first := append([]rune(r.value)[:pos], []rune(value)...)
		r.value = string(append(first, []rune(r.value)[pos:]...))
		r.valueNil = false
		r.len = Count(r.value)
	} else {
		if pos < r.left.len {
			r.left.Insert(pos, value)
			r.len = r.left.len + r.right.len
		} else {
			r.right.Insert(pos-r.left.len, value)
		}
	}

	r.Adjust()
}
