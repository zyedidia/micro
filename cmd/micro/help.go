package main

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
