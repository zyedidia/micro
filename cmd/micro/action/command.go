package action

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/zyedidia/micro/cmd/micro/buffer"
	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/micro/cmd/micro/shell"
	"github.com/zyedidia/micro/cmd/micro/shellwords"
	"github.com/zyedidia/micro/cmd/micro/util"
)

// A Command contains information about how to execute a command
// It has the action for that command as well as a completer function
type Command struct {
	action    func(*BufPane, []string)
	completer buffer.Completer
}

var commands map[string]Command

func InitCommands() {
	commands = map[string]Command{
		"set":        Command{(*BufPane).SetCmd, OptionValueComplete},
		"setlocal":   Command{(*BufPane).SetLocalCmd, OptionValueComplete},
		"show":       Command{(*BufPane).ShowCmd, OptionComplete},
		"showkey":    Command{(*BufPane).ShowKeyCmd, nil},
		"run":        Command{(*BufPane).RunCmd, nil},
		"bind":       Command{(*BufPane).BindCmd, nil},
		"unbind":     Command{(*BufPane).UnbindCmd, nil},
		"quit":       Command{(*BufPane).QuitCmd, nil},
		"save":       Command{(*BufPane).SaveCmd, nil},
		"replace":    Command{(*BufPane).ReplaceCmd, nil},
		"replaceall": Command{(*BufPane).ReplaceAllCmd, nil},
		"vsplit":     Command{(*BufPane).VSplitCmd, buffer.FileComplete},
		"hsplit":     Command{(*BufPane).HSplitCmd, buffer.FileComplete},
		"tab":        Command{(*BufPane).NewTabCmd, buffer.FileComplete},
		"help":       Command{(*BufPane).HelpCmd, HelpComplete},
		"eval":       Command{(*BufPane).EvalCmd, nil},
		"togglelog":  Command{(*BufPane).ToggleLogCmd, nil},
		"plugin":     Command{(*BufPane).PluginCmd, nil},
		"reload":     Command{(*BufPane).ReloadCmd, nil},
		"reopen":     Command{(*BufPane).ReopenCmd, nil},
		"cd":         Command{(*BufPane).CdCmd, buffer.FileComplete},
		"pwd":        Command{(*BufPane).PwdCmd, nil},
		"open":       Command{(*BufPane).OpenCmd, buffer.FileComplete},
		"tabswitch":  Command{(*BufPane).TabSwitchCmd, nil},
		"term":       Command{(*BufPane).TermCmd, nil},
		"memusage":   Command{(*BufPane).MemUsageCmd, nil},
		"retab":      Command{(*BufPane).RetabCmd, nil},
		"raw":        Command{(*BufPane).RawCmd, nil},
	}
}

// MakeCommand is a function to easily create new commands
// This can be called by plugins in Lua so that plugins can define their own commands
// func MakeCommand(name, function string, completions ...Completion) {
// 	action := commandActions[function]
// 	// if _, ok := commandActions[function]; !ok {
// 	// If the user seems to be binding a function that doesn't exist
// 	// We hope that it's a lua function that exists and bind it to that
// 	// action = LuaFunctionCommand(function)
// 	// }
//
// 	commands[name] = Command{action, completions}
// }

