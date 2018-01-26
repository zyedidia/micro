// +build linux darwin dragonfly openbsd_amd64 freebsd

package main

import (
	"github.com/zyedidia/micro/cmd/micro/shellwords"
)

const TermEmuSupported = true

func RunTermEmulator(input string, wait bool, getOutput bool, callback string) error {
	args, err := shellwords.Split(input)
	if err != nil {
		return err
	}
	err = CurView().StartTerminal(args, wait, getOutput, callback)
	return err
}
