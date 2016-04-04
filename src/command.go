package main

import (
	"os"
	"regexp"
	"strings"
)

// HandleCommand handles input from the user
func HandleCommand(input string, view *View) {
	inputCmd := strings.Split(input, " ")[0]
	args := strings.Split(input, " ")[1:]

	commands := []string{"set", "quit", "save"}

	i := 0
	cmd := inputCmd

	for _, c := range commands {
		if strings.HasPrefix(c, inputCmd) {
			i++
			cmd = c
		}
	}
	if i == 1 {
		inputCmd = cmd
	}

	switch inputCmd {
	case "set":
		SetOption(view, args)
	case "quit":
		if view.CanClose("Quit anyway? ") {
			screen.Fini()
			os.Exit(0)
		}
	case "save":
		view.Save()
	case "replace":
		r := regexp.MustCompile(`"[^"\\]*(?:\\.[^"\\]*)*"|[^\s]*`)
		replaceCmd := r.FindAllString(strings.Join(args, " "), -1)
		if len(replaceCmd) < 2 {
			messenger.Error("Invalid replace statement: " + strings.Join(args, " "))
			return
		}

		var flags string
		if len(replaceCmd) == 3 {
			// The user included some flags
			flags = replaceCmd[2]
		}

		search := string(replaceCmd[0])
		replace := string(replaceCmd[1])

		if strings.HasPrefix(search, `"`) && strings.HasSuffix(search, `"`) {
			search = search[1 : len(search)-1]
		}
		if strings.HasPrefix(replace, `"`) && strings.HasSuffix(replace, `"`) {
			replace = replace[1 : len(replace)-1]
		}

		search = strings.Replace(search, `\"`, `"`, -1)
		replace = strings.Replace(replace, `\"`, `"`, -1)

		// messenger.Error(search + " -> " + replace)

		regex, err := regexp.Compile(search)
		if err != nil {
			messenger.Error(err.Error())
			return
		}

		found := false
		for {
			match := regex.FindStringIndex(view.buf.text)
			if match == nil {
				break
			}
			found = true
			if strings.Contains(flags, "c") {
				// 	// The 'check' flag was used
				// 	if messenger.YesNoPrompt("Perform replacement?") {
				// 		view.eh.Replace(match[0], match[1], replace)
				// 	} else {
				// 		continue
				// 	}
			}
			view.eh.Replace(match[0], match[1], replace)
		}
		if !found {
			messenger.Message("Nothing matched " + search)
		}
	default:
		messenger.Error("Unknown command: " + cmd)
	}
}
