package screen

import (
	"fmt"
	"os"
	"sync"

	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/micro/internal/util"
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
var DrawChan chan bool

func Lock() {
	lock.Lock()
}

func Unlock() {
	lock.Unlock()
}

func Redraw() {
	DrawChan <- true
}

func ShowFakeCursor(x, y int) {
	r, _, _, _ := Screen.GetContent(x, y)
	Screen.SetContent(x, y, r, nil, config.DefStyle.Reverse(true))
}

func ShowCursor(x, y int) {
	if util.FakeCursor {
		ShowFakeCursor(x, y)
	} else {
		Screen.ShowCursor(x, y)
	}
}

// TempFini shuts the screen down temporarily
func TempFini() bool {
	screenWasNil := Screen == nil

	if !screenWasNil {
		Screen.Fini()
		Lock()
		Screen = nil
	}
	return screenWasNil
}

// TempStart restarts the screen after it was temporarily disabled
func TempStart(screenWasNil bool) {
	if !screenWasNil {
		Init()
		Unlock()
	}
}

// Init creates and initializes the tcell screen
func Init() {
	DrawChan = make(chan bool, 8)

	// Should we enable true color?
	truecolor := os.Getenv("MICRO_TRUECOLOR") == "1"

	if !truecolor {
		os.Setenv("TCELL_TRUECOLOR", "disable")
	}

	// Initilize tcell
	var err error
	Screen, err = tcell.NewScreen()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Fatal: Micro could not initialize a Screen.")
		os.Exit(1)
	}
	if err = Screen.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if config.GetGlobalOption("mouse").(bool) {
		Screen.EnableMouse()
	}
}
