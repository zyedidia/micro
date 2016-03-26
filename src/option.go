package main

import (
	"strconv"
	"strings"
)

// The options that the user can set
var options map[string]interface{}

func InitOptions() {
	options = make(map[string]interface{})
	options["tabsize"] = 4
	options["colorscheme"] = "default"
}

func SetOption(view *View) {
	choice, canceled := messenger.Prompt("Option: ")
	if !canceled {
		split := strings.Split(choice, "=")
		if len(split) == 2 {
			option := strings.TrimSpace(split[0])
			value := strings.TrimSpace(split[1])
			if _, exists := options[option]; exists {
				if option == "tabsize" {
					tsize, err := strconv.Atoi(value)
					if err != nil {
						messenger.Error("Invalid value for " + option)
						return
					}
					options[option] = tsize
				}
				if option == "colorscheme" {
					options[option] = value
					LoadSyntaxFiles()
					view.buf.UpdateRules()
				}
			} else {
				messenger.Error("Option " + option + " does not exist")
			}
		} else {
			messenger.Error("Invalid option, please use option = value")
		}
	}
}
