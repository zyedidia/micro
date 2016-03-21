package main

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/mattn/go-isatty"
	"github.com/zyedidia/tcell"
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

	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
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

	s.SetStyle(defStyle)
	s.EnableMouse()

	v := NewView(NewBuffer(string(input), filename), s)

	// Initially everything needs to be drawn
	redraw := 2
	for {
		if redraw == 2 {
			v.matches = Match(v.buf.rules, v.buf, v)
			s.Clear()
			v.Display()
			v.cursor.Display()
			v.sl.Display()
			s.Show()
		} else if redraw == 1 {
			v.cursor.Display()
			v.sl.Display()
			s.Show()
		}

		event := s.PollEvent()

		switch e := event.(type) {
		case *tcell.EventKey:
			if e.Key() == tcell.KeyCtrlQ {
				s.Fini()
				os.Exit(0)
			}
		}

		redraw = v.HandleEvent(event)
	}
}
