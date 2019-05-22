package main

// ScrollBar represents an optional scrollbar that can be used
type ScrollBar struct {
	view *View
}

// Display shows the scrollbar
func (sb *ScrollBar) Display() {
	style := defStyle.Reverse(true)

	height := sb.height()

	// If the height of the scrollbar is bigger than the view, don't render it
	if height >= sb.view.Height {
		return
	}

	x := sb.view.x + sb.view.Width - 1
	y := sb.view.y + sb.pos()

	for i := 0; i < height; i++ {
		screen.SetContent(x, y+i, ' ', nil, style)
	}
}

func (sb *ScrollBar) height() int {
	// The ratio between the height of the file and the height of the view
	ratio := float32(sb.view.Buf.NumLines) / float32(sb.view.Height)

	// Get height of the scrollbar so that the ratio between the view height
	// and the scrollbar height have the same ratio
	return int(float32(sb.view.Height) / ratio)
}

func (sb *ScrollBar) pos() int {
	numlines := sb.view.Buf.NumLines
	h := sb.view.Height
	filepercent := float32(sb.view.Topline) / float32(numlines)

	return int(filepercent * float32(h))
}
