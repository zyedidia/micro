// +build plan9 nacl windows

package main

func (*BufActionHandler) Suspend() bool {
	// TODO: error message
	return false
}
