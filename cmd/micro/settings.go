package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// The options that the user can set
var settings Settings

// All the possible settings
// This map maps the name of the setting in the Settings struct
// to the name that the user will actually use (the one in the json file)
var possibleSettings = map[string]string{
	"colorscheme":  "Colorscheme",
	"tabsize":      "TabSize",
	"autoindent":   "AutoIndent",
	"syntax":       "Syntax",
	"tabsToSpaces": "TabsToSpaces",
	"ruler":        "Ruler",
	"gofmt":        "GoFmt",
	"goimports":    "GoImports",
	"multicursor":  "MultiCursor",
}

// The Settings struct contains the settings for micro
type Settings struct {
	Colorscheme  string `json:"colorscheme"`
	TabSize      int    `json:"tabsize"`
	AutoIndent   bool   `json:"autoindent"`
	Syntax       bool   `json:"syntax"`
	TabsToSpaces bool   `json:"tabsToSpaces"`
	Ruler        bool   `json:"ruler"`
	GoFmt        bool   `json:"gofmt"`
	GoImports    bool   `json:"goimports"`
	MultiCursor  bool   `json:"multiCursor"`
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

		err = json.Unmarshal(input, &settings)
		if err != nil {
			TermMessage("Error reading settings.json:", err.Error())
		}
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
		Ruler:        true,
		GoFmt:        false,
		GoImports:    false,
		MultiCursor:  false,
	}
}

// SetOption prompts the user to set an option and checks that the response is valid
func SetOption(view *View, args []string) {
	filename := configDir + "/settings.json"
	if len(args) == 2 {
		option := strings.TrimSpace(args[0])
		value := strings.TrimSpace(args[1])

		mutable := reflect.ValueOf(&settings).Elem()
		field := mutable.FieldByName(possibleSettings[option])
		if !field.IsValid() {
			messenger.Error(option + " is not a valid option")
			return
		}
		kind := field.Type().Kind()
		if kind == reflect.Bool {
			b, err := ParseBool(value)
			if err != nil {
				messenger.Error("Invalid value for " + option)
				return
			}
			field.SetBool(b)
		} else if kind == reflect.String {
			field.SetString(value)
		} else if kind == reflect.Int {
			i, err := strconv.Atoi(value)
			if err != nil {
				messenger.Error("Invalid value for " + option)
				return
			}
			field.SetInt(int64(i))
		}

		if option == "colorscheme" {
			LoadSyntaxFiles()
			view.buf.UpdateRules()
		}

		err := WriteSettings(filename)
		if err != nil {
			messenger.Error("Error writing to settings.json: " + err.Error())
			return
		}
	} else {
		messenger.Error("No value given")
	}
}
