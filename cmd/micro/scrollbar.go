package main

// ScrollBar represents an optional scrollbar that can be used
type ScrollBar struct {
	view      *View
	Scrolling bool
}

// Display shows the scrollbar
func (sb *ScrollBar) Display() {
	style := defStyle.Reverse(true)
	screen.SetContent(sb.view.x+sb.view.Width-1, sb.view.y+sb.pos(), ' ', nil, style)
}

func (sb *ScrollBar) HandleMousePress(x, y int) bool {
	v := sb.view

	if (x == v.x+v.Width-1 && v.mouseReleased) || sb.Scrolling {
		numlines := v.Buf.NumLines
		h := v.Height
		filepercent := float32(y) / float32(h)
		v.Topline = int(filepercent * float32(numlines))

		sb.Scrolling = true

		v.mouseReleased = false
		return true
	}
	return false
}

func (sb *ScrollBar) pos() int {
	numlines := sb.view.Buf.NumLines
	h := sb.view.Height
	filepercent := float32(sb.view.Topline) / float32(numlines)

	return int(filepercent * float32(h))
}
