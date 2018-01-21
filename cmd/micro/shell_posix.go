// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package main

import (
	"github.com/zyedidia/micro/cmd/micro/shellwords"
)

const TermEmuSupported = true

func RunTermEmulator(input string, wait bool, getOutput bool) error {
	args, err := shellwords.Split(input)
	if err != nil {
		return err
	}
	err = CurView().StartTerminal(args, wait, false, "")
	return err
}
