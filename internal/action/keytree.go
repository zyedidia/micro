package action

type KeyAction func(Pane) bool
type MouseAction func(Pane, *MouseEvent) bool
type KeyAnyAction func(Pane, keys []KeyEvent)

type KeyTreeNode struct {
	children map[Event]KeyTreeNode

	// action KeyAction
	// any KeyAnyAction
}
