package main

import (
	"github.com/gdamore/tcell"
	"strings"
)

type View struct {
	cursor  Cursor
	topline int
	height  int
	width   int

	buf *Buffer
	sl  Statusline

	mouseReleased bool

	s tcell.Screen
}

func newView(buf *Buffer, s tcell.Screen) *View {
	w, h := s.Size()
	return newViewWidthHeight(buf, s, w, h)
}

func newViewWidthHeight(buf *Buffer, s tcell.Screen, w, h int) *View {
	v := new(View)

	v.buf = buf
	v.s = s

	v.topline = 0
	v.height = h - 2
	v.width = w
	v.cursor = Cursor{
		x:   0,
		y:   0,
		loc: 0,
		v:   v,
	}

	v.sl = Statusline{
		v: v,
	}

	return v
}

// Returns an int describing how the screen needs to be redrawn
// 0: Screen does not need to be redrawn
// 1: Only the cursor/statusline needs to be redrawn
// 2: Everything needs to be redrawn
func (v *View) handleEvent(event tcell.Event) int {
	var ret int
	switch e := event.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyUp:
			v.cursor.up()
			ret = 1
		case tcell.KeyDown:
			v.cursor.down()
			ret = 1
		case tcell.KeyLeft:
			v.cursor.left()
			ret = 1
		case tcell.KeyRight:
			v.cursor.right()
			ret = 1
		case tcell.KeyEnter:
			v.buf.insert(v.cursor.loc, "\n")
			v.cursor.right()
			ret = 2
		case tcell.KeySpace:
			v.buf.insert(v.cursor.loc, " ")
			v.cursor.right()
			ret = 2
		case tcell.KeyBackspace2:
			if v.cursor.loc > 0 {
				v.cursor.left()
				v.buf.remove(v.cursor.loc, v.cursor.loc+1)
				ret = 2
			}
		case tcell.KeyTab:
			v.buf.insert(v.cursor.loc, "\t")
			v.cursor.right()
			ret = 2
		case tcell.KeyCtrlS:
			err := v.buf.save()
			if err != nil {
				// Error!
			}
			// Need to redraw the status line
			ret = 1
		case tcell.KeyRune:
			v.buf.insert(v.cursor.loc, string(e.Rune()))
			v.cursor.right()
			ret = 2
		}
	case *tcell.EventMouse:
		x, y := e.Position()
		y += v.topline
		// Position always seems to be off by one
		x--
		y--

		button := e.Buttons()

		switch button {
		case tcell.Button1:
			if y-v.topline > v.height-1 {
				y = v.height + v.topline - 1
			}
			if y > len(v.buf.lines) {
				y = len(v.buf.lines) - 1
			}
			if x > count(v.buf.lines[y]) {
				x = count(v.buf.lines[y])
			}

			x = v.cursor.getCharPos(y, x)
			d := v.cursor.distance(x, y)
			v.cursor.loc += d
			v.cursor.x = x
			v.cursor.y = y

			if v.mouseReleased {
				v.cursor.selectionStart = v.cursor.loc
			}
			v.cursor.selectionEnd = v.cursor.loc
			v.mouseReleased = false
			ret = 2
		case tcell.ButtonNone:
			v.mouseReleased = true
		case tcell.WheelUp:
			if v.topline > 0 {
				v.topline--
				return 2
			} else {
				return 0
			}
		case tcell.WheelDown:
			if v.topline < len(v.buf.lines)-v.height {
				v.topline++
				return 2
			} else {
				return 0
			}
		}
	}

	cy := v.cursor.y
	if cy < v.topline {
		v.topline = cy
		ret = 2
	}
	if cy > v.topline+v.height-1 {
		v.topline = cy - v.height + 1
		ret = 2
	}

	return ret
}

func (v *View) display() {
	var charNum int
	for lineN := 0; lineN < v.height; lineN++ {
		if lineN+v.topline >= len(v.buf.lines) {
			break
		}
		line := strings.Replace(v.buf.lines[lineN+v.topline], "\t", emptyString(tabSize), -1)
		for colN, ch := range line {
			st := tcell.StyleDefault
			if v.cursor.hasSelection() && charNum >= v.cursor.selectionStart && charNum <= v.cursor.selectionEnd {
				st = st.Reverse(true)
			}

			v.s.SetContent(colN, lineN, ch, nil, st)
			charNum++
		}
		charNum++
	}
}
