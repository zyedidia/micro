package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/flynn/json5"
	"github.com/zyedidia/glob"
)

type optionValidator func(string, interface{}) error

// The options that the user can set
var globalSettings map[string]interface{}

var invalidSettings bool

// Options with validators
var optionValidators = map[string]optionValidator{
	"tabsize":      validatePositiveValue,
	"scrollmargin": validateNonNegativeValue,
	"scrollspeed":  validateNonNegativeValue,
	"colorscheme":  validateColorscheme,
	"colorcolumn":  validateNonNegativeValue,
	"fileformat":   validateLineEnding,
}

// InitGlobalSettings initializes the options map and sets all options to their default values
func InitGlobalSettings() {
	invalidSettings = false
	defaults := DefaultGlobalSettings()
	var parsed map[string]interface{}

	filename := configDir + "/settings.json"
	writeSettings := false
	if _, e := os.Stat(filename); e == nil {
		input, err := ioutil.ReadFile(filename)
		if !strings.HasPrefix(string(input), "null") {
			if err != nil {
				TermMessage("Error reading settings.json file: " + err.Error())
				invalidSettings = true
				return
			}

			err = json5.Unmarshal(input, &parsed)
			if err != nil {
				TermMessage("Error reading settings.json:", err.Error())
				invalidSettings = true
			}
		} else {
			writeSettings = true
		}
	}

	globalSettings = make(map[string]interface{})
	for k, v := range defaults {
		globalSettings[k] = v
	}
	for k, v := range parsed {
		if !strings.HasPrefix(reflect.TypeOf(v).String(), "map") {
			globalSettings[k] = v
		}
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) || writeSettings {
		err := WriteSettings(filename)
		if err != nil {
			TermMessage("Error writing settings.json file: " + err.Error())
		}
	}
}

// InitLocalSettings scans the json in settings.json and sets the options locally based
// on whether the buffer matches the glob
func InitLocalSettings(buf *Buffer) {
	invalidSettings = false
	var parsed map[string]interface{}

	filename := configDir + "/settings.json"
	if _, e := os.Stat(filename); e == nil {
		input, err := ioutil.ReadFile(filename)
		if err != nil {
			TermMessage("Error reading settings.json file: " + err.Error())
			invalidSettings = true
			return
		}

		err = json5.Unmarshal(input, &parsed)
		if err != nil {
			TermMessage("Error reading settings.json:", err.Error())
			invalidSettings = true
		}
	}

	for k, v := range parsed {
		if strings.HasPrefix(reflect.TypeOf(v).String(), "map") {
			if strings.HasPrefix(k, "ft:") {
				if buf.Settings["filetype"].(string) == k[3:] {
					for k1, v1 := range v.(map[string]interface{}) {
						buf.Settings[k1] = v1
					}
				}
			} else {
				g, err := glob.Compile(k)
				if err != nil {
					TermMessage("Error with glob setting ", k, ": ", err)
					continue
				}

				if g.MatchString(buf.Path) {
					for k1, v1 := range v.(map[string]interface{}) {
						buf.Settings[k1] = v1
					}
				}
			}
		}
	}
}

