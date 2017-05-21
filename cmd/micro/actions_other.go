// +build !linux

package main

func (v *View) Suspend(usePlugin bool) bool { 
	TermError("Suspend is only supported on Linux")

	return false
}
