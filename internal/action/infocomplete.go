package action

import (
	"bytes"
	"sort"
	"strings"

	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/util"
)

// This file is meant (for now) for autocompletion in command mode, not
// while coding. This helps micro autocomplete commands and then filenames
// for example with `vsplit filename`.

// CommandComplete autocompletes commands
func CommandComplete(b *buffer.Buffer) ([]string, []string) {
	c := b.GetActiveCursor()
	input, argstart := buffer.GetArg(b)

	var suggestions []string
	for cmd := range commands {
		if strings.HasPrefix(cmd, input) {
			suggestions = append(suggestions, cmd)
		}
	}

	sort.Strings(suggestions)
	completions := make([]string, len(suggestions))
	for i := range suggestions {
		completions[i] = util.SliceEndStr(suggestions[i], c.X-argstart)
	}

	return completions, suggestions
}

// HelpComplete autocompletes help topics
func HelpComplete(b *buffer.Buffer) ([]string, []string) {
	c := b.GetActiveCursor()
	input, argstart := buffer.GetArg(b)

	var suggestions []string

	for _, file := range config.ListRuntimeFiles(config.RTHelp) {
		topic := file.Name()
		if strings.HasPrefix(topic, input) {
			suggestions = append(suggestions, topic)
		}
	}

	sort.Strings(suggestions)
	completions := make([]string, len(suggestions))
	for i := range suggestions {
		completions[i] = util.SliceEndStr(suggestions[i], c.X-argstart)
	}
	return completions, suggestions
}

// colorschemeComplete tab-completes names of colorschemes.
// This is just a heper value for OptionValueComplete
func colorschemeComplete(input string) (string, []string) {
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
func OptionComplete(b *buffer.Buffer) ([]string, []string) {
	c := b.GetActiveCursor()
	input, argstart := buffer.GetArg(b)

	var suggestions []string
	for option := range config.GlobalSettings {
		if strings.HasPrefix(option, input) {
			suggestions = append(suggestions, option)
		}
	}
	// for option := range localSettings {
	// 	if strings.HasPrefix(option, input) && !contains(suggestions, option) {
	// 		suggestions = append(suggestions, option)
	// 	}
	// }

	sort.Strings(suggestions)
	completions := make([]string, len(suggestions))
	for i := range suggestions {
		completions[i] = util.SliceEndStr(suggestions[i], c.X-argstart)
	}
	return completions, suggestions
}

// OptionValueComplete completes values for various options
func OptionValueComplete(b *buffer.Buffer) ([]string, []string) {
	c := b.GetActiveCursor()
	l := b.LineBytes(c.Y)
	l = util.SliceStart(l, c.X)
	input, argstart := buffer.GetArg(b)

	completeValue := false
	args := bytes.Split(l, []byte{' '})
	if len(args) >= 2 {
		// localSettings := config.DefaultLocalSettings()
		for option := range config.GlobalSettings {
			if option == string(args[len(args)-2]) {
				completeValue = true
				break
			}
		}
		// for option := range localSettings {
		// 	if option == string(args[len(args)-2]) {
		// 		completeValue = true
		// 		break
		// 	}
		// }
	}
	if !completeValue {
		return OptionComplete(b)
	}

	inputOpt := string(args[len(args)-2])

	inputOpt = strings.TrimSpace(inputOpt)
	var suggestions []string
	// localSettings := config.DefaultLocalSettings()
	var optionVal interface{}
	for k, option := range config.GlobalSettings {
		if k == inputOpt {
			optionVal = option
		}
	}
	// for k, option := range localSettings {
	// 	if k == inputOpt {
	// 		optionVal = option
	// 	}
	// }

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
			_, suggestions = colorschemeComplete(input)
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
	sort.Strings(suggestions)

	completions := make([]string, len(suggestions))
	for i := range suggestions {
		completions[i] = util.SliceEndStr(suggestions[i], c.X-argstart)
	}
	return completions, suggestions
}

// PluginCmdComplete autocompletes the plugin command
func PluginCmdComplete(b *buffer.Buffer) ([]string, []string) {
	c := b.GetActiveCursor()
	input, argstart := buffer.GetArg(b)

	var suggestions []string
	for _, cmd := range PluginCmds {
		if strings.HasPrefix(cmd, input) {
			suggestions = append(suggestions, cmd)
		}
	}

	sort.Strings(suggestions)
	completions := make([]string, len(suggestions))
	for i := range suggestions {
		completions[i] = util.SliceEndStr(suggestions[i], c.X-argstart)
	}
	return completions, suggestions
}

// PluginComplete completes values for the plugin command
func PluginComplete(b *buffer.Buffer) ([]string, []string) {
	c := b.GetActiveCursor()
	l := b.LineBytes(c.Y)
	l = util.SliceStart(l, c.X)
	input, argstart := buffer.GetArg(b)

	completeValue := false
	args := bytes.Split(l, []byte{' '})
	if len(args) >= 2 {
		for _, cmd := range PluginCmds {
			if cmd == string(args[len(args)-2]) {
				completeValue = true
				break
			}
		}
	}
	if !completeValue {
		return PluginCmdComplete(b)
	}

	var suggestions []string
	for _, pl := range config.Plugins {
		if strings.HasPrefix(pl.Name, input) {
			suggestions = append(suggestions, pl.Name)
		}
	}
	sort.Strings(suggestions)

	completions := make([]string, len(suggestions))
	for i := range suggestions {
		completions[i] = util.SliceEndStr(suggestions[i], c.X-argstart)
	}
	return completions, suggestions
}

// PluginNameComplete completes with the names of loaded plugins
// func PluginNameComplete(b *buffer.Buffer) ([]string, []string) {
// 	c := b.GetActiveCursor()
// 	input, argstart := buffer.GetArg(b)
//
// 	var suggestions []string
// 	for _, pp := range config.GetAllPluginPackages(nil) {
// 		if strings.HasPrefix(pp.Name, input) {
// 			suggestions = append(suggestions, pp.Name)
// 		}
// 	}
//
// 	sort.Strings(suggestions)
// 	completions := make([]string, len(suggestions))
// 	for i := range suggestions {
// 		completions[i] = util.SliceEndStr(suggestions[i], c.X-argstart)
// 	}
// 	return completions, suggestions
// }

// // MakeCompletion registers a function from a plugin for autocomplete commands
// func MakeCompletion(function string) Completion {
// 	pluginCompletions = append(pluginCompletions, LuaFunctionComplete(function))
// 	return Completion(-len(pluginCompletions))
// }
