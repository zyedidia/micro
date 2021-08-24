package config

import (
	"embed"
	"os"
	"strings"
)

//go:generate go run syntax/make_headers.go syntax

//go:embed colorschemes help plugins syntax
var runtime embed.FS

// AssetDir lists file names in folder
func AssetDir(name string) ([]string, error) {
	name = strings.TrimLeft(name, "runtime/")
	entries, err := runtime.ReadDir(name)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries), len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names, nil
}

// Asset returns a file content
func Asset(name string) ([]byte, error) {
	name = strings.TrimLeft(name, "runtime/")
	return runtime.ReadFile(name)
}
