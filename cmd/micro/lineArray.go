package main

import (
	"bufio"
	"io"
	"unicode/utf8"

	"github.com/zyedidia/micro/cmd/micro/highlight"
)

func runeToByteIndex(n int, txt []byte) int {
	if n == 0 {
		return 0
	}

	count := 0
	i := 0
	for len(txt) > 0 {
		_, size := utf8.DecodeRune(txt)

		txt = txt[size:]
		count += size
		i++

		if i == n {
			break
		}
	}
	return count
}

type Line struct {
	data []byte

	state       highlight.State
	match       highlight.LineMatch
	rehighlight bool
}

// A LineArray simply stores and array of lines and makes it easy to insert
// and delete in it
type LineArray struct {
	lines []Line
}

// NewLineArray returns a new line array from an array of bytes
func NewLineArray(reader io.Reader) *LineArray {
	la := new(LineArray)
	br := bufio.NewReader(reader)

	i := 0
	for {
		data, err := br.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				la.lines = append(la.lines, Line{data[:len(data)], nil, nil, false})
			}
			// Last line was read
			break
		} else {
			la.lines = append(la.lines, Line{data[:len(data)-1], nil, nil, false})
		}
		i++
	}

	return la
}

// Returns the String representation of the LineArray
func (la *LineArray) String() string {
	str := ""
	for i, l := range la.lines {
		str += string(l.data)
		if i != len(la.lines)-1 {
			str += "\n"
		}
	}
	return str
}

// NewlineBelow adds a newline below the given line number
func (la *LineArray) NewlineBelow(y int) {
	la.lines = append(la.lines, Line{[]byte(" "), nil, nil, false})
	copy(la.lines[y+2:], la.lines[y+1:])
	la.lines[y+1] = Line{[]byte(""), la.lines[y].state, nil, false}
}

// inserts a byte array at a given location
func (la *LineArray) insert(pos Loc, value []byte) {
	x, y := runeToByteIndex(pos.X, la.lines[pos.Y].data), pos.Y
	// x, y := pos.x, pos.y
	for i := 0; i < len(value); i++ {
		if value[i] == '\n' {
			la.Split(Loc{x, y})
			x = 0
			y++
			continue
		}
		la.insertByte(Loc{x, y}, value[i])
		x++
	}
}

// inserts a byte at a given location
func (la *LineArray) insertByte(pos Loc, value byte) {
	la.lines[pos.Y].data = append(la.lines[pos.Y].data, 0)
	copy(la.lines[pos.Y].data[pos.X+1:], la.lines[pos.Y].data[pos.X:])
	la.lines[pos.Y].data[pos.X] = value
}

// JoinLines joins the two lines a and b
func (la *LineArray) JoinLines(a, b int) {
	la.insert(Loc{len(la.lines[a].data), a}, la.lines[b].data)
	la.DeleteLine(b)
}

// Split splits a line at a given position
func (la *LineArray) Split(pos Loc) {
	la.NewlineBelow(pos.Y)
	la.insert(Loc{0, pos.Y + 1}, la.lines[pos.Y].data[pos.X:])
	la.lines[pos.Y+1].state = la.lines[pos.Y].state
	la.lines[pos.Y].state = nil
	la.lines[pos.Y].match = nil
	la.lines[pos.Y+1].match = nil
	la.lines[pos.Y].rehighlight = true
	la.DeleteToEnd(Loc{pos.X, pos.Y})
}

// removes from start to end
func (la *LineArray) remove(start, end Loc) string {
	sub := la.Substr(start, end)
	startX := runeToByteIndex(start.X, la.lines[start.Y].data)
	endX := runeToByteIndex(end.X, la.lines[end.Y].data)
	if start.Y == end.Y {
		la.lines[start.Y].data = append(la.lines[start.Y].data[:startX], la.lines[start.Y].data[endX:]...)
	} else {
		for i := start.Y + 1; i <= end.Y-1; i++ {
			la.DeleteLine(start.Y + 1)
		}
		la.DeleteToEnd(Loc{startX, start.Y})
		la.DeleteFromStart(Loc{endX - 1, start.Y + 1})
		la.JoinLines(start.Y, start.Y+1)
	}
	return sub
}

// DeleteToEnd deletes from the end of a line to the position
func (la *LineArray) DeleteToEnd(pos Loc) {
	la.lines[pos.Y].data = la.lines[pos.Y].data[:pos.X]
}

// DeleteFromStart deletes from the start of a line to the position
func (la *LineArray) DeleteFromStart(pos Loc) {
	la.lines[pos.Y].data = la.lines[pos.Y].data[pos.X+1:]
}

// DeleteLine deletes the line number
func (la *LineArray) DeleteLine(y int) {
	la.lines = la.lines[:y+copy(la.lines[y:], la.lines[y+1:])]
}

// DeleteByte deletes the byte at a position
func (la *LineArray) DeleteByte(pos Loc) {
	la.lines[pos.Y].data = la.lines[pos.Y].data[:pos.X+copy(la.lines[pos.Y].data[pos.X:], la.lines[pos.Y].data[pos.X+1:])]
}

// Substr returns the string representation between two locations
func (la *LineArray) Substr(start, end Loc) string {
	startX := runeToByteIndex(start.X, la.lines[start.Y].data)
	endX := runeToByteIndex(end.X, la.lines[end.Y].data)
	if start.Y == end.Y {
		return string(la.lines[start.Y].data[startX:endX])
	}
	var str string
	str += string(la.lines[start.Y].data[startX:]) + "\n"
	for i := start.Y + 1; i <= end.Y-1; i++ {
		str += string(la.lines[i].data) + "\n"
	}
	str += string(la.lines[end.Y].data[:endX])
	return str
}

func (la *LineArray) LineData() [][]byte {
	lines := make([][]byte, len(la.lines))
	for i, l := range la.lines {
		lines[i] = l.data
	}
	return lines
}

func (la *LineArray) State(lineN int) highlight.State {
	return la.lines[lineN].state
}

func (la *LineArray) SetState(lineN int, s highlight.State) {
	la.lines[lineN].state = s
}

func (la *LineArray) SetMatch(lineN int, m highlight.LineMatch) {
	la.lines[lineN].match = m
}

func (la *LineArray) Match(lineN int) highlight.LineMatch {
	return la.lines[lineN].match
}
