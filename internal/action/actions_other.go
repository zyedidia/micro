// +build plan9 nacl windows

package action

func (*BufPane) Suspend() bool {
	InfoBar.Error("Suspend is only supported on BSD/Linux")
	return false
}

func (*BufPane) Abort() bool {
	InfoBar.Error("Abort is only supported on BSD/Linux")
	return false
}