// WriteSettings writes the settings to the specified filename as JSON
func WriteSettings(filename string) error {
	if invalidSettings {
		// Do not write the settings if there was an error when reading them
		return nil
	}

	var err error
	if _, e := os.Stat(configDir); e == nil {
		parsed := make(map[string]interface{})

		filename := configDir + "/settings.json"
		for k, v := range globalSettings {
			parsed[k] = v
		}
		if _, e := os.Stat(filename); e == nil {
			input, err := ioutil.ReadFile(filename)
			if string(input) != "null" {
				if err != nil {
					return err
				}

				err = json5.Unmarshal(input, &parsed)
				if err != nil {
					TermMessage("Error reading settings.json:", err.Error())
					invalidSettings = true
				}

				for k, v := range parsed {
					if !strings.HasPrefix(reflect.TypeOf(v).String(), "map") {
						if _, ok := globalSettings[k]; ok {
							parsed[k] = globalSettings[k]
						}
					}
				}
			}
		}

		txt, _ := json.MarshalIndent(parsed, "", "    ")
		err = ioutil.WriteFile(filename, append(txt, '\n'), 0644)
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

// GetGlobalOption returns the global value of the given option
func GetGlobalOption(name string) interface{} {
	return globalSettings[name]
}

// GetLocalOption returns the local value of the given option
func GetLocalOption(name string, buf *Buffer) interface{} {
	return buf.Settings[name]
}

// GetOption returns the value of the given option
// If there is a local version of the option, it returns that
// otherwise it will return the global version
func GetOption(name string) interface{} {
	if GetLocalOption(name, CurView().Buf) != nil {
		return GetLocalOption(name, CurView().Buf)
	}
	return GetGlobalOption(name)
}

// DefaultGlobalSettings returns the default global settings for micro
// Note that colorscheme is a global only option
func DefaultGlobalSettings() map[string]interface{} {
	return map[string]interface{}{
		"autoindent":     true,
		"autosave":       false,
		"basename":       false,
		"colorcolumn":    float64(0),
		"colorscheme":    "default",
		"cursorline":     true,
		"eofnewline":     false,
		"fastdirty":      true,
		"fileformat":     "unix",
		"hidehelp":       false,
		"ignorecase":     false,
		"indentchar":     " ",
		"infobar":        true,
		"keepautoindent": false,
		"keymenu":        false,
		"matchbrace":     false,
		"matchbraceleft": false,
		"mouse":          true,
		"pluginchannels": []string{"https://raw.githubusercontent.com/micro-editor/plugin-channel/master/channel.json"},
		"pluginrepos":    []string{},
		"rmtrailingws":   false,
		"ruler":          true,
		"savecursor":     false,
		"savehistory":    true,
		"saveundo":       false,
		"scrollbar":      false,
		"scrollmargin":   float64(3),
		"scrollspeed":    float64(2),
		"softwrap":       false,
		"smartpaste":     true,
		"splitbottom":    true,
		"splitright":     true,
		"statusline":     true,
		"sucmd":          "sudo",
		"syntax":         true,
		"tabmovement":    false,
		"tabsize":        float64(4),
		"tabstospaces":   false,
		"termtitle":      false,
		"useprimary":     true,
	}
}

// DefaultLocalSettings returns the default local settings
// Note that filetype is a local only option
func DefaultLocalSettings() map[string]interface{} {
	return map[string]interface{}{
		"autoindent":     true,
		"autosave":       false,
		"basename":       false,
		"colorcolumn":    float64(0),
		"cursorline":     true,
		"eofnewline":     false,
		"fastdirty":      true,
		"fileformat":     "unix",
		"filetype":       "Unknown",
		"hidehelp":       false,
		"ignorecase":     false,
		"indentchar":     " ",
		"keepautoindent": false,
		"matchbrace":     false,
		"matchbraceleft": false,
		"rmtrailingws":   false,
		"ruler":          true,
		"savecursor":     false,
		"saveundo":       false,
		"scrollbar":      false,
		"scrollmargin":   float64(3),
		"scrollspeed":    float64(2),
		"softwrap":       false,
		"smartpaste":     true,
		"splitbottom":    true,
		"splitright":     true,
		"statusline":     true,
		"syntax":         true,
		"tabmovement":    false,
		"tabsize":        float64(4),
		"tabstospaces":   false,
		"useprimary":     true,
	}
}

// SetOption attempts to set the given option to the value
// By default it will set the option as global, but if the option
// is local only it will set the local version
// Use setlocal to force an option to be set locally
func SetOption(option, value string) error {
	if _, ok := globalSettings[option]; !ok {
		if _, ok := CurView().Buf.Settings[option]; !ok {
			return errors.New("Invalid option")
		}
		SetLocalOption(option, value, CurView())
		return nil
	}

	var nativeValue interface{}

	kind := reflect.TypeOf(globalSettings[option]).Kind()
	if kind == reflect.Bool {
		b, err := ParseBool(value)
		if err != nil {
			return errors.New("Invalid value")
		}
		nativeValue = b
	} else if kind == reflect.String {
		nativeValue = value
	} else if kind == reflect.Float64 {
		i, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("Invalid value")
		}
		nativeValue = float64(i)
	} else {
		return errors.New("Option has unsupported value type")
	}

	if err := optionIsValid(option, nativeValue); err != nil {
		return err
	}

	globalSettings[option] = nativeValue

	if option == "colorscheme" {
		// LoadSyntaxFiles()
		InitColorscheme()
		for _, tab := range tabs {
			for _, view := range tab.Views {
				view.Buf.UpdateRules()
			}
		}
	}

	if option == "infobar" || option == "keymenu" {
		for _, tab := range tabs {
			tab.Resize()
		}
	}

	if option == "mouse" {
		if !nativeValue.(bool) {
			screen.DisableMouse()
		} else {
			screen.EnableMouse()
		}
	}

	if len(tabs) != 0 {
		if _, ok := CurView().Buf.Settings[option]; ok {
			for _, tab := range tabs {
				for _, view := range tab.Views {
					SetLocalOption(option, value, view)
				}
			}
		}
	}

	return nil
}

// SetLocalOption sets the local version of this option
func SetLocalOption(option, value string, view *View) error {
	buf := view.Buf
	if _, ok := buf.Settings[option]; !ok {
		return errors.New("Invalid option")
	}

	var nativeValue interface{}

	kind := reflect.TypeOf(buf.Settings[option]).Kind()
	if kind == reflect.Bool {
		b, err := ParseBool(value)
		if err != nil {
			return errors.New("Invalid value")
		}
		nativeValue = b
	} else if kind == reflect.String {
		nativeValue = value
	} else if kind == reflect.Float64 {
		i, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("Invalid value")
		}
		nativeValue = float64(i)
	} else {
		return errors.New("Option has unsupported value type")
	}

	if err := optionIsValid(option, nativeValue); err != nil {
		return err
	}

	if option == "fastdirty" {
		// If it is being turned off, we have to hash every open buffer
		var empty [md5.Size]byte
		var wg sync.WaitGroup

		for _, tab := range tabs {
			for _, v := range tab.Views {
				if !nativeValue.(bool) {
					if v.Buf.origHash == empty {
						wg.Add(1)

						go func(b *Buffer) { // calculate md5 hash of the file
							defer wg.Done()

							if file, e := os.Open(b.AbsPath); e == nil {
								defer file.Close()

								h := md5.New()

								if _, e = io.Copy(h, file); e == nil {
									h.Sum(b.origHash[:0])
								}
							}
						}(v.Buf)
					}
				} else {
					v.Buf.IsModified = v.Buf.Modified()
				}
			}
		}

		wg.Wait()
	}

	buf.Settings[option] = nativeValue

	if option == "statusline" {
		view.ToggleStatusLine()
	}

	if option == "filetype" {
		// LoadSyntaxFiles()
		InitColorscheme()
		buf.UpdateRules()
	}

	if option == "fileformat" {
		buf.IsModified = true
	}

	if option == "syntax" {
		if !nativeValue.(bool) {
			buf.ClearMatches()
		} else {
			if buf.highlighter != nil {
				buf.highlighter.HighlightStates(buf)
			}
		}
	}

	return nil
}

