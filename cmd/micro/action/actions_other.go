// +build plan9 nacl windows

package action

func (*BufHandler) Suspend() bool {
	// TODO: error message
	return false
}
