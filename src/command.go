package main

import (
	"os"
	"strings"
)

// HandleCommand handles input from the user
func HandleCommand(input string, view *View) {
	inputCmd := strings.Split(input, " ")[0]
	args := strings.Split(input, " ")[1:]

	commands := []string{"set", "quit", "save"}

	i := 0
	cmd := inputCmd

	for _, c := range commands {
		if strings.HasPrefix(c, inputCmd) {
			i++
			cmd = c
		}
	}
	if i == 1 {
		inputCmd = cmd
	}

	switch inputCmd {
	case "set":
		SetOption(view, args)
	case "quit":
		if view.CanClose("Quit anyway? ") {
			screen.Fini()
			os.Exit(0)
		}
	case "save":
		view.Save()
	}
}
