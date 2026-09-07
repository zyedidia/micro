package main

import (
	"testing"

	"github.com/micro-editor/micro/v2/internal/action"
	"github.com/micro-editor/micro/v2/internal/buffer"
	"github.com/micro-editor/tcell/v2"
)

func TestNestedSplitMouseResize(t *testing.T) {
	newBuffer := func(text string) *buffer.Buffer {
		b := buffer.NewBufferFromString(text, "", buffer.BTDefault)
		b.Settings["backup"] = false
		t.Cleanup(b.Close)
		return b
	}
	tab := action.NewTabFromBuffer(0, 0, 80, 24, newBuffer("first"))
	first := tab.CurPane()
	second := first.VSplitIndex(newBuffer("second"), true)
	third := second.HSplitIndex(newBuffer("third"), true)
	fourth := third.VSplitIndex(newBuffer("fourth"), true)
	activeID := fourth.ID()
	mouse := func(x, y int, button tcell.ButtonMask) {
		tab.HandleEvent(tcell.NewEventMouse(x, y, button, tcell.ModNone, ""))
	}

	// Repeatedly drag the initial divider toward the right screen edge.
	mouse(40, 0, tcell.Button1)
	for x := 41; x < 80; x++ {
		mouse(x, 0, tcell.Button1)
		for _, p := range tab.Panes {
			n := tab.GetNode(p.ID())
			if n == nil || n.W < 1 || n.H < 1 {
				t.Fatalf("mouse drag to %d lost pane %d", x, p.ID())
			}
		}
	}
	mouse(79, 0, tcell.ButtonNone)
	if tab.CurPane().ID() != activeID || len(tab.Panes) != 4 {
		t.Fatal("resize changed focus or pane membership")
	}

	// The nested divider still renders and can be selected with the mouse.
	n := tab.GetNode(third.ID())
	divider := buffer.Loc{X: n.X + n.W, Y: n.Y}
	if tab.GetMouseSplitNode(divider) != n {
		t.Fatal("nested divider is no longer reachable")
	}
	sim.Clear()
	tab.UIWindow.Display()
	if char, _, _, _ := sim.GetContent(divider.X, divider.Y); char != '|' {
		t.Fatalf("nested divider is not visible: %q", char)
	}

	// Drag the parent back: all four panes regain their original width.
	outer := tab.GetNode(first.ID())
	mouse(outer.X+outer.W, 0, tcell.Button1)
	mouse(40, 0, tcell.Button1)
	mouse(40, 0, tcell.ButtonNone)
	if tab.GetNode(third.ID()).W != 20 || tab.GetNode(fourth.ID()).W != 20 {
		t.Fatal("nested panes did not recover their proportions")
	}
	if third.Buf.Line(0) != "third" || fourth.Buf.Line(0) != "fourth" {
		t.Fatal("resize changed buffer contents")
	}
}
