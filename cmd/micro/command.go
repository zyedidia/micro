package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"

	"github.com/mitchellh/go-homedir"
)

type Command struct {
	action      func([]string)
	completions []Completion
}

type StrCommand struct {
	action      string
	completions []Completion
}

var commands map[string]Command

var commandActions = map[string]func([]string){
	"Set":      Set,
	"SetLocal": SetLocal,
	"Show":     Show,
	"Run":      Run,
	"Bind":     Bind,
	"Quit":     Quit,
	"Save":     Save,
	"Replace":  Replace,
	"VSplit":   VSplit,
	"HSplit":   HSplit,
	"Tab":      NewTab,
	"Help":     Help,
}

// InitCommands initializes the default commands
func InitCommands() {
	commands = make(map[string]Command)

	defaults := DefaultCommands()
	parseCommands(defaults)
}

func parseCommands(userCommands map[string]StrCommand) {
	for k, v := range userCommands {
		MakeCommand(k, v.action, v.completions...)
	}
}

// MakeCommand is a function to easily create new commands
// This can be called by plugins in Lua so that plugins can define their own commands
func MakeCommand(name, function string, completions ...Completion) {
	action := commandActions[function]
	if _, ok := commandActions[function]; !ok {
		// If the user seems to be binding a function that doesn't exist
		// We hope that it's a lua function that exists and bind it to that
		action = LuaFunctionCommand(function)
	}

	commands[name] = Command{action, completions}
}

// DefaultCommands returns a map containing micro's default commands
func DefaultCommands() map[string]StrCommand {
	return map[string]StrCommand{
		"set":      {"Set", []Completion{OptionCompletion, NoCompletion}},
		"setlocal": {"SetLocal", []Completion{OptionCompletion, NoCompletion}},
		"show":     {"Show", []Completion{OptionCompletion, NoCompletion}},
		"bind":     {"Bind", []Completion{NoCompletion}},
		"run":      {"Run", []Completion{NoCompletion}},
		"quit":     {"Quit", []Completion{NoCompletion}},
		"save":     {"Save", []Completion{NoCompletion}},
		"replace":  {"Replace", []Completion{NoCompletion}},
		"vsplit":   {"VSplit", []Completion{FileCompletion, NoCompletion}},
		"hsplit":   {"HSplit", []Completion{FileCompletion, NoCompletion}},
		"tab":      {"Tab", []Completion{FileCompletion, NoCompletion}},
		"help":     {"Help", []Completion{HelpCompletion, NoCompletion}},
	}
}

// Help tries to open the given help page in a horizontal split
func Help(args []string) {
	if len(args) < 1 {
		// Open the default help if the user just typed "> help"
		CurView().openHelp("help")
	} else {
		helpPage := args[0]
		if _, ok := helpPages[helpPage]; ok {
			CurView().openHelp(helpPage)
		} else {
			messenger.Error("Sorry, no help for ", helpPage)
		}
	}
}

// VSplit opens a vertical split with file given in the first argument
// If no file is given, it opens an empty buffer in a new split
func VSplit(args []string) {
	if len(args) == 0 {
		CurView().VSplit(NewBuffer([]byte{}, ""))
	} else {
		filename := args[0]
		home, _ := homedir.Dir()
		filename = strings.Replace(filename, "~", home, 1)
		file, err := ioutil.ReadFile(filename)

		var buf *Buffer
		if err != nil {
			// File does not exist -- create an empty buffer with that name
			buf = NewBuffer([]byte{}, filename)
		} else {
			buf = NewBuffer(file, filename)
		}
		CurView().VSplit(buf)
	}
}

// HSplit opens a horizontal split with file given in the first argument
// If no file is given, it opens an empty buffer in a new split
func HSplit(args []string) {
	if len(args) == 0 {
		CurView().HSplit(NewBuffer([]byte{}, ""))
	} else {
		filename := args[0]
		home, _ := homedir.Dir()
		filename = strings.Replace(filename, "~", home, 1)
		file, err := ioutil.ReadFile(filename)

		var buf *Buffer
		if err != nil {
			// File does not exist -- create an empty buffer with that name
			buf = NewBuffer([]byte{}, filename)
		} else {
			buf = NewBuffer(file, filename)
		}
		CurView().HSplit(buf)
	}
}

// NewTab opens the given file in a new tab
func NewTab(args []string) {
	if len(args) == 0 {
		CurView().AddTab(true)
	} else {
		filename := args[0]
		home, _ := homedir.Dir()
		filename = strings.Replace(filename, "~", home, 1)
		file, _ := ioutil.ReadFile(filename)

		tab := NewTabFromView(NewView(NewBuffer(file, filename)))
		tab.SetNum(len(tabs))
		tabs = append(tabs, tab)
		curTab++
		if len(tabs) == 2 {
			for _, t := range tabs {
				for _, v := range t.views {
					v.ToggleTabbar()
				}
			}
		}
	}
}

