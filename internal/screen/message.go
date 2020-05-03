package screen

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// TermMessage sends a message to the user in the terminal. This usually occurs before
// micro has been fully initialized -- ie if there is an error in the syntax highlighting
// regular expressions
// The function must be called when the Screen is not initialized
// This will write the message, and wait for the user
// to press and key to continue
func TermMessage(msg ...interface{}) {
	screenb := TempFini()

	fmt.Println(msg...)
	fmt.Print("\nPress enter to continue")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	TempStart(screenb)
}

// TermPrompt prints a prompt and requests the user for a response
// The result is matched against a list of options and the index of
// the match is returned
// If wait is true, the prompt re-prompts until a valid option is
// chosen, otherwise if wait is false, -1 is returned for no match
func TermPrompt(prompt string, options []string, wait bool) int {
	screenb := TempFini()

	idx := -1
	// same behavior as do { ... } while (wait && idx == -1)
	for ok := true; ok; ok = wait && idx == -1 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(prompt)
		resp, _ := reader.ReadString('\n')
		resp = strings.TrimSpace(resp)

		for i, opt := range options {
			if resp == opt {
				idx = i
			}
		}

		if wait && idx == -1 {
			fmt.Println("\nInvalid choice.")
		}
	}

	TempStart(screenb)

	return idx
}

// TermError sends an error to the user in the terminal. Like TermMessage except formatted
// as an error
func TermError(filename string, lineNum int, err string) {
	TermMessage(filename + ", " + strconv.Itoa(lineNum) + ": " + err)
}

// Password a password
type Password struct {
	Secret   string
	Prompted bool
}

// TermPassword gets the password for the encrypted file
func TermPassword(filename string) (pass Password) {
	if _, e := os.Stat(filename); e != nil {
		return
	}

	for {
		fmt.Printf("Password for %v: ", filename)
		password, err := terminal.ReadPassword(syscall.Stdin)
		fmt.Printf("\n")
		if err != nil {
			fmt.Println(err)
			continue
		}
		pass.Secret = string(password)
		pass.Prompted = true
		break
	}

	return pass
}
