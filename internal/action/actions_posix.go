//go:build linux || darwin || dragonfly || solaris || openbsd || netbsd || freebsd

package action

import (
	"syscall"

	"github.com/zyedidia/micro/v2/internal/screen"
)

// Suspend sends micro to the background. This is the same as pressing CtrlZ in most unix programs.
// This only works on linux and has no default binding.
// This code was adapted from the suspend code in nsf/godit
func (*BufPane) Suspend() bool {
	screenb := screen.TempFini()

	// suspend the process
	pid := syscall.Getpid()
	err := syscall.Kill(pid, syscall.SIGSTOP)
	if err != nil {
		screen.TermMessage(err)
	}

	screen.TempStart(screenb)

	return false
}
