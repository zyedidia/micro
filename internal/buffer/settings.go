package buffer

import (
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
)

func (b *Buffer) SetOptionNative(option string, nativeValue interface{}) error {
	b.Settings[option] = nativeValue

	if option == "fastdirty" {
		if !nativeValue.(bool) {
			e := calcHash(b, &b.origHash)
			if e == ErrFileTooLarge {
				b.Settings["fastdirty"] = false
			}
		}
	} else if option == "statusline" {
		screen.Redraw()
	} else if option == "filetype" {
		b.UpdateRules()
	} else if option == "fileformat" {
		switch b.Settings["fileformat"].(string) {
		case "unix":
			b.Endings = FFUnix
		case "dos":
			b.Endings = FFDos
		}
		b.isModified = true
	} else if option == "syntax" {
		if !nativeValue.(bool) {
			b.ClearMatches()
		} else {
			b.UpdateRules()
		}
	} else if option == "encoding" {
		b.isModified = true
	} else if option == "readonly" && b.Type.Kind == BTDefault.Kind {
		b.Type.Readonly = nativeValue.(bool)
	}

	return nil
}

// SetOption sets a given option to a value just for this buffer
func (b *Buffer) SetOption(option, value string) error {
	if _, ok := b.Settings[option]; !ok {
		return config.ErrInvalidOption
	}

	nativeValue, err := config.GetNativeValue(option, b.Settings[option], value)
	if err != nil {
		return err
	}

	return b.SetOptionNative(option, nativeValue)
}
