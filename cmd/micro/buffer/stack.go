package buffer

// TEStack is a simple implementation of a LIFO stack for text events
type TEStack struct {
	Top  *Element
	Size int
}

// An Element which is stored in the Stack
type Element struct {
	Value *TextEvent
	Next  *Element
}

// Len returns the stack's length
func (s *TEStack) Len() int {
	return s.Size
}

// Push a new element onto the stack
func (s *TEStack) Push(value *TextEvent) {
	s.Top = &Element{value, s.Top}
	s.Size++
}

// Pop removes the top element from the stack and returns its value
// If the stack is empty, return nil
func (s *TEStack) Pop() (value *TextEvent) {
	if s.Size > 0 {
		value, s.Top = s.Top.Value, s.Top.Next
		s.Size--
		return
	}
	return nil
}

// Peek returns the top element of the stack without removing it
func (s *TEStack) Peek() *TextEvent {
	if s.Size > 0 {
		return s.Top.Value
	}
	return nil
}
