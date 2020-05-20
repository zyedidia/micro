package buffer

import (
	"bufio"
	"bytes"
	"io"
	"sync"

	"github.com/zyedidia/micro/v2/internal/util"
	"github.com/zyedidia/micro/v2/pkg/highlight"
)

// Finds the byte index of the nth rune in a byte slice
func runeToByteIndex(n int, txt []byte) int {
	if n == 0 {
		return 0
	}

	count := 0
	i := 0
	for len(txt) > 0 {
		_, _, size := util.DecodeCharacter(txt)

		txt = txt[size:]
		count += size
		i++

		if i == n {
			break
		}
	}
	return count
}

// A Line contains the data in bytes as well as a highlight state, match
// and a flag for whether the highlighting needs to be updated
type Line struct {
	data []byte

	state       highlight.State
	match       highlight.LineMatch
	rehighlight bool
	lock        sync.Mutex
}

const (
	// Line ending file formats
	FFAuto = 0 // Autodetect format
	FFUnix = 1 // LF line endings (unix style '\n')
	FFDos  = 2 // CRLF line endings (dos style '\r\n')
)

type FileFormat byte

// A LineArray simply stores and array of lines and makes it easy to insert
// and delete in it
type LineArray struct {
	lines    []Line
	Endings  FileFormat
	initsize uint64
}

// Append efficiently appends lines together
// It allocates an additional 10000 lines if the original estimate
// is incorrect
func Append(slice []Line, data ...Line) []Line {
	l := len(slice)
	if l+len(data) > cap(slice) { // reallocate
		newSlice := make([]Line, (l+len(data))+10000)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0 : l+len(data)]
	for i, c := range data {
		slice[l+i] = c
	}
	return slice
}

// NewLineArray returns a new line array from an array of bytes
func NewLineArray(size uint64, endings FileFormat, reader io.Reader) *LineArray {
	la := new(LineArray)

	la.lines = make([]Line, 0, 1000)
	la.initsize = size

	br := bufio.NewReader(reader)
	var loaded int

	n := 0
	for {
		data, err := br.ReadBytes('\n')
		// Detect the line ending by checking to see if there is a '\r' char
		// before the '\n'
		// Even if the file format is set to DOS, the '\r' is removed so
		// that all lines end with '\n'
		dlen := len(data)
		if dlen > 1 && data[dlen-2] == '\r' {
			data = append(data[:dlen-2], '\n')
			if endings == FFAuto {
				la.Endings = FFDos
			}
			dlen = len(data)
		} else if dlen > 0 {
			if endings == FFAuto {
				la.Endings = FFUnix
			}
		}

		// If we are loading a large file (greater than 1000) we use the file
		// size and the length of the first 1000 lines to try to estimate
		// how many lines will need to be allocated for the rest of the file
		// We add an extra 10000 to the original estimate to be safe and give
		// plenty of room for expansion
		if n >= 1000 && loaded >= 0 {
			totalLinesNum := int(float64(size) * (float64(n) / float64(loaded)))
			newSlice := make([]Line, len(la.lines), totalLinesNum+10000)
			copy(newSlice, la.lines)
			la.lines = newSlice
			loaded = -1
		}

		// Counter for the number of bytes in the first 1000 lines
		if loaded >= 0 {
			loaded += dlen
		}

		if err != nil {
			if err == io.EOF {
				la.lines = Append(la.lines, Line{
					data:        data[:],
					state:       nil,
					match:       nil,
					rehighlight: false,
				})
			}
			// Last line was read
			break
		} else {
			la.lines = Append(la.lines, Line{
				data:        data[:dlen-1],
				state:       nil,
				match:       nil,
				rehighlight: false,
			})
		}
		n++
	}

	return la
}

// Bytes returns the string that should be written to disk when
// the line array is saved
func (la *LineArray) Bytes() []byte {
	b := new(bytes.Buffer)
	// initsize should provide a good estimate
	b.Grow(int(la.initsize + 4096))
	for i, l := range la.lines {
		b.Write(l.data)
		if i != len(la.lines)-1 {
			if la.Endings == FFDos {
				b.WriteByte('\r')
			}
			b.WriteByte('\n')
		}
	}
	return b.Bytes()
}

// newlineBelow adds a newline below the given line number
func (la *LineArray) newlineBelow(y int) {
	la.lines = append(la.lines, Line{
		data:        []byte{' '},
		state:       nil,
		match:       nil,
		rehighlight: false,
	})
	copy(la.lines[y+2:], la.lines[y+1:])
	la.lines[y+1] = Line{
		data:        []byte{},
		state:       la.lines[y].state,
		match:       nil,
		rehighlight: false,
	}
}

// Inserts a byte array at a given location
func (la *LineArray) insert(pos Loc, value []byte) {
	x, y := runeToByteIndex(pos.X, la.lines[pos.Y].data), pos.Y
	for i := 0; i < len(value); i++ {
		if value[i] == '\n' {
			la.split(Loc{x, y})
			x = 0
			y++
			continue
		}
		la.insertByte(Loc{x, y}, value[i])
		x++
	}
}

// InsertByte inserts a byte at a given location
func (la *LineArray) insertByte(pos Loc, value byte) {
	la.lines[pos.Y].data = append(la.lines[pos.Y].data, 0)
	copy(la.lines[pos.Y].data[pos.X+1:], la.lines[pos.Y].data[pos.X:])
	la.lines[pos.Y].data[pos.X] = value
}

