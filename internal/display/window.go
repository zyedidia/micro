package display

import (
	"github.com/zyedidia/micro/v2/internal/buffer"
)

type View struct {
	X, Y          int // X,Y location of the view
	Width, Height int // Width and height of the view

	// Start line of the view (for vertical scroll)
	StartLine SLoc

	// Start column of the view (for horizontal scroll)
	// note that since the starting column of every line is different if the view
	// is scrolled, StartCol is a visual index (will be the same for every line)
	StartCol int
}

type Window interface {
	Display()
	Clear()
	Relocate() bool
	GetView() *View
	SetView(v *View)
	LocFromVisual(vloc buffer.Loc) buffer.Loc
	Resize(w, h int)
	SetActive(b bool)
	IsActive() bool
}

type BWindow interface {
	Window
	SoftWrap
	SetBuffer(b *buffer.Buffer)
	BufView() View
}
