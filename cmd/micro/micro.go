package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-errors/errors"
	isatty "github.com/mattn/go-isatty"
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
	// These variables should be set by the linker when compiling

	// Version is the version number or commit hash
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

// LoadInput determines which files should be loaded into buffers
// based on the input stored in flag.Args()
func LoadInput() []*buffer.Buffer {
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
	args := flag.Args()
	buffers := make([]*buffer.Buffer, 0, len(args))

	if len(args) > 0 {
		// Option 1
		// We go through each file and load it
		for i := 0; i < len(args); i++ {
			if strings.HasPrefix(args[i], "+") {
				if strings.Contains(args[i], ":") {
					split := strings.Split(args[i], ":")
					*flagStartPos = split[0][1:] + "," + split[1]
				} else {
					*flagStartPos = args[i][1:] + ",0"
				}
				continue
			}

			buf, err := buffer.NewBufferFromFile(args[i], buffer.BTDefault)
			if err != nil {
				util.TermMessage(err)
				continue
			}
			// If the file didn't exist, input will be empty, and we'll open an empty buffer
			buffers = append(buffers, buf)
		}
	} else if !isatty.IsTerminal(os.Stdin.Fd()) {
		// Option 2
		// The input is not a terminal, so something is being piped in
		// and we should read from stdin
		input, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			util.TermMessage("Error reading from stdin: ", err)
			input = []byte{}
		}
		buffers = append(buffers, buffer.NewBufferFromString(string(input), filename, buffer.BTDefault))
	} else {
		// Option 3, just open an empty buffer
		buffers = append(buffers, buffer.NewBufferFromString(string(input), filename, buffer.BTDefault))
	}

	return buffers
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
	action.InitCommands()

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

	b := LoadInput()[0]
	width, height := screen.Screen.Size()
	ep := action.NewBufEditPane(0, 0, width, height-1, b)

	action.InitGlobals()

	// Here is the event loop which runs in a separate thread
	go func() {
		events = make(chan tcell.Event)
		for {
			screen.Lock()
			events <- screen.Screen.PollEvent()
			screen.Unlock()
		}
	}()

	for {
		// Display everything
		screen.Screen.Fill(' ', config.DefStyle)
		screen.Screen.HideCursor()
		ep.Display()
		action.InfoBar.Display()
		screen.Screen.Show()

		var event tcell.Event

		// Check for new events
		select {
		case event = <-events:
		}

		if event != nil {
			if action.InfoBar.HasPrompt {
				action.InfoBar.HandleEvent(event)
			} else {
				ep.HandleEvent(event)
			}
		}
	}
}
