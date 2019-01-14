package action

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/micro/cmd/micro/shell"
	"github.com/zyedidia/micro/cmd/micro/shellwords"
	"github.com/zyedidia/micro/cmd/micro/util"
)

// A Command contains an action (a function to call) as well as information about how to autocomplete the command
type Command struct {
	action      func(*BufHandler, []string)
	completions []Completion
}

// A StrCommand is similar to a command but keeps the name of the action
type StrCommand struct {
	action      string
	completions []Completion
}

var commands map[string]Command

var commandActions = map[string]func(*BufHandler, []string){
	"Set":        (*BufHandler).SetCmd,
	"SetLocal":   (*BufHandler).SetLocalCmd,
	"Show":       (*BufHandler).ShowCmd,
	"ShowKey":    (*BufHandler).ShowKeyCmd,
	"Run":        (*BufHandler).RunCmd,
	"Bind":       (*BufHandler).BindCmd,
	"Unbind":     (*BufHandler).UnbindCmd,
	"Quit":       (*BufHandler).QuitCmd,
	"Save":       (*BufHandler).SaveCmd,
	"Replace":    (*BufHandler).ReplaceCmd,
	"ReplaceAll": (*BufHandler).ReplaceAllCmd,
	"VSplit":     (*BufHandler).VSplitCmd,
	"HSplit":     (*BufHandler).HSplitCmd,
	"Tab":        (*BufHandler).NewTabCmd,
	"Help":       (*BufHandler).HelpCmd,
	"Eval":       (*BufHandler).EvalCmd,
	"ToggleLog":  (*BufHandler).ToggleLogCmd,
	"Plugin":     (*BufHandler).PluginCmd,
	"Reload":     (*BufHandler).ReloadCmd,
	"Cd":         (*BufHandler).CdCmd,
	"Pwd":        (*BufHandler).PwdCmd,
	"Open":       (*BufHandler).OpenCmd,
	"TabSwitch":  (*BufHandler).TabSwitchCmd,
	"Term":       (*BufHandler).TermCmd,
	"MemUsage":   (*BufHandler).MemUsageCmd,
	"Retab":      (*BufHandler).RetabCmd,
	"Raw":        (*BufHandler).RawCmd,
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
		"unbind":     {"Unbind", []Completion{NoCompletion}},
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
				MainTab().CurPane().HandleCommand(resp)
			}
		})
		return false
	}
}

// CommandAction returns a bindable function which executes the
// given command
func CommandAction(cmd string) BufKeyAction {
	return func(h *BufHandler) bool {
		MainTab().CurPane().HandleCommand(cmd)
		return false
	}
}

// PluginCmd installs, removes, updates, lists, or searches for given plugins
func (h *BufHandler) PluginCmd(args []string) {
}

// RetabCmd changes all spaces to tabs or all tabs to spaces
// depending on the user's settings
func (h *BufHandler) RetabCmd(args []string) {
}

// RawCmd opens a new raw view which displays the escape sequences micro
// is receiving in real-time
func (h *BufHandler) RawCmd(args []string) {
}

// TabSwitchCmd switches to a given tab either by name or by number
func (h *BufHandler) TabSwitchCmd(args []string) {
	if len(args) > 0 {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			// Check for tab with this name

			found := false
			for i, t := range Tabs.List {
				if t.Panes[t.active].Name() == args[0] {
					Tabs.SetActive(i)
					found = true
				}
			}
			if !found {
				InfoBar.Error("Could not find tab: ", err)
			}
		} else {
			num--
			if num >= 0 && num < len(Tabs.List) {
				Tabs.SetActive(num)
			} else {
				InfoBar.Error("Invalid tab index")
			}
		}
	}
}

// CdCmd changes the current working directory
func (h *BufHandler) CdCmd(args []string) {
	if len(args) > 0 {
		path, err := util.ReplaceHome(args[0])
		if err != nil {
			InfoBar.Error(err)
			return
		}
		err = os.Chdir(path)
		if err != nil {
			InfoBar.Error(err)
			return
		}
		wd, _ := os.Getwd()
		for _, b := range buffer.OpenBuffers {
			if len(b.Path) > 0 {
				b.Path, _ = util.MakeRelative(b.AbsPath, wd)
				if p, _ := filepath.Abs(b.Path); !strings.Contains(p, wd) {
					b.Path = b.AbsPath
				}
			}
		}
	}
}

// MemUsageCmd prints micro's memory usage
// Alloc shows how many bytes are currently in use
// Sys shows how many bytes have been requested from the operating system
// NumGC shows how many times the GC has been run
// Note that Go commonly reserves more memory from the OS than is currently in-use/required
// Additionally, even if Go returns memory to the OS, the OS does not always claim it because
// there may be plenty of memory to spare
func (h *BufHandler) MemUsageCmd(args []string) {
	InfoBar.Message(util.GetMemStats())
}

