package util

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/zyedidia/micro/cmd/micro/screen"
)

// TermMessage sends a message to the user in the terminal. This usually occurs before
// micro has been fully initialized -- ie if there is an error in the syntax highlighting
// regular expressions
// The function must be called when the Screen is not initialized
// This will write the message, and wait for the user
// to press and key to continue
func TermMessage(msg ...interface{}) {
	screenWasNil := screen.Screen == nil
	if !screenWasNil {
		screen.Screen.Fini()
		screen.Screen = nil
	}

	fmt.Println(msg...)
	fmt.Print("\nPress enter to continue")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	if !screenWasNil {
		screen.Init()
	}
}

// TermError sends an error to the user in the terminal. Like TermMessage except formatted
// as an error
func TermError(filename string, lineNum int, err string) {
	TermMessage(filename + ", " + strconv.Itoa(lineNum) + ": " + err)
}
