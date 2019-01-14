package buffer

import (
	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/screen"
)

// SetOption sets a given option to a value just for this buffer
func (b *Buffer) SetOption(option, value string) error {
	if _, ok := b.Settings[option]; !ok {
		return config.ErrInvalidOption
	}

	nativeValue, err := config.GetNativeValue(option, b.Settings[option], value)
	if err != nil {
		return err
	}

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
		b.isModified = true
	} else if option == "syntax" {
		if !nativeValue.(bool) {
			b.ClearMatches()
		} else {
			b.UpdateRules()
		}
	}

	return nil
}
