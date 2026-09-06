package views

import (
	"fmt"
	"testing"
)

func TestHSplit(t *testing.T) {
	root := NewRoot(0, 0, 80, 80)
	n1 := root.VSplit(true)
	root.GetNode(n1).VSplit(true)
	root.GetNode(root.id).ResizeSplit(7)
	root.Resize(120, 120)

	fmt.Println(root.String())
}

func TestIDWithSplits(t *testing.T) {
	root := NewRoot(0, 0, 80, 80)
	id := root.ID()
	if id == 0 {
		t.Fatal("root id is 0")
	}

	newid := root.VSplit(true)
	if got := root.ID(); got != id {
		t.Errorf("root id changed from %d to %d after split", id, got)
	}
	if root.GetNode(id) == nil || root.GetNode(newid) == nil {
		t.Error("leaf lookup by id broken after split")
	}
}
