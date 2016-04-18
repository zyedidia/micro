package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// The options that the user can set
var settings Settings

// All the possible settings
var possibleSettings = []string{"colorscheme", "tabsize", "autoindent", "syntax", "tabsToSpaces"}

// The Settings struct contains the settings for micro
type Settings struct {
	Colorscheme  string `json:"colorscheme"`
	TabSize      int    `json:"tabsize"`
	AutoIndent   bool   `json:"autoindent"`
	Syntax       bool   `json:"syntax"`
	TabsToSpaces bool   `json:"tabsToSpaces"`
}

// InitSettings initializes the options map and sets all options to their default values
func InitSettings() {
	filename := configDir + "/settings.json"
	if _, e := os.Stat(filename); e == nil {
		input, err := ioutil.ReadFile(filename)
		if err != nil {
			TermMessage("Error reading settings.json file: " + err.Error())
			return
		}

		json.Unmarshal(input, &settings)
	} else {
		settings = DefaultSettings()
		err := WriteSettings(filename)
		if err != nil {
			TermMessage("Error writing settings.json file: " + err.Error())
		}
	}
}

// WriteSettings writes the settings to the specified filename as JSON
func WriteSettings(filename string) error {
	var err error
	if _, e := os.Stat(configDir); e == nil {
		txt, _ := json.MarshalIndent(settings, "", "    ")
		err = ioutil.WriteFile(filename, txt, 0644)
	}
	return err
}

// DefaultSettings returns the default settings for micro
func DefaultSettings() Settings {
	return Settings{
		Colorscheme:  "default",
		TabSize:      4,
		AutoIndent:   true,
		Syntax:       true,
		TabsToSpaces: false,
	}
}

// SetOption prompts the user to set an option and checks that the response is valid
func SetOption(view *View, args []string) {
	filename := configDir + "/settings.json"
	if len(args) == 2 {
		option := strings.TrimSpace(args[0])
		value := strings.TrimSpace(args[1])

		if Contains(possibleSettings, option) {
			if option == "tabsize" {
				tsize, err := strconv.Atoi(value)
				if err != nil {
					messenger.Error("Invalid value for " + option)
					return
				}
				settings.TabSize = tsize
			} else if option == "colorscheme" {
				settings.Colorscheme = value
				LoadSyntaxFiles()
				view.buf.UpdateRules()
			} else if option == "syntax" {
				if value == "on" {
					settings.Syntax = true
				} else if value == "off" {
					settings.Syntax = false
				} else {
					messenger.Error("Invalid value for " + option)
					return
				}
				LoadSyntaxFiles()
				view.buf.UpdateRules()
			} else if option == "tabsToSpaces" {
				if value == "on" {
					settings.TabsToSpaces = true
				} else if value == "off" {
					settings.TabsToSpaces = false
				} else {
					messenger.Error("Invalid value for " + option)
					return
				}
			}
			err := WriteSettings(filename)
			if err != nil {
				messenger.Error("Error writing to settings.json: " + err.Error())
				return
			}
		} else {
			messenger.Error("Option " + option + " does not exist")
		}
	} else {
		messenger.Error("Invalid option, please use option value")
	}
}
