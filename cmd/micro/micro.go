package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/go-errors/errors"
	isatty "github.com/mattn/go-isatty"
	"github.com/zyedidia/micro/internal/action"
	"github.com/zyedidia/micro/internal/buffer"
	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/micro/internal/screen"
	"github.com/zyedidia/micro/internal/shell"
	"github.com/zyedidia/micro/internal/util"
	"github.com/zyedidia/tcell"
)

var (
	// Event channel
	events   chan tcell.Event
	autosave chan bool

	// Command line flags
	flagVersion   = flag.Bool("version", false, "Show the version number and information")
	flagConfigDir = flag.String("config-dir", "", "Specify a custom location for the configuration directory")
	flagOptions   = flag.Bool("options", false, "Show all option help")
	optionFlags   map[string]*string
)

func InitFlags() {
	flag.Usage = func() {
		fmt.Println("Usage: micro [OPTIONS] [FILE]...")
		fmt.Println("-config-dir dir")
		fmt.Println("    \tSpecify a custom location for the configuration directory")
		fmt.Println("[FILE]:LINE:COL")
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

	optionFlags = make(map[string]*string)

	for k, v := range config.DefaultAllSettings() {
		optionFlags[k] = flag.String(k, "", fmt.Sprintf("The %s option. Default value: '%v'.", k, v))
	}

	flag.Parse()

	if *flagVersion {
		// If -version was passed
		fmt.Println("Version:", util.Version)
		fmt.Println("Commit hash:", util.CommitHash)
		fmt.Println("Compiled on", util.CompileDate)
		os.Exit(0)
	}

	if *flagOptions {
		// If -options was passed
		var keys []string
		m := config.DefaultAllSettings()
		for k, _ := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := m[k]
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
			buf, err := buffer.NewBufferFromFile(args[i], buffer.BTDefault)
			if err != nil {
				screen.TermMessage(err)
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
			screen.TermMessage("Error reading from stdin: ", err)
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
		screen.TermMessage(err)
	}

	config.InitRuntimeFiles()
	err = config.ReadSettings()
	if err != nil {
		screen.TermMessage(err)
	}
	config.InitGlobalSettings()

	// flag options
	for k, v := range optionFlags {
		if *v != "" {
			nativeValue, err := config.GetNativeValue(k, config.DefaultAllSettings()[k], *v)
			if err != nil {
				screen.TermMessage(err)
				continue
			}
			config.GlobalSettings[k] = nativeValue
		}
	}

	action.InitBindings()
	action.InitCommands()

	err = config.InitColorscheme()
	if err != nil {
		screen.TermMessage(err)
	}

	err = config.LoadAllPlugins()
	if err != nil {
		screen.TermMessage(err)
	}
	err = config.RunPluginFn("init")
	if err != nil {
		screen.TermMessage(err)
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

	b := LoadInput()
	action.InitTabs(b)
	action.InitGlobals()

	// Here is the event loop which runs in a separate thread
	go func() {
		events = make(chan tcell.Event)
		for {
			screen.Lock()
			e := screen.Screen.PollEvent()
			screen.Unlock()
			if e != nil {
				events <- e
			}
		}
	}()

	for {
		// Display everything
		screen.Screen.Fill(' ', config.DefStyle)
		screen.Screen.HideCursor()
		action.Tabs.Display()
		for _, ep := range action.MainTab().Panes {
			ep.Display()
		}
		action.MainTab().Display()
		action.InfoBar.Display()
		screen.Screen.Show()

		var event tcell.Event

		// Check for new events
		select {
		case f := <-shell.Jobs:
			// If a new job has finished while running in the background we should execute the callback
			f.Function(f.Output, f.Args...)
		case <-config.Autosave:
			for _, b := range buffer.OpenBuffers {
				b.Save()
			}
		case <-shell.CloseTerms:
		case event = <-events:
		case <-screen.DrawChan:
		}

		if action.InfoBar.HasPrompt {
			action.InfoBar.HandleEvent(event)
		} else {
			action.Tabs.HandleEvent(event)
		}
	}
}
