package views

import (
	"fmt"
	"log"
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

	Kind SplitType

	parent   *Node
	children []*Node

	// Nodes can be marked as non resizable if they shouldn't be rescaled
	// when the terminal window is resized or when a new split is added
	// Only the splits on the edges of the screen can be marked as non resizable
	canResize bool
	// A node may also be marked with proportional scaling. This means that when
	// the window is resized the split maintains its proportions
	propScale bool

	propW, propH float64
	id           uint64
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
func (n *Node) Children() []*Node {
	return n.children
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

func NewNode(Kind SplitType, x, y, w, h int, parent *Node, id uint64) *Node {
	n := new(Node)
	n.Kind = Kind
	n.canResize = true
	n.propScale = true
	n.X, n.Y, n.W, n.H = x, y, w, h
	n.children = make([]*Node, 0)
	n.parent = parent
	n.id = id
	if parent != nil {
		n.propW, n.propH = float64(w)/float64(parent.W), float64(h)/float64(parent.H)
	} else {
		n.propW, n.propH = 1, 1
	}

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
	c1, c2 := n.children[i], n.children[i+1]
	toth := c1.H + c2.H
	if size >= toth {
		return false
	}
	c2.Y = size
	c1.Resize(c1.W, size)
	c2.Resize(c2.W, toth-size)
	n.propW = float64(size) / float64(n.parent.W)
	return true
}
func (n *Node) hResizeSplit(i int, size int) bool {
	if i < 0 || i >= len(n.children)-1 {
		return false
	}
	c1, c2 := n.children[i], n.children[i+1]
	totw := c1.W + c2.W
	if size >= totw {
		return false
	}
	c2.X = size
	c1.Resize(size, c1.H)
	c2.Resize(totw-size, c2.H)
	n.propH = float64(size) / float64(n.parent.H)
	return true
}

func (n *Node) ResizeSplit(size int) bool {
	ind := 0
	for i, c := range n.parent.children {
		if c.id == n.id {
			ind = i
		}
	}
	if n.parent.Kind == STVert {
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
		n.alignSize()
		return newid
	} else {
		numr := 0
		numnr := 0
		nonrh := 0
		for _, c := range n.children {
			if !c.CanResize() {
				nonrh += c.H
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

		y := n.Y
		for _, c := range n.children {
			c.Y = y
			if c.CanResize() {
				c.Resize(c.W, height)
			}
			y += c.H
		}
		n.alignSize()
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
		n.alignSize()
		return newid
	} else {
		numr := 0
		numnr := 0
		nonrw := 0
		for _, c := range n.children {
			if !c.CanResize() {
				nonrw += c.W
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

		x := n.X
		for _, c := range n.children {
			c.X = x
			if c.CanResize() {
				c.Resize(width, c.H)
			}
			x += c.W
		}
		n.alignSize()
		return newid
	}
}

func (n *Node) HSplit(bottom bool) uint64 {
	if !n.IsLeaf() {
		return 0
	}
	if n.Kind == STUndef {
		n.Kind = STVert
	}
	if n.Kind == STVert {
		return n.vHSplit(0, bottom)
	}
	return n.hHSplit(bottom)
}

func (n *Node) VSplit(right bool) uint64 {
	if !n.IsLeaf() {
		return 0
	}
	if n.Kind == STUndef {
		n.Kind = STHoriz
	}
	if n.Kind == STVert {
		return n.vVSplit(right)
	}
	return n.hVSplit(0, right)
}

func (n *Node) Resize(w, h int) {
	if n.IsLeaf() {
		n.W, n.H = w, h
	} else {
		x, y := n.X, n.Y
		for i, c := range n.children {
			cW := int(float64(w) * c.propW)
			if c.IsLeaf() && i != len(n.children)-1 {
				cW++
			}
			cH := int(float64(h) * c.propH)
			log.Println(c.id, c.propW, c.propH, cW, cH, w, h)
			c.Resize(cW, cH)
			c.X = x
			c.Y = y
			if n.Kind == STHoriz {
				x += cW
			} else {
				y += cH
			}
		}
		n.alignSize()
		n.W, n.H = w, h
	}
}

func (n *Node) alignSize() {
	if len(n.children) == 0 {
		return
	}

	totw, toth := 0, 0
	for i, c := range n.children {
		if n.Kind == STHoriz {
			if i != len(n.children)-1 {
				c.Resize(c.W-1, c.H)
			}
			totw += c.W
		} else {
			toth += c.H
		}
	}
	if n.Kind == STVert && toth != n.H {
		last := n.children[len(n.children)-1]
		last.Resize(last.W, last.H+n.H-toth)
	} else if n.Kind == STHoriz && totw != n.W {
		last := n.children[len(n.children)-1]
		last.Resize(last.W+n.W-totw, last.H)
	}
}

func (n *Node) Unsplit() {

}

func (n *Node) String() string {
	var strf func(n *Node, ident int) string
	strf = func(n *Node, ident int) string {
		marker := "|"
		if n.Kind == STHoriz {
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