// CommandEditAction returns a bindable function that opens a prompt with
// the given string and executes the command when the user presses
// enter
func CommandEditAction(prompt string) BufKeyAction {
	return func(h *BufPane) bool {
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
	return func(h *BufPane) bool {
		MainTab().CurPane().HandleCommand(cmd)
		return false
	}
}

// PluginCmd installs, removes, updates, lists, or searches for given plugins
func (h *BufPane) PluginCmd(args []string) {
}

// RetabCmd changes all spaces to tabs or all tabs to spaces
// depending on the user's settings
func (h *BufPane) RetabCmd(args []string) {
	h.Buf.Retab()
}

// RawCmd opens a new raw view which displays the escape sequences micro
// is receiving in real-time
func (h *BufPane) RawCmd(args []string) {
	width, height := screen.Screen.Size()
	iOffset := config.GetInfoBarOffset()
	tp := NewTabFromPane(0, 0, width, height-iOffset, NewRawPane())
	Tabs.AddTab(tp)
	Tabs.SetActive(len(Tabs.List) - 1)
}

// TabSwitchCmd switches to a given tab either by name or by number
func (h *BufPane) TabSwitchCmd(args []string) {
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
func (h *BufPane) CdCmd(args []string) {
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
func (h *BufPane) MemUsageCmd(args []string) {
	InfoBar.Message(util.GetMemStats())
}

// PwdCmd prints the current working directory
func (h *BufPane) PwdCmd(args []string) {
	wd, err := os.Getwd()
	if err != nil {
		InfoBar.Message(err.Error())
	} else {
		InfoBar.Message(wd)
	}
}

// OpenCmd opens a new buffer with a given filename
func (h *BufPane) OpenCmd(args []string) {
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
func (h *BufPane) ToggleLogCmd(args []string) {
}

// ReloadCmd reloads all files (syntax files, colorschemes...)
func (h *BufPane) ReloadCmd(args []string) {
}

// ReopenCmd reopens the buffer (reload from disk)
func (h *BufPane) ReopenCmd(args []string) {
	if h.Buf.Modified() {
		InfoBar.YNPrompt("Save file before reopen?", func(yes, canceled bool) {
			if !canceled && yes {
				h.Save()
				h.Buf.ReOpen()
			} else if !canceled {
				h.Buf.ReOpen()
			}
		})
	} else {
		h.Buf.ReOpen()
	}
}

func (h *BufPane) openHelp(page string) error {
	if data, err := config.FindRuntimeFile(config.RTHelp, page).Data(); err != nil {
		return errors.New(fmt.Sprint("Unable to load help text", page, "\n", err))
	} else {
		helpBuffer := buffer.NewBufferFromString(string(data), page+".md", buffer.BTHelp)
		helpBuffer.SetName("Help " + page)

		if h.Buf.Type == buffer.BTHelp {
			h.OpenBuffer(helpBuffer)
		} else {
			h.HSplitBuf(helpBuffer)
		}
	}
	return nil
}

// HelpCmd tries to open the given help page in a horizontal split
func (h *BufPane) HelpCmd(args []string) {
	if len(args) < 1 {
		// Open the default help if the user just typed "> help"
		h.openHelp("help")
	} else {
		if config.FindRuntimeFile(config.RTHelp, args[0]) != nil {
			err := h.openHelp(args[0])
			if err != nil {
				InfoBar.Error(err)
			}
		} else {
			InfoBar.Error("Sorry, no help for ", args[0])
		}
	}
}

// VSplitCmd opens a vertical split with file given in the first argument
// If no file is given, it opens an empty buffer in a new split
func (h *BufPane) VSplitCmd(args []string) {
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
func (h *BufPane) HSplitCmd(args []string) {
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
func (h *BufPane) EvalCmd(args []string) {
}

// NewTabCmd opens the given file in a new tab
func (h *BufPane) NewTabCmd(args []string) {
	width, height := screen.Screen.Size()
	iOffset := config.GetInfoBarOffset()
	if len(args) > 0 {
		for _, a := range args {
			b, err := buffer.NewBufferFromFile(a, buffer.BTDefault)
			if err != nil {
				InfoBar.Error(err)
				return
			}
			tp := NewTabFromBuffer(0, 0, width, height-1-iOffset, b)
			Tabs.AddTab(tp)
			Tabs.SetActive(len(Tabs.List) - 1)
		}
	} else {
		b := buffer.NewBufferFromString("", "", buffer.BTDefault)
		tp := NewTabFromBuffer(0, 0, width, height-iOffset, b)
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

	if option == "infobar" || option == "keymenu" {
		Tabs.Resize()
	}

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
func (h *BufPane) SetCmd(args []string) {
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
func (h *BufPane) SetLocalCmd(args []string) {
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
func (h *BufPane) ShowCmd(args []string) {
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
func (h *BufPane) ShowKeyCmd(args []string) {
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
func (h *BufPane) BindCmd(args []string) {
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
func (h *BufPane) UnbindCmd(args []string) {
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
func (h *BufPane) RunCmd(args []string) {
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
func (h *BufPane) QuitCmd(args []string) {
	h.Quit()
}

// SaveCmd saves the buffer optionally with an argument file name
func (h *BufPane) SaveCmd(args []string) {
	if len(args) == 0 {
		h.Save()
	} else {
		h.Buf.SaveAs(args[0])
	}
}

// ReplaceCmd runs search and replace
func (h *BufPane) ReplaceCmd(args []string) {
	if len(args) < 2 || len(args) > 4 {
		// We need to find both a search and replace expression
		InfoBar.Error("Invalid replace statement: " + strings.Join(args, " "))
		return
	}

	all := false
	noRegex := false

	if len(args) > 2 {
		for _, arg := range args[2:] {
			switch arg {
			case "-a":
				all = true
			case "-l":
				noRegex = true
			default:
				InfoBar.Error("Invalid flag: " + arg)
				return
			}
		}
	}

	search := args[0]

	if noRegex {
		search = regexp.QuoteMeta(search)
	}

	replace := []byte(args[1])
	replaceStr := args[1]

	var regex *regexp.Regexp
	var err error
	if h.Buf.Settings["ignorecase"].(bool) {
		regex, err = regexp.Compile("(?im)" + search)
	} else {
		regex, err = regexp.Compile("(?m)" + search)
	}
	if err != nil {
		// There was an error with the user's regex
		InfoBar.Error(err)
		return
	}

	nreplaced := 0
	start := h.Buf.Start()
	end := h.Buf.End()
	if h.Cursor.HasSelection() {
		start = h.Cursor.CurSelection[0]
		end = h.Cursor.CurSelection[1]
	}
	if all {
		nreplaced = h.Buf.ReplaceRegex(start, end, regex, replace)
	} else {
		inRange := func(l buffer.Loc) bool {
			return l.GreaterEqual(start) && l.LessThan(end)
		}

		searchLoc := start
		searching := true
		var doReplacement func()
		doReplacement = func() {
			locs, found, err := h.Buf.FindNext(search, start, end, searchLoc, true, !noRegex)
			if err != nil {
				InfoBar.Error(err)
				return
			}
			if !found || !inRange(locs[0]) || !inRange(locs[1]) {
				h.Cursor.ResetSelection()
				h.Buf.RelocateCursors()
				return
			}

			h.Cursor.SetSelectionStart(locs[0])
			h.Cursor.SetSelectionEnd(locs[1])

			InfoBar.YNPrompt("Perform replacement (y,n,esc)", func(yes, canceled bool) {
				if !canceled && yes {
					h.Buf.Replace(locs[0], locs[1], replaceStr)
					searchLoc = locs[0]
					searchLoc.X += utf8.RuneCount(replace)
					h.Cursor.Loc = searchLoc
					nreplaced++
				} else if !canceled && !yes {
					searchLoc = locs[0]
					searchLoc.X += utf8.RuneCount(replace)
				} else if canceled {
					h.Cursor.ResetSelection()
					h.Buf.RelocateCursors()
					return
				}
				if searching {
					doReplacement()
				}
			})
		}
		doReplacement()
	}

	h.Buf.RelocateCursors()

	if nreplaced > 1 {
		InfoBar.Message("Replaced ", nreplaced, " occurrences of ", search)
	} else if nreplaced == 1 {
		InfoBar.Message("Replaced ", nreplaced, " occurrence of ", search)
	} else {
		InfoBar.Message("Nothing matched ", search)
	}
}

// ReplaceAllCmd replaces search term all at once
func (h *BufPane) ReplaceAllCmd(args []string) {
	// aliased to Replace command
	h.ReplaceCmd(append(args, "-a"))
}

// TermCmd opens a terminal in the current view
func (h *BufPane) TermCmd(args []string) {
	ps := MainTab().Panes

	if len(args) == 0 {
		sh := os.Getenv("SHELL")
		if sh == "" {
			InfoBar.Error("Shell environment not found")
			return
		}
		args = []string{sh}
	}

	term := func(i int, newtab bool) {

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
		MainTab().Panes[i] = NewTermPane(v.X, v.Y, v.Width, v.Height, t, id)
		MainTab().SetActive(i)
	}

	// If there is only one open file we make a new tab instead of overwriting it
	newtab := len(MainTab().Panes) == 1 && len(Tabs.List) == 1

	if newtab {
		term(0, true)
		return
	}

	for i, p := range ps {
		if p.ID() == h.ID() {
			if h.Buf.Modified() {
				InfoBar.YNPrompt("Save changes to "+h.Buf.GetName()+" before closing? (y,n,esc)", func(yes, canceled bool) {
					if !canceled && !yes {
						term(i, false)
					} else if !canceled && yes {
						h.Save()
						term(i, false)
					}
				})
			} else {
				term(i, false)
			}
		}
	}
}

// HandleCommand handles input from the user
func (h *BufPane) HandleCommand(input string) {
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
