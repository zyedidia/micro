// +build plan9 nacl windows

package main

const TermEmuSupported = false

func RunTermEmulator(input string, wait bool, getOutput bool) string {
	return "Unsupported"
}
