package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/zyedidia/glob"
	"github.com/micro-editor/json5"
	"github.com/zyedidia/micro/v2/internal/util"
	"golang.org/x/text/encoding/htmlindex"
)

type optionValidator func(string, interface{}) error

// a list of settings that need option validators
var optionValidators = map[string]optionValidator{
	"autosave":        validateNonNegativeValue,
	"clipboard":       validateChoice,
	"colorcolumn":     validateNonNegativeValue,
	"colorscheme":     validateColorscheme,
	"detectlimit":     validateNonNegativeValue,
	"encoding":        validateEncoding,
	"fileformat":      validateChoice,
	"helpsplit":       validateChoice,
	"matchbracestyle": validateChoice,
	"multiopen":       validateChoice,
	"pageoverlap":     validateNonNegativeValue,
	"reload":          validateChoice,
	"scrollmargin":    validateNonNegativeValue,
	"scrollspeed":     validateNonNegativeValue,
	"tabsize":         validatePositiveValue,
}

// a list of settings with pre-defined choices
var OptionChoices = map[string][]string{
	"clipboard":       {"internal", "external", "terminal"},
	"fileformat":      {"unix", "dos"},
	"helpsplit":       {"hsplit", "vsplit"},
	"matchbracestyle": {"underline", "highlight"},
	"multiopen":       {"tab", "hsplit", "vsplit"},
	"reload":          {"prompt", "auto", "disabled"},
}

// a list of settings that can be globally and locally modified and their
// default values
var defaultCommonSettings = map[string]interface{}{
	"autoindent":      true,
	"autosu":          false,
	"backup":          true,
	"backupdir":       "",
	"basename":        false,
	"colorcolumn":     float64(0),
	"cursorline":      true,
	"detectlimit":     float64(100),
	"diffgutter":      false,
	"encoding":        "utf-8",
	"eofnewline":      true,
	"fastdirty":       false,
	"fileformat":      defaultFileFormat(),
	"filetype":        "unknown",
	"hlsearch":        false,
	"hltaberrors":     false,
	"hltrailingws":    false,
	"ignorecase":      true,
	"incsearch":       true,
	"indentchar":      " ",
	"keepautoindent":  false,
	"matchbrace":      true,
	"matchbraceleft":  true,
	"matchbracestyle": "underline",
	"mkparents":       false,
	"pageoverlap":     float64(2),
	"permbackup":      false,
	"readonly":        false,
	"relativeruler":   false,
	"reload":          "prompt",
	"rmtrailingws":    false,
	"ruler":           true,
	"savecursor":      false,
	"saveundo":        false,
	"scrollbar":       false,
	"scrollmargin":    float64(3),
	"scrollspeed":     float64(2),
	"smartpaste":      true,
	"softwrap":        false,
	"splitbottom":     true,
	"splitright":      true,
	"statusformatl":   "$(filename) $(modified)$(overwrite)($(line),$(col)) $(status.paste)| ft:$(opt:filetype) | $(opt:fileformat) | $(opt:encoding)",
	"statusformatr":   "$(bind:ToggleKeyMenu): bindings, $(bind:ToggleHelp): help",
	"statusline":      true,
	"syntax":          true,
	"tabmovement":     false,
	"tabsize":         float64(4),
	"tabstospaces":    false,
	"useprimary":      true,
	"wordwrap":        false,
}

// a list of settings that should only be globally modified and their
// default values
var DefaultGlobalOnlySettings = map[string]interface{}{
	"autosave":       float64(0),
	"clipboard":      "external",
	"colorscheme":    "default",
	"divchars":       "|-",
	"divreverse":     true,
	"fakecursor":     false,
	"helpsplit":      "hsplit",
	"infobar":        true,
	"keymenu":        false,
	"mouse":          true,
	"multiopen":      "tab",
	"parsecursor":    false,
	"paste":          false,
	"pluginchannels": []string{"https://raw.githubusercontent.com/micro-editor/plugin-channel/master/channel.json"},
	"pluginrepos":    []string{},
	"savehistory":    true,
	"scrollbarchar":  "|",
	"sucmd":          "sudo",
	"tabhighlight":   false,
	"tabreverse":     true,
	"xterm":          false,
}

// a list of settings that should never be globally modified
var LocalSettings = []string{
	"filetype",
	"readonly",
}

var (
	ErrInvalidOption = errors.New("Invalid option")
	ErrInvalidValue  = errors.New("Invalid value")

	// The options that the user can set
	GlobalSettings map[string]interface{}

	// This is the raw parsed json
	parsedSettings     map[string]interface{}
	settingsParseError bool

	// ModifiedSettings is a map of settings which should be written to disk
	// because they have been modified by the user in this session
	ModifiedSettings map[string]bool

	// VolatileSettings is a map of settings which should not be written to disk
	// because they have been temporarily set for this session only
	VolatileSettings map[string]bool
)

