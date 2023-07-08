package config

import (
	"embed"
	"path/filepath"
	"strings"
)

//go:generate go run syntax/make_headers.go syntax

//go:embed colorschemes help plugins syntax
var runtime embed.FS

func fixPath(name string) string {
	return strings.TrimLeft(filepath.ToSlash(name), "runtime/")
}

// AssetDir lists file names in folder
func AssetDir(name string) ([]string, error) {
	name = fixPath(name)
	entries, err := runtime.ReadDir(name)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names, nil
}

// Asset returns a file content
func Asset(name string) ([]byte, error) {
	name = fixPath(name)
	return runtime.ReadFile(name)
}
