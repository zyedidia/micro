package optionprovider

// Option describes a multiple choice selection.
type Option struct {
	T, H string
}

// New creates a new option value.
func New(text, hint string) Option {
	return Option{T: text, H: hint}
}

// Text is the string that will be inserted.
func (o Option) Text() string {
	return o.T
}

// Hint describes the text, e.g. if it's a method.
func (o Option) Hint() string {
	return o.H
}
