package main

import "strings"

const (
	// ExtensionArmorGPG armored gpg encrypted file extension
	ExtensionArmorGPG = ".asc"
	// ExtensionGPG gpg encrypted file extension
	ExtensionGPG = ".gpg"
)

// Encrypted determines if a file is encrypted
func Encrypted(filename string) bool {
	return strings.HasSuffix(filename, ExtensionArmorGPG) ||
		strings.HasSuffix(filename, ExtensionGPG)
}
