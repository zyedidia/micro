package main

import (
	"sort"
<<<<<<< HEAD
=======
	"strconv"
>>>>>>> dc272633 (UI Tweaks)
	"path/filepath"

	"github.com/zyedidia/tcell"
)

var (
	tabBarLen int
	currentTabOffset int
	previousTabOffset int
	tabBarOffset int
	ScrollOffset int
)

type Tab struct {
	// This contains all the views in this tab
	// There is generally only one view per tab, but you can have
	// multiple views with splits
	views []*View
	// This is the current view for this tab
	CurView int

	tree *SplitTree
}

// NewTabFromView creates a new tab and puts the given view in the tab
func NewTabFromView(v *View) *Tab {
	t := new(Tab)
	t.views = append(t.views, v)
	t.views[0].Num = 0

	t.tree = new(SplitTree)
	t.tree.kind = VerticalSplit
	t.tree.children = []Node{NewLeafNode(t.views[0], t.tree)}

	w, h := screen.Size()
	t.tree.width = w
	t.tree.height = h

	if globalSettings["infobar"].(bool) {
		t.tree.height--
	}

	//Reset the Scroll Offset so that the tabbar centers correctly on the new tab.
	ScrollOffset = 0

	t.Resize()

	return t
}

// SetNum sets all this tab's views to have the correct tab number
func (t *Tab) SetNum(num int) {
	t.tree.tabNum = num
	for _, v := range t.views {
		v.TabNum = num
	}
}

func (t *Tab) Cleanup() {
	t.tree.Cleanup()
}

func (t *Tab) Resize() {
	w, h := screen.Size()
	t.tree.width = w
	t.tree.height = h

	if globalSettings["infobar"].(bool) {
		t.tree.height--
	}

	t.tree.ResizeSplits()

	for i, v := range t.views {
		v.Num = i
	}
}

// CurView returns the current view
func CurView() *View {
	curTab := tabs[curTab]
	return curTab.views[curTab.CurView]
}

// TabbarString returns the string that should be displayed in the tabbar
// It also returns a map containing which indicies correspond to which tab number
// This is useful when we know that the mouse click has occurred at an x location
// but need to know which tab that corresponds to to accurately change the tab
func TabbarString() (string, map[int]int) {
	str := ""
	indicies := make(map[int]int)
	for i, t := range tabs {
		if i == curTab {
			str += "["
		} else {
			str += " "
		}
<<<<<<< HEAD
		//To address issue 556.2
=======
		if globalSettings["numberedtabs"].(bool){
			str += "(" + strconv.Itoa(i + 1) + ")"
		}
>>>>>>> dc272633 (UI Tweaks)
		_, name := filepath.Split(t.views[t.CurView].Buf.GetName())
		str += name
		if i == curTab {
			str += "]"
		} else {
			str += " "
		}
		indicies[len(str)-1] = i + 1
		str += " "
	}
	return str, indicies
}

// TabbarHandleMouseEvent checks the given mouse event if it is clicking on the tabbar
// If it is it changes the current tab accordingly
// This function returns true if the tab is changed
func TabbarHandleMouseEvent(event tcell.Event) bool {
	// There is no tabbar displayed if there are less than 2 tabs
	if len(tabs) <= 1 {
		return false
	}

	switch e := event.(type) {
	case *tcell.EventMouse:
		button := e.Buttons()
		// Must be a left click
		if button == tcell.Button1 {
			x, y := e.Position()
			if y != 0 {
				return false
			}
			str, indicies := TabbarString()
			if x+tabBarOffset >= len(str) {
				return false
			}
			var tabnum int
			var keys []int
			for k := range indicies {
				keys = append(keys, k)
			}
			sort.Ints(keys)
			for _, k := range keys {
				if x+tabBarOffset <= k {
					tabnum = indicies[k] - 1
					break
				}
			}
			curTab = tabnum
			return true
		}
		
		//Close tab on right click
		if button == tcell.Button3 {
			x, y := e.Position()
			if y != 0 {
				return false
			}
			str, indicies := TabbarString()
			if x + tabBarOffset >= len(str) {
				return false
			}
			var tabnum int
			var keys []int
			for k := range indicies {
				keys = append(keys, k)
			}
			sort.Ints(keys)
			for _, k := range keys {
				if x+tabBarOffset <= k {
					tabnum = indicies[k] - 1
					break
				}
			}
			c := 0
			for i := range tabs[tabnum].views {
				tabs[tabnum].views[i-c].Quit(false)
				c++
			}
			return true
		}
		
		//Scroll left on mousewheel up
		if button == tcell.WheelUp {
			_, y := e.Position()
			if y != 0 {
				return false
			}
			w, _ := screen.Size()
			if w > tabBarLen {
				return true
			}
			//If there is nothing to the left to scroll to, ignore scroll event completely.
			//leftBuffer := previousTabOffset + (currentTabOffset - previousTabOffset)/2
			//if leftBuffer <= 0 {
			//	return true
			//}
			ScrollOffset--
			return true
		}
		
		//Scroll right on mousewheel down
		if button == tcell.WheelDown {
			_, y := e.Position()
			if y != 0 {
				return false
			}
			w, _ := screen.Size()
			if w > tabBarLen {
				return true
			}
			//If there is nothing to the right to scroll to, ignore scroll event completely
			//rightBuffer := currentTabOffset + (currentTabOffset - previousTabOffset)/2
			//if rightBuffer <= 0 {
			//	return true
			//}
			ScrollOffset++
			return true
		}
	}
	
	ScrollOffset = 0
	return false
}

