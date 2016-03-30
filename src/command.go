package main

import (
	"strings"
)

// HandleCommand handles input from the user
func HandleCommand(input string, view *View) {
	cmd := strings.Split(input, " ")[0]
	args := strings.Split(input, " ")[1:]
	if cmd == "set" {
		SetOption(view, args)
	}
}