func init() {
	ModifiedSettings = make(map[string]bool)
	VolatileSettings = make(map[string]bool)
}

func validateParsedSettings() error {
	var err error
	defaults := DefaultAllSettings()
	for k, v := range parsedSettings {
		if strings.HasPrefix(reflect.TypeOf(v).String(), "map") {
			if strings.HasPrefix(k, "ft:") {
				for k1, v1 := range v.(map[string]interface{}) {
					if _, ok := defaults[k1]; ok {
						if e := verifySetting(k1, v1, defaults[k1]); e != nil {
							err = e
							parsedSettings[k].(map[string]interface{})[k1] = defaults[k1]
							continue
						}
					}
				}
			} else {
				if _, e := glob.Compile(k); e != nil {
					err = errors.New("Error with glob setting " + k + ": " + e.Error())
					delete(parsedSettings, k)
					continue
				}
				for k1, v1 := range v.(map[string]interface{}) {
					if _, ok := defaults[k1]; ok {
						if e := verifySetting(k1, v1, defaults[k1]); e != nil {
							err = e
							parsedSettings[k].(map[string]interface{})[k1] = defaults[k1]
							continue
						}
					}
				}
			}
			continue
		}

		if k == "autosave" {
			// if autosave is a boolean convert it to float
			s, ok := v.(bool)
			if ok {
				if s {
					parsedSettings["autosave"] = 8.0
				} else {
					parsedSettings["autosave"] = 0.0
				}
			}
			continue
		}
		if _, ok := defaults[k]; ok {
			if e := verifySetting(k, v, defaults[k]); e != nil {
				err = e
				parsedSettings[k] = defaults[k]
				continue
			}
		}
	}
	return err
}

