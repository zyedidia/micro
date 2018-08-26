package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-errors/errors"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/zyedidia/micro/cmd/micro/terminfo"
	"github.com/zyedidia/tcell"
)

const (
	doubleClickThreshold = 400 // How many milliseconds to wait before a second click is not a double click
	undoThreshold        = 500 // If two events are less than n milliseconds apart, undo both of them
	autosaveTime         = 8   // Number of seconds to wait before autosaving
)

var (
	// The main screen
	screen tcell.Screen

	// Where the user's configuration is
	// This should be $XDG_CONFIG_HOME/micro
	// If $XDG_CONFIG_HOME is not set, it is ~/.config/micro
	configDir string

	// Version is the version number or commit hash
	// These variables should be set by the linker when compiling
	Version = "0.0.0-unknown"
	// CommitHash is the commit this version was built on
	CommitHash = "Unknown"
	// CompileDate is the date this binary was compiled on
	CompileDate = "Unknown"
	// Debug logging
	Debug = "ON"

	// Event channel
	events   chan tcell.Event
	autosave chan bool

	// How many redraws have happened
	numRedraw uint

	// Command line flags
	flagVersion   = flag.Bool("version", false, "Show the version number and information")
	flagStartPos  = flag.String("startpos", "", "LINE,COL to start the cursor at when opening a buffer.")
	flagConfigDir = flag.String("config-dir", "", "Specify a custom location for the configuration directory")
	flagOptions   = flag.Bool("options", false, "Show all option help")
)

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

	if len(*flagConfigDir) > 0 {
		if _, err := os.Stat(*flagConfigDir); os.IsNotExist(err) {
			TermMessage("Error: " + *flagConfigDir + " does not exist. Defaulting to " + configDir + ".")
		} else {
			configDir = *flagConfigDir
			return
		}
	}

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

	tcelldb := os.Getenv("TCELLDB")
	os.Setenv("TCELLDB", configDir+"/.tcelldb")

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
		if err == tcell.ErrTermNotFound {
			err = terminfo.WriteDB(configDir + "/.tcelldb")
			if err != nil {
				fmt.Println(err)
				fmt.Println("Fatal: Micro could not create tcelldb")
				os.Exit(1)
			}
			screen, err = tcell.NewScreen()
			if err != nil {
				fmt.Println(err)
				fmt.Println("Fatal: Micro could not initialize a screen.")
				os.Exit(1)
			}
		} else {
			fmt.Println(err)
			fmt.Println("Fatal: Micro could not initialize a screen.")
			os.Exit(1)
		}
	}
	if err = screen.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Now we can put the TERM back to what it was before
	if truecolor {
		os.Setenv("TERM", oldTerm)
	}

	if GetGlobalOption("mouse").(bool) {
		screen.EnableMouse()
	}

	os.Setenv("TCELLDB", tcelldb)

	// screen.SetStyle(defStyle)
}

func InitFlags() {
	flag.Usage = func() {
		fmt.Println("Usage: micro [OPTIONS] [FILE]...")
		fmt.Println("-config-dir dir")
		fmt.Println("    \tSpecify a custom location for the configuration directory")
		fmt.Println("-startpos LINE,COL")
		fmt.Println("+LINE:COL")
		fmt.Println("    \tSpecify a line and column to start the cursor at when opening a buffer")
		fmt.Println("    \tThis can also be done by opening file:LINE:COL")
		fmt.Println("-options")
		fmt.Println("    \tShow all option help")
		fmt.Println("-version")
		fmt.Println("    \tShow the version number and information")

		fmt.Print("\nMicro's options can also be set via command line arguments for quick\nadjustments. For real configuration, please use the settings.json\nfile (see 'help options').\n\n")
		fmt.Println("-option value")
		fmt.Println("    \tSet `option` to `value` for this session")
		fmt.Println("    \tFor example: `micro -syntax off file.c`")
		fmt.Println("\nUse `micro -options` to see the full list of configuration options")
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

	if *flagOptions {
		// If -options was passed
		for k, v := range DefaultGlobalSettings() {
			fmt.Printf("-%s value\n", k)
			fmt.Printf("    \tDefault value: '%v'\n", v)
		}
		os.Exit(0)
	}
}

func main() {
	var err error

	InitLog()
	InitFlags()
	InitConfigDir()
	InitRuntimeFiles()
	err = ReadSettings()
	if err != nil {
		TermMessage(err)
	}
	InitGlobalSettings()
	err = InitColorscheme()
	if err != nil {
		TermMessage(err)
	}

	InitScreen()

	// If we have an error, we can exit cleanly and not completely
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

	b, err := NewBufferFromFile(os.Args[1])

	if err != nil {
		TermMessage(err)
	}

	width, height := screen.Size()

	w := NewWindow(0, 0, width/2, height/2, b)

	for i := 0; i < 5; i++ {
		screen.Clear()
		w.DisplayBuffer()
		w.DisplayStatusLine()
		screen.Show()
		time.Sleep(200 * time.Millisecond)
		w.StartLine++
	}

	// time.Sleep(2 * time.Second)

	screen.Fini()
}
