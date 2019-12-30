package info

// A GutterMessage is a message displayed on the side of the editor
type GutterMessage struct {
	lineNum int
	msg     string
	kind    int
}

// These are the different types of messages
const (
	// GutterInfo represents a simple info message
	GutterInfo = iota
	// GutterWarning represents a compiler warning
	GutterWarning
	// GutterError represents a compiler error
	GutterError
)