// SetOptionAndSettings sets the given option and saves the option setting to the settings config file
func SetOptionAndSettings(option, value string) {
	filename := configDir + "/settings.json"

	err := SetOption(option, value)

	if err != nil {
		messenger.Error(err.Error())
		return
	}

	err = WriteSettings(filename)
	if err != nil {
		messenger.Error("Error writing to settings.json: " + err.Error())
		return
	}
}

func optionIsValid(option string, value interface{}) error {
	if validator, ok := optionValidators[option]; ok {
		return validator(option, value)
	}

	return nil
}

// Option validators

func validatePositiveValue(option string, value interface{}) error {
	tabsize, ok := value.(float64)

	if !ok {
		return errors.New("Expected numeric type for " + option)
	}

	if tabsize < 1 {
		return errors.New(option + " must be greater than 0")
	}

	return nil
}

func validateNonNegativeValue(option string, value interface{}) error {
	nativeValue, ok := value.(float64)

	if !ok {
		return errors.New("Expected numeric type for " + option)
	}

	if nativeValue < 0 {
		return errors.New(option + " must be non-negative")
	}

	return nil
}

func validateColorscheme(option string, value interface{}) error {
	colorscheme, ok := value.(string)

	if !ok {
		return errors.New("Expected string type for colorscheme")
	}

	if !ColorschemeExists(colorscheme) {
		return errors.New(colorscheme + " is not a valid colorscheme")
	}

	return nil
}

func validateLineEnding(option string, value interface{}) error {
	endingType, ok := value.(string)

	if !ok {
		return errors.New("Expected string type for file format")
	}

	if endingType != "unix" && endingType != "dos" {
		return errors.New("File format must be either 'unix' or 'dos'")
	}

	return nil
}
