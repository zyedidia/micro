package action

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/zyedidia/json5"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/tcell"
)

func createBindingsIfNotExist(fname string) {
	if _, e := os.Stat(fname); os.IsNotExist(e) {
		ioutil.WriteFile(fname, []byte("{}"), 0644)
	}
}

// InitBindings intializes the bindings map by reading from bindings.json
func InitBindings() {
	config.Bindings = DefaultBindings()

	var parsed map[string]string
	defaults := DefaultBindings()

	filename := filepath.Join(config.ConfigDir, "bindings.json")
	createBindingsIfNotExist(filename)

	if _, e := os.Stat(filename); e == nil {
		input, err := ioutil.ReadFile(filename)
		if err != nil {
			screen.TermMessage("Error reading bindings.json file: " + err.Error())
			return
		}

		err = json5.Unmarshal(input, &parsed)
		if err != nil {
			screen.TermMessage("Error reading bindings.json:", err.Error())
		}
	}

	for k, v := range defaults {
		BindKey(k, v)
	}
	for k, v := range parsed {
		BindKey(k, v)
	}
}

func BindKey(k, v string) {
	event, ok := findEvent(k)
	if !ok {
		screen.TermMessage(k, "is not a bindable event")
	}

	switch e := event.(type) {
	case KeyEvent:
		BufMapKey(e, v)
	case MouseEvent:
		BufMapMouse(e, v)
	case RawEvent:
		BufMapKey(e, v)
	}

	config.Bindings[k] = v
}

// findEvent will find binding Key 'b' using string 'k'
func findEvent(k string) (b Event, ok bool) {
	modifiers := tcell.ModNone

	// First, we'll strip off all the modifiers in the name and add them to the
	// ModMask
modSearch:
	for {
		switch {
		case strings.HasPrefix(k, "-"):
			// We optionally support dashes between modifiers
			k = k[1:]
		case strings.HasPrefix(k, "Ctrl") && k != "CtrlH":
			// CtrlH technically does not have a 'Ctrl' modifier because it is really backspace
			k = k[4:]
			modifiers |= tcell.ModCtrl
		case strings.HasPrefix(k, "Alt"):
			k = k[3:]
			modifiers |= tcell.ModAlt
		case strings.HasPrefix(k, "Shift"):
			k = k[5:]
			modifiers |= tcell.ModShift
		case strings.HasPrefix(k, "\x1b"):
			screen.Screen.RegisterRawSeq(k)
			return RawEvent{
				esc: k,
			}, true
		default:
			break modSearch
		}
	}

	if len(k) == 0 {
		return KeyEvent{}, false
	}

	// Control is handled in a special way, since the terminal sends explicitly
	// marked escape sequences for control keys
	// We should check for Control keys first
	if modifiers&tcell.ModCtrl != 0 {
		// see if the key is in bindingKeys with the Ctrl prefix.
		k = string(unicode.ToUpper(rune(k[0]))) + k[1:]
		if code, ok := keyEvents["Ctrl"+k]; ok {
			var r tcell.Key
			// Special case for escape, for some reason tcell doesn't send it with the esc character
			if code < 256 && code != 27 {
				r = code
			}
			// It is, we're done.
			return KeyEvent{
				code: code,
				mod:  modifiers,
				r:    rune(r),
			}, true
		}
	}

	// See if we can find the key in bindingKeys
	if code, ok := keyEvents[k]; ok {
		var r tcell.Key
		// Special case for escape, for some reason tcell doesn't send it with the esc character
		if code < 256 && code != 27 {
			r = code
		}
		return KeyEvent{
			code: code,
			mod:  modifiers,
			r:    rune(r),
		}, true
	}

	// See if we can find the key in bindingMouse
	if code, ok := mouseEvents[k]; ok {
		return MouseEvent{
			btn: code,
			mod: modifiers,
		}, true
	}

	// If we were given one character, then we've got a rune.
	if len(k) == 1 {
		return KeyEvent{
			code: tcell.KeyRune,
			mod:  modifiers,
			r:    rune(k[0]),
		}, true
	}

	// We don't know what happened.
	return KeyEvent{}, false
}

