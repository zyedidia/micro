package shell

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"github.com/zyedidia/micro/internal/buffer"
	"github.com/zyedidia/micro/internal/config"
	ulua "github.com/zyedidia/micro/internal/lua"
	"github.com/zyedidia/micro/internal/screen"
	"github.com/zyedidia/terminal"
	luar "layeh.com/gopher-luar"
)

type TermType int
type CallbackFunc func(string)

const (
	TTClose   = iota // Should be closed
	TTRunning        // Currently running a command
	TTDone           // Finished running a command
)

var CloseTerms chan bool

func init() {
	CloseTerms = make(chan bool)
}

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
	callback  CallbackFunc
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
func (t *Terminal) Start(execCmd []string, getOutput bool, wait bool, callback string, userargs []interface{}) error {
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

	luaFn := strings.Split(callback, ".")
	if len(luaFn) >= 2 {
		plName, plFn := luaFn[0], luaFn[1]
		pl := config.FindPlugin(plName)
		if pl != nil {
			t.callback = func(out string) {
				var luaArgs []lua.LValue
				luaArgs = append(luaArgs, luar.New(ulua.L, out))
				for _, v := range userargs {
					luaArgs = append(luaArgs, luar.New(ulua.L, v))
				}
				_, err := pl.Call(plFn, luaArgs...)
				if err != nil {
					screen.TermMessage(err)
				}
			}
		}
	}

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
		CloseTerms <- true
	}
}

// Close sets the Status to TTClose indicating that the terminal
// is done and should be closed
func (t *Terminal) Close() {
	t.Status = TTClose
	// call the lua function that the user has given as a callback
	if t.getOutput {
		if t.callback != nil {
			t.callback(t.output.String())
		}
	}
}

// WriteString writes a given string to this terminal's pty
func (t *Terminal) WriteString(str string) {
	t.Term.File().WriteString(str)
}
