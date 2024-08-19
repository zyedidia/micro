package buffer

import (
	"crypto/md5"
	"reflect"

	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
)

func (b *Buffer) ReloadSettings(reloadFiletype bool) {
	settings := config.ParsedSettings()

	if _, ok := b.LocalSettings["filetype"]; !ok && reloadFiletype {
		// need to update filetype before updating other settings based on it
		b.Settings["filetype"] = "unknown"
		if v, ok := settings["filetype"]; ok {
			b.Settings["filetype"] = v
		}
	}

	// update syntax rules, which will also update filetype if needed
	b.UpdateRules()
	settings["filetype"] = b.Settings["filetype"]

	config.InitLocalSettings(settings, b.Path)
	for k, v := range config.DefaultCommonSettings() {
		if k == "filetype" {
			// prevent recursion
			continue
		}
		if _, ok := config.VolatileSettings[k]; ok {
			// reload should not override volatile settings
			continue
		}
		if _, ok := b.LocalSettings[k]; ok {
			// reload should not override local settings
			continue
		}
		if _, ok := settings[k]; ok {
			b.DoSetOptionNative(k, settings[k])
		} else {
			b.DoSetOptionNative(k, v)
		}
	}
}

func (b *Buffer) DoSetOptionNative(option string, nativeValue interface{}) {
	if reflect.DeepEqual(b.Settings[option], nativeValue) {
		return
	}

	b.Settings[option] = nativeValue

	if option == "fastdirty" {
		if !nativeValue.(bool) {
			if b.Size() > LargeFileThreshold {
				b.Settings["fastdirty"] = true
			} else {
				if !b.isModified {
					calcHash(b, &b.origHash)
				} else {
					// prevent using an old stale origHash value
					b.origHash = [md5.Size]byte{}
				}
			}
		}
	} else if option == "statusline" {
		screen.Redraw()
	} else if option == "filetype" {
		b.ReloadSettings(false)
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
	} else if option == "hlsearch" {
		for _, buf := range OpenBuffers {
			if b.SharedBuffer == buf.SharedBuffer {
				buf.HighlightSearch = nativeValue.(bool)
			}
		}
	} else {
		for _, pl := range config.Plugins {
			if option == pl.Name {
				if nativeValue.(bool) {
					if !pl.Loaded {
						pl.Load()
					}
					_, err := pl.Call("init")
					if err != nil && err != config.ErrNoSuchFunction {
						screen.TermMessage(err)
					}
				} else if !nativeValue.(bool) && pl.Loaded {
					_, err := pl.Call("deinit")
					if err != nil && err != config.ErrNoSuchFunction {
						screen.TermMessage(err)
					}
				}
			}
		}
	}

	if b.OptionCallback != nil {
		b.OptionCallback(option, nativeValue)
	}
}

func (b *Buffer) SetOptionNative(option string, nativeValue interface{}) error {
	if err := config.OptionIsValid(option, nativeValue); err != nil {
		return err
	}

	b.DoSetOptionNative(option, nativeValue)
	b.LocalSettings[option] = true

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
