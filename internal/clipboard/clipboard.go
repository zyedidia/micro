package clipboard

import (
	"errors"
	"log"

	"github.com/zyedidia/clipper"
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

// CurrentMethod is the selected clipboard method. External falls back to internal
// storage if no system clipboard is available.
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

// Each detection attempt owns its result. The worker never changes the selected
// method, and readers must wait for done before accessing clipboard or err.
type clipboardDetection struct {
	done      chan struct{}
	clipboard clipper.Clipboard
	err       error
}

var external *clipboardDetection

// detect starts a probe or reuses the current result. Explicit initialization
// refreshes completed probes so changes to installed tools can be picked up.
// Like the clipboard registers and selected method, external is owned by the
// editor goroutine; only the result inside a detection is written by its worker.
func detect(refresh bool) *clipboardDetection {
	if external != nil {
		select {
		case <-external.done:
			if !refresh {
				return external
			}
		default:
			return external
		}
	}

	clips := make([]clipper.Clipboard, 0, len(clipper.Clipboards)+1)
	clips = append(clips, &clipper.Custom{Name: "micro-clip"})
	clips = append(clips, clipper.Clipboards...)
	d := &clipboardDetection{done: make(chan struct{})}
	external = d
	go func() {
		d.clipboard, d.err = clipper.GetClipboard(clips...)
		close(d.done)
	}()
	return d
}

// Initialize waits for clipboard detection using the given method. A completed
// external probe is refreshed; a probe already in progress is shared.
func Initialize(m Method) error {
	if m != External {
		return nil
	}
	d := detect(true)
	<-d.done
	return d.err
}

// InitAsync starts clipboard detection without waiting, logging any failure.
// Only external system clipboard operations need to wait for the result.
func InitAsync(m Method) {
	if m != External {
		return
	}
	d := detect(false)
	go func() {
		<-d.done
		if d.err != nil {
			log.Println(d.err, " or change 'clipboard' option")
		}
	}()
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
func ReadMulti(r Register, num, ncursors int) (string, error) {
	clip, err := Read(r)
	if err != nil {
		return "", err
	}
	if ValidMulti(r, clip, ncursors) {
		return multi.getText(r, num), nil
	}
	return clip, nil
}

// WriteMulti writes text to a clipboard register for a certain multi-cursor
func WriteMulti(text string, r Register, num int, ncursors int) error {
	return writeMulti(text, r, num, ncursors, CurrentMethod)
}

// ValidMulti checks if the internal multi-clipboard is valid and up-to-date
// with the system clipboard
func ValidMulti(r Register, clip string, ncursors int) bool {
	return multi.isValid(r, clip, ncursors)
}

func writeMulti(text string, r Register, num int, ncursors int, m Method) error {
	multi.writeText(text, r, num, ncursors)
	return write(multi.getAllText(r), r, m)
}

func read(r Register, m Method) (string, error) {
	switch m {
	case External:
		if r != ClipboardReg && r != PrimaryReg {
			return internal.read(r), nil
		}
		d := detect(false)
		<-d.done
		if d.err != nil {
			return internal.read(r), nil
		}
		switch r {
		case ClipboardReg:
			b, e := d.clipboard.ReadAll(clipper.RegClipboard)
			return string(b), e
		case PrimaryReg:
			b, e := d.clipboard.ReadAll(clipper.RegPrimary)
			return string(b), e
		}
	case Internal:
		return internal.read(r), nil
	case Terminal:
		switch r {
		case ClipboardReg:
			// terminal paste works by sending an esc sequence to the
			// terminal to trigger a paste event
			return terminal.read("clipboard")
		case PrimaryReg:
			return terminal.read("primary")
		default:
			return internal.read(r), nil
		}
	}
	return "", errors.New("Invalid clipboard method")
}

func write(text string, r Register, m Method) error {
	switch m {
	case External:
		if r != ClipboardReg && r != PrimaryReg {
			internal.write(text, r)
			return nil
		}
		d := detect(false)
		<-d.done
		if d.err != nil {
			internal.write(text, r)
			return nil
		}
		switch r {
		case ClipboardReg:
			return d.clipboard.WriteAll(clipper.RegClipboard, []byte(text))
		case PrimaryReg:
			return d.clipboard.WriteAll(clipper.RegPrimary, []byte(text))
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
