package main

import "testing"

func TestStack(t *testing.T) {
	stack := new(Stack)

	if stack.Len() != 0 {
		t.Errorf("Len failed")
	}
	stack.Push(5)
	stack.Push("test")
	stack.Push(10)
	if stack.Len() != 3 {
		t.Errorf("Len failed")
	}

	var popped interface{}
	popped = stack.Pop()
	if popped != 10 {
		t.Errorf("Pop failed")
	}

	popped = stack.Pop()
	if popped != "test" {
		t.Errorf("Pop failed")
	}

	stack.Push("test")
	popped = stack.Pop()
	if popped != "test" {
		t.Errorf("Pop failed")
	}
	stack.Pop()
	popped = stack.Pop()
	if popped != nil {
		t.Errorf("Pop failed")
	}
}
