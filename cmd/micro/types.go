package main

import "strings"

const (
	// ExtensionArmorGPG armored gpg encrypted file extension
	ExtensionArmorGPG = "asc"
	// ExtensionGPG gpg encrypted file extension
	ExtensionGPG = "gpg"
)

// Encrypted determines if a file is encrypted
func Encrypted(filename string) bool {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		for _, part := range parts[1:] {
			if part == ExtensionArmorGPG || part == ExtensionGPG {
				return true
			}
		}
	}
	return false
}
