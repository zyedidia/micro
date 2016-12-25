package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-errors/errors"
	"layeh.com/gopher-luar"
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
	autosaveTime         = 8   // Number of seconds to wait before autosaving
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
	Version     = "0.0.0-unknown"
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
	events   chan tcell.Event
	autosave chan bool
)

// LoadInput determines which files should be loaded into buffers
// based on the input stored in flag.Args()
func LoadInput() []*Buffer {
	// There are a number of ways micro should start given its input

	// 1. If it is given a files in flag.Args(), it should open those

	// 2. If there is no input file and the input is not a terminal, that means
	// something is being piped in and the stdin should be opened in an
	// empty buffer

	// 3. If there is no input file and the input is a terminal, an empty buffer
	// should be opened

	var filename string
	var input []byte
	var err error
	var buffers []*Buffer

	if len(flag.Args()) > 0 {
		// Option 1
		// We go through each file and load it
		for i := 0; i < len(flag.Args()); i++ {
			filename = flag.Args()[i]

			// Check that the file exists
			var input *os.File
			if _, e := os.Stat(filename); e == nil {
				// If it exists we load it into a buffer
				input, err = os.Open(filename)
				stat, _ := input.Stat()
				defer input.Close()
				if err != nil {
					TermMessage(err)
					continue
				}
				if stat.IsDir() {
					TermMessage("Cannot read", filename, "because it is a directory")
					continue
				}
			}
			// If the file didn't exist, input will be empty, and we'll open an empty buffer
			if input != nil {
				buffers = append(buffers, NewBuffer(input, filename))
			} else {
				buffers = append(buffers, NewBuffer(strings.NewReader(""), filename))
			}
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
		buffers = append(buffers, NewBuffer(strings.NewReader(string(input)), filename))
	} else {
		// Option 3, just open an empty buffer
		buffers = append(buffers, NewBuffer(strings.NewReader(string(input)), filename))
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

func LoadAll() {
	// Find the user's configuration directory (probably $XDG_CONFIG_HOME/micro)
	InitConfigDir()

	// Build a list of available Extensions (Syntax, Colorscheme etc.)
	InitRuntimeFiles()

	// Load the user's settings
	InitGlobalSettings()

	InitCommands()
	InitBindings()

	LoadSyntaxFiles()

	for _, tab := range tabs {
		for _, v := range tab.views {
			v.Buf.UpdateRules()
			if v.Buf.Settings["syntax"].(bool) {
				v.matches = Match(v)
			}
		}
	}
}

// Passing -version as a flag will have micro print out the version number
var flagVersion = flag.Bool("version", false, "Show the version number and information")
var flagStartPos = flag.String("startpos", "", "LINE,COL to start the cursor at when opening a buffer.")

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: micro [OPTIONS] [FILE]...")
		fmt.Print("Micro's options can be set via command line arguments for quick adjustments. For real configuration, please use the bindings.json file (see 'help options').\n\n")
		flag.PrintDefaults()
	}

	optionFlags := make(map[string]*string)

	for k, v := range DefaultGlobalSettings() {
		optionFlags[k] = flag.String(k, "", fmt.Sprintf("The %s option. Default value: '%v'", k, v))
	}

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

	// Build a list of available Extensions (Syntax, Colorscheme etc.)
	InitRuntimeFiles()

	// Load the user's settings
	InitGlobalSettings()

	InitCommands()
	InitBindings()

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
	if len(buffers) == 0 {
		screen.Fini()
		os.Exit(1)
	}
	for _, buf := range buffers {
		// For each buffer we create a new tab and place the view in that tab
		tab := NewTabFromView(NewView(buf))
		tab.SetNum(len(tabs))
		tabs = append(tabs, tab)
		for _, t := range tabs {
			for _, v := range t.views {
				v.Center(false)
			}

			t.Resize()
		}
	}

	for k, v := range optionFlags {
		if *v != "" {
			SetOption(k, *v)
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
	L.SetGlobal("MakeCompletion", luar.New(L, MakeCompletion))
	L.SetGlobal("NewBuffer", luar.New(L, NewBufferFromString))
	L.SetGlobal("RuneStr", luar.New(L, func(r rune) string {
		return string(r)
	}))
	L.SetGlobal("Loc", luar.New(L, func(x, y int) Loc {
		return Loc{x, y}
	}))
	L.SetGlobal("JoinPaths", luar.New(L, filepath.Join))
	L.SetGlobal("DirectoryName", luar.New(L, filepath.Dir))
	L.SetGlobal("configDir", luar.New(L, configDir))
	L.SetGlobal("Reload", luar.New(L, LoadAll))
	L.SetGlobal("ByteOffset", luar.New(L, ByteOffset))
	L.SetGlobal("ToCharPos", luar.New(L, ToCharPos))

	// Used for asynchronous jobs
	L.SetGlobal("JobStart", luar.New(L, JobStart))
	L.SetGlobal("JobSpawn", luar.New(L, JobSpawn))
	L.SetGlobal("JobSend", luar.New(L, JobSend))
	L.SetGlobal("JobStop", luar.New(L, JobStop))

	// Extension Files
	L.SetGlobal("ReadRuntimeFile", luar.New(L, PluginReadRuntimeFile))
	L.SetGlobal("ListRuntimeFiles", luar.New(L, PluginListRuntimeFiles))
	L.SetGlobal("AddRuntimeFile", luar.New(L, PluginAddRuntimeFile))
	L.SetGlobal("AddRuntimeFilesFromDirectory", luar.New(L, PluginAddRuntimeFilesFromDirectory))

	jobs = make(chan JobFunction, 100)
	events = make(chan tcell.Event, 100)
	autosave = make(chan bool)

	LoadPlugins()

	// Load the syntax files, including the colorscheme
	LoadSyntaxFiles()

	for _, t := range tabs {
		for _, v := range t.views {
			v.Buf.FindFileType()
			v.Buf.UpdateRules()
			for pl := range loadedPlugins {
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

	go func() {
		for {
			time.Sleep(autosaveTime * time.Second)
			if globalSettings["autosave"].(bool) {
				autosave <- true
			}
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
		case <-autosave:
			CurView().Save(true)
		case event = <-events:
		}

		for event != nil {
			switch e := event.(type) {
			case *tcell.EventMouse:
				if e.Buttons() == tcell.Button1 {
					// If the user left clicked we check a couple things
					_, h := screen.Size()
					x, y := e.Position()
					if y == h-1 && messenger.message != "" && globalSettings["infobar"].(bool) {
						// If the user clicked in the bottom bar, and there is a message down there
						// we copy it to the clipboard.
						// Often error messages are displayed down there so it can be useful to easily
						// copy the message
						clipboard.WriteAll(messenger.message, "primary")
						break
					}

					if CurView().mouseReleased {
						// We loop through each view in the current tab and make sure the current view
						// is the one being clicked in
						for _, v := range tabs[curTab].views {
							if x >= v.x && x < v.x+v.Width && y >= v.y && y < v.y+v.Height {
								tabs[curTab].CurView = v.Num
							}
						}
					}
				}
			}

			// This function checks the mouse event for the possibility of changing the current tab
			// If the tab was changed it returns true
			if TabbarHandleMouseEvent(event) {
				break
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

			select {
			case event = <-events:
			default:
				event = nil
			}

		}
	}
}