// TryBindKey tries to bind a key by writing to config.ConfigDir/bindings.json
// Returns true if the keybinding already existed and a possible error
func TryBindKey(k, v string, overwrite bool) (bool, error) {
	var e error
	var parsed map[string]string

	filename := filepath.Join(config.ConfigDir, "bindings.json")
	createBindingsIfNotExist(filename)
	if _, e = os.Stat(filename); e == nil {
		input, err := ioutil.ReadFile(filename)
		if err != nil {
			return false, errors.New("Error reading bindings.json file: " + err.Error())
		}

		err = json5.Unmarshal(input, &parsed)
		if err != nil {
			return false, errors.New("Error reading bindings.json: " + err.Error())
		}

		key, ok := findEvent(k)
		if !ok {
			return false, errors.New("Invalid event " + k)
		}

		found := false
		for ev := range parsed {
			if e, ok := findEvent(ev); ok {
				if e == key {
					if overwrite {
						parsed[ev] = v
					}
					found = true
					break
				}
			}
		}

		if found && !overwrite {
			return true, nil
		} else if !found {
			parsed[k] = v
		}

		BindKey(k, v)

		txt, _ := json.MarshalIndent(parsed, "", "    ")
		return true, ioutil.WriteFile(filename, append(txt, '\n'), 0644)
	}
	return false, e
}

// UnbindKey removes the binding for a key from the bindings.json file
func UnbindKey(k string) error {
	var e error
	var parsed map[string]string

	filename := filepath.Join(config.ConfigDir, "bindings.json")
	createBindingsIfNotExist(filename)
	if _, e = os.Stat(filename); e == nil {
		input, err := ioutil.ReadFile(filename)
		if err != nil {
			return errors.New("Error reading bindings.json file: " + err.Error())
		}

		err = json5.Unmarshal(input, &parsed)
		if err != nil {
			return errors.New("Error reading bindings.json: " + err.Error())
		}

		key, ok := findEvent(k)
		if !ok {
			return errors.New("Invalid event " + k)
		}

		for ev := range parsed {
			if e, ok := findEvent(ev); ok {
				if e == key {
					delete(parsed, ev)
					break
				}
			}
		}

		defaults := DefaultBindings()
		if a, ok := defaults[k]; ok {
			BindKey(k, a)
		} else if _, ok := config.Bindings[k]; ok {
			BufUnmap(key)
			delete(config.Bindings, k)
		}

		txt, _ := json.MarshalIndent(parsed, "", "    ")
		return ioutil.WriteFile(filename, append(txt, '\n'), 0644)
	}
	return e
}

var mouseEvents = map[string]tcell.ButtonMask{
	"MouseLeft":       tcell.Button1,
	"MouseMiddle":     tcell.Button2,
	"MouseRight":      tcell.Button3,
	"MouseWheelUp":    tcell.WheelUp,
	"MouseWheelDown":  tcell.WheelDown,
	"MouseWheelLeft":  tcell.WheelLeft,
	"MouseWheelRight": tcell.WheelRight,
}

