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
	cmd.Stdout = outputBytes
	cmd.Stderr = outputBytes
	cmd.Start()
	err := cmd.Wait() // wait for command to finish
	outstring := outputBytes.String()
	return outstring, err
}

// HandleShellCommand runs the shell command
// The openTerm argument specifies whether a terminal should be opened (for viewing output
// or interacting with stdin)
func HandleShellCommand(input string, openTerm bool) {
	inputCmd := strings.Split(input, " ")[0]
	if !openTerm {
		// Simply run the command in the background and notify the user when it's done
		messenger.Message("Running...")
		go func() {
			output, err := RunShellCommand(input)
			totalLines := strings.Split(output, "\n")

			if len(totalLines) < 3 {
				if err == nil {
					messenger.Message(inputCmd, " exited without error")
				} else {
					messenger.Message(inputCmd, " exited with error: ", err, ": ", output)
				}
			} else {
				messenger.Message(output)
			}
			// We have to make sure to redraw
			RedrawAll()
		}()
	} else {
		// Shut down the screen because we're going to interact directly with the shell
		screen.Fini()
		screen = nil

		args := strings.Split(input, " ")[1:]

		// Set up everything for the command
		cmd := exec.Command(inputCmd, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// This is a trap for Ctrl-C so that it doesn't kill micro
		// Instead we trap Ctrl-C to kill the program we're running
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for range c {
				cmd.Process.Kill()
			}
		}()

		// Start the command
		cmd.Start()
		cmd.Wait()

		// This is just so we don't return right away and let the user press enter to return
		TermMessage("")

		// Start the screen back up
		InitScreen()
	}
}

// HandleCommand handles input from the user
func HandleCommand(input string) {
	inputCmd := strings.Split(input, " ")[0]
	args := strings.Split(input, " ")[1:]

	switch inputCmd {
	case "set":
		// Set an option and we have to set it for every view
		for _, view := range views {
			SetOption(view, args)
		}
	case "run":
		// Run a shell command in the background (openTerm is false)
		HandleShellCommand(strings.Join(args, " "), false)
	case "quit":
		// This is a bit weird because micro only has one view for now so there is no way to close
		// a single view
		// Currently if multiple views were open, it would close all of them, and not check the non-mainviews
		// for unsaved changes. This, and the behavior of Ctrl-Q need to be changed when splits are implemented
		if views[mainView].CanClose("Quit anyway? (yes, no, save) ") {
			screen.Fini()
			os.Exit(0)
		}
	case "save":
		// Save the main view
		views[mainView].Save()
	case "replace":
		// This is a regex to parse the replace expression
		// We allow no quotes if there are no spaces, but if you want to search
		// for or replace an expression with spaces, you can add double quotes
		r := regexp.MustCompile(`"[^"\\]*(?:\\.[^"\\]*)*"|[^\s]*`)
		replaceCmd := r.FindAllString(strings.Join(args, " "), -1)
		if len(replaceCmd) < 2 {
			// We need to find both a search and replace expression
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

		// If the search and replace expressions have quotes, we need to remove those
		if strings.HasPrefix(search, `"`) && strings.HasSuffix(search, `"`) {
			search = search[1 : len(search)-1]
		}
		if strings.HasPrefix(replace, `"`) && strings.HasSuffix(replace, `"`) {
			replace = replace[1 : len(replace)-1]
		}

		// We replace all escaped double quotes to real double quotes
		search = strings.Replace(search, `\"`, `"`, -1)
		replace = strings.Replace(replace, `\"`, `"`, -1)
		// Replace some things so users can actually insert newlines and tabs in replacements
		replace = strings.Replace(replace, "\\n", "\n", -1)
		replace = strings.Replace(replace, "\\t", "\t", -1)

		regex, err := regexp.Compile(search)
		if err != nil {
			// There was an error with the user's regex
			messenger.Error(err.Error())
			return
		}

		view := views[mainView]

		found := false
		for {
			match := regex.FindStringIndex(view.Buf.String())
			if match == nil {
				break
			}
			found = true
			if strings.Contains(flags, "c") {
				// The 'check' flag was used
				Search(search, view, true)
				view.Relocate()
				if settings["syntax"].(bool) {
					view.matches = Match(view)
				}
				RedrawAll()
				choice, canceled := messenger.YesNoPrompt("Perform replacement? (y,n)")
				if canceled {
					if view.Cursor.HasSelection() {
						view.Cursor.SetLoc(view.Cursor.curSelection[0])
						view.Cursor.ResetSelection()
					}
					messenger.Reset()
					return
				}
				if choice {
					view.Cursor.DeleteSelection()
					view.Buf.Insert(match[0], replace)
					view.Cursor.ResetSelection()
					messenger.Reset()
				} else {
					if view.Cursor.HasSelection() {
						searchStart = view.Cursor.curSelection[1]
					} else {
						searchStart = ToCharPos(view.Cursor.x, view.Cursor.y, view.Buf)
					}
					continue
				}
			} else {
				view.Buf.Replace(match[0], match[1], replace)
			}
		}
		if !found {
			messenger.Message("Nothing matched " + search)
		}
	default:
		messenger.Error("Unknown command: " + inputCmd)
	}
}
