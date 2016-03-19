package main

import (
	"github.com/gdamore/tcell"
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

func (v *View) scrollUp(n int) {
	// Try to scroll by n but if it would overflow, scroll by 1
	if v.topline-n >= 0 {
		v.topline -= n
	} else if v.topline > 0 {
		v.topline--
	}
}

func (v *View) scrollDown(n int) {
	// Try to scroll by n but if it would overflow, scroll by 1
	if v.topline+n <= len(v.buf.lines)-v.height {
		v.topline += n
	} else if v.topline < len(v.buf.lines)-v.height {
		v.topline++
	}
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
			if v.cursor.hasSelection() {
				v.cursor.deleteSelected()
				v.cursor.resetSelection()
				ret = 2
			} else if v.cursor.loc > 0 {
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
			if v.cursor.hasSelection() {
				v.cursor.deleteSelected()
				v.cursor.resetSelection()
			}
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
				v.scrollDown(1)
				y = v.height + v.topline - 1
			}
			if y > len(v.buf.lines) {
				y = len(v.buf.lines) - 2
			}

			x = v.cursor.getCharPosInLine(y, x)
			if x > count(v.buf.lines[y]) {
				x = count(v.buf.lines[y])
			}
			d := v.cursor.distance(x, y)
			v.cursor.loc += d
			v.cursor.x = x
			v.cursor.y = y

			if v.mouseReleased {
				v.cursor.selectionStart = v.cursor.loc
				v.cursor.selectionStartX = v.cursor.x
				v.cursor.selectionStartY = v.cursor.y
			}
			v.cursor.selectionEnd = v.cursor.loc
			v.mouseReleased = false
			return 2
		case tcell.ButtonNone:
			v.mouseReleased = true
			return 0
		case tcell.WheelUp:
			v.scrollUp(2)
			return 2
		case tcell.WheelDown:
			v.scrollDown(2)
			return 2
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
	charNum := v.cursor.loc + v.cursor.distance(0, v.topline)
	for lineN := 0; lineN < v.height; lineN++ {
		if lineN+v.topline >= len(v.buf.lines) {
			break
		}
		// line := strings.Replace(v.buf.lines[lineN+v.topline], "\t", emptyString(tabSize), -1)
		line := v.buf.lines[lineN+v.topline]
		tabchars := 0
		for colN, ch := range line {
			st := tcell.StyleDefault
			if v.cursor.hasSelection() &&
				(charNum >= v.cursor.selectionStart && charNum <= v.cursor.selectionEnd ||
					charNum <= v.cursor.selectionStart && charNum >= v.cursor.selectionEnd) {
				st = st.Reverse(true)
			}

			if ch == '\t' {
				v.s.SetContent(colN+tabchars, lineN, ' ', nil, st)
				for i := 0; i < tabSize-1; i++ {
					tabchars++
					v.s.SetContent(colN+tabchars, lineN, ' ', nil, st)
				}
			} else {
				v.s.SetContent(colN+tabchars, lineN, ch, nil, st)
			}
			charNum++
		}
		charNum++
	}
}
