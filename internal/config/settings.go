package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/zyedidia/glob"
	"github.com/zyedidia/json5"
	"github.com/zyedidia/micro/v2/internal/util"
	"golang.org/x/text/encoding/htmlindex"
)

type optionValidator func(string, interface{}) error

var (
	ErrInvalidOption = errors.New("Invalid option")
	ErrInvalidValue  = errors.New("Invalid value")

	// The options that the user can set
	GlobalSettings map[string]interface{}

	// This is the raw parsed json
	parsedSettings map[string]interface{}
)

func init() {
	parsedSettings = make(map[string]interface{})
}

// Options with validators
var optionValidators = map[string]optionValidator{
	"autosave":     validateNonNegativeValue,
	"tabsize":      validatePositiveValue,
	"scrollmargin": validateNonNegativeValue,
	"scrollspeed":  validateNonNegativeValue,
	"colorscheme":  validateColorscheme,
	"colorcolumn":  validateNonNegativeValue,
	"fileformat":   validateLineEnding,
	"encoding":     validateEncoding,
}

func ReadSettings() error {
	filename := filepath.Join(ConfigDir, "settings.json")
	if _, e := os.Stat(filename); e == nil {
		input, err := ioutil.ReadFile(filename)
		if err != nil {
			return errors.New("Error reading settings.json file: " + err.Error())
		}
		if !strings.HasPrefix(string(input), "null") {
			// Unmarshal the input into the parsed map
			err = json5.Unmarshal(input, &parsedSettings)
			if err != nil {
				return errors.New("Error reading settings.json: " + err.Error())
			}

			// check if autosave is a boolean and convert it to float if so
			if v, ok := parsedSettings["autosave"]; ok {
				s, ok := v.(bool)
				if ok {
					if s {
						parsedSettings["autosave"] = 8.0
					} else {
						parsedSettings["autosave"] = 0.0
					}
				}
			}
		}
	}
	return nil
}

// InitGlobalSettings initializes the options map and sets all options to their default values
// Must be called after ReadSettings
func InitGlobalSettings() {
	GlobalSettings = DefaultGlobalSettings()

	for k, v := range parsedSettings {
		if !strings.HasPrefix(reflect.TypeOf(v).String(), "map") {
			GlobalSettings[k] = v
		}
	}
}

// InitLocalSettings scans the json in settings.json and sets the options locally based
// on whether the filetype or path matches ft or glob local settings
// Must be called after ReadSettings
func InitLocalSettings(settings map[string]interface{}, path string) error {
	var parseError error
	for k, v := range parsedSettings {
		if strings.HasPrefix(reflect.TypeOf(v).String(), "map") {
			if strings.HasPrefix(k, "ft:") {
				if settings["filetype"].(string) == k[3:] {
					for k1, v1 := range v.(map[string]interface{}) {
						settings[k1] = v1
					}
				}
			} else {
				g, err := glob.Compile(k)
				if err != nil {
					parseError = errors.New("Error with glob setting " + k + ": " + err.Error())
					continue
				}

				if g.MatchString(path) {
					for k1, v1 := range v.(map[string]interface{}) {
						settings[k1] = v1
					}
				}
			}
		}
	}
	return parseError
}

// WriteSettings writes the settings to the specified filename as JSON
func WriteSettings(filename string) error {
	var err error
	if _, e := os.Stat(ConfigDir); e == nil {
		for k, v := range GlobalSettings {
			parsedSettings[k] = v
		}

		txt, _ := json.MarshalIndent(parsedSettings, "", "    ")
		err = ioutil.WriteFile(filename, append(txt, '\n'), 0644)
	}
	return err
}

func OverwriteSettings(filename string) error {
	var err error
	if _, e := os.Stat(ConfigDir); e == nil {
		txt, _ := json.MarshalIndent(GlobalSettings, "", "    ")
		err = ioutil.WriteFile(filename, append(txt, '\n'), 0644)
	}
	return err
}

// RegisterCommonOptionPlug creates a new option (called pl.name). This is meant to be called by plugins to add options.
func RegisterCommonOptionPlug(pl string, name string, defaultvalue interface{}) error {
	name = pl + "." + name
	if v, ok := GlobalSettings[name]; !ok {
		defaultCommonSettings[name] = defaultvalue
		GlobalSettings[name] = defaultvalue
		err := WriteSettings(filepath.Join(ConfigDir, "settings.json"))
		if err != nil {
			return errors.New("Error writing settings.json file: " + err.Error())
		}
	} else {
		defaultCommonSettings[name] = v
	}
	return nil
}

// RegisterGlobalOptionPlug creates a new global-only option (named pl.name)
func RegisterGlobalOptionPlug(pl string, name string, defaultvalue interface{}) error {
	return RegisterGlobalOption(pl+"."+name, defaultvalue)
}

