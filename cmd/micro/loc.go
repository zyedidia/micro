package main

import "unicode/utf8"

// Loc stores a location
type Loc struct {
	X, Y int
}

// LessThan returns true if b is smaller
func (l Loc) LessThan(b Loc) bool {
	if l.Y < b.Y {
		return true
	}
	return l.Y == b.Y && l.X < b.X
}

// GreaterThan returns true if b is bigger
func (l Loc) GreaterThan(b Loc) bool {
	if l.Y > b.Y {
		return true
	}
	return l.Y == b.Y && l.X > b.X
}

// GreaterEqual returns true if b is greater than or equal to b
func (l Loc) GreaterEqual(b Loc) bool {
	if l.Y > b.Y {
		return true
	}
	if l.Y == b.Y && l.X > b.X {
		return true
	}
	return l == b
}

// LessEqual returns true if b is less than or equal to b
func (l Loc) LessEqual(b Loc) bool {
	if l.Y < b.Y {
		return true
	}
	if l.Y == b.Y && l.X < b.X {
		return true
	}
	return l == b
}

// The following functions require a buffer to know where newlines are

// Diff returns the distance between two locations
func Diff(a, b Loc, buf *Buffer) int {
	if a.Y == b.Y {
		if a.X > b.X {
			return a.X - b.X
		}
		return b.X - a.X
	}

	// Make sure a is guaranteed to be less than b
	if b.LessThan(a) {
		a, b = b, a
	}

	loc := 0
	for i := a.Y + 1; i < b.Y; i++ {
		// + 1 for the newline
		loc += utf8.RuneCount(buf.LineBytes(i)) + 1
	}
	loc += utf8.RuneCount(buf.LineBytes(a.Y)) - a.X + b.X + 1
	return loc
}

// This moves the location one character to the right
func (l Loc) right(buf *Buffer) Loc {
	if l == buf.End() {
		return Loc{l.X + 1, l.Y}
	}
	var res Loc
	if l.X < utf8.RuneCount(buf.LineBytes(l.Y)) {
		res = Loc{l.X + 1, l.Y}
	} else {
		res = Loc{0, l.Y + 1}
	}
	return res
}

// This moves the given location one character to the left
func (l Loc) left(buf *Buffer) Loc {
	if l == buf.Start() {
		return Loc{l.X - 1, l.Y}
	}
	var res Loc
	if l.X > 0 {
		res = Loc{l.X - 1, l.Y}
	} else {
		res = Loc{utf8.RuneCount(buf.LineBytes(l.Y - 1)), l.Y - 1}
	}
	return res
}

// Move moves the cursor n characters to the left or right
// It moves the cursor left if n is negative
func (l Loc) Move(n int, buf *Buffer) Loc {
	if n > 0 {
		for i := 0; i < n; i++ {
			l = l.right(buf)
		}
		return l
	}
	for i := 0; i < Abs(n); i++ {
		l = l.left(buf)
	}
	return l
}