// PwdCmd prints the current working directory
func (h *BufHandler) PwdCmd(args []string) {
	wd, err := os.Getwd()
	if err != nil {
		InfoBar.Message(err.Error())
	} else {
		InfoBar.Message(wd)
	}
}

// OpenCmd opens a new buffer with a given filename
func (h *BufHandler) OpenCmd(args []string) {
	if len(args) > 0 {
		filename := args[0]
		// the filename might or might not be quoted, so unquote first then join the strings.
		args, err := shellwords.Split(filename)
		if err != nil {
			InfoBar.Error("Error parsing args ", err)
			return
		}
		filename = strings.Join(args, " ")

		open := func() {
			b, err := buffer.NewBufferFromFile(filename, buffer.BTDefault)
			if err != nil {
				InfoBar.Error(err)
				return
			}
			h.OpenBuffer(b)
		}
		if h.Buf.Modified() {
			InfoBar.YNPrompt("Save changes to "+h.Buf.GetName()+" before closing? (y,n,esc)", func(yes, canceled bool) {
				if !canceled && !yes {
					open()
				} else if !canceled && yes {
					h.Save()
					open()
				}
			})
		} else {
			open()
		}
	} else {
		InfoBar.Error("No filename")
	}
}

// ToggleLogCmd toggles the log view
func (h *BufHandler) ToggleLogCmd(args []string) {
}

// ReloadCmd reloads all files (syntax files, colorschemes...)
func (h *BufHandler) ReloadCmd(args []string) {
}

// HelpCmd tries to open the given help page in a horizontal split
func (h *BufHandler) HelpCmd(args []string) {
}

// VSplitCmd opens a vertical split with file given in the first argument
// If no file is given, it opens an empty buffer in a new split
func (h *BufHandler) VSplitCmd(args []string) {
	if len(args) == 0 {
		// Open an empty vertical split
		h.VSplitAction()
		return
	}

	buf, err := buffer.NewBufferFromFile(args[0], buffer.BTDefault)
	if err != nil {
		InfoBar.Error(err)
		return
	}

	h.VSplitBuf(buf)
}

// HSplitCmd opens a horizontal split with file given in the first argument
// If no file is given, it opens an empty buffer in a new split
func (h *BufHandler) HSplitCmd(args []string) {
	if len(args) == 0 {
		// Open an empty horizontal split
		h.HSplitAction()
		return
	}

	buf, err := buffer.NewBufferFromFile(args[0], buffer.BTDefault)
	if err != nil {
		InfoBar.Error(err)
		return
	}

	h.HSplitBuf(buf)
}

// EvalCmd evaluates a lua expression
func (h *BufHandler) EvalCmd(args []string) {
}

// NewTabCmd opens the given file in a new tab
func (h *BufHandler) NewTabCmd(args []string) {
	width, height := screen.Screen.Size()
	if len(args) > 0 {
		for _, a := range args {
			b, err := buffer.NewBufferFromFile(a, buffer.BTDefault)
			if err != nil {
				InfoBar.Error(err)
				return
			}
			tp := NewTabFromBuffer(0, 0, width, height-1, b)
			Tabs.AddTab(tp)
			Tabs.SetActive(len(Tabs.List) - 1)
		}
	} else {
		b := buffer.NewBufferFromString("", "", buffer.BTDefault)
		tp := NewTabFromBuffer(0, 0, width, height-1, b)
		Tabs.AddTab(tp)
		Tabs.SetActive(len(Tabs.List) - 1)
	}
}

func SetGlobalOption(option, value string) error {
	if _, ok := config.GlobalSettings[option]; !ok {
		return config.ErrInvalidOption
	}

	nativeValue, err := config.GetNativeValue(option, config.GlobalSettings[option], value)
	if err != nil {
		return err
	}

	config.GlobalSettings[option] = nativeValue

	if option == "colorscheme" {
		// LoadSyntaxFiles()
		config.InitColorscheme()
		for _, b := range buffer.OpenBuffers {
			b.UpdateRules()
		}
	}

	// TODO: info and keymenu option change
	// if option == "infobar" || option == "keymenu" {
	// 	for _, tab := range tabs {
	// 		tab.Resize()
	// 	}
	// }

	if option == "mouse" {
		if !nativeValue.(bool) {
			screen.Screen.DisableMouse()
		} else {
			screen.Screen.EnableMouse()
		}
	}

	for _, b := range buffer.OpenBuffers {
		b.SetOption(option, value)
	}

	config.WriteSettings(config.ConfigDir + "/settings.json")

	return nil
}

