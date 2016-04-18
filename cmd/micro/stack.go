package main

// Stack is a simple implementation of a LIFO stack
type Stack struct {
	top  *Element
	size int
}

// An Element which is stored in the Stack
type Element struct {
	value interface{} // All types satisfy the empty interface, so we can store anything here.
	next  *Element
}

// Len returns the stack's length
func (s *Stack) Len() int {
	return s.size
}

// Push a new element onto the stack
func (s *Stack) Push(value interface{}) {
	s.top = &Element{value, s.top}
	s.size++
}

// Pop removes the top element from the stack and returns its value
// If the stack is empty, return nil
func (s *Stack) Pop() (value interface{}) {
	if s.size > 0 {
		value, s.top = s.top.value, s.top.next
		s.size--
		return
	}
	return nil
}

// Peek returns the top element of the stack without removing it
func (s *Stack) Peek() interface{} {
	if s.size > 0 {
		return s.top.value
	}
	return nil
}
