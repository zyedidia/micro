package views

import (
	"fmt"
	"strings"
)

type SplitType uint8

const (
	STVert  = 0
	STHoriz = 1
	STUndef = 2
)

var idcounter uint64

func NewID() uint64 {
	idcounter++
	return idcounter
}

type View struct {
	X, Y int
	W, H int
}

type Node struct {
	View

	kind SplitType

	parent   *Node
	children []*Node

	// Nodes can be marked as non resizable if they shouldn't be rescaled
	// when the terminal window is resized or when a new split is added
	// Only the splits on the edges of the screen can be marked as non resizable
	canResize bool
	// A node may also be marked with proportional scaling. This means that when
	// the window is resized the split maintains its proportions
	propScale bool

	id uint64
}

func (n *Node) ID() uint64 {
	if n.IsLeaf() {
		return n.id
	}
	return 0
}
func (n *Node) CanResize() bool {
	return n.canResize
}
func (n *Node) PropScale() bool {
	return n.propScale
}
func (n *Node) SetResize(b bool) {
	n.canResize = b
}
func (n *Node) SetPropScale(b bool) {
	n.propScale = b
}
func (n *Node) GetView() View {
	return n.View
}
func (n *Node) SetView(v View) {
	n.X, n.Y, n.W, n.H = v.X, v.Y, v.W, v.H
}

func (n *Node) GetNode(id uint64) *Node {
	if n.id == id && n.IsLeaf() {
		return n
	}
	for _, c := range n.children {
		if c.id == id && c.IsLeaf() {
			return c
		}
		gc := c.GetNode(id)
		if gc != nil {
			return gc
		}
	}
	return nil
}

func NewNode(kind SplitType, x, y, w, h int, parent *Node, id uint64) *Node {
	n := new(Node)
	n.kind = kind
	n.canResize = true
	n.propScale = true
	n.X, n.Y, n.W, n.H = x, y, w, h
	n.children = make([]*Node, 0)
	n.parent = parent
	n.id = id

	return n
}

func NewRoot(x, y, w, h int) *Node {
	n1 := NewNode(STUndef, x, y, w, h, nil, NewID())

	return n1
}

func (n *Node) IsLeaf() bool {
	return len(n.children) == 0
}

func (n *Node) vResizeSplit(i int, size int) bool {
	if i < 0 || i >= len(n.children)-1 {
		return false
	}
	v1, v2 := n.children[i].GetView(), n.children[i+1].GetView()
	toth := v1.H + v2.H
	if size >= toth {
		return false
	}
	v1.H, v2.H = size, toth-size
	v2.Y = size
	n.children[i].SetView(v1)
	n.children[i+1].SetView(v2)
	return true
}
func (n *Node) hResizeSplit(i int, size int) bool {
	if i < 0 || i >= len(n.children)-1 {
		return false
	}
	v1, v2 := n.children[i].GetView(), n.children[i+1].GetView()
	totw := v1.W + v2.W
	if size >= totw {
		return false
	}
	v1.W, v2.W = size, totw-size
	v2.X = size
	n.children[i].SetView(v1)
	n.children[i+1].SetView(v2)
	return true
}

func (n *Node) ResizeSplit(size int) bool {
	ind := 0
	for i, c := range n.parent.children {
		if c.id == n.id {
			ind = i
		}
	}
	if n.parent.kind == STVert {
		return n.parent.vResizeSplit(ind, size)
	}
	return n.parent.hResizeSplit(ind, size)
}

