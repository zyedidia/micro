package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
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

var commandActions map[string]func([]string)

func init() {
	commandActions = map[string]func([]string){
		"Set":       Set,
		"SetLocal":  SetLocal,
		"Show":      Show,
		"Run":       Run,
		"Bind":      Bind,
		"Quit":      Quit,
		"Save":      Save,
		"Replace":   Replace,
		"VSplit":    VSplit,
		"HSplit":    HSplit,
		"Tab":       NewTab,
		"Help":      Help,
		"Eval":      Eval,
		"ToggleLog": ToggleLog,
		"Plugin":    PluginCmd,
		"Reload":    Reload,
		"Cd":        Cd,
		"Pwd":       Pwd,
		"Open":      Open,
	}
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
		"eval":     {"Eval", []Completion{NoCompletion}},
		"log":      {"ToggleLog", []Completion{NoCompletion}},
		"plugin":   {"Plugin", []Completion{PluginCmdCompletion, PluginNameCompletion}},
		"reload":   {"Reload", []Completion{NoCompletion}},
		"cd":       {"Cd", []Completion{FileCompletion}},
		"pwd":      {"Pwd", []Completion{NoCompletion}},
		"open":     {"Open", []Completion{FileCompletion}},
	}
}

// PluginCmd installs, removes, updates, lists, or searches for given plugins
func PluginCmd(args []string) {
	if len(args) >= 1 {
		switch args[0] {
		case "install":
			installedVersions := GetInstalledVersions(false)
			for _, plugin := range args[1:] {
				pp := GetAllPluginPackages().Get(plugin)
				if pp == nil {
					messenger.Error("Unknown plugin \"" + plugin + "\"")
				} else if err := pp.IsInstallable(); err != nil {
					messenger.Error("Error installing ", plugin, ": ", err)
				} else {
					for _, installed := range installedVersions {
						if pp.Name == installed.pack.Name {
							if pp.Versions[0].Version.Compare(installed.Version) == 1 {
								messenger.Error(pp.Name, " is already installed but out-of-date: use 'plugin update ", pp.Name, "' to update")
							} else {
								messenger.Error(pp.Name, " is already installed")
							}
						}
					}
					pp.Install()
				}
			}
		case "remove":
			removed := ""
			for _, plugin := range args[1:] {
				// check if the plugin exists.
				if _, ok := loadedPlugins[plugin]; ok {
					UninstallPlugin(plugin)
					removed += plugin + " "
					continue
				}
			}
			if !IsSpaces(removed) {
				messenger.Message("Removed ", removed)
			} else {
				messenger.Error("The requested plugins do not exist")
			}
		case "update":
			UpdatePlugins(args[1:])
		case "list":
			plugins := GetInstalledVersions(false)
			messenger.AddLog("----------------")
			messenger.AddLog("The following plugins are currently installed:\n")
			for _, p := range plugins {
				messenger.AddLog(fmt.Sprintf("%s (%s)", p.pack.Name, p.Version))
			}
			messenger.AddLog("----------------")
			if len(plugins) > 0 {
				if CurView().Type != vtLog {
					ToggleLog([]string{})
				}
			}
		case "search":
			plugins := SearchPlugin(args[1:])
			messenger.Message(len(plugins), " plugins found")
			for _, p := range plugins {
				messenger.AddLog("----------------")
				messenger.AddLog(p.String())
			}
			messenger.AddLog("----------------")
			if len(plugins) > 0 {
				if CurView().Type != vtLog {
					ToggleLog([]string{})
				}
			}
		case "available":
			packages := GetAllPluginPackages()
			messenger.AddLog("Available Plugins:")
			for _, pkg := range packages {
				messenger.AddLog(pkg.Name)
			}
			if CurView().Type != vtLog {
				ToggleLog([]string{})
			}
		}
	} else {
		messenger.Error("Not enough arguments")
	}
}

