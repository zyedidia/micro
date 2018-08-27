package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-errors/errors"
	"github.com/zyedidia/micro/cmd/micro/action"
	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/micro/cmd/micro/util"
	"github.com/zyedidia/tcell"
)

const (
	doubleClickThreshold = 400 // How many milliseconds to wait before a second click is not a double click
	autosaveTime         = 8   // Number of seconds to wait before autosaving
)

var (
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

	for k, v := range config.DefaultGlobalSettings() {
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
		for k, v := range config.DefaultGlobalSettings() {
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
	err = config.InitConfigDir(*flagConfigDir)
	if err != nil {
		util.TermMessage(err)
	}
	config.InitRuntimeFiles()
	err = config.ReadSettings()
	if err != nil {
		util.TermMessage(err)
	}
	config.InitGlobalSettings()
	action.InitBindings()
	err = config.InitColorscheme()
	if err != nil {
		util.TermMessage(err)
	}

	screen.Init()

	// If we have an error, we can exit cleanly and not completely
	// mess up the terminal being worked in
	// In other words we need to shut down tcell before the program crashes
	defer func() {
		if err := recover(); err != nil {
			screen.Screen.Fini()
			fmt.Println("Micro encountered an error:", err)
			// Print the stack trace too
			fmt.Print(errors.Wrap(err, 2).ErrorStack())
			os.Exit(1)
		}
	}()

	action.TryBindKey("Ctrl-z", "Undo", true)

	b, err := buffer.NewBufferFromFile(os.Args[1])

	if err != nil {
		util.TermMessage(err)
	}

	width, height := screen.Screen.Size()
	w := NewWindow(0, 0, width, height-1, b)

	a := action.NewBufHandler(b)

	// Here is the event loop which runs in a separate thread
	go func() {
		events = make(chan tcell.Event)
		for {
			// TODO: fix race condition with screen.Screen = nil
			events <- screen.Screen.PollEvent()
		}
	}()

	for {
		// Display everything
		screen.Screen.Fill(' ', config.DefStyle)
		w.DisplayBuffer()
		w.DisplayStatusLine()
		screen.Screen.Show()

		var event tcell.Event

		// Check for new events
		select {
		case event = <-events:
		}

		if event != nil {
			a.HandleEvent(event)
		}
	}

	screen.Screen.Fini()
}
