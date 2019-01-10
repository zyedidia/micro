package action

import (
	"log"
	"os"

	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/shellwords"
	"github.com/zyedidia/micro/cmd/micro/util"
)

// A Command contains an action (a function to call) as well as information about how to autocomplete the command
type Command struct {
	action      func([]string)
	completions []Completion
}

// A StrCommand is similar to a command but keeps the name of the action
type StrCommand struct {
	action      string
	completions []Completion
}

var commands map[string]Command

var commandActions map[string]func([]string)

func init() {
	commandActions = map[string]func([]string){
		"Set":        Set,
		"SetLocal":   SetLocal,
		"Show":       Show,
		"ShowKey":    ShowKey,
		"Run":        Run,
		"Bind":       Bind,
		"Quit":       Quit,
		"Save":       Save,
		"Replace":    Replace,
		"ReplaceAll": ReplaceAll,
		"VSplit":     VSplit,
		"HSplit":     HSplit,
		"Tab":        NewTab,
		"Help":       Help,
		"Eval":       Eval,
		"ToggleLog":  ToggleLog,
		"Plugin":     PluginCmd,
		"Reload":     Reload,
		"Cd":         Cd,
		"Pwd":        Pwd,
		"Open":       Open,
		"TabSwitch":  TabSwitch,
		"Term":       Term,
		"MemUsage":   MemUsage,
		"Retab":      Retab,
		"Raw":        Raw,
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
	// if _, ok := commandActions[function]; !ok {
	// If the user seems to be binding a function that doesn't exist
	// We hope that it's a lua function that exists and bind it to that
	// action = LuaFunctionCommand(function)
	// }

	commands[name] = Command{action, completions}
}

// DefaultCommands returns a map containing micro's default commands
func DefaultCommands() map[string]StrCommand {
	return map[string]StrCommand{
		"set":        {"Set", []Completion{OptionCompletion, OptionValueCompletion}},
		"setlocal":   {"SetLocal", []Completion{OptionCompletion, OptionValueCompletion}},
		"show":       {"Show", []Completion{OptionCompletion, NoCompletion}},
		"showkey":    {"ShowKey", []Completion{NoCompletion}},
		"bind":       {"Bind", []Completion{NoCompletion}},
		"run":        {"Run", []Completion{NoCompletion}},
		"quit":       {"Quit", []Completion{NoCompletion}},
		"save":       {"Save", []Completion{NoCompletion}},
		"replace":    {"Replace", []Completion{NoCompletion}},
		"replaceall": {"ReplaceAll", []Completion{NoCompletion}},
		"vsplit":     {"VSplit", []Completion{FileCompletion, NoCompletion}},
		"hsplit":     {"HSplit", []Completion{FileCompletion, NoCompletion}},
		"tab":        {"Tab", []Completion{FileCompletion, NoCompletion}},
		"help":       {"Help", []Completion{HelpCompletion, NoCompletion}},
		"eval":       {"Eval", []Completion{NoCompletion}},
		"log":        {"ToggleLog", []Completion{NoCompletion}},
		"plugin":     {"Plugin", []Completion{PluginCmdCompletion, PluginNameCompletion}},
		"reload":     {"Reload", []Completion{NoCompletion}},
		"cd":         {"Cd", []Completion{FileCompletion}},
		"pwd":        {"Pwd", []Completion{NoCompletion}},
		"open":       {"Open", []Completion{FileCompletion}},
		"tabswitch":  {"TabSwitch", []Completion{NoCompletion}},
		"term":       {"Term", []Completion{NoCompletion}},
		"memusage":   {"MemUsage", []Completion{NoCompletion}},
		"retab":      {"Retab", []Completion{NoCompletion}},
		"raw":        {"Raw", []Completion{NoCompletion}},
	}
}

// CommandEditAction returns a bindable function that opens a prompt with
// the given string and executes the command when the user presses
// enter
func CommandEditAction(prompt string) BufKeyAction {
	return func(h *BufHandler) bool {
		InfoBar.Prompt("> ", prompt, "Command", nil, func(resp string, canceled bool) {
			if !canceled {
				HandleCommand(resp)
			}
		})
		return false
	}
}

// CommandAction returns a bindable function which executes the
// given command
func CommandAction(cmd string) BufKeyAction {
	return func(h *BufHandler) bool {
		HandleCommand(cmd)
		return false
	}
}

// PluginCmd installs, removes, updates, lists, or searches for given plugins
func PluginCmd(args []string) {
}

// Retab changes all spaces to tabs or all tabs to spaces
// depending on the user's settings
func Retab(args []string) {
}

// Raw opens a new raw view which displays the escape sequences micro
// is receiving in real-time
func Raw(args []string) {
}

// TabSwitch switches to a given tab either by name or by number
func TabSwitch(args []string) {
}

// Cd changes the current working directory
func Cd(args []string) {
}

// MemUsage prints micro's memory usage
// Alloc shows how many bytes are currently in use
// Sys shows how many bytes have been requested from the operating system
// NumGC shows how many times the GC has been run
// Note that Go commonly reserves more memory from the OS than is currently in-use/required
// Additionally, even if Go returns memory to the OS, the OS does not always claim it because
// there may be plenty of memory to spare
func MemUsage(args []string) {
	InfoBar.Message(util.GetMemStats())
}

// Pwd prints the current working directory
func Pwd(args []string) {
	wd, err := os.Getwd()
	if err != nil {
		InfoBar.Message(err.Error())
	} else {
		InfoBar.Message(wd)
	}
}

// Open opens a new buffer with a given filename
func Open(args []string) {
}

// ToggleLog toggles the log view
func ToggleLog(args []string) {
}

// Reload reloads all files (syntax files, colorschemes...)
func Reload(args []string) {
}

// Help tries to open the given help page in a horizontal split
func Help(args []string) {
}

// VSplit opens a vertical split with file given in the first argument
// If no file is given, it opens an empty buffer in a new split
func VSplit(args []string) {
	buf, err := buffer.NewBufferFromFile(args[0], buffer.BTDefault)
	if err != nil {
		InfoBar.Error(err)
		return
	}

	log.Println("loaded")
	MainTab().CurPane().vsplit(buf)
}

// HSplit opens a horizontal split with file given in the first argument
// If no file is given, it opens an empty buffer in a new split
func HSplit(args []string) {
	buf, err := buffer.NewBufferFromFile(args[0], buffer.BTDefault)
	if err != nil {
		InfoBar.Error(err)
		return
	}

	MainTab().CurPane().hsplit(buf)
}

// Eval evaluates a lua expression
func Eval(args []string) {
}

// NewTab opens the given file in a new tab
func NewTab(args []string) {
}

// Set sets an option
func Set(args []string) {
}

// SetLocal sets an option local to the buffer
func SetLocal(args []string) {
}

// Show shows the value of the given option
func Show(args []string) {
}

// ShowKey displays the action that a key is bound to
func ShowKey(args []string) {
	if len(args) < 1 {
		InfoBar.Error("Please provide a key to show")
		return
	}

	if action, ok := Bindings[args[0]]; ok {
		InfoBar.Message(action)
	} else {
		InfoBar.Message(args[0], " has no binding")
	}
}

// Bind creates a new keybinding
func Bind(args []string) {
}

// Run runs a shell command in the background
func Run(args []string) {
}

// Quit closes the main view
func Quit(args []string) {
}

// Save saves the buffer in the main view
func Save(args []string) {
}

// Replace runs search and replace
func Replace(args []string) {
}

// ReplaceAll replaces search term all at once
func ReplaceAll(args []string) {
}

// Term opens a terminal in the current view
func Term(args []string) {
}

// HandleCommand handles input from the user
func HandleCommand(input string) {
	args, err := shellwords.Split(input)
	if err != nil {
		InfoBar.Error("Error parsing args ", err)
		return
	}

	inputCmd := args[0]

	if _, ok := commands[inputCmd]; !ok {
		InfoBar.Error("Unknown command ", inputCmd)
	} else {
		commands[inputCmd].action(args[1:])
	}
}
