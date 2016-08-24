package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
)

// The options that the user can set
var globalSettings map[string]interface{}

// InitSettings initializes the options map and sets all options to their default values
func InitSettings() {
	defaults := DefaultSettings()
	var parsed map[string]interface{}

	filename := configDir + "/settings.json"
	if _, e := os.Stat(filename); e == nil {
		input, err := ioutil.ReadFile(filename)
		if err != nil {
			TermMessage("Error reading settings.json file: " + err.Error())
			return
		}

		err = json.Unmarshal(input, &parsed)
		if err != nil {
			TermMessage("Error reading settings.json:", err.Error())
		}
	}

	globalSettings = make(map[string]interface{})
	for k, v := range defaults {
		globalSettings[k] = v
	}
	for k, v := range parsed {
		globalSettings[k] = v
	}

	err := WriteSettings(filename)
	if err != nil {
		TermMessage("Error writing settings.json file: " + err.Error())
	}
}

// WriteSettings writes the settings to the specified filename as JSON
func WriteSettings(filename string) error {
	var err error
	if _, e := os.Stat(configDir); e == nil {
		txt, _ := json.MarshalIndent(globalSettings, "", "    ")
		err = ioutil.WriteFile(filename, txt, 0644)
	}
	return err
}

// AddOption creates a new option. This is meant to be called by plugins to add options.
func AddOption(name string, value interface{}) {
	globalSettings[name] = value
	err := WriteSettings(configDir + "/settings.json")
	if err != nil {
		TermMessage("Error writing settings.json file: " + err.Error())
	}
}

// GetOption returns the specified option. This is meant to be called by plugins to add options.
func GetGlobalOption(name string) interface{} {
	return globalSettings[name]
}

func GetLocalOption(name string, buf *Buffer) interface{} {
	return buf.Settings[name]
}

func GetOption(name string) interface{} {
	if GetLocalOption(name, CurView().Buf) != nil {
		return GetLocalOption(name, CurView().Buf)
	}
	return GetGlobalOption(name)
}

// DefaultSettings returns the default settings for micro
func DefaultSettings() map[string]interface{} {
	return map[string]interface{}{
		"autoindent":   true,
		"colorscheme":  "monokai",
		"cursorline":   false,
		"ignorecase":   false,
		"indentchar":   " ",
		"ruler":        true,
		"savecursor":   false,
		"saveundo":     false,
		"scrollspeed":  float64(2),
		"scrollmargin": float64(3),
		"statusline":   true,
		"syntax":       true,
		"tabsize":      float64(4),
		"tabstospaces": false,
	}
}

// SetOption attempts to set the given option to the value
func SetGlobalOption(option, value string) error {
	if _, ok := globalSettings[option]; !ok {
		return errors.New("Invalid option")
	}

	kind := reflect.TypeOf(globalSettings[option]).Kind()
	if kind == reflect.Bool {
		b, err := ParseBool(value)
		if err != nil {
			return errors.New("Invalid value")
		}
		globalSettings[option] = b
	} else if kind == reflect.String {
		globalSettings[option] = value
	} else if kind == reflect.Float64 {
		i, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("Invalid value")
		}
		globalSettings[option] = float64(i)
	}

	for _, tab := range tabs {
		for _, view := range tab.views {
			SetLocalOption(option, value, view)
		}
	}

	return nil
}

func SetLocalOption(option, value string, view *View) error {
	buf := view.Buf
	if _, ok := buf.Settings[option]; !ok {
		return errors.New("Invalid option")
	}

	kind := reflect.TypeOf(buf.Settings[option]).Kind()
	if kind == reflect.Bool {
		b, err := ParseBool(value)
		if err != nil {
			return errors.New("Invalid value")
		}
		buf.Settings[option] = b
	} else if kind == reflect.String {
		buf.Settings[option] = value
	} else if kind == reflect.Float64 {
		i, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("Invalid value")
		}
		buf.Settings[option] = float64(i)
	}

	if option == "colorscheme" {
		LoadSyntaxFiles()
		buf.UpdateRules()
		if buf.Settings["syntax"].(bool) {
			view.matches = Match(view)
		}
	}

	if option == "statusline" {
		view.ToggleStatusLine()
		if buf.Settings["syntax"].(bool) {
			view.matches = Match(view)
		}
	}

	return nil
}

// SetOptionAndSettings sets the given option and saves the option setting to the settings config file
func SetOptionAndSettings(option, value string) {
	filename := configDir + "/settings.json"

	err := SetGlobalOption(option, value)

	if err != nil {
		messenger.Message(err.Error())
		return
	}

	err = WriteSettings(filename)
	if err != nil {
		messenger.Error("Error writing to settings.json: " + err.Error())
		return
	}
}
