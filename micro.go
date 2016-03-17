package main

import (
	"fmt"
	"github.com/mattn/go-isatty"
	"io/ioutil"
	"os"

	"github.com/gdamore/tcell"
)

const (
	tabSize = 4
)

func main() {
	var input []byte
	var filename string

	if len(os.Args) > 1 {
		filename = os.Args[1]
		var err error
		input, err = ioutil.ReadFile(filename)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else if !isatty.IsTerminal(os.Stdin.Fd()) {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("Error reading stdin")
			os.Exit(1)
		}
		input = bytes
	}

	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	s.EnableMouse()

	v := newViewFromBuffer(newBuffer(string(input), filename), s)

	// Initially everything needs to be drawn
	redraw := 2
	for {
		if redraw == 2 {
			s.Clear()
			v.display()
			v.cursor.display()
			s.Show()
		} else if redraw == 1 {
			v.cursor.display()
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

		redraw = v.handleEvent(event)
	}
}