// joinLines joins the two lines a and b
func (la *LineArray) joinLines(a, b int) {
	la.insert(Loc{len(la.lines[a].data), a}, la.lines[b].data)
	la.deleteLine(b)
}

// split splits a line at a given position
func (la *LineArray) split(pos Loc) {
	la.newlineBelow(pos.Y)
	la.insert(Loc{0, pos.Y + 1}, la.lines[pos.Y].data[pos.X:])
	la.lines[pos.Y+1].state = la.lines[pos.Y].state
	la.lines[pos.Y].state = nil
	la.lines[pos.Y].match = nil
	la.lines[pos.Y+1].match = nil
	la.lines[pos.Y].rehighlight = true
	la.deleteToEnd(Loc{pos.X, pos.Y})
}

// removes from start to end
func (la *LineArray) remove(start, end Loc) []byte {
	sub := la.Substr(start, end)
	startX := runeToByteIndex(start.X, la.lines[start.Y].data)
	endX := runeToByteIndex(end.X, la.lines[end.Y].data)
	if start.Y == end.Y {
		la.lines[start.Y].data = append(la.lines[start.Y].data[:startX], la.lines[start.Y].data[endX:]...)
	} else {
		la.deleteLines(start.Y+1, end.Y-1)
		la.deleteToEnd(Loc{startX, start.Y})
		la.deleteFromStart(Loc{endX - 1, start.Y + 1})
		la.joinLines(start.Y, start.Y+1)
	}
	return sub
}

// deleteToEnd deletes from the end of a line to the position
func (la *LineArray) deleteToEnd(pos Loc) {
	la.lines[pos.Y].data = la.lines[pos.Y].data[:pos.X]
}

// deleteFromStart deletes from the start of a line to the position
func (la *LineArray) deleteFromStart(pos Loc) {
	la.lines[pos.Y].data = la.lines[pos.Y].data[pos.X+1:]
}

// deleteLine deletes the line number
func (la *LineArray) deleteLine(y int) {
	la.lines = la.lines[:y+copy(la.lines[y:], la.lines[y+1:])]
}

func (la *LineArray) deleteLines(y1, y2 int) {
	la.lines = la.lines[:y1+copy(la.lines[y1:], la.lines[y2+1:])]
}

// DeleteByte deletes the byte at a position
func (la *LineArray) deleteByte(pos Loc) {
	la.lines[pos.Y].data = la.lines[pos.Y].data[:pos.X+copy(la.lines[pos.Y].data[pos.X:], la.lines[pos.Y].data[pos.X+1:])]
}

// Substr returns the string representation between two locations
func (la *LineArray) Substr(start, end Loc) []byte {
	startX := runeToByteIndex(start.X, la.lines[start.Y].data)
	endX := runeToByteIndex(end.X, la.lines[end.Y].data)
	if start.Y == end.Y {
		src := la.lines[start.Y].data[startX:endX]
		dest := make([]byte, len(src))
		copy(dest, src)
		return dest
	}
	str := make([]byte, 0, len(la.lines[start.Y+1].data)*(end.Y-start.Y))
	str = append(str, la.lines[start.Y].data[startX:]...)
	str = append(str, '\n')
	for i := start.Y + 1; i <= end.Y-1; i++ {
		str = append(str, la.lines[i].data...)
		str = append(str, '\n')
	}
	str = append(str, la.lines[end.Y].data[:endX]...)
	return str
}

// LinesNum returns the number of lines in the buffer
func (la *LineArray) LinesNum() int {
	return len(la.lines)
}

// Start returns the start of the buffer
func (la *LineArray) Start() Loc {
	return Loc{0, 0}
}

// End returns the location of the last character in the buffer
func (la *LineArray) End() Loc {
	numlines := len(la.lines)
	return Loc{util.CharacterCount(la.lines[numlines-1].data), numlines - 1}
}

// LineBytes returns line n as an array of bytes
func (la *LineArray) LineBytes(n int) []byte {
	if n >= len(la.lines) || n < 0 {
		return []byte{}
	}
	return la.lines[n].data
}

// State gets the highlight state for the given line number
func (la *LineArray) State(lineN int) highlight.State {
	la.lines[lineN].lock.Lock()
	defer la.lines[lineN].lock.Unlock()
	return la.lines[lineN].state
}

// SetState sets the highlight state at the given line number
func (la *LineArray) SetState(lineN int, s highlight.State) {
	la.lines[lineN].lock.Lock()
	defer la.lines[lineN].lock.Unlock()
	la.lines[lineN].state = s
}

// SetMatch sets the match at the given line number
func (la *LineArray) SetMatch(lineN int, m highlight.LineMatch) {
	la.lines[lineN].lock.Lock()
	defer la.lines[lineN].lock.Unlock()
	la.lines[lineN].match = m
}

// Match retrieves the match for the given line number
func (la *LineArray) Match(lineN int) highlight.LineMatch {
	la.lines[lineN].lock.Lock()
	defer la.lines[lineN].lock.Unlock()
	return la.lines[lineN].match
}

func (la *LineArray) Rehighlight(lineN int) bool {
	la.lines[lineN].lock.Lock()
	defer la.lines[lineN].lock.Unlock()
	return la.lines[lineN].rehighlight
}

func (la *LineArray) SetRehighlight(lineN int, on bool) {
	la.lines[lineN].lock.Lock()
	defer la.lines[lineN].lock.Unlock()
	la.lines[lineN].rehighlight = on
}