func Cd(args []string) {
	if len(args) > 0 {
		home, _ := homedir.Dir()
		path := strings.Replace(args[0], "~", home, 1)
		os.Chdir(path)
		for _, tab := range tabs {
			for _, view := range tab.views {
				wd, _ := os.Getwd()
				view.Buf.Path, _ = MakeRelative(view.Buf.AbsPath, wd)
				if p, _ := filepath.Abs(view.Buf.Path); !strings.Contains(p, wd) {
					view.Buf.Path = view.Buf.AbsPath
				}
			}
		}
	}
}

func Pwd(args []string) {
	wd, err := os.Getwd()
	if err != nil {
		messenger.Message(err.Error())
	} else {
		messenger.Message(wd)
	}
}

func Open(args []string) {
	if len(args) > 0 {
		filename := args[0]
		// the filename might or might not be quoted, so unquote first then join the strings.
		filename = strings.Join(SplitCommandArgs(filename), " ")

		CurView().Open(filename)
	} else {
		messenger.Error("No filename")
	}
}

func ToggleLog(args []string) {
	buffer := messenger.getBuffer()
	if CurView().Type != vtLog {
		CurView().HSplit(buffer)
		CurView().Type = vtLog
		RedrawAll()
		buffer.Cursor.Loc = buffer.Start()
		CurView().Relocate()
		buffer.Cursor.Loc = buffer.End()
		CurView().Relocate()
	} else {
		CurView().Quit(true)
	}
}

func Reload(args []string) {
	LoadAll()
}

// Help tries to open the given help page in a horizontal split
func Help(args []string) {
	if len(args) < 1 {
		// Open the default help if the user just typed "> help"
		CurView().openHelp("help")
	} else {
		helpPage := args[0]
		if FindRuntimeFile(RTHelp, helpPage) != nil {
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
		CurView().VSplit(NewBuffer(strings.NewReader(""), ""))
	} else {
		filename := args[0]
		home, _ := homedir.Dir()
		filename = strings.Replace(filename, "~", home, 1)
		file, err := os.Open(filename)
		defer file.Close()

		var buf *Buffer
		if err != nil {
			// File does not exist -- create an empty buffer with that name
			buf = NewBuffer(strings.NewReader(""), filename)
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
		CurView().HSplit(NewBuffer(strings.NewReader(""), ""))
	} else {
		filename := args[0]
		home, _ := homedir.Dir()
		filename = strings.Replace(filename, "~", home, 1)
		file, err := os.Open(filename)
		defer file.Close()

		var buf *Buffer
		if err != nil {
			// File does not exist -- create an empty buffer with that name
			buf = NewBuffer(strings.NewReader(""), filename)
		} else {
			buf = NewBuffer(file, filename)
		}
		CurView().HSplit(buf)
	}
}

// Eval evaluates a lua expression
func Eval(args []string) {
	if len(args) >= 1 {
		err := L.DoString(args[0])
		if err != nil {
			messenger.Error(err)
		}
	} else {
		messenger.Error("Not enough arguments")
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
		file, _ := os.Open(filename)
		defer file.Close()

		tab := NewTabFromView(NewView(NewBuffer(file, filename)))
		tab.SetNum(len(tabs))
		tabs = append(tabs, tab)
		curTab = len(tabs) - 1
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

	regex, err := regexp.Compile("(?m)" + search)
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
		bufStr := view.Buf.String()
		matches := regex.FindAllStringIndex(bufStr, -1)
		if matches != nil && len(matches) > 0 {
			prevMatchCount := runePos(matches[0][0], bufStr)
			searchCount := runePos(matches[0][1], bufStr) - prevMatchCount
			from := FromCharPos(matches[0][0], view.Buf)
			to := from.Move(searchCount, view.Buf)
			adjust := Count(replace) - searchCount
			view.Buf.Replace(from, to, replace)
			if len(matches) > 1 {
				for _, match := range matches[1:] {
					found++
					matchCount := runePos(match[0], bufStr)
					searchCount = runePos(match[1], bufStr) - matchCount
					from = from.Move(matchCount-prevMatchCount+adjust, view.Buf)
					to = from.Move(searchCount, view.Buf)
					view.Buf.Replace(from, to, replace)
					prevMatchCount = matchCount
					adjust = Count(replace) - searchCount
				}
			}
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
