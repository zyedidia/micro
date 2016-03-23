package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/go-errors/errors"
	"github.com/mattn/go-isatty"
	"io/ioutil"
	"os"
)

const (
	tabSize      = 4
	synLinesUp   = 75
	synLinesDown = 75
)

func main() {
	var input []byte
	var filename string

	if len(os.Args) > 1 {
		filename = os.Args[1]
		if _, err := os.Stat(filename); err == nil {
			var err error
			input, err = ioutil.ReadFile(filename)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	} else if !isatty.IsTerminal(os.Stdin.Fd()) {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("Error reading stdin")
			os.Exit(1)
		}
		input = bytes
	}

	LoadSyntaxFiles()

	truecolor := os.Getenv("MICRO_TRUECOLOR") == "1"

	oldTerm := os.Getenv("TERM")
	if truecolor {
		os.Setenv("TERM", "xterm-truecolor")
	}

	s, e := tcell.NewTerminfoScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	if truecolor {
		os.Setenv("TERM", oldTerm)
	}

	defer func() {
		if err := recover(); err != nil {
			s.Fini()
			fmt.Println("Micro encountered an error:", err)
			fmt.Print(errors.Wrap(err, 2).ErrorStack())
			os.Exit(1)
		}
	}()

	defStyle := tcell.StyleDefault.
		Background(tcell.ColorDefault).
		Foreground(tcell.ColorDefault)

	if _, ok := colorscheme["default"]; ok {
		defStyle = colorscheme["default"]
	}

	s.SetStyle(defStyle)
	s.EnableMouse()

	m := NewMessenger(s)
	v := NewView(NewBuffer(string(input), filename), m, s)

	// Initially everything needs to be drawn
	redraw := 2
	for {
		if redraw == 2 {
			v.matches = Match(v.buf.rules, v.buf, v)
			s.Clear()
			v.Display()
			v.cursor.Display()
			v.sl.Display()
			m.Display()
			s.Show()
		} else if redraw == 1 {
			v.cursor.Display()
			v.sl.Display()
			m.Display()
			s.Show()
		}

		event := s.PollEvent()
		redraw = v.HandleEvent(event)
	}
}
