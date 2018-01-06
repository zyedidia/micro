package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

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
	term      *terminal.VT
	title     string
	status    int
	selection [2]Loc
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
func (t *Terminal) Start(execCmd []string, view *View) error {
	if len(execCmd) <= 0 {
		return nil
	}

	cmd := exec.Command(execCmd[0], execCmd[1:]...)
	term, _, err := terminal.Start(&t.state, cmd)
	if err != nil {
		return err
	}
	t.term = term
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

// Stop stops execution of the terminal and sets the status
// to VTDone
func (t *Terminal) Stop() {
	t.term.File().Close()
	t.term.Close()
	t.status = VTDone
}

// Close sets the status to VTIdle indicating that the terminal
// is ready for a new command to execute
func (t *Terminal) Close() {
	t.status = VTIdle
}

// WriteString writes a given string to this terminal's pty
func (t *Terminal) WriteString(str string) {
	t.term.File().WriteString(str)
}

// Display displays this terminal in a view
func (t *Terminal) Display(v *View) {
	divider := 0
	if v.x != 0 {
		divider = 1
		dividerStyle := defStyle
		if style, ok := colorscheme["divider"]; ok {
			dividerStyle = style
		}
		for i := 0; i < v.Height; i++ {
			screen.SetContent(v.x, v.y+i, '|', nil, dividerStyle.Reverse(true))
		}
	}
	t.state.Lock()
	defer t.state.Unlock()

	var l Loc
	for y := 0; y < v.Height; y++ {
		for x := 0; x < v.Width; x++ {
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

			screen.SetContent(v.x+x+divider, v.y+y, c, nil, st)
		}
	}
	if t.state.CursorVisible() && tabs[curTab].CurView == v.Num {
		curx, cury := t.state.Cursor()
		screen.ShowCursor(curx+v.x+divider, cury+v.y)
	}
}
