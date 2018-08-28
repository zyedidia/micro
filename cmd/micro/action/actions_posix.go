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
	screen.TempFini()

	// suspend the process
	pid := syscall.Getpid()
	err := syscall.Kill(pid, syscall.SIGSTOP)
	if err != nil {
		util.TermMessage(err)
	}

	screen.TempStart()

	return false
}