// SetCmd sets an option
func (h *BufHandler) SetCmd(args []string) {
	if len(args) < 2 {
		InfoBar.Error("Not enough arguments")
		return
	}

	option := args[0]
	value := args[1]

	err := SetGlobalOption(option, value)
	if err == config.ErrInvalidOption {
		err := h.Buf.SetOption(option, value)
		if err != nil {
			InfoBar.Error(err)
		}
	} else if err != nil {
		InfoBar.Error(err)
	}
}

// SetLocalCmd sets an option local to the buffer
func (h *BufHandler) SetLocalCmd(args []string) {
	if len(args) < 2 {
		InfoBar.Error("Not enough arguments")
		return
	}

	option := args[0]
	value := args[1]

	err := h.Buf.SetOption(option, value)
	if err != nil {
		InfoBar.Error(err)
	}

}

// ShowCmd shows the value of the given option
func (h *BufHandler) ShowCmd(args []string) {
	if len(args) < 1 {
		InfoBar.Error("Please provide an option to show")
		return
	}

	var option interface{}
	if opt, ok := h.Buf.Settings[args[0]]; ok {
		option = opt
	} else if opt, ok := config.GlobalSettings[args[0]]; ok {
		option = opt
	}

	if option == nil {
		InfoBar.Error(args[0], " is not a valid option")
		return
	}

	InfoBar.Message(option)
}

// ShowKeyCmd displays the action that a key is bound to
func (h *BufHandler) ShowKeyCmd(args []string) {
	if len(args) < 1 {
		InfoBar.Error("Please provide a key to show")
		return
	}

	if action, ok := config.Bindings[args[0]]; ok {
		InfoBar.Message(action)
	} else {
		InfoBar.Message(args[0], " has no binding")
	}
}

// BindCmd creates a new keybinding
func (h *BufHandler) BindCmd(args []string) {
	if len(args) < 2 {
		InfoBar.Error("Not enough arguments")
		return
	}

	_, err := TryBindKey(args[0], args[1], true)
	if err != nil {
		InfoBar.Error(err)
	}
}

// UnbindCmd binds a key to its default action
func (h *BufHandler) UnbindCmd(args []string) {
	if len(args) < 1 {
		InfoBar.Error("Not enough arguements")
		return
	}

	err := UnbindKey(args[0])
	if err != nil {
		InfoBar.Error(err)
	}
}

// RunCmd runs a shell command in the background
func (h *BufHandler) RunCmd(args []string) {
	runf, err := shell.RunBackgroundShell(shellwords.Join(args...))
	if err != nil {
		InfoBar.Error(err)
	} else {
		go func() {
			InfoBar.Message(runf())
			screen.Redraw()
		}()
	}
}

// QuitCmd closes the main view
func (h *BufHandler) QuitCmd(args []string) {
	h.Quit()
}

// SaveCmd saves the buffer in the main view
func (h *BufHandler) SaveCmd(args []string) {
	h.Save()
}

// ReplaceCmd runs search and replace
func (h *BufHandler) ReplaceCmd(args []string) {
}

// ReplaceAllCmd replaces search term all at once
func (h *BufHandler) ReplaceAllCmd(args []string) {
}

// TermCmd opens a terminal in the current view
func (h *BufHandler) TermCmd(args []string) {
	ps := MainTab().Panes

	if len(args) == 0 {
		sh := os.Getenv("SHELL")
		if sh == "" {
			InfoBar.Error("Shell environment not found")
			return
		}
		args = []string{sh}
	}

	term := func(i int) {
		// If there is only one open file we make a new tab instead of overwriting it
		newtab := len(MainTab().Panes) == 1 && len(Tabs.List) == 1

		t := new(shell.Terminal)
		t.Start(args, false, true)

		id := h.ID()
		if newtab {
			h.AddTab()
			i = 0
			id = MainTab().Panes[0].ID()
		} else {
			MainTab().Panes[i].Close()
		}

		v := h.GetView()
		MainTab().Panes[i] = NewTermHandler(v.X, v.Y, v.Width, v.Height, t, id)
		MainTab().SetActive(i)
	}

	for i, p := range ps {
		if p.ID() == h.ID() {
			if h.Buf.Modified() {
				InfoBar.YNPrompt("Save changes to "+h.Buf.GetName()+" before closing? (y,n,esc)", func(yes, canceled bool) {
					if !canceled && !yes {
						term(i)
					} else if !canceled && yes {
						h.Save()
						term(i)
					}
				})
			} else {
				term(i)
			}
		}
	}
}

// HandleCommand handles input from the user
func (h *BufHandler) HandleCommand(input string) {
	args, err := shellwords.Split(input)
	if err != nil {
		InfoBar.Error("Error parsing args ", err)
		return
	}

	inputCmd := args[0]

	if _, ok := commands[inputCmd]; !ok {
		InfoBar.Error("Unknown command ", inputCmd)
	} else {
		commands[inputCmd].action(h, args[1:])
	}
}