// DisplayTabs displays the tabbar at the top of the editor if there are multiple tabs
func DisplayTabs() {
	if len(tabs) <= 1 {
		if !globalSettings["tabbaralways"].(bool) {
			return
		}
	}

	str, indicies := TabbarString()

	tabBarStyle := defStyle.Reverse(true)
	if style, ok := colorscheme["tabbar"]; ok {
		tabBarStyle = style
	}

	// Maybe there is a unicode filename?
	fileRunes := []rune(str)
	tabBarLen = len(fileRunes)
	w, _ := screen.Size()
	tooWide := (w < tabBarLen)

	// if the entire tab-bar is longer than the screen is wide,
	// then it should be truncated appropriately to keep the
	// active tab visible on the UI.
	if tooWide == true {
		// first we have to work out where the selected tab is
		// out of the total length of the tab bar. this is done
		// by extracting the hit-areas from the indicies map
		// that was constructed by `TabbarString()`
		var keys []int
		for offset := range indicies {
			keys = append(keys, offset)
		}
		// sort them to be in ascending order so that values will
		// correctly reflect the displayed ordering of the tabs
		sort.Ints(keys)
		// record the offset of each tab and the previous tab so
		// we can find the position of the tab's hit-box.
		for _, k := range keys {
			tabIndex := indicies[k] - 1
			if tabIndex == curTab {
				currentTabOffset = k
				break
			}
			// this is +2 because there are two padding spaces that aren't accounted
			// for in the display. please note that this is for cosmetic purposes only.
			previousTabOffset = k + 2
		}
		// get the width of the hitbox of the active tab, from there calculate the offsets
		// to the left and right of it to approximately center it on the tab bar display.
		centeringOffset := (w - (currentTabOffset - previousTabOffset) + ScrollOffset)
		//centeringOffset := (w - (currentTabOffset - previousTabOffset) )
		leftBuffer := previousTabOffset - (centeringOffset / 2)
		rightBuffer := currentTabOffset + (centeringOffset / 2)

		// check to make sure we haven't overshot the bounds of the string,
		// if we have, then take that remainder and put it on the left side
		overshotRight := rightBuffer - tabBarLen
		if overshotRight > 0 {
			leftBuffer = leftBuffer + overshotRight
		}

		overshotLeft := leftBuffer - 0
		if overshotLeft < 0 {
			leftBuffer = 0
			rightBuffer = leftBuffer + (w - 1)
		} else {
			rightBuffer = leftBuffer + (w - 2)
		}

		if rightBuffer > tabBarLen - 1 {
			rightBuffer = tabBarLen - 1
		}

		// construct a new buffer of text to put the
		// newly formatted tab bar text into.
		var displayText []rune

		// if the left-side of the tab bar isn't at the start
		// of the constructed tab bar text, then show that are
		// more tabs to the left by displaying a "<"
		if leftBuffer != 0 {
			displayText = append(displayText, '<')
		}
		// copy the runes in from the original tab bar text string
		// into the new display buffer
		for x := leftBuffer; x < rightBuffer; x++ {
			displayText = append(displayText, fileRunes[x])
		}
		// if there is more text to the right of the right-most
		// column in the tab bar text, then indicate there are more
		// tabs to the right by displaying a ">"
		if rightBuffer < tabBarLen - 1 {
			displayText = append(displayText, '>')
		}

		// now store the offset from zero of the left-most text
		// that is being displayed. This is to ensure that when
		// clicking on the tab bar, the correct tab gets selected.
		tabBarOffset = leftBuffer

		// use the constructed buffer as the display buffer to print
		// onscreen.
		fileRunes = displayText
	} else {
		tabBarOffset = 0
	}

	// iterate over the width of the terminal display and for each column,
	// write a character into the tab display area with the appropriate style.
	for x := 0; x < w; x++ {
		if x < len(fileRunes) {
			screen.SetContent(x, 0, fileRunes[x], nil, tabBarStyle)
		} else {
			screen.SetContent(x, 0, ' ', nil, tabBarStyle)
		}
	}
}