func ReadSettings() error {
	parsedSettings = make(map[string]interface{})
	filename := filepath.Join(ConfigDir, "settings.json")
	if _, e := os.Stat(filename); e == nil {
		input, err := ioutil.ReadFile(filename)
		if err != nil {
			settingsParseError = true
			return errors.New("Error reading settings.json file: " + err.Error())
		}
		if !strings.HasPrefix(string(input), "null") {
			// Unmarshal the input into the parsed map
			err = json5.Unmarshal(input, &parsedSettings)
			if err != nil {
				settingsParseError = true
				return errors.New("Error reading settings.json: " + err.Error())
			}
			err = validateParsedSettings()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ParsedSettings() map[string]interface{} {
	s := make(map[string]interface{})
	for k, v := range parsedSettings {
		s[k] = v
	}
	return s
}

func verifySetting(option string, value interface{}, def interface{}) error {
	var interfaceArr []interface{}
	valType := reflect.TypeOf(value)
	defType := reflect.TypeOf(def)
	assignable := false

	switch option {
	case "pluginrepos", "pluginchannels":
		assignable = valType.AssignableTo(reflect.TypeOf(interfaceArr))
	default:
		assignable = defType.AssignableTo(valType)
	}
	if !assignable {
		return fmt.Errorf("Error: setting '%s' has incorrect type (%s), using default value: %v (%s)", option, valType, def, defType)
	}

	if err := OptionIsValid(option, value); err != nil {
		return err
	}

	return nil
}

// InitGlobalSettings initializes the options map and sets all options to their default values
// Must be called after ReadSettings
func InitGlobalSettings() error {
	var err error
	GlobalSettings = DefaultAllSettings()

	for k, v := range parsedSettings {
		if !strings.HasPrefix(reflect.TypeOf(v).String(), "map") {
			GlobalSettings[k] = v
		}
	}
	return err
}

// UpdatePathGlobLocals scans the already parsed settings and sets the options locally
// based on whether the path matches a glob
// Must be called after ReadSettings
func UpdatePathGlobLocals(settings map[string]interface{}, path string) {
	for k, v := range parsedSettings {
		if strings.HasPrefix(reflect.TypeOf(v).String(), "map") && !strings.HasPrefix(k, "ft:") {
			g, _ := glob.Compile(k)
			if g.MatchString(path) {
				for k1, v1 := range v.(map[string]interface{}) {
					settings[k1] = v1
				}
			}
		}
	}
}

// UpdateFileTypeLocals scans the already parsed settings and sets the options locally
// based on whether the filetype matches to "ft:"
// Must be called after ReadSettings
func UpdateFileTypeLocals(settings map[string]interface{}, filetype string) {
	for k, v := range parsedSettings {
		if strings.HasPrefix(reflect.TypeOf(v).String(), "map") && strings.HasPrefix(k, "ft:") {
			if filetype == k[3:] {
				for k1, v1 := range v.(map[string]interface{}) {
					if k1 != "filetype" {
						settings[k1] = v1
					}
				}
			}
		}
	}
}

func WriteJson(filename string, v interface{}) error {
		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		enc := json.NewEncoder(file)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "    ")
		return enc.Encode(v)
}

// WriteSettings writes the settings to the specified filename as JSON
func WriteSettings(filename string) error {
	if settingsParseError {
		// Don't write settings if there was a parse error
		// because this will delete the settings.json if it
		// is invalid. Instead we should allow the user to fix
		// it manually.
		return nil
	}

	var err error
	if _, e := os.Stat(ConfigDir); e == nil {
		defaults := DefaultAllSettings()

		// remove any options froms parsedSettings that have since been marked as default
		for k, v := range parsedSettings {
			if !strings.HasPrefix(reflect.TypeOf(v).String(), "map") {
				cur, okcur := GlobalSettings[k]
				_, vol := VolatileSettings[k]
				if def, ok := defaults[k]; ok && okcur && !vol && reflect.DeepEqual(cur, def) {
					delete(parsedSettings, k)
				}
			}
		}

		// add any options to parsedSettings that have since been marked as non-default
		for k, v := range GlobalSettings {
			if def, ok := defaults[k]; !ok || !reflect.DeepEqual(v, def) {
				if _, wr := ModifiedSettings[k]; wr {
					parsedSettings[k] = v
				}
			}
		}

		err = WriteJson(filename, parsedSettings)
	}
	return err
}

// OverwriteSettings writes the current settings to settings.json and
// resets any user configuration of local settings present in settings.json
func OverwriteSettings(filename string) error {
	settings := make(map[string]interface{})

	var err error
	if _, e := os.Stat(ConfigDir); e == nil {
		defaults := DefaultAllSettings()
		for k, v := range GlobalSettings {
			if def, ok := defaults[k]; !ok || !reflect.DeepEqual(v, def) {
				if _, wr := ModifiedSettings[k]; wr {
					settings[k] = v
				}
			}
		}

		err = WriteJson(filename, settings)
	}
	return err
}

// RegisterCommonOptionPlug creates a new option (called pl.name). This is meant to be called by plugins to add options.
func RegisterCommonOptionPlug(pl string, name string, defaultvalue interface{}) error {
	return RegisterCommonOption(pl+"."+name, defaultvalue)
}

// RegisterGlobalOptionPlug creates a new global-only option (named pl.name)
func RegisterGlobalOptionPlug(pl string, name string, defaultvalue interface{}) error {
	return RegisterGlobalOption(pl+"."+name, defaultvalue)
}

// RegisterCommonOption creates a new option
func RegisterCommonOption(name string, defaultvalue interface{}) error {
	if _, ok := GlobalSettings[name]; !ok {
		GlobalSettings[name] = defaultvalue
	}
	defaultCommonSettings[name] = defaultvalue
	return nil
}

// RegisterGlobalOption creates a new global-only option
func RegisterGlobalOption(name string, defaultvalue interface{}) error {
	if _, ok := GlobalSettings[name]; !ok {
		GlobalSettings[name] = defaultvalue
	}
	DefaultGlobalOnlySettings[name] = defaultvalue
	return nil
}

// GetGlobalOption returns the global value of the given option
func GetGlobalOption(name string) interface{} {
	return GlobalSettings[name]
}

func defaultFileFormat() string {
	if runtime.GOOS == "windows" {
		return "dos"
	}
	return "unix"
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

// DefaultCommonSettings returns a map of all common buffer settings
// and their default values
func DefaultCommonSettings() map[string]interface{} {
	commonsettings := make(map[string]interface{})
	for k, v := range defaultCommonSettings {
		commonsettings[k] = v
	}
	return commonsettings
}

// DefaultAllSettings returns a map of all common buffer & global-only settings
// and their default values
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
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, ErrInvalidValue
		}
		native = f
	} else {
		return nil, ErrInvalidValue
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
	nativeValue, ok := value.(float64)

	if !ok {
		return errors.New("Expected numeric type for " + option)
	}

	if nativeValue < 1 {
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

func validateChoice(option string, value interface{}) error {
	if choices, ok := OptionChoices[option]; ok {
		val, ok := value.(string)
		if !ok {
			return errors.New("Expected string type for " + option)
		}

		for _, v := range choices {
			if val == v {
				return nil
			}
		}

		choicesStr := strings.Join(choices, ", ")
		return errors.New(option + " must be one of: " + choicesStr)
	}

	return errors.New("Option has no pre-defined choices")
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

func validateEncoding(option string, value interface{}) error {
	_, err := htmlindex.Get(value.(string))
	return err
}
