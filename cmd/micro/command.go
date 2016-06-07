package main

import (
	"bytes"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
)

var commands map[string]func([]string)

var commandActions = map[string]func([]string){
	"Set":     Set,
	"Run":     Run,
	"Bind":    Bind,
	"Quit":    Quit,
	"Save":    Save,
	"Replace": Replace,
}

// InitCommands initializes the default commands
func InitCommands() {
	commands = make(map[string]func([]string))

	defaults := DefaultCommands()
	parseCommands(defaults)
}

func parseCommands(userCommands map[string]string) {
	for k, v := range userCommands {
		MakeCommand(k, v)
	}
}

// MakeCommand is a function to easily create new commands
// This can be called by plugins in Lua so that plugins can define their own commands
func MakeCommand(name, function string) {
	action := commandActions[function]
	if _, ok := commandActions[function]; !ok {
		// If the user seems to be binding a function that doesn't exist
		// We hope that it's a lua function that exists and bind it to that
		action = LuaFunctionCommand(function)
	}

	commands[name] = action
}

// DefaultCommands returns a map containing micro's default commands
func DefaultCommands() map[string]string {
	return map[string]string{
		"set":     "Set",
		"bind":    "Bind",
		"run":     "Run",
		"quit":    "Quit",
		"save":    "Save",
		"replace": "Replace",
	}
}

// Set sets an option
func Set(args []string) {
	// Set an option and we have to set it for every view
	for _, view := range views {
		SetOption(view, args)
	}
}

// Bind creates a new keybinding
func Bind(args []string) {
	if len(args) != 2 {
		messenger.Error("Incorrect number of arguments")
		return
	}
	BindKey(args[0], args[1])
}

// Run runs a shell command in the background
func Run(args []string) {
	// Run a shell command in the background (openTerm is false)
	HandleShellCommand(strings.Join(args, " "), false)
}

// Quit closes the main view
func Quit(args []string) {
	// Close the main view
	views[mainView].Quit()
}

// Save saves the buffer in the main view
func Save(args []string) {
	// Save the main view
	views[mainView].Save()
}

// Replace runs search and replace
func Replace(args []string) {
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
					view.Cursor.Loc = view.Cursor.CurSelection[0]
					view.Cursor.ResetSelection()
				}
				messenger.Reset()
				return
			}
			if choice {
				view.Cursor.DeleteSelection()
				view.Buf.Insert(FromCharPos(match[0], view.Buf), replace)
				view.Cursor.ResetSelection()
				messenger.Reset()
			} else {
				if view.Cursor.HasSelection() {
					searchStart = ToCharPos(view.Cursor.CurSelection[1], view.Buf)
				} else {
					searchStart = ToCharPos(view.Cursor.Loc, view.Buf)
				}
				continue
			}
		} else {
			view.Buf.Replace(FromCharPos(match[0], view.Buf), FromCharPos(match[1], view.Buf), replace)
		}
	}
	if !found {
		messenger.Message("Nothing matched " + search)
	}
}

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

	if _, ok := commands[inputCmd]; !ok {
		messenger.Error("Unkown command ", inputCmd)
	} else {
		commands[inputCmd](args)
	}
}
