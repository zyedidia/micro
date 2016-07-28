package main

var helpPages map[string]string

var helpFiles = []string{
	"help",
	"keybindings",
	"plugins",
	"colors",
	"options",
	"commands",
}

// LoadHelp loads the help text from inside the binary
func LoadHelp() {
	helpPages = make(map[string]string)
	for _, file := range helpFiles {
		data, err := Asset("runtime/help/" + file + ".md")
		if err != nil {
			TermMessage("Unable to load help text", file)
		}
		helpPages[file] = string(data)
	}
}
