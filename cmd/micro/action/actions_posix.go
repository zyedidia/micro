// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package action

import (
	"syscall"

	"github.com/zyedidia/micro/cmd/micro/screen"
	"github.com/zyedidia/micro/cmd/micro/util"
)

// Suspend sends micro to the background. This is the same as pressing CtrlZ in most unix programs.
// This only works on linux and has no default binding.
// This code was adapted from the suspend code in nsf/godit
func (*BufHandler) Suspend() bool {
	screenWasNil := screen.Screen == nil

	if !screenWasNil {
		screen.Screen.Fini()
		screen.Screen = nil
	}

	// suspend the process
	pid := syscall.Getpid()
	err := syscall.Kill(pid, syscall.SIGSTOP)
	if err != nil {
		util.TermMessage(err)
	}

	if !screenWasNil {
		screen.Init()
	}

	return false
}
