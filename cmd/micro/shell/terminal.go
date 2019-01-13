package shell

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/terminal"
)

type TermType int

const (
	TTClose   = iota // Should be closed
	TTRunning        // Currently running a command
	TTDone           // Finished running a command
)

// A Terminal holds information for the terminal emulator
type Terminal struct {
	State     terminal.State
	Term      *terminal.VT
	title     string
	Status    TermType
	Selection [2]buffer.Loc
	wait      bool
	getOutput bool
	output    *bytes.Buffer
	callback  string
}

// HasSelection returns whether this terminal has a valid selection
func (t *Terminal) HasSelection() bool {
	return t.Selection[0] != t.Selection[1]
}

func (t *Terminal) Name() string {
	return t.title
}

// GetSelection returns the selected text
func (t *Terminal) GetSelection(width int) string {
	start := t.Selection[0]
	end := t.Selection[1]
	if start.GreaterThan(end) {
		start, end = end, start
	}
	var ret string
	var l buffer.Loc
	for y := start.Y; y <= end.Y; y++ {
		for x := 0; x < width; x++ {
			l.X, l.Y = x, y
			if l.GreaterEqual(start) && l.LessThan(end) {
				c, _, _ := t.State.Cell(x, y)
				ret += string(c)
			}
		}
	}
	return ret
}

// Start begins a new command in this terminal with a given view
func (t *Terminal) Start(execCmd []string, getOutput bool, wait bool) error {
	if len(execCmd) <= 0 {
		return nil
	}

	cmd := exec.Command(execCmd[0], execCmd[1:]...)
	t.output = nil
	if getOutput {
		t.output = bytes.NewBuffer([]byte{})
	}
	Term, _, err := terminal.Start(&t.State, cmd, t.output)
	if err != nil {
		return err
	}
	t.Term = Term
	t.getOutput = getOutput
	t.Status = TTRunning
	t.title = execCmd[0] + ":" + strconv.Itoa(cmd.Process.Pid)
	t.wait = wait

	go func() {
		for {
			err := Term.Parse()
			if err != nil {
				fmt.Fprintln(os.Stderr, "[Press enter to close]")
				break
			}
			screen.Redraw()
		}
		t.Stop()
	}()

	return nil
}

// Stop stops execution of the terminal and sets the Status
// to TTDone
func (t *Terminal) Stop() {
	t.Term.File().Close()
	t.Term.Close()
	if t.wait {
		t.Status = TTDone
	} else {
		t.Close()
	}
}

// Close sets the Status to TTClose indicating that the terminal
// is done and should be closed
func (t *Terminal) Close() {
	t.Status = TTClose
	// call the lua function that the user has given as a callback
	if t.getOutput {
		// TODO: plugin callback on Term emulator
		// _, err := Call(t.callback, t.output.String())
		// if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
		// 	TermMessage(err)
		// }
	}
}

// WriteString writes a given string to this terminal's pty
func (t *Terminal) WriteString(str string) {
	t.Term.File().WriteString(str)
}