// Set sets an option
func Set(args []string) {
	if len(args) < 2 {
		messenger.Error("Not enough arguments")
		return
	}

	option := strings.TrimSpace(args[0])
	value := strings.TrimSpace(args[1])

	SetOptionAndSettings(option, value)
}

// SetLocal sets an option local to the buffer
func SetLocal(args []string) {
	if len(args) < 2 {
		messenger.Error("Not enough arguments")
		return
	}

	option := strings.TrimSpace(args[0])
	value := strings.TrimSpace(args[1])

	err := SetLocalOption(option, value, CurView())
	if err != nil {
		messenger.Error(err.Error())
	}
}

// Show shows the value of the given option
func Show(args []string) {
	if len(args) < 1 {
		messenger.Error("Please provide an option to show")
		return
	}

	option := GetOption(args[0])

	if option == nil {
		messenger.Error(args[0], " is not a valid option")
		return
	}

	messenger.Message(option)
}

// Bind creates a new keybinding
func Bind(args []string) {
	if len(args) < 2 {
		messenger.Error("Not enough arguments")
		return
	}
	BindKey(args[0], args[1])
}

// Run runs a shell command in the background
func Run(args []string) {
	// Run a shell command in the background (openTerm is false)
	HandleShellCommand(JoinCommandArgs(args...), false, true)
}

// Quit closes the main view
func Quit(args []string) {
	// Close the main view
	CurView().Quit(true)
}

// Save saves the buffer in the main view
func Save(args []string) {
	if len(args) == 0 {
		// Save the main view
		CurView().Save(true)
	} else {
		CurView().Buf.SaveAs(args[0])
	}
}

// Replace runs search and replace
func Replace(args []string) {
	if len(args) < 2 {
		// We need to find both a search and replace expression
		messenger.Error("Invalid replace statement: " + strings.Join(args, " "))
		return
	}

	var flags string
	if len(args) == 3 {
		// The user included some flags
		flags = args[2]
	}

	search := string(args[0])
	replace := string(args[1])

	regex, err := regexp.Compile(search)
	if err != nil {
		// There was an error with the user's regex
		messenger.Error(err.Error())
		return
	}

	view := CurView()

	found := 0
	if strings.Contains(flags, "c") {
		for {
			// The 'check' flag was used
			Search(search, view, true)
			if !view.Cursor.HasSelection() {
				break
			}
			view.Relocate()
			if view.Buf.Settings["syntax"].(bool) {
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
				break
			}
			if choice {
				view.Cursor.DeleteSelection()
				view.Buf.Insert(view.Cursor.Loc, replace)
				view.Cursor.ResetSelection()
				messenger.Reset()
				found++
			} else {
				if view.Cursor.HasSelection() {
					searchStart = ToCharPos(view.Cursor.CurSelection[1], view.Buf)
				} else {
					searchStart = ToCharPos(view.Cursor.Loc, view.Buf)
				}
				continue
			}
		}
	} else {
		for {
			match := regex.FindStringIndex(view.Buf.String())
			if match == nil {
				break
			}
			found++
			view.Buf.Replace(FromCharPos(match[0], view.Buf), FromCharPos(match[1], view.Buf), replace)
		}
	}
	view.Cursor.Relocate()

	if found > 1 {
		messenger.Message("Replaced ", found, " occurrences of ", search)
	} else if found == 1 {
		messenger.Message("Replaced ", found, " occurrence of ", search)
	} else {
		messenger.Message("Nothing matched ", search)
	}
}

// RunShellCommand executes a shell command and returns the output/error
func RunShellCommand(input string) (string, error) {
	inputCmd := SplitCommandArgs(input)[0]
	args := SplitCommandArgs(input)[1:]

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
func HandleShellCommand(input string, openTerm bool, waitToFinish bool) string {
	inputCmd := SplitCommandArgs(input)[0]
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

		args := SplitCommandArgs(input)[1:]

		// Set up everything for the command
		var outputBuf bytes.Buffer
		cmd := exec.Command(inputCmd, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = io.MultiWriter(os.Stdout, &outputBuf)
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

		cmd.Start()
		err := cmd.Wait()

		output := outputBuf.String()
		if err != nil {
			output = err.Error()
		}

		if waitToFinish {
			// This is just so we don't return right away and let the user press enter to return
			TermMessage("")
		}

		// Start the screen back up
		InitScreen()

		return output
	}
	return ""
}

// HandleCommand handles input from the user
func HandleCommand(input string) {
	args := SplitCommandArgs(input)
	inputCmd := args[0]

	if _, ok := commands[inputCmd]; !ok {
		messenger.Error("Unknown command ", inputCmd)
	} else {
		commands[inputCmd].action(args[1:])
	}
}
