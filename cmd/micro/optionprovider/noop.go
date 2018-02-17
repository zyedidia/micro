package optionprovider

var noopOptions = []Option{}

// Noop is an option provider that does nothing.
func Noop(l func(s string, values ...interface{}), buffer []byte, startOffset, currentOffset int) (options []Option, startOffsetDelta int, err error) {
	return noopOptions, 0, nil
}
