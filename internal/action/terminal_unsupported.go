// +build !linux,!darwin,!freebsd,!dragonfly,!openbsd_amd64

package action

import "errors"

// TermEmuSupported is a constant that marks if the terminal emulator is supported
const TermEmuSupported = false

// RunTermEmulator returns an error for unsupported systems (non-unix systems
func RunTermEmulator(input string, wait bool, getOutput bool, callback func(out string, userargs []interface{}), userargs []interface{}) error {
	return errors.New("Unsupported operating system")
}
