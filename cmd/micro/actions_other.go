// +build !linux

package main

func (v *View) Suspend(usePlugin bool) bool {
	messenger.Error("Suspend is only supported on Linux")

	return false
}