var keyEvents = map[string]tcell.Key{
	"Up":             tcell.KeyUp,
	"Down":           tcell.KeyDown,
	"Right":          tcell.KeyRight,
	"Left":           tcell.KeyLeft,
	"UpLeft":         tcell.KeyUpLeft,
	"UpRight":        tcell.KeyUpRight,
	"DownLeft":       tcell.KeyDownLeft,
	"DownRight":      tcell.KeyDownRight,
	"Center":         tcell.KeyCenter,
	"PageUp":         tcell.KeyPgUp,
	"PageDown":       tcell.KeyPgDn,
	"Home":           tcell.KeyHome,
	"End":            tcell.KeyEnd,
	"Insert":         tcell.KeyInsert,
	"Delete":         tcell.KeyDelete,
	"Help":           tcell.KeyHelp,
	"Exit":           tcell.KeyExit,
	"Clear":          tcell.KeyClear,
	"Cancel":         tcell.KeyCancel,
	"Print":          tcell.KeyPrint,
	"Pause":          tcell.KeyPause,
	"Backtab":        tcell.KeyBacktab,
	"F1":             tcell.KeyF1,
	"F2":             tcell.KeyF2,
	"F3":             tcell.KeyF3,
	"F4":             tcell.KeyF4,
	"F5":             tcell.KeyF5,
	"F6":             tcell.KeyF6,
	"F7":             tcell.KeyF7,
	"F8":             tcell.KeyF8,
	"F9":             tcell.KeyF9,
	"F10":            tcell.KeyF10,
	"F11":            tcell.KeyF11,
	"F12":            tcell.KeyF12,
	"F13":            tcell.KeyF13,
	"F14":            tcell.KeyF14,
	"F15":            tcell.KeyF15,
	"F16":            tcell.KeyF16,
	"F17":            tcell.KeyF17,
	"F18":            tcell.KeyF18,
	"F19":            tcell.KeyF19,
	"F20":            tcell.KeyF20,
	"F21":            tcell.KeyF21,
	"F22":            tcell.KeyF22,
	"F23":            tcell.KeyF23,
	"F24":            tcell.KeyF24,
	"F25":            tcell.KeyF25,
	"F26":            tcell.KeyF26,
	"F27":            tcell.KeyF27,
	"F28":            tcell.KeyF28,
	"F29":            tcell.KeyF29,
	"F30":            tcell.KeyF30,
	"F31":            tcell.KeyF31,
	"F32":            tcell.KeyF32,
	"F33":            tcell.KeyF33,
	"F34":            tcell.KeyF34,
	"F35":            tcell.KeyF35,
	"F36":            tcell.KeyF36,
	"F37":            tcell.KeyF37,
	"F38":            tcell.KeyF38,
	"F39":            tcell.KeyF39,
	"F40":            tcell.KeyF40,
	"F41":            tcell.KeyF41,
	"F42":            tcell.KeyF42,
	"F43":            tcell.KeyF43,
	"F44":            tcell.KeyF44,
	"F45":            tcell.KeyF45,
	"F46":            tcell.KeyF46,
	"F47":            tcell.KeyF47,
	"F48":            tcell.KeyF48,
	"F49":            tcell.KeyF49,
	"F50":            tcell.KeyF50,
	"F51":            tcell.KeyF51,
	"F52":            tcell.KeyF52,
	"F53":            tcell.KeyF53,
	"F54":            tcell.KeyF54,
	"F55":            tcell.KeyF55,
	"F56":            tcell.KeyF56,
	"F57":            tcell.KeyF57,
	"F58":            tcell.KeyF58,
	"F59":            tcell.KeyF59,
	"F60":            tcell.KeyF60,
	"F61":            tcell.KeyF61,
	"F62":            tcell.KeyF62,
	"F63":            tcell.KeyF63,
	"F64":            tcell.KeyF64,
	"CtrlSpace":      tcell.KeyCtrlSpace,
	"CtrlA":          tcell.KeyCtrlA,
	"CtrlB":          tcell.KeyCtrlB,
	"CtrlC":          tcell.KeyCtrlC,
	"CtrlD":          tcell.KeyCtrlD,
	"CtrlE":          tcell.KeyCtrlE,
	"CtrlF":          tcell.KeyCtrlF,
	"CtrlG":          tcell.KeyCtrlG,
	"CtrlH":          tcell.KeyCtrlH,
	"CtrlI":          tcell.KeyCtrlI,
	"CtrlJ":          tcell.KeyCtrlJ,
	"CtrlK":          tcell.KeyCtrlK,
	"CtrlL":          tcell.KeyCtrlL,
	"CtrlM":          tcell.KeyCtrlM,
	"CtrlN":          tcell.KeyCtrlN,
	"CtrlO":          tcell.KeyCtrlO,
	"CtrlP":          tcell.KeyCtrlP,
	"CtrlQ":          tcell.KeyCtrlQ,
	"CtrlR":          tcell.KeyCtrlR,
	"CtrlS":          tcell.KeyCtrlS,
	"CtrlT":          tcell.KeyCtrlT,
	"CtrlU":          tcell.KeyCtrlU,
	"CtrlV":          tcell.KeyCtrlV,
	"CtrlW":          tcell.KeyCtrlW,
	"CtrlX":          tcell.KeyCtrlX,
	"CtrlY":          tcell.KeyCtrlY,
	"CtrlZ":          tcell.KeyCtrlZ,
	"CtrlLeftSq":     tcell.KeyCtrlLeftSq,
	"CtrlBackslash":  tcell.KeyCtrlBackslash,
	"CtrlRightSq":    tcell.KeyCtrlRightSq,
	"CtrlCarat":      tcell.KeyCtrlCarat,
	"CtrlUnderscore": tcell.KeyCtrlUnderscore,
	"Tab":            tcell.KeyTab,
	"Esc":            tcell.KeyEsc,
	"Escape":         tcell.KeyEscape,
	"Enter":          tcell.KeyEnter,
	"Backspace":      tcell.KeyBackspace2,
	"OldBackspace":   tcell.KeyBackspace,

	// I renamed these keys to PageUp and PageDown but I don't want to break someone's keybindings
	"PgUp":   tcell.KeyPgUp,
	"PgDown": tcell.KeyPgDn,
}
