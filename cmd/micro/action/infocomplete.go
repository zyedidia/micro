package action

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/util"
)

// Completion represents a type of completion
type Completion int

const (
	NoCompletion Completion = iota
	FileCompletion
	CommandCompletion
	HelpCompletion
	OptionCompletion
	PluginCmdCompletion
	PluginNameCompletion
	OptionValueCompletion
)

var pluginCompletions []func(string) []string

// This file is meant (for now) for autocompletion in command mode, not
// while coding. This helps micro autocomplete commands and then filenames
// for example with `vsplit filename`.

// CommandComplete autocompletes commands
func CommandComplete(b *buffer.Buffer) (string, []string) {
	c := b.GetActiveCursor()
	l := b.LineBytes(c.Y)
	l = util.SliceStart(l, c.X)

	args := bytes.Split(l, []byte{' '})
	input := string(args[len(args)-1])
	argstart := 0
	for i, a := range args {
		if i == len(args)-1 {
			break
		}
		argstart += utf8.RuneCount(a) + 1
	}

	var suggestions []string
	for cmd := range commands {
		if strings.HasPrefix(cmd, input) {
			suggestions = append(suggestions, cmd)
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		chosen = util.SliceEndStr(suggestions[0], c.X-argstart)
	}
	return chosen, suggestions
}

// HelpComplete autocompletes help topics
func HelpComplete(input string) (string, []string) {
	var suggestions []string

	for _, file := range config.ListRuntimeFiles(config.RTHelp) {
		topic := file.Name()
		if strings.HasPrefix(topic, input) {
			suggestions = append(suggestions, topic)
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}
	return chosen, suggestions
}

// ColorschemeComplete tab-completes names of colorschemes.
func ColorschemeComplete(input string) (string, []string) {
	var suggestions []string
	files := config.ListRuntimeFiles(config.RTColorscheme)

	for _, f := range files {
		if strings.HasPrefix(f.Name(), input) {
			suggestions = append(suggestions, f.Name())
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}

	return chosen, suggestions
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// OptionComplete autocompletes options
func OptionComplete(input string) (string, []string) {
	var suggestions []string
	localSettings := config.DefaultLocalSettings()
	for option := range config.GlobalSettings {
		if strings.HasPrefix(option, input) {
			suggestions = append(suggestions, option)
		}
	}
	for option := range localSettings {
		if strings.HasPrefix(option, input) && !contains(suggestions, option) {
			suggestions = append(suggestions, option)
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}
	return chosen, suggestions
}

// OptionValueComplete completes values for various options
func OptionValueComplete(inputOpt, input string) (string, []string) {
	inputOpt = strings.TrimSpace(inputOpt)
	var suggestions []string
	localSettings := config.DefaultLocalSettings()
	var optionVal interface{}
	for k, option := range config.GlobalSettings {
		if k == inputOpt {
			optionVal = option
		}
	}
	for k, option := range localSettings {
		if k == inputOpt {
			optionVal = option
		}
	}

	switch optionVal.(type) {
	case bool:
		if strings.HasPrefix("on", input) {
			suggestions = append(suggestions, "on")
		} else if strings.HasPrefix("true", input) {
			suggestions = append(suggestions, "true")
		}
		if strings.HasPrefix("off", input) {
			suggestions = append(suggestions, "off")
		} else if strings.HasPrefix("false", input) {
			suggestions = append(suggestions, "false")
		}
	case string:
		switch inputOpt {
		case "colorscheme":
			_, suggestions = ColorschemeComplete(input)
		case "fileformat":
			if strings.HasPrefix("unix", input) {
				suggestions = append(suggestions, "unix")
			}
			if strings.HasPrefix("dos", input) {
				suggestions = append(suggestions, "dos")
			}
		case "sucmd":
			if strings.HasPrefix("sudo", input) {
				suggestions = append(suggestions, "sudo")
			}
			if strings.HasPrefix("doas", input) {
				suggestions = append(suggestions, "doas")
			}
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}
	return chosen, suggestions
}

// // MakeCompletion registers a function from a plugin for autocomplete commands
// func MakeCompletion(function string) Completion {
// 	pluginCompletions = append(pluginCompletions, LuaFunctionComplete(function))
// 	return Completion(-len(pluginCompletions))
// }
//
// // PluginComplete autocompletes from plugin function
// func PluginComplete(complete Completion, input string) (chosen string, suggestions []string) {
// 	idx := int(-complete) - 1
//
// 	if len(pluginCompletions) <= idx {
// 		return "", nil
// 	}
// 	suggestions = pluginCompletions[idx](input)
//
// 	if len(suggestions) == 1 {
// 		chosen = suggestions[0]
// 	}
// 	return
// }
//
// // PluginCmdComplete completes with possible choices for the `> plugin` command
// func PluginCmdComplete(input string) (chosen string, suggestions []string) {
// 	for _, cmd := range []string{"install", "remove", "search", "update", "list"} {
// 		if strings.HasPrefix(cmd, input) {
// 			suggestions = append(suggestions, cmd)
// 		}
// 	}
//
// 	if len(suggestions) == 1 {
// 		chosen = suggestions[0]
// 	}
// 	return chosen, suggestions
// }
//
// // PluginnameComplete completes with the names of loaded plugins
// func PluginNameComplete(input string) (chosen string, suggestions []string) {
// 	for _, pp := range GetAllPluginPackages() {
// 		if strings.HasPrefix(pp.Name, input) {
// 			suggestions = append(suggestions, pp.Name)
// 		}
// 	}
//
// 	if len(suggestions) == 1 {
// 		chosen = suggestions[0]
// 	}
// 	return chosen, suggestions
// }
