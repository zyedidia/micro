package main

import (
	"io/ioutil"
	"os"
	"strings"
)

func FileComplete(input string) (string, []string) {
	dirs := strings.Split(input, "/")
	var files []os.FileInfo
	var err error
	if len(dirs) > 1 {
		files, err = ioutil.ReadDir(strings.Join(dirs[:len(dirs)-1], "/"))
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
			name += "/"
		}
		if strings.HasPrefix(name, dirs[len(dirs)-1]) {
			suggestions = append(suggestions, name)
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		if len(dirs) > 1 {
			chosen = strings.Join(dirs[:len(dirs)-1], "/") + "/" + suggestions[0]
		} else {
			chosen = suggestions[0]
		}
	}

	return chosen, suggestions
}

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

func HelpComplete(input string) (string, []string) {
	var suggestions []string

	for _, topic := range helpFiles {
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

func OptionComplete(input string) (string, []string) {
	var suggestions []string
	for option := range settings {
		if strings.HasPrefix(option, input) {
			suggestions = append(suggestions, option)
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}
	return chosen, suggestions
}
