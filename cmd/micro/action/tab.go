package action

import (
	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/display"
	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/micro/cmd/micro/views"
	"github.com/zyedidia/tcell"
)

type TabList struct {
	*display.TabWindow
	List   []*TabPane
	Active int
}

func NewTabList(bufs []*buffer.Buffer) *TabList {
	w, h := screen.Screen.Size()
	tl := new(TabList)
	tl.List = make([]*TabPane, len(bufs))
	if len(bufs) > 1 {
		for i, b := range bufs {
			tl.List[i] = NewTabPane(0, 1, w, h-2, b)
		}
	} else {
		tl.List[0] = NewTabPane(0, 0, w, h-1, bufs[0])
	}
	tl.TabWindow = display.NewTabWindow(w, 0)
	tl.Names = make([]string, len(bufs))
	tl.UpdateNames()

	return tl
}

func (t *TabList) UpdateNames() {
	t.Names = t.Names[:0]
	for _, p := range t.List {
		t.Names = append(t.Names, p.Panes[p.active].Buf.GetName())
	}
}

func (t *TabList) HandleEvent(event tcell.Event) {
	switch e := event.(type) {
	case *tcell.EventResize:
		w, h := screen.Screen.Size()
		InfoBar.Resize(w, h-1)
		if len(t.List) > 1 {
			for _, p := range t.List {
				p.Node.Resize(w, h-2)
				p.Resize()
			}
		} else {
			t.List[0].Node.Resize(w, h-2)
			t.List[0].Resize()
		}
	case *tcell.EventMouse:
		switch e.Buttons() {
		case tcell.Button1:
		}

	}
	t.List[t.Active].HandleEvent(event)
}

func (t *TabList) Display() {
	if len(t.List) > 1 {
		t.TabWindow.Display()
	}
}

var Tabs *TabList

func InitTabs(bufs []*buffer.Buffer) {
	Tabs = NewTabList(bufs)
}

func MainTab() *TabPane {
	return Tabs.List[Tabs.Active]
}

// A TabPane represents a single tab
// It consists of a list of edit panes (the open buffers),
// a split tree (stored as just the root node), and a uiwindow
// to display the UI elements like the borders between splits
type TabPane struct {
	*views.Node
	*display.UIWindow
	Panes  []*EditPane
	active int

	resizing *views.Node // node currently being resized
}

// HandleEvent takes a tcell event and usually dispatches it to the current
// active pane. However if the event is a resize or a mouse event where the user
// is interacting with the UI (resizing splits) then the event is consumed here
// If the event is a mouse event in a pane, that pane will become active and get
// the event
func (t *TabPane) HandleEvent(event tcell.Event) {
	switch e := event.(type) {
	case *tcell.EventMouse:
		switch e.Buttons() {
		case tcell.Button1:
			mx, my := e.Position()

			resizeID := t.GetMouseSplitID(buffer.Loc{mx, my})
			if t.resizing != nil {
				var size int
				if t.resizing.Kind == views.STVert {
					size = mx - t.resizing.X
				} else {
					size = my - t.resizing.Y + 1
				}
				t.resizing.ResizeSplit(size)
				t.Resize()
				return
			}

			if resizeID != 0 {
				t.resizing = t.GetNode(uint64(resizeID))
				return
			}

			for i, p := range t.Panes {
				v := p.GetView()
				inpane := mx >= v.X && mx < v.X+v.Width && my >= v.Y && my < v.Y+v.Height
				if inpane {
					t.active = i
					p.SetActive(true)
				} else {
					p.SetActive(false)
				}
			}
		case tcell.ButtonNone:
			t.resizing = nil
		}

	}
	t.Panes[t.active].HandleEvent(event)
}

// SetActive changes the currently active pane to the specified index
func (t *TabPane) SetActive(i int) {
	t.active = i
	for j, p := range t.Panes {
		if j == i {
			p.SetActive(true)
		} else {
			p.SetActive(false)
		}
	}
}

// GetPane returns the pane with the given split index
func (t *TabPane) GetPane(splitid uint64) int {
	for i, p := range t.Panes {
		if p.splitID == splitid {
			return i
		}
	}
	return 0
}

// Remove pane removes the pane with the given index
func (t *TabPane) RemovePane(i int) {
	copy(t.Panes[i:], t.Panes[i+1:])
	t.Panes[len(t.Panes)-1] = nil // or the zero value of T
	t.Panes = t.Panes[:len(t.Panes)-1]
}

// Resize resizes all panes according to their corresponding split nodes
func (t *TabPane) Resize() {
	for i, p := range t.Panes {
		n := t.GetNode(p.splitID)
		pv := p.GetView()
		offset := 0
		if i != 0 {
			offset = 1
		}
		pv.X, pv.Y = n.X+offset, n.Y
		p.SetView(pv)
		p.Resize(n.W-offset, n.H)
	}
}

// CurPane returns the currently active pane
func (t *TabPane) CurPane() *EditPane {
	return t.Panes[t.active]
}
