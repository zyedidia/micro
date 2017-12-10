// +build plan9 nacl windows

package main

func (v *View) Suspend(usePlugin bool) bool {
	messenger.Error("Suspend is only supported on Posix")

	return false
}
