// +build plan9 nacl windows

package action

func (*BufHandler) Suspend() bool {
	InfoBar.Error("Suspend is only supported on BSD/Linux")
	return false
}