func (n *Node) vVSplit(right bool) uint64 {
	ind := 0
	for i, c := range n.parent.children {
		if c.id == n.id {
			ind = i
		}
	}
	return n.parent.hVSplit(ind, right)
}
func (n *Node) hHSplit(bottom bool) uint64 {
	ind := 0
	for i, c := range n.parent.children {
		if c.id == n.id {
			ind = i
		}
	}
	return n.parent.vHSplit(ind, bottom)
}
func (n *Node) vHSplit(i int, right bool) uint64 {
	if n.IsLeaf() {
		newid := NewID()
		hn1 := NewNode(STHoriz, n.X, n.Y, n.W, n.H/2, n, n.id)
		hn2 := NewNode(STHoriz, n.X, n.Y+hn1.H, n.W, n.H/2, n, newid)
		if !right {
			hn1.id, hn2.id = hn2.id, hn1.id
		}

		n.children = append(n.children, hn1, hn2)
		return newid
	} else {
		numr := 0
		numnr := 0
		nonrh := 0
		for _, c := range n.children {
			view := c.GetView()
			if !c.CanResize() {
				nonrh += view.H
				numnr++
			} else {
				numr++
			}
		}

		// if there are no resizable splits make them all resizable
		if numr == 0 {
			numr = numnr
		}

		height := (n.H - nonrh) / (numr + 1)

		newid := NewID()
		hn := NewNode(STHoriz, n.X, 0, n.W, height, n, newid)
		n.children = append(n.children, nil)
		inspos := i
		if right {
			inspos++
		}
		copy(n.children[inspos+1:], n.children[inspos:])
		n.children[inspos] = hn

		y := 0
		for _, c := range n.children {
			view := c.GetView()
			if c.CanResize() {
				view.H = height
				view.Y = y
			} else {
				view.Y = y
			}
			y += view.H
			c.SetView(view)
		}
		return newid
	}
}
func (n *Node) hVSplit(i int, right bool) uint64 {
	if n.IsLeaf() {
		newid := NewID()
		vn1 := NewNode(STVert, n.X, n.Y, n.W/2, n.H, n, n.id)
		vn2 := NewNode(STVert, n.X+vn1.W, n.Y, n.W/2, n.H, n, newid)
		if !right {
			vn1.id, vn2.id = vn2.id, vn1.id
		}

		n.children = append(n.children, vn1, vn2)
		return newid
	} else {
		numr := 0
		numnr := 0
		nonrw := 0
		for _, c := range n.children {
			view := c.GetView()
			if !c.CanResize() {
				nonrw += view.W
				numnr++
			} else {
				numr++
			}
		}

		// if there are no resizable splits make them all resizable
		if numr == 0 {
			numr = numnr
		}

		width := (n.W - nonrw) / (numr + 1)

		newid := NewID()
		vn := NewNode(STVert, 0, n.Y, width, n.H, n, newid)
		n.children = append(n.children, nil)
		inspos := i
		if right {
			inspos++
		}
		copy(n.children[inspos+1:], n.children[inspos:])
		n.children[inspos] = vn

		x := 0
		for _, c := range n.children {
			view := c.GetView()
			if c.CanResize() {
				view.W = width
				view.X = x
			} else {
				view.X = x
			}
			x += view.W
			c.SetView(view)
		}
		return newid
	}
}

func (n *Node) HSplit(bottom bool) uint64 {
	if !n.IsLeaf() {
		return 0
	}
	if n.kind == STUndef {
		n.kind = STVert
	}
	if n.kind == STVert {
		return n.vHSplit(0, bottom)
	}
	return n.hHSplit(bottom)
}

func (n *Node) VSplit(right bool) uint64 {
	if !n.IsLeaf() {
		return 0
	}
	if n.kind == STUndef {
		n.kind = STHoriz
	}
	if n.kind == STVert {
		return n.vVSplit(right)
	}
	return n.hVSplit(0, right)
}

func (n *Node) Resize(w, h int) {
	propW, propH := float64(w)/float64(n.W), float64(h)/float64(n.H)
	x, y := n.X, n.Y
	for _, c := range n.children {
		cW := int(float64(c.W) * propW)
		cH := int(float64(c.H) * propH)
		c.Resize(cW, cH)
		c.X = x
		c.Y = y
		if n.kind == STHoriz {
			x += cW
		} else {
			y += cH
		}
	}
	n.W, n.H = w, h
}

func (n *Node) Unsplit() {

}

func (n *Node) String() string {
	var strf func(n *Node, ident int) string
	strf = func(n *Node, ident int) string {
		marker := "|"
		if n.kind == STHoriz {
			marker = "-"
		}
		str := fmt.Sprint(strings.Repeat("\t", ident), marker, n.View, n.id)
		if n.IsLeaf() {
			str += "🍁"
		}
		str += "\n"
		for _, c := range n.children {
			str += strf(c, ident+1)
		}
		return str
	}
	return strf(n, 0)
}
