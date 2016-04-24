package main

var helpTxt string

// LoadHelp loads the help text from inside the binary
func LoadHelp() {
	data, err := Asset("runtime/help/help.md")
	if err != nil {
		TermMessage("Unable to load help text")
		return
	}
	helpTxt = string(data)
}
