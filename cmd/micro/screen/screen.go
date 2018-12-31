package screen

import (
	"fmt"
	"os"
	"sync"

	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/terminfo"
	"github.com/zyedidia/tcell"
)

// Screen is the tcell screen we use to draw to the terminal
// Synchronization is used because we poll the screen on a separate
// thread and sometimes the screen is shut down by the main thread
// (for example on TermMessage) so we don't want to poll a nil/shutdown
// screen. TODO: maybe we should worry about polling and drawing at the
// same time too.
var Screen tcell.Screen
var lock sync.Mutex

func Lock() {
	lock.Lock()
}

func Unlock() {
	lock.Unlock()
}

var screenWasNil bool

// TempFini shuts the screen down temporarily
func TempFini() {
	screenWasNil = Screen == nil

	if !screenWasNil {
		Lock()
		Screen.Fini()
		Screen = nil
	}
}

// TempStart restarts the screen after it was temporarily disabled
func TempStart() {
	if !screenWasNil {
		Init()
		Unlock()
	}
}

// Init creates and initializes the tcell screen
func Init() {
	// Should we enable true color?
	truecolor := os.Getenv("MICRO_TRUECOLOR") == "1"

	tcelldb := os.Getenv("TCELLDB")
	os.Setenv("TCELLDB", config.ConfigDir+"/.tcelldb")

	// In order to enable true color, we have to set the TERM to `xterm-truecolor` when
	// initializing tcell, but after that, we can set the TERM back to whatever it was
	oldTerm := os.Getenv("TERM")
	if truecolor {
		os.Setenv("TERM", "xterm-truecolor")
	}

	// Initilize tcell
	var err error
	Screen, err = tcell.NewScreen()
	if err != nil {
		if err == tcell.ErrTermNotFound {
			err = terminfo.WriteDB(config.ConfigDir + "/.tcelldb")
			if err != nil {
				fmt.Println(err)
				fmt.Println("Fatal: Micro could not create terminal database file", config.ConfigDir+"/.tcelldb")
				os.Exit(1)
			}
			Screen, err = tcell.NewScreen()
			if err != nil {
				fmt.Println(err)
				fmt.Println("Fatal: Micro could not initialize a Screen.")
				os.Exit(1)
			}
		} else {
			fmt.Println(err)
			fmt.Println("Fatal: Micro could not initialize a Screen.")
			os.Exit(1)
		}
	}
	if err = Screen.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Now we can put the TERM back to what it was before
	if truecolor {
		os.Setenv("TERM", oldTerm)
	}

	if config.GetGlobalOption("mouse").(bool) {
		Screen.EnableMouse()
	}

	os.Setenv("TCELLDB", tcelldb)
}
