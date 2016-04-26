package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/zyedidia/tcell"
	"github.com/zyedidia/tcell/encoding"
	"github.com/go-errors/errors"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/go-homedir"
)

const (
	synLinesUp           = 75  // How many lines up to look to do syntax highlighting
	synLinesDown         = 75  // How many lines down to look to do syntax highlighting
	doubleClickThreshold = 400 // How many milliseconds to wait before a second click is not a double click
	undoThreshold        = 500 // If two events are less than n milliseconds apart, undo both of them
)

var (
	// The main screen
	screen tcell.Screen

	// Object to send messages and prompts to the user
	messenger *Messenger

	// The default style
	defStyle tcell.Style

	// Where the user's configuration is
	// This should be $XDG_CONFIG_HOME/micro
	// If $XDG_CONFIG_HOME is not set, it is ~/.config/micro
	configDir string

	// Version is the version number.
	// This should be set by the linker
	Version = "Unknown"

	// Is the help screen open
	helpOpen = false
)

// LoadInput loads the file input for the editor
func LoadInput() (string, []byte, error) {
	// There are a number of ways micro should start given its input
	// 1. If it is given a file in os.Args, it should open that

	// 2. If there is no input file and the input is not a terminal, that means
	// something is being piped in and the stdin should be opened in an
	// empty buffer

	// 3. If there is no input file and the input is a terminal, an empty buffer
	// should be opened

	// These are empty by default so if we get to option 3, we can just returns the
	// default values
	var filename string
	var input []byte
	var err error

	if len(os.Args) > 1 {
		// Option 1
		filename = os.Args[1]
		// Check that the file exists
		if _, e := os.Stat(filename); e == nil {
			input, err = ioutil.ReadFile(filename)
		}
	} else if !isatty.IsTerminal(os.Stdin.Fd()) {
		// Option 2
		// The input is not a terminal, so something is being piped in
		// and we should read from stdin
		input, err = ioutil.ReadAll(os.Stdin)
	}

	// Option 3, or just return whatever we got
	return filename, input, err
}

// InitConfigDir finds the configuration directory for micro according to the
// XDG spec.
// If no directory is found, it creates one.
func InitConfigDir() {
	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgHome == "" {
		home, err := homedir.Dir()
		if err != nil {
			TermMessage("Error finding your home directory\nCan't load syntax files")
			return
		}
		xdgHome = home + "/.config"
	}
	configDir = xdgHome + "/micro"

	if _, err := os.Stat(xdgHome); os.IsNotExist(err) {
		err = os.Mkdir(xdgHome, os.ModePerm)
		if err != nil {
			TermMessage("Error creating XDG_CONFIG_HOME directory: " + err.Error())
		}
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.Mkdir(configDir, os.ModePerm)
		if err != nil {
			TermMessage("Error creating configuration directory: " + err.Error())
		}
	}
}

// InitScreen creates and initializes the tcell screen
func InitScreen() {
	// Should we enable true color?
	truecolor := os.Getenv("MICRO_TRUECOLOR") == "1"

	// In order to enable true color, we have to set the TERM to `xterm-truecolor` when
	// initializing tcell, but after that, we can set the TERM back to whatever it was
	oldTerm := os.Getenv("TERM")
	if truecolor {
		os.Setenv("TERM", "xterm-truecolor")
	}

	// Initilize tcell
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err = screen.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Now we can put the TERM back to what it was before
	if truecolor {
		os.Setenv("TERM", oldTerm)
	}

	// Default style
	defStyle = tcell.StyleDefault.
		Foreground(tcell.ColorDefault).
		Background(tcell.ColorDefault)

	// There may be another default style defined in the colorscheme
	if style, ok := colorscheme["default"]; ok {
		defStyle = style
	}

	screen.SetStyle(defStyle)
	screen.EnableMouse()
}

// Redraw redraws the screen and the given view
func Redraw(view *View) {
	screen.Clear()
	view.Display()
	messenger.Display()
	screen.Show()
}

var flagVersion = flag.Bool("version", false, "Show version number")

func main() {
	flag.Parse()
	if *flagVersion {
		fmt.Println("Micro version:", Version)
		os.Exit(0)
	}

	filename, input, err := LoadInput()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	encoding.Register()
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	// Find the user's configuration directory (probably $XDG_CONFIG_HOME/micro)
	InitConfigDir()
	// Load the user's settings
	InitSettings()
	InitBindings()
	// Load the syntax files, including the colorscheme
	LoadSyntaxFiles()
	// Load the help files
	LoadHelp()

	buf := NewBuffer(string(input), filename)

	InitScreen()

	// This is just so if we have an error, we can exit cleanly and not completely
	// mess up the terminal being worked in
	defer func() {
		if err := recover(); err != nil {
			screen.Fini()
			fmt.Println("Micro encountered an error:", err)
			// Print the stack trace too
			fmt.Print(errors.Wrap(err, 2).ErrorStack())
			os.Exit(1)
		}
	}()

	messenger = new(Messenger)
	view := NewView(buf)

	for {
		// Display everything
		Redraw(view)

		// Wait for the user's action
		event := screen.PollEvent()

		if searching {
			HandleSearchEvent(event, view)
		} else {
			// Check if we should quit
			switch e := event.(type) {
			case *tcell.EventKey:
				switch e.Key() {
				case tcell.KeyCtrlQ:
					// Make sure not to quit if there are unsaved changes
					if helpOpen {
						view.OpenBuffer(buf)
						helpOpen = false
					} else {
						if view.CanClose("Quit anyway? (yes, no, save) ") {
							screen.Fini()
							os.Exit(0)
						}
					}
				case tcell.KeyCtrlE:
					input, canceled := messenger.Prompt("> ")
					if !canceled {
						HandleCommand(input, view)
					}
				case tcell.KeyCtrlB:
					input, canceled := messenger.Prompt("$ ")
					if !canceled {
						HandleShellCommand(input, view, true)
					}
				case tcell.KeyCtrlG:
					if !helpOpen {
						helpBuffer := NewBuffer(helpTxt, "help.md")
						helpBuffer.name = "Help"
						helpOpen = true
						view.OpenBuffer(helpBuffer)
					} else {
						view.OpenBuffer(buf)
						helpOpen = false
					}
				}
			}

			// Send it to the view
			view.HandleEvent(event)
		}
	}
}
