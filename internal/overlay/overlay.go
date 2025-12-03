package overlay

import (
	"github.com/mattn/go-runewidth"
	"github.com/micro-editor/tcell/v2"
	"github.com/zyedidia/micro/v2/internal/action"
	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/display"
	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/micro/v2/internal/util"
)

type OverlayHandle int
type OverlayFunction func()

type Rect struct {
	X, Y, W, H int
}

var overlay_handle = OverlayHandle(0)
var overlays = make(map[OverlayHandle]OverlayFunction)

func DisplayOverlays() {
	// Should an OverlayFunction create or destroy an overlay, that would modify
	// the overlays map while we are iterating through it.
	// For this reason, we copy the overlays map into temp_overlays.

	temp_overlays := make(map[OverlayHandle]OverlayFunction, len(overlays))

	for h, o := range overlays {
		temp_overlays[h] = o
	}

	for _, draw_fn := range temp_overlays {
		draw_fn()
	}
}

// CreateOverlay creates and registers a new overlay, and returns
// the OverlayHandle associated with it.
func CreateOverlay(draw OverlayFunction) OverlayHandle {
	overlay_handle++
	overlays[overlay_handle] = draw
	return overlay_handle
}

// DestroyOverlay destroys/deregisters an existing overlay via its handle.
func DestroyOverlay(overlay OverlayHandle) {
	delete(overlays, overlay)
}

// DrawRect draws a flat styled rectangle to the provided screen coordinates.
func DrawRect(x, y, w, h int, style tcell.Style) {
	for yy := 0; yy < h; yy++ {
		for xx := 0; xx < w; xx++ {
			screen.SetContent(x+xx, y+yy, ' ', nil, style)
		}
	}
}

// DrawText draws styled clipped text to the provided screen coordinates.
func DrawText(text string, x, y, w, h int, style tcell.Style) {
	DrawRect(x, y, w, h, style)

	tabsize := util.IntOpt(config.GlobalSettings["tabsize"])
	text_bytes := []byte(text)
	xx := 0
	yy := 0

	for len(text_bytes) > 0 {
		r, combc, size := util.DecodeCharacter(text_bytes)
		text_bytes = text_bytes[size:]
		width := 0

		switch r {
		case '\t':
			width = tabsize - (xx % tabsize)
		case '\n':
			xx = 0
			yy++
			continue
		default:
			width = runewidth.RuneWidth(r)
		}

		if yy > h {
			break
		}

		if xx+width <= w {
			screen.SetContent(x+xx, y+yy, r, combc, style)
		}

		xx += width
	}
}

// BufPaneScreenRect returns the bounds of a BufPane in screen coordinates.
func BufPaneScreenRect(bp *action.BufPane) Rect {
	// NOTE: This function is a very thin wrapper around bp.GetView(). As such,
	//       it is maybe a candidate for removal?
	v := bp.GetView()
	return Rect{
		X: v.X,
		Y: v.Y,
		W: v.Width,
		H: v.Height,
	}
}

// BufPaneScreenLoc converts a Loc in the buffer displayed in
// a bufpane to screen coordinates.
func BufPaneScreenLoc(bp *action.BufPane, loc buffer.Loc) buffer.Loc {
	gutter := 0
	bw, ok := bp.BWindow.(*display.BufWindow)
	if ok {
		gutter = bw.GutterOffset()
	}

	v := bp.GetView()
	vloc := bp.VLocFromLoc(loc)
	top := v.StartLine
	yoff := bp.Diff(top, vloc.SLoc)

	return buffer.Loc{
		X: v.X + gutter + vloc.VisualX,
		Y: v.Y + yoff,
	}
}