// RegisterGlobalOption creates a new global-only option
func RegisterGlobalOption(name string, defaultvalue interface{}) error {
	if v, ok := GlobalSettings[name]; !ok {
		DefaultGlobalOnlySettings[name] = defaultvalue
		GlobalSettings[name] = defaultvalue
		err := WriteSettings(filepath.Join(ConfigDir, "settings.json"))
		if err != nil {
			return errors.New("Error writing settings.json file: " + err.Error())
		}
	} else {
		DefaultGlobalOnlySettings[name] = v
	}
	return nil
}

// GetGlobalOption returns the global value of the given option
func GetGlobalOption(name string) interface{} {
	return GlobalSettings[name]
}

var defaultCommonSettings = map[string]interface{}{
	"autoindent":     true,
	"autosu":         false,
	"backup":         true,
	"basename":       false,
	"colorcolumn":    float64(0),
	"cursorline":     true,
	"diffgutter":     false,
	"encoding":       "utf-8",
	"eofnewline":     true,
	"fastdirty":      false,
	"fileformat":     "unix",
	"filetype":       "unknown",
	"ignorecase":     false,
	"indentchar":     " ",
	"keepautoindent": false,
	"matchbrace":     true,
	"mkparents":      false,
	"readonly":       false,
	"rmtrailingws":   false,
	"ruler":          true,
	"relativeruler":  false,
	"savecursor":     false,
	"saveundo":       false,
	"scrollbar":      false,
	"scrollmargin":   float64(3),
	"scrollspeed":    float64(2),
	"smartpaste":     true,
	"softwrap":       false,
	"splitbottom":    true,
	"splitright":     true,
	"statusformatl":  "$(filename) $(modified)($(line),$(col)) $(status.paste)| ft:$(opt:filetype) | $(opt:fileformat) | $(opt:encoding)",
	"statusformatr":  "$(bind:ToggleKeyMenu): bindings, $(bind:ToggleHelp): help",
	"statusline":     true,
	"syntax":         true,
	"tabmovement":    false,
	"tabsize":        float64(4),
	"tabstospaces":   false,
	"useprimary":     true,
}

func GetInfoBarOffset() int {
	offset := 0
	if GetGlobalOption("infobar").(bool) {
		offset++
	}
	if GetGlobalOption("keymenu").(bool) {
		offset += 2
	}
	return offset
}

// DefaultCommonSettings returns the default global settings for micro
// Note that colorscheme is a global only option
func DefaultCommonSettings() map[string]interface{} {
	commonsettings := make(map[string]interface{})
	for k, v := range defaultCommonSettings {
		commonsettings[k] = v
	}
	return commonsettings
}

// a list of settings that should only be globally modified and their
// default values
var DefaultGlobalOnlySettings = map[string]interface{}{
	"autosave":       float64(0),
	"colorscheme":    "default",
	"divchars":       "|-",
	"divreverse":     true,
	"infobar":        true,
	"keymenu":        false,
	"mouse":          true,
	"parsecursor":    false,
	"paste":          false,
	"savehistory":    true,
	"sucmd":          "sudo",
	"pluginchannels": []string{"https://raw.githubusercontent.com/micro-editor/plugin-channel/master/channel.json"},
	"pluginrepos":    []string{},
	"xterm":          false,
}

// a list of settings that should never be globally modified
var LocalSettings = []string{
	"filetype",
	"readonly",
}

// DefaultGlobalSettings returns the default global settings for micro
// Note that colorscheme is a global only option
func DefaultGlobalSettings() map[string]interface{} {
	globalsettings := make(map[string]interface{})
	for k, v := range defaultCommonSettings {
		globalsettings[k] = v
	}
	for k, v := range DefaultGlobalOnlySettings {
		globalsettings[k] = v
	}
	return globalsettings
}

// DefaultAllSettings returns a map of all settings and their
// default values (both common and global settings)
func DefaultAllSettings() map[string]interface{} {
	allsettings := make(map[string]interface{})
	for k, v := range defaultCommonSettings {
		allsettings[k] = v
	}
	for k, v := range DefaultGlobalOnlySettings {
		allsettings[k] = v
	}
	return allsettings
}

// GetNativeValue parses and validates a value for a given option
func GetNativeValue(option string, realValue interface{}, value string) (interface{}, error) {
	var native interface{}
	kind := reflect.TypeOf(realValue).Kind()
	if kind == reflect.Bool {
		b, err := util.ParseBool(value)
		if err != nil {
			return nil, ErrInvalidValue
		}
		native = b
	} else if kind == reflect.String {
		native = value
	} else if kind == reflect.Float64 {
		i, err := strconv.Atoi(value)
		if err != nil {
			return nil, ErrInvalidValue
		}
		native = float64(i)
	} else {
		return nil, ErrInvalidValue
	}

	if err := OptionIsValid(option, native); err != nil {
		return nil, err
	}
	return native, nil
}

// OptionIsValid checks if a value is valid for a certain option
func OptionIsValid(option string, value interface{}) error {
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

func validateEncoding(option string, value interface{}) error {
	_, err := htmlindex.Get(value.(string))
	return err
}
