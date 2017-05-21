package main

import "syscall"

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
	tid := syscall.Gettid()
	err := syscall.Tgkill(pid, tid, syscall.SIGSTOP)
	if err != nil {
		panic(err)
	}

	if !screenWasNil {
		InitScreen()
	}

	if usePlugin {
		return PostActionCall("Suspend", v)
	}
	return true
}
