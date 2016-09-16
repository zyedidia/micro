package main

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

const (
	FILE_ColorScheme = "colorscheme"
	FILE_Syntax      = "syntax"
	FILE_Help        = "help"
)

// RuntimeFile allows the program to read runtime data like colorschemes or syntax files
type RuntimeFile interface {
	// Name returns a name of the file without paths or extensions
	Name() string
	// Data returns the content of the file.
	Data() ([]byte, error)
}

// allFiles contains all available files, mapped by filetype
var allFiles map[string][]RuntimeFile

// some file on filesystem
type realFile string

// some asset file
type assetFile string

// some file on filesystem but with a different name
type namedFile struct {
	realFile
	name string
}

func (rf realFile) Name() string {
	fn := filepath.Base(string(rf))
	return fn[:len(fn)-len(filepath.Ext(fn))]
}

func (rf realFile) Data() ([]byte, error) {
	return ioutil.ReadFile(string(rf))
}

func (af assetFile) Name() string {
	fn := path.Base(string(af))
	return fn[:len(fn)-len(path.Ext(fn))]
}

func (af assetFile) Data() ([]byte, error) {
	return Asset(string(af))
}

func (nf namedFile) Name() string {
	return nf.name
}

// AddRuntimeFile registers a file for the given filetype
func AddRuntimeFile(fileType string, file RuntimeFile) {
	if allFiles == nil {
		allFiles = make(map[string][]RuntimeFile)
	}
	allFiles[fileType] = append(allFiles[fileType], file)
}

// AddRuntimeFilesFromDirectory registers each file from the given directory for
// the filetype which matches the file-pattern
func AddRuntimeFilesFromDirectory(fileType, directory, pattern string) {
	files, _ := ioutil.ReadDir(directory)
	for _, f := range files {
		if ok, _ := filepath.Match(pattern, f.Name()); !f.IsDir() && ok {
			fullPath := filepath.Join(directory, f.Name())
			AddRuntimeFile(fileType, realFile(fullPath))
		}
	}
}

// AddRuntimeFilesFromDirectory registers each file from the given asset-directory for
// the filetype which matches the file-pattern
func AddRuntimeFilesFromAssets(fileType, directory, pattern string) {
	files, err := AssetDir(directory)
	if err != nil {
		return
	}
	for _, f := range files {
		if ok, _ := path.Match(pattern, f); ok {
			AddRuntimeFile(fileType, assetFile(path.Join(directory, f)))
		}
	}
}

// FindRuntimeFile finds a runtime file of the given filetype and name
// will return nil if no file was found
func FindRuntimeFile(fileType, name string) RuntimeFile {
	for _, f := range ListRuntimeFiles(fileType) {
		if f.Name() == name {
			return f
		}
	}
	return nil
}

// Lists all known runtime files for the given filetype
func ListRuntimeFiles(fileType string) []RuntimeFile {
	if files, ok := allFiles[fileType]; ok {
		return files
	} else {
		return []RuntimeFile{}
	}
}

// Initializes all assets file and the config directory
func InitRuntimeFiles() {
	add := func(fileType, dir, pattern string) {
		AddRuntimeFilesFromDirectory(fileType, filepath.Join(configDir, dir), pattern)
		AddRuntimeFilesFromAssets(fileType, path.Join("runtime", dir), pattern)
	}

	add(FILE_ColorScheme, "colorschemes", "*.micro")
	add(FILE_Syntax, "syntax", "*.micro")
	add(FILE_Help, "help", "*.md")
}

// Allows plugin scripts to read the content of a runtime file
func PluginReadRuntimeFile(fileType, name string) string {
	if file := FindRuntimeFile(fileType, name); file != nil {
		if data, err := file.Data(); err == nil {
			return string(data)
		}
	}
	return ""
}

// Allows plugins to lists all runtime files of the given type
func PluginListRuntimeFiles(fileType string) []string {
	files := ListRuntimeFiles(fileType)
	result := make([]string, len(files))
	for i, f := range files {
		result[i] = f.Name()
	}
	return result
}

func PluginAddRuntimeFile(plugin, filetype, filePath string) {
	fullpath := filepath.Join(configDir, "plugins", plugin, filePath)
	if _, err := os.Stat(fullpath); err == nil {
		AddRuntimeFile(filetype, realFile(fullpath))
	} else {
		fullpath = path.Join("runtime", "plugins", plugin, filePath)
		AddRuntimeFile(filetype, assetFile(fullpath))
	}
}
