package optionprovider

var noopOptions = []Option{}

// Noop is an option provider that does nothing.
func Noop(buffer []byte, offset int) (options []Option, err error) {
	return noopOptions, nil
}
