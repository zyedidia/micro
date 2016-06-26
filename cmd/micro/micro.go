package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/atotto/clipboard"
	"github.com/go-errors/errors"
	"github.com/layeh/gopher-luar"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/go-homedir"
	"github.com/yuin/gopher-lua"
	"github.com/zyedidia/tcell"
	"github.com/zyedidia/tcell/encoding"
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

	// The default highlighting style
	// This simply defines the default foreground and background colors
	defStyle tcell.Style

	// Where the user's configuration is
	// This should be $XDG_CONFIG_HOME/micro
	// If $XDG_CONFIG_HOME is not set, it is ~/.config/micro
	configDir string

	// Version is the version number or commit hash
	// This should be set by the linker
	Version = "Unknown"

	// L is the lua state
	// This is the VM that runs the plugins
	L *lua.LState

	// The list of views
	tabs []*Tab
	// This is the currently open tab
	// It's just an index to the tab in the tabs array
	curTab int

	jobs   chan JobFunction
	events chan tcell.Event
)

// LoadInput loads the file input for the editor
func LoadInput() []*Buffer {
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
	var buffers []*Buffer

	if len(os.Args) > 1 {
		// Option 1
		for i := 1; i < len(os.Args); i++ {
			filename = os.Args[i]
			// Check that the file exists
			if _, e := os.Stat(filename); e == nil {
				input, err = ioutil.ReadFile(filename)
				if err != nil {
					TermMessage(err)
					continue
				}
			}
			buffers = append(buffers, NewBuffer(input, filename))
		}
	} else if !isatty.IsTerminal(os.Stdin.Fd()) {
		// Option 2
		// The input is not a terminal, so something is being piped in
		// and we should read from stdin
		input, err = ioutil.ReadAll(os.Stdin)
		buffers = append(buffers, NewBuffer(input, filename))
	} else {
		// Option 3, just open an empty buffer
		buffers = append(buffers, NewBuffer(input, filename))
	}

	return buffers
}

// InitConfigDir finds the configuration directory for micro according to the XDG spec.
// If no directory is found, it creates one.
func InitConfigDir() {
	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgHome == "" {
		// The user has not set $XDG_CONFIG_HOME so we should act like it was set to ~/.config
		home, err := homedir.Dir()
		if err != nil {
			TermMessage("Error finding your home directory\nCan't load config files")
			return
		}
		xdgHome = home + "/.config"
	}
	configDir = xdgHome + "/micro"

	if _, err := os.Stat(xdgHome); os.IsNotExist(err) {
		// If the xdgHome doesn't exist we should create it
		err = os.Mkdir(xdgHome, os.ModePerm)
		if err != nil {
			TermMessage("Error creating XDG_CONFIG_HOME directory: " + err.Error())
		}
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// If the micro specific config directory doesn't exist we should create that too
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

	screen.SetStyle(defStyle)
	screen.EnableMouse()
}

// RedrawAll redraws everything -- all the views and the messenger
func RedrawAll() {
	messenger.Clear()
	DisplayTabs()
	messenger.Display()
	for _, v := range tabs[curTab].views {
		v.Display()
	}
	screen.Show()
}

var flagVersion = flag.Bool("version", false, "Show version number")

func main() {
	flag.Parse()
	if *flagVersion {
		fmt.Println("Micro version:", Version)
		os.Exit(0)
	}

	L = lua.NewState()
	defer L.Close()

	// Some encoding stuff in case the user isn't using UTF-8
	encoding.Register()
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	// Find the user's configuration directory (probably $XDG_CONFIG_HOME/micro)
	InitConfigDir()
	// Load the user's settings
	InitSettings()
	InitCommands()
	InitBindings()
	// Load the syntax files, including the colorscheme
	LoadSyntaxFiles()
	// Load the help files
	LoadHelp()

	InitScreen()

	// This is just so if we have an error, we can exit cleanly and not completely
	// mess up the terminal being worked in
	// In other words we need to shut down tcell before the program crashes
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
	messenger.history = make(map[string][]string)

	buffers := LoadInput()
	for _, buf := range buffers {
		tab := NewTabFromView(NewView(buf))
		tab.SetNum(len(tabs))
		tabs = append(tabs, tab)
		for _, t := range tabs {
			for _, v := range t.views {
				v.Resize(screen.Size())
			}
		}
	}

	L.SetGlobal("OS", luar.New(L, runtime.GOOS))
	L.SetGlobal("tabs", luar.New(L, tabs))
	L.SetGlobal("curTab", luar.New(L, curTab))
	L.SetGlobal("messenger", luar.New(L, messenger))
	L.SetGlobal("GetOption", luar.New(L, GetOption))
	L.SetGlobal("AddOption", luar.New(L, AddOption))
	L.SetGlobal("BindKey", luar.New(L, BindKey))
	L.SetGlobal("MakeCommand", luar.New(L, MakeCommand))
	L.SetGlobal("CurView", luar.New(L, CurView))
	L.SetGlobal("IsWordChar", luar.New(L, IsWordChar))

	L.SetGlobal("JobStart", luar.New(L, JobStart))
	L.SetGlobal("JobSend", luar.New(L, JobSend))
	L.SetGlobal("JobStop", luar.New(L, JobStop))

	LoadPlugins()

	jobs = make(chan JobFunction, 100)
	events = make(chan tcell.Event)

	go func() {
		for {
			events <- screen.PollEvent()
		}
	}()

	for {
		// Display everything
		RedrawAll()

		var event tcell.Event
		select {
		case f := <-jobs:
			f.function(f.output, f.args...)
			continue
		case event = <-events:
		}

		switch e := event.(type) {
		case *tcell.EventMouse:
			if e.Buttons() == tcell.Button1 {
				_, h := screen.Size()
				_, y := e.Position()
				if y == h-1 && messenger.message != "" {
					clipboard.WriteAll(messenger.message)
					continue
				}
			}
		}

		if TabbarHandleMouseEvent(event) {
			continue
		}

		if searching {
			// Since searching is done in real time, we need to redraw every time
			// there is a new event in the search bar
			HandleSearchEvent(event, CurView())
		} else {
			// Send it to the view
			CurView().HandleEvent(event)
		}
	}
}
