package clipboard

import (
	"errors"

	"github.com/zyedidia/clipboard"
)

type Method int

const (
	// External relies on external tools for accessing the clipboard
	// These include xclip, xsel, wl-clipboard for linux, pbcopy/pbpaste on Mac,
	// and Syscalls on Windows.
	External Method = iota
	// Terminal uses the terminal to manage the clipboard via OSC 52. Many
	// terminals do not support OSC 52, in which case this method won't work.
	Terminal
	// Internal just manages the clipboard with an internal buffer and doesn't
	// attempt to interface with the system clipboard
	Internal
)

// CurrentMethod is the method used to store clipboard information
var CurrentMethod Method = Internal

// A Register is a buffer used to store text. The system clipboard has the 'clipboard'
// and 'primary' (linux-only) registers, but other registers may be used internal to micro.
type Register int

const (
	// ClipboardReg is the main system clipboard
	ClipboardReg Register = -1
	// PrimaryReg is the system primary clipboard (linux only)
	PrimaryReg = -2
)

// Initialize attempts to initialize the clipboard using the given method
func Initialize(m Method) error {
	var err error
	switch m {
	case External:
		err = clipboard.Initialize()
	}
	return err
}

// SetMethod changes the clipboard access method
func SetMethod(m string) Method {
	switch m {
	case "internal":
		CurrentMethod = Internal
	case "external":
		CurrentMethod = External
	case "terminal":
		CurrentMethod = Terminal
	}
	return CurrentMethod
}

// Read reads from a clipboard register
func Read(r Register) (string, error) {
	return read(r, CurrentMethod)
}

// Write writes text to a clipboard register
func Write(text string, r Register) error {
	return write(text, r, CurrentMethod)
}

// ReadMulti reads text from a clipboard register for a certain multi-cursor
func ReadMulti(r Register, num int) (string, error) {
	s := multi.getText(r, num)
	return s, nil
}

// WriteMulti writes text to a clipboard register for a certain multi-cursor
func WriteMulti(text string, r Register, num int) error {
	return writeMulti(text, r, num, CurrentMethod)
}

// ValidMulti checks if the internal multi-clipboard is valid and up-to-date
// with the system clipboard
func ValidMulti(r Register, ncursors int) bool {
	clip, err := Read(r)
	if err != nil {
		return false
	}
	return multi.isValid(r, ncursors, clip)
}

func writeMulti(text string, r Register, num int, m Method) error {
	multi.writeText(text, r, num)
	return write(multi.getAllText(r), r, m)
}

func read(r Register, m Method) (string, error) {
	switch m {
	case External:
		switch r {
		case ClipboardReg:
			return clipboard.ReadAll("clipboard")
		case PrimaryReg:
			return clipboard.ReadAll("primary")
		default:
			return internal.read(r), nil
		}
	case Internal:
		return internal.read(r), nil
	case Terminal:
		switch r {
		case ClipboardReg:
			// terminal paste works by sending an esc sequence to the
			// terminal to trigger a paste event
			err := terminal.read("clipboard")
			return "", err
		case PrimaryReg:
			err := terminal.read("primary")
			return "", err
		default:
			return internal.read(r), nil
		}
	}
	return "", errors.New("Invalid clipboard method")
}

func write(text string, r Register, m Method) error {
	switch m {
	case External:
		switch r {
		case ClipboardReg:
			return clipboard.WriteAll(text, "clipboard")
		case PrimaryReg:
			return clipboard.WriteAll(text, "primary")
		default:
			internal.write(text, r)
		}
	case Internal:
		internal.write(text, r)
	case Terminal:
		switch r {
		case ClipboardReg:
			return terminal.write(text, "c")
		case PrimaryReg:
			return terminal.write(text, "p")
		default:
			internal.write(text, r)
		}
	}
	return nil
}
