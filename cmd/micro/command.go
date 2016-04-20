package main

import (
	"bytes"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
)

// RunShellCommand executes a shell command and returns the output/error
func RunShellCommand(input string) (string, error) {
	inputCmd := strings.Split(input, " ")[0]
	args := strings.Split(input, " ")[1:]

	cmd := exec.Command(inputCmd, args...)
	outputBytes := &bytes.Buffer{}

	cmd.Stdout = outputBytes // send output to buffer
	cmd.Start()
	err := cmd.Wait() // wait for command to finish
	outstring := outputBytes.String()
	return outstring, err
}

// HandleShellCommand runs the shell command and outputs to DisplayBlock
func HandleShellCommand(input string, view *View, openTerm bool) {
	inputCmd := strings.Split(input, " ")[0]
	if !openTerm {
		messenger.Message("Running...")
		go func() {
			output, err := RunShellCommand(input)
			totalLines := strings.Split(output, "\n")

			if len(totalLines) < 3 {
				if err == nil {
					messenger.Message(inputCmd, " exited without error")
				} else {
					messenger.Message(inputCmd, " exited with error: ", err)
				}
			} else {
				messenger.Message(output)
			}
			Redraw(view)
		}()
	} else {
		screen.Fini()

		args := strings.Split(input, " ")[1:]

		cmd := exec.Command(inputCmd, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for range c {
				cmd.Process.Kill()
			}
		}()

		cmd.Start()
		cmd.Wait()

		TermMessage("")

		InitScreen()
	}
}

// HandleCommand handles input from the user
func HandleCommand(input string, view *View) {
	inputCmd := strings.Split(input, " ")[0]
	args := strings.Split(input, " ")[1:]

	commands := []string{"set", "quit", "save", "replace", "run"}

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
	case "run":
		HandleShellCommand(strings.Join(args, " "), view, false)
	case "quit":
		if view.CanClose("Quit anyway? (yes, no, save) ") {
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
		messenger.Error("Unknown command: " + inputCmd)
	}
}
