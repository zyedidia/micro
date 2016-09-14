package main

import (
	"io/ioutil"
)

type HelpPage interface {
	HelpFile() ([]byte, error)
}

var helpPages map[string]HelpPage = map[string]HelpPage{
	"help":        assetHelpPage("help"),
	"keybindings": assetHelpPage("keybindings"),
	"plugins":     assetHelpPage("plugins"),
	"colors":      assetHelpPage("colors"),
	"options":     assetHelpPage("options"),
	"commands":    assetHelpPage("commands"),
	"tutorial":    assetHelpPage("tutorial"),
}

type assetHelpPage string

func (file assetHelpPage) HelpFile() ([]byte, error) {
	return Asset("runtime/help/" + string(file) + ".md")
}

type fileHelpPage string

func (file fileHelpPage) HelpFile() ([]byte, error) {
	return ioutil.ReadFile(string(file))
}

func AddPluginHelp(name, file string) {
	if _, exists := helpPages[name]; exists {
		return
	}
	helpPages[name] = fileHelpPage(file)
}
