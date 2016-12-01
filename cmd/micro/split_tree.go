package main

// SpltType specifies whether a split is horizontal or vertical
type SplitType bool

const (
	// VerticalSplit type
	VerticalSplit = false
	// HorizontalSplit type
	HorizontalSplit = true
)

// A Node on the split tree
type Node interface {
	VSplit(buf *Buffer, splitIndex int)
	HSplit(buf *Buffer, splitIndex int)
	String() string
}

// A LeafNode is an actual split so it contains a view
type LeafNode struct {
	view *View

	parent *SplitTree
}

// NewLeafNode returns a new leaf node containing the given view
func NewLeafNode(v *View, parent *SplitTree) *LeafNode {
	n := new(LeafNode)
	n.view = v
	n.view.splitNode = n
	n.parent = parent
	return n
}

// A SplitTree is a Node itself and it contains other nodes
type SplitTree struct {
	kind SplitType

	parent   *SplitTree
	children []Node

	x int
	y int

	width      int
	height     int
	lockWidth  bool
	lockHeight bool

	tabNum int
}

// VSplit creates a vertical split
func (l *LeafNode) VSplit(buf *Buffer, splitIndex int) {
	if splitIndex < 0 {
		splitIndex = 0
	}

	tab := tabs[l.parent.tabNum]
	if l.parent.kind == VerticalSplit {
		if splitIndex > len(l.parent.children) {
			splitIndex = len(l.parent.children)
		}

		newView := NewView(buf)
		newView.TabNum = l.parent.tabNum

		l.parent.children = append(l.parent.children, nil)
		copy(l.parent.children[splitIndex+1:], l.parent.children[splitIndex:])
		l.parent.children[splitIndex] = NewLeafNode(newView, l.parent)

		tab.views = append(tab.views, nil)
		copy(tab.views[splitIndex+1:], tab.views[splitIndex:])
		tab.views[splitIndex] = newView

		tab.CurView = splitIndex
	} else {
		if splitIndex > 1 {
			splitIndex = 1
		}

		s := new(SplitTree)
		s.kind = VerticalSplit
		s.parent = l.parent
		s.tabNum = l.parent.tabNum
		newView := NewView(buf)
		newView.TabNum = l.parent.tabNum
		if splitIndex == 1 {
			s.children = []Node{l, NewLeafNode(newView, s)}
		} else {
			s.children = []Node{NewLeafNode(newView, s), l}
		}
		l.parent.children[search(l.parent.children, l)] = s
		l.parent = s

		tab.views = append(tab.views, nil)
		copy(tab.views[splitIndex+1:], tab.views[splitIndex:])
		tab.views[splitIndex] = newView

		tab.CurView = splitIndex
	}

	tab.Resize()
}

// HSplit creates a horizontal split
func (l *LeafNode) HSplit(buf *Buffer, splitIndex int) {
	if splitIndex < 0 {
		splitIndex = 0
	}

	tab := tabs[l.parent.tabNum]
	if l.parent.kind == HorizontalSplit {
		if splitIndex > len(l.parent.children) {
			splitIndex = len(l.parent.children)
		}

		newView := NewView(buf)
		newView.TabNum = l.parent.tabNum

		l.parent.children = append(l.parent.children, nil)
		copy(l.parent.children[splitIndex+1:], l.parent.children[splitIndex:])
		l.parent.children[splitIndex] = NewLeafNode(newView, l.parent)

		tab.views = append(tab.views, nil)
		copy(tab.views[splitIndex+1:], tab.views[splitIndex:])
		tab.views[splitIndex] = newView

		tab.CurView = splitIndex
	} else {
		if splitIndex > 1 {
			splitIndex = 1
		}

		s := new(SplitTree)
		s.kind = HorizontalSplit
		s.tabNum = l.parent.tabNum
		s.parent = l.parent
		newView := NewView(buf)
		newView.TabNum = l.parent.tabNum
		newView.Num = len(tab.views)
		if splitIndex == 1 {
			s.children = []Node{l, NewLeafNode(newView, s)}
		} else {
			s.children = []Node{NewLeafNode(newView, s), l}
		}
		l.parent.children[search(l.parent.children, l)] = s
		l.parent = s

		tab.views = append(tab.views, nil)
		copy(tab.views[splitIndex+1:], tab.views[splitIndex:])
		tab.views[splitIndex] = newView

		tab.CurView = splitIndex
	}

	tab.Resize()
}

