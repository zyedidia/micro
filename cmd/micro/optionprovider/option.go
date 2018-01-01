package optionprovider

// Option describes a multiple choice selection.
type Option interface {
	// Text is the string that will be inserted.
	Text() string
	// Hint describes the text, e.g. if it's a method.
	Hint() string
}
