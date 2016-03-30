package main

import (
	"os"
	"strings"
)

// HandleCommand handles input from the user
func HandleCommand(input string, view *View) {
	cmd := strings.Split(input, " ")[0]
	args := strings.Split(input, " ")[1:]
	switch cmd {
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
