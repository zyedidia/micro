package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/zyedidia/clipboard"
	"github.com/zyedidia/tcell"
	"github.com/zyedidia/terminal"
)

const (
	VTIdle    = iota // Waiting for a new command
	VTRunning        // Currently running a command
	VTDone           // Finished running a command
)

// A Terminal holds information for the terminal emulator
type Terminal struct {
	state     terminal.State
	view      *View
	vtOld     ViewType
	term      *terminal.VT
	title     string
	status    int
	selection [2]Loc
	wait      bool
	getOutput bool
	output    *bytes.Buffer
	callback  string
}

// HasSelection returns whether this terminal has a valid selection
func (t *Terminal) HasSelection() bool {
	return t.selection[0] != t.selection[1]
}

// GetSelection returns the selected text
func (t *Terminal) GetSelection(width int) string {
	start := t.selection[0]
	end := t.selection[1]
	if start.GreaterThan(end) {
		start, end = end, start
	}
	var ret string
	var l Loc
	for y := start.Y; y <= end.Y; y++ {
		for x := 0; x < width; x++ {
			l.X, l.Y = x, y
			if l.GreaterEqual(start) && l.LessThan(end) {
				c, _, _ := t.state.Cell(x, y)
				ret += string(c)
			}
		}
	}
	return ret
}

// Start begins a new command in this terminal with a given view
func (t *Terminal) Start(execCmd []string, view *View, getOutput bool) error {
	if len(execCmd) <= 0 {
		return nil
	}

	cmd := exec.Command(execCmd[0], execCmd[1:]...)
	t.output = nil
	if getOutput {
		t.output = bytes.NewBuffer([]byte{})
	}
	term, _, err := terminal.Start(&t.state, cmd, t.output)
	if err != nil {
		return err
	}
	t.term = term
	t.view = view
	t.getOutput = getOutput
	t.vtOld = view.Type
	t.status = VTRunning
	t.title = execCmd[0] + ":" + strconv.Itoa(cmd.Process.Pid)

	go func() {
		for {
			err := term.Parse()
			if err != nil {
				fmt.Fprintln(os.Stderr, "[Press enter to close]")
				break
			}
			updateterm <- true
		}
		closeterm <- view.Num
	}()

	return nil
}

// Resize informs the terminal of a resize event
func (t *Terminal) Resize(width, height int) {
	t.term.Resize(width, height)
}

// HandleEvent handles a tcell event by forwarding it to the terminal emulator
// If the event is a mouse event and the program running in the emulator
// does not have mouse support, the emulator will support selections and
// copy-paste
func (t *Terminal) HandleEvent(event tcell.Event) {
	if e, ok := event.(*tcell.EventKey); ok {
		if t.status == VTDone {
			switch e.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlQ, tcell.KeyEnter:
				t.Close()
				t.view.Type = vtDefault
			default:
			}
		}
		if e.Key() == tcell.KeyCtrlC && t.HasSelection() {
			clipboard.WriteAll(t.GetSelection(t.view.Width), "clipboard")
			messenger.Message("Copied selection to clipboard")
		} else if t.status != VTDone {
			t.WriteString(event.EscSeq())
		}
	} else if e, ok := event.(*tcell.EventMouse); !ok || t.state.Mode(terminal.ModeMouseMask) {
		t.WriteString(event.EscSeq())
	} else {
		x, y := e.Position()
		x -= t.view.x
		y += t.view.y

		if e.Buttons() == tcell.Button1 {
			if !t.view.mouseReleased {
				// drag
				t.selection[1].X = x
				t.selection[1].Y = y
			} else {
				t.selection[0].X = x
				t.selection[0].Y = y
				t.selection[1].X = x
				t.selection[1].Y = y
			}

			t.view.mouseReleased = false
		} else if e.Buttons() == tcell.ButtonNone {
			if !t.view.mouseReleased {
				t.selection[1].X = x
				t.selection[1].Y = y
			}
			t.view.mouseReleased = true
		}
	}
}

// Stop stops execution of the terminal and sets the status
// to VTDone
func (t *Terminal) Stop() {
	t.term.File().Close()
	t.term.Close()
	if t.wait {
		t.status = VTDone
	} else {
		t.Close()
		t.view.Type = t.vtOld
	}
}

// Close sets the status to VTIdle indicating that the terminal
// is ready for a new command to execute
func (t *Terminal) Close() {
	t.status = VTIdle
	// call the lua function that the user has given as a callback
	if t.getOutput {
		_, err := Call(t.callback, t.output.String())
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
		}
	}
}

// WriteString writes a given string to this terminal's pty
func (t *Terminal) WriteString(str string) {
	t.term.File().WriteString(str)
}

// Display displays this terminal in a view
func (t *Terminal) Display() {
	divider := 0
	if t.view.x != 0 {
		divider = 1
		dividerStyle := defStyle
		if style, ok := colorscheme["divider"]; ok {
			dividerStyle = style
		}
		for i := 0; i < t.view.Height; i++ {
			screen.SetContent(t.view.x, t.view.y+i, '|', nil, dividerStyle.Reverse(true))
		}
	}
	t.state.Lock()
	defer t.state.Unlock()

	var l Loc
	for y := 0; y < t.view.Height; y++ {
		for x := 0; x < t.view.Width; x++ {
			l.X, l.Y = x, y
			c, f, b := t.state.Cell(x, y)

			fg, bg := int(f), int(b)
			if f == terminal.DefaultFG {
				fg = int(tcell.ColorDefault)
			}
			if b == terminal.DefaultBG {
				bg = int(tcell.ColorDefault)
			}
			st := tcell.StyleDefault.Foreground(GetColor256(int(fg))).Background(GetColor256(int(bg)))

			if l.LessThan(t.selection[1]) && l.GreaterEqual(t.selection[0]) || l.LessThan(t.selection[0]) && l.GreaterEqual(t.selection[1]) {
				st = st.Reverse(true)
			}

			screen.SetContent(t.view.x+x+divider, t.view.y+y, c, nil, st)
		}
	}
	if t.state.CursorVisible() && tabs[curTab].CurView == t.view.Num {
		curx, cury := t.state.Cursor()
		screen.ShowCursor(curx+t.view.x+divider, cury+t.view.y)
	}
}