// Delete deletes a split
func (l *LeafNode) Delete() {
	i := search(l.parent.children, l)

	copy(l.parent.children[i:], l.parent.children[i+1:])
	l.parent.children[len(l.parent.children)-1] = nil
	l.parent.children = l.parent.children[:len(l.parent.children)-1]

	tab := tabs[l.parent.tabNum]
	j := findView(tab.views, l.view)
	copy(tab.views[j:], tab.views[j+1:])
	tab.views[len(tab.views)-1] = nil // or the zero value of T
	tab.views = tab.views[:len(tab.views)-1]

	for i, v := range tab.views {
		v.Num = i
	}
	if tab.CurView > 0 {
		tab.CurView--
	}
}

// Cleanup rearranges all the parents after a split has been deleted
func (s *SplitTree) Cleanup() {
	for i, node := range s.children {
		if n, ok := node.(*SplitTree); ok {
			if len(n.children) == 1 {
				if child, ok := n.children[0].(*LeafNode); ok {
					s.children[i] = child
					child.parent = s
					continue
				}
			}
			n.Cleanup()
		}
	}
}

// ResizeSplits resizes all the splits correctly
func (s *SplitTree) ResizeSplits() {
	lockedWidth := 0
	lockedHeight := 0
	lockedChildren := 0
	for _, node := range s.children {
		if n, ok := node.(*LeafNode); ok {
			if s.kind == VerticalSplit {
				if n.view.LockWidth {
					lockedWidth += n.view.Width
					lockedChildren++
				}
			} else {
				if n.view.LockHeight {
					lockedHeight += n.view.Height
					lockedChildren++
				}
			}
		} else if n, ok := node.(*SplitTree); ok {
			if s.kind == VerticalSplit {
				if n.lockWidth {
					lockedWidth += n.width
					lockedChildren++
				}
			} else {
				if n.lockHeight {
					lockedHeight += n.height
					lockedChildren++
				}
			}
		}
	}
	x, y := 0, 0
	for _, node := range s.children {
		if n, ok := node.(*LeafNode); ok {
			if s.kind == VerticalSplit {
				if !n.view.LockWidth {
					n.view.Width = (s.width - lockedWidth) / (len(s.children) - lockedChildren)
				}
				n.view.Height = s.height

				n.view.x = s.x + x
				n.view.y = s.y
				x += n.view.Width
			} else {
				if !n.view.LockHeight {
					n.view.Height = (s.height - lockedHeight) / (len(s.children) - lockedChildren)
				}
				n.view.Width = s.width

				n.view.y = s.y + y
				n.view.x = s.x
				y += n.view.Height
			}
			if n.view.Buf.Settings["statusline"].(bool) {
				n.view.Height--
			}

			n.view.ToggleTabbar()
			n.view.matches = Match(n.view)
		} else if n, ok := node.(*SplitTree); ok {
			if s.kind == VerticalSplit {
				if !n.lockWidth {
					n.width = (s.width - lockedWidth) / (len(s.children) - lockedChildren)
				}
				n.height = s.height

				n.x = s.x + x
				n.y = s.y
				x += n.width
			} else {
				if !n.lockHeight {
					n.height = (s.height - lockedHeight) / (len(s.children) - lockedChildren)
				}
				n.width = s.width

				n.y = s.y + y
				n.x = s.x
				y += n.height
			}
			n.ResizeSplits()
		}
	}
}

func (l *LeafNode) String() string {
	return l.view.Buf.GetName()
}

func search(haystack []Node, needle Node) int {
	for i, x := range haystack {
		if x == needle {
			return i
		}
	}
	return 0
}

func findView(haystack []*View, needle *View) int {
	for i, x := range haystack {
		if x == needle {
			return i
		}
	}
	return 0
}

// VSplit is here just to make SplitTree fit the Node interface
func (s *SplitTree) VSplit(buf *Buffer, splitIndex int) {}

// HSplit is here just to make SplitTree fit the Node interface
func (s *SplitTree) HSplit(buf *Buffer, splitIndex int) {}

func (s *SplitTree) String() string {
	str := "["
	for _, child := range s.children {
		str += child.String() + ", "
	}
	return str + "]"
}
