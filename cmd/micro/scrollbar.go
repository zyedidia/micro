package main

type ScrollBar struct {
	view *View
}

func (sb *ScrollBar) Display() {
	style := defStyle.Reverse(true)
	screen.SetContent(sb.view.x+sb.view.Width-1, sb.view.y+sb.Pos(), ' ', nil, style)
}

func (sb *ScrollBar) Pos() int {
	numlines := sb.view.Buf.NumLines
	h := sb.view.Height
	filepercent := float32(sb.view.Topline) / float32(numlines)

	return int(filepercent * float32(h))
}
