// +build plan9 nacl windows

package main

import "errors"

const TermEmuSupported = false

func RunTermEmulator(input string, wait bool, getOutput bool) error {
	return errors.New("Unsupported operating system")
}
