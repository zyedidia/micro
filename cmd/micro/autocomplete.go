package main

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
)

var pluginCompletions []func(string) []string

// This file is meant (for now) for autocompletion in command mode, not
// while coding. This helps micro autocomplete commands and then filenames
// for example with `vsplit filename`.

// FileComplete autocompletes filenames
func FileComplete(input string) (string, []string) {
	var sep string = string(os.PathSeparator)
	dirs := strings.Split(input, sep)

	var files []os.FileInfo
	var err error
	if len(dirs) > 1 {
		home, _ := homedir.Dir()

		directories := strings.Join(dirs[:len(dirs)-1], sep) + sep

		if strings.HasPrefix(directories, "~") {
			directories = strings.Replace(directories, "~", home, 1)
		}
		files, err = ioutil.ReadDir(directories)
	} else {
		files, err = ioutil.ReadDir(".")
	}

	var suggestions []string
	if err != nil {
		return "", suggestions
	}
	for _, f := range files {
		name := f.Name()
		if f.IsDir() {
			name += sep
		}
		if strings.HasPrefix(name, dirs[len(dirs)-1]) {
			suggestions = append(suggestions, name)
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		if len(dirs) > 1 {
			chosen = strings.Join(dirs[:len(dirs)-1], sep) + sep + suggestions[0]
		} else {
			chosen = suggestions[0]
		}
	} else {
		if len(dirs) > 1 {
			chosen = strings.Join(dirs[:len(dirs)-1], sep) + sep
		}
	}

	return chosen, suggestions
}

// CommandComplete autocompletes commands
func CommandComplete(input string) (string, []string) {
	var suggestions []string
	for cmd := range commands {
		if strings.HasPrefix(cmd, input) {
			suggestions = append(suggestions, cmd)
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}
	return chosen, suggestions
}

// HelpComplete autocompletes help topics
func HelpComplete(input string) (string, []string) {
	var suggestions []string

	for _, file := range ListRuntimeFiles(FILE_Help) {
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
	localSettings := DefaultLocalSettings()
	for option := range globalSettings {
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

// MakeCompletion registeres a function from a plugin for autocomplete commands
func MakeCompletion(function string) Completion {
	pluginCompletions = append(pluginCompletions, LuaFunctionComplete(function))
	return Completion(-len(pluginCompletions))
}

// PluginComplete autocompletes from plugin function
func PluginComplete(complete Completion, input string) (chosen string, suggestions []string) {
	idx := int(-complete) - 1

	if len(pluginCompletions) <= idx {
		return "", nil
	}
	suggestions = pluginCompletions[idx](input)

	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}
	return
}
