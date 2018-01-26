// +build !linux,!darwin,!freebsd,!dragonfly,!openbsd_amd64

package main

import "errors"

const TermEmuSupported = false

func RunTermEmulator(input string, wait bool, getOutput bool) error {
	return errors.New("Unsupported operating system")
}
