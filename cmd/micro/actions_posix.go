// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package main

import "syscall"

// Suspend sends micro to the background. This is the same as pressing CtrlZ in most unix programs.
// This only works on linux and has no default binding.
// This code was adapted from the suspend code in nsf/godit
func (v *View) Suspend(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Suspend", v) {
		return false
	}

	screenWasNil := screen == nil

	if !screenWasNil {
		screen.Fini()
		screen = nil
	}

	// suspend the process
	pid := syscall.Getpid()
	err := syscall.Kill(pid, syscall.SIGSTOP)
	if err != nil {
		TermMessage(err)
	}

	if !screenWasNil {
		InitScreen()
	}

	if usePlugin {
		return PostActionCall("Suspend", v)
	}
	return true
}
