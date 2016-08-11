package main

type SplitType bool

const (
	VerticalSplit   = false
	HorizontalSplit = true
)

type Node interface {
	VSplit(buf *Buffer)
	HSplit(buf *Buffer)
	String() string
}

type LeafNode struct {
	view *View

	parent *SplitTree
}

func NewLeafNode(v *View, parent *SplitTree) *LeafNode {
	n := new(LeafNode)
	n.view = v
	n.view.splitNode = n
	n.parent = parent
	return n
}

type SplitTree struct {
	kind SplitType

	parent   *SplitTree
	children []Node

	x int
	y int

	width  int
	height int

	tabNum int
}

func (l *LeafNode) VSplit(buf *Buffer) {
	tab := tabs[l.parent.tabNum]
	if l.parent.kind == VerticalSplit {
		newView := NewView(buf)
		newView.TabNum = l.parent.tabNum
		newView.Num = len(tab.views)
		l.parent.children = append(l.parent.children, NewLeafNode(newView, l.parent))

		tab.curView++
		tab.views = append(tab.views, newView)
	} else {
		s := new(SplitTree)
		s.kind = VerticalSplit
		s.parent = l.parent
		newView := NewView(buf)
		newView.TabNum = l.parent.tabNum
		newView.Num = len(tab.views)
		s.children = []Node{l, NewLeafNode(newView, s)}
		l.parent.children[search(l.parent.children, l)] = s

		tab.curView++
		tab.views = append(tab.views, newView)
	}
}

func (l *LeafNode) HSplit(buf *Buffer) {
	tab := tabs[l.parent.tabNum]
	if l.parent.kind == HorizontalSplit {
		newView := NewView(buf)
		newView.TabNum = l.parent.tabNum
		newView.Num = len(tab.views)
		l.parent.children = append(l.parent.children, NewLeafNode(newView, l.parent))

		tab.curView++
		tab.views = append(tab.views, newView)
	} else {
		s := new(SplitTree)
		s.kind = HorizontalSplit
		s.parent = l.parent
		newView := NewView(buf)
		newView.TabNum = l.parent.tabNum
		newView.Num = len(tab.views)
		s.children = []Node{l, NewLeafNode(newView, s)}
		l.parent.children[search(l.parent.children, l)] = s

		tab.curView++
		tab.views = append(tab.views, newView)
	}
}

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
	if tab.curView > 0 {
		tab.curView--
	}
}

func (s *SplitTree) Cleanup() {
	for i, node := range s.children {
		if n, ok := node.(*SplitTree); ok {
			if len(n.children) == 1 {
				if _, ok := n.children[0].(*LeafNode); ok {
					s.children[i] = n.children[0]
				}
			}
			n.Cleanup()
		}
	}
}

func (s *SplitTree) ResizeSplits() {
	for i, node := range s.children {
		if n, ok := node.(*LeafNode); ok {
			if s.kind == VerticalSplit {
				n.view.width = s.width / len(s.children)
				n.view.height = s.height

				n.view.x = s.x + n.view.width*i
				n.view.y = s.y
			} else {
				n.view.height = s.height / len(s.children)
				n.view.width = s.width

				n.view.y = s.y + n.view.height*i
				n.view.x = s.x
			}
			n.view.matches = Match(n.view)
		} else if n, ok := node.(*SplitTree); ok {
			if s.kind == VerticalSplit {
				n.width = s.width / len(s.children)
				n.height = s.height

				n.x = s.x + n.width*i
				n.y = s.y
			} else {
				n.height = s.height / len(s.children)
				n.width = s.width

				n.y = s.y + n.height*i
				n.x = s.x
			}
			n.ResizeSplits()
		}
	}
}

func (l *LeafNode) String() string {
	return l.view.Buf.Name
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

func (s *SplitTree) VSplit(buf *Buffer) {}
func (s *SplitTree) HSplit(buf *Buffer) {}

func (s *SplitTree) String() string {
	str := "["
	for _, child := range s.children {
		str += child.String() + ", "
	}
	return str + "]"
}
