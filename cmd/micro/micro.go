package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/go-errors/errors"
	"github.com/layeh/gopher-luar"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/go-homedir"
	"github.com/yuin/gopher-lua"
	"github.com/zyedidia/clipboard"
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
	// These variables should be set by the linker when compiling
	Version     = "Unknown"
	CommitHash  = "Unknown"
	CompileDate = "Unknown"

	// L is the lua state
	// This is the VM that runs the plugins
	L *lua.LState

	// The list of views
	tabs []*Tab
	// This is the currently open tab
	// It's just an index to the tab in the tabs array
	curTab int

	// Channel of jobs running in the background
	jobs chan JobFunction
	// Event channel
	events chan tcell.Event
)

// LoadInput determines which files should be loaded into buffers
// based on the input stored in os.Args
func LoadInput() []*Buffer {
	// There are a number of ways micro should start given its input

	// 1. If it is given a files in os.Args, it should open those

	// 2. If there is no input file and the input is not a terminal, that means
	// something is being piped in and the stdin should be opened in an
	// empty buffer

	// 3. If there is no input file and the input is a terminal, an empty buffer
	// should be opened

	var filename string
	var input []byte
	var err error
	var buffers []*Buffer

	if len(os.Args) > 1 {
		// Option 1
		// We go through each file and load it
		for i := 1; i < len(os.Args); i++ {
			filename = os.Args[i]

			// Need to skip arguments that are not filenames
			if filename == "-cursor" {
				i++ // also skip the LINE,COL for -cursor
				continue
			}

			// Check that the file exists
			if _, e := os.Stat(filename); e == nil {
				// If it exists we load it into a buffer
				input, err = ioutil.ReadFile(filename)
				if err != nil {
					TermMessage(err)
					input = []byte{}
					filename = ""
				}
			}
			// If the file didn't exist, input will be empty, and we'll open an empty buffer
			buffers = append(buffers, NewBuffer(input, filename))
		}
	} else if !isatty.IsTerminal(os.Stdin.Fd()) {
		// Option 2
		// The input is not a terminal, so something is being piped in
		// and we should read from stdin
		input, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			TermMessage("Error reading from stdin: ", err)
			input = []byte{}
		}
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
	for _, v := range tabs[curTab].views {
		v.Display()
	}
	DisplayTabs()
	messenger.Display()
	screen.Show()
}

// Passing -version as a flag will have micro print out the version number
var flagVersion = flag.Bool("version", false, "Show the version number")

// Passing -cursor LINE,COL will start the cursor at position LINE,COL
var flagLineColumn = flag.String("cursor", "", "Start the cursor at position `LINE,COL`")

func main() {
	flag.Parse()
	if *flagVersion {
		// If -version was passed
		fmt.Println("Version:", Version)
		fmt.Println("Commit hash:", CommitHash)
		fmt.Println("Compiled on", CompileDate)
		os.Exit(0)
	}

	// Start the Lua VM for running plugins
	L = lua.NewState()
	defer L.Close()

	// Some encoding stuff in case the user isn't using UTF-8
	encoding.Register()
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	// Find the user's configuration directory (probably $XDG_CONFIG_HOME/micro)
	InitConfigDir()

	// Load the user's settings
	InitGlobalSettings()
	InitCommands()
	InitBindings()

	// Load the syntax files, including the colorscheme
	LoadSyntaxFiles()

	// Load the help files
	LoadHelp()

	// Start the screen
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

	// Create a new messenger
	// This is used for sending the user messages in the bottom of the editor
	messenger = new(Messenger)
	messenger.history = make(map[string][]string)

	// Now we load the input
	buffers := LoadInput()
	for _, buf := range buffers {
		// For each buffer we create a new tab and place the view in that tab
		tab := NewTabFromView(NewView(buf))
		tab.SetNum(len(tabs))
		tabs = append(tabs, tab)
		for _, t := range tabs {
			for _, v := range t.views {
				v.Center(false)
				if globalSettings["syntax"].(bool) {
					v.matches = Match(v)
				}
			}
		}
	}

	// Load all the plugin stuff
	// We give plugins access to a bunch of variables here which could be useful to them
	L.SetGlobal("OS", luar.New(L, runtime.GOOS))
	L.SetGlobal("tabs", luar.New(L, tabs))
	L.SetGlobal("curTab", luar.New(L, curTab))
	L.SetGlobal("messenger", luar.New(L, messenger))
	L.SetGlobal("GetOption", luar.New(L, GetOption))
	L.SetGlobal("AddOption", luar.New(L, AddOption))
	L.SetGlobal("SetOption", luar.New(L, SetOption))
	L.SetGlobal("SetLocalOption", luar.New(L, SetLocalOption))
	L.SetGlobal("BindKey", luar.New(L, BindKey))
	L.SetGlobal("MakeCommand", luar.New(L, MakeCommand))
	L.SetGlobal("CurView", luar.New(L, CurView))
	L.SetGlobal("IsWordChar", luar.New(L, IsWordChar))
	L.SetGlobal("HandleCommand", luar.New(L, HandleCommand))
	L.SetGlobal("HandleShellCommand", luar.New(L, HandleShellCommand))
	L.SetGlobal("GetLeadingWhitespace", luar.New(L, GetLeadingWhitespace))

	// Used for asynchronous jobs
	L.SetGlobal("JobStart", luar.New(L, JobStart))
	L.SetGlobal("JobSend", luar.New(L, JobSend))
	L.SetGlobal("JobStop", luar.New(L, JobStop))

	LoadPlugins()

	jobs = make(chan JobFunction, 100)
	events = make(chan tcell.Event)

	for _, t := range tabs {
		for _, v := range t.views {
			for _, pl := range loadedPlugins {
				_, err := Call(pl+".onViewOpen", v)
				if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
					TermMessage(err)
					continue
				}
			}
			if v.Buf.Settings["syntax"].(bool) {
				v.matches = Match(v)
			}
		}
	}

	// Here is the event loop which runs in a separate thread
	go func() {
		for {
			events <- screen.PollEvent()
		}
	}()

	for {
		// Display everything
		RedrawAll()

		var event tcell.Event

		// Check for new events
		select {
		case f := <-jobs:
			// If a new job has finished while running in the background we should execute the callback
			f.function(f.output, f.args...)
			continue
		case event = <-events:
		}

		switch e := event.(type) {
		case *tcell.EventMouse:
			if e.Buttons() == tcell.Button1 {
				// If the user left clicked we check a couple things
				_, h := screen.Size()
				x, y := e.Position()
				if y == h-1 && messenger.message != "" {
					// If the user clicked in the bottom bar, and there is a message down there
					// we copy it to the clipboard.
					// Often error messages are displayed down there so it can be useful to easily
					// copy the message
					clipboard.WriteAll(messenger.message)
					continue
				}

				// We loop through each view in the current tab and make sure the current view
				// it the one being clicked in
				for _, v := range tabs[curTab].views {
					if x >= v.x && x < v.x+v.width && y >= v.y && y < v.y+v.height {
						tabs[curTab].curView = v.Num
					}
				}
			}
		}

		// This function checks the mouse event for the possibility of changing the current tab
		// If the tab was changed it returns true
		if TabbarHandleMouseEvent(event) {
			continue
		}

		if searching {
			// Since searching is done in real time, we need to redraw every time
			// there is a new event in the search bar so we need a special function
			// to run instead of the standard HandleEvent.
			HandleSearchEvent(event, CurView())
		} else {
			// Send it to the view
			CurView().HandleEvent(event)
		}
	}
}
