package main

// FromCharPos converts from a character position to an x, y position
func FromCharPos(loc int, buf *Buffer) Loc {
	charNum := 0
	x, y := 0, 0

	lineLen := Count(buf.Line(y)) + 1
	for charNum+lineLen <= loc {
		charNum += lineLen
		y++
		lineLen = Count(buf.Line(y)) + 1
	}
	x = loc - charNum

	return Loc{x, y}
}

// ToCharPos converts from an x, y position to a character position
func ToCharPos(start Loc, buf *Buffer) int {
	x, y := start.X, start.Y
	loc := 0
	for i := 0; i < y; i++ {
		// + 1 for the newline
		loc += Count(buf.Line(i)) + 1
	}
	loc += x
	return loc
}

// InBounds returns whether the given location is a valid character position in the given buffer
func InBounds(pos Loc, buf *Buffer) bool {
	if pos.Y < 0 || pos.Y >= buf.NumLines || pos.X < 0 || pos.X > Count(buf.Line(pos.Y)) {
		return false
	}

	return true
}

// ByteOffset is just like ToCharPos except it counts bytes instead of runes
func ByteOffset(pos Loc, buf *Buffer) int {
	x, y := pos.X, pos.Y
	loc := 0
	for i := 0; i < y; i++ {
		// + 1 for the newline
		loc += len(buf.Line(i)) + 1
	}
	loc += len(buf.Line(y)[:x])
	return loc
}

// Loc stores a location
type Loc struct {
	X, Y int
}

// Inside returns true if l is inside the square made by p1 & p2
func (l Loc) Inside(p1, p2 Loc) bool {
	if p1.X > p2.X {
		p1.X, p2.X = p2.X, p1.X
	}
	if p1.Y > p2.Y {
		p1.Y, p2.Y = p2.Y, p1.Y
	}
	return l.Y >= p1.Y && l.X >= p1.X && l.Y <= p2.Y && l.X <= p2.X
}

// LessThan returns true if b is smaller
func (l Loc) LessThan(b Loc) bool {
	if l.Y < b.Y {
		return true
	}
	if l.Y == b.Y && l.X < b.X {
		return true
	}
	return false
}

// GreaterThan returns true if b is bigger
func (l Loc) GreaterThan(b Loc) bool {
	if l.Y > b.Y {
		return true
	}
	if l.Y == b.Y && l.X > b.X {
		return true
	}
	return false
}

// GreaterEqual returns true if b is greater than or equal to b
func (l Loc) GreaterEqual(b Loc) bool {
	if l.Y > b.Y {
		return true
	}
	if l.Y == b.Y && l.X > b.X {
		return true
	}
	if l == b {
		return true
	}
	return false
}

// LessEqual returns true if b is less than or equal to b
func (l Loc) LessEqual(b Loc) bool {
	if l.Y < b.Y {
		return true
	}
	if l.Y == b.Y && l.X < b.X {
		return true
	}
	if l == b {
		return true
	}
	return false
}

// This moves the location one character to the right
func (l Loc) right(buf *Buffer) Loc {
	if l == buf.End() {
		return Loc{l.X + 1, l.Y}
	}
	var res Loc
	if l.X < Count(buf.Line(l.Y)) {
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
		res = Loc{Count(buf.Line(l.Y - 1)), l.Y - 1}
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
