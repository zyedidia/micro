package main

import (
	"io/ioutil"
	"path"
	"path/filepath"
)

const (
	FILE_ColorScheme = "colorscheme"
	FILE_Syntax      = "syntax"
	FILE_Help        = "help"
)

// ExtensionFile allows the program to read runtime data like colorschemes or syntax files
type ExtensionFile interface {
	// Name returns a name of the file without paths or extensions
	Name() string
	// Data returns the content of the file.
	Data() ([]byte, error)
}

// allFiles contains all available files, mapped by filetype
var allFiles map[string][]ExtensionFile

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

func AddFile(fileType string, file ExtensionFile) {
	if allFiles == nil {
		allFiles = make(map[string][]ExtensionFile)
	}
	allFiles[fileType] = append(allFiles[fileType], file)
}

func AddFilesFromDirectory(fileType, directory, pattern string) {
	files, _ := ioutil.ReadDir(directory)
	for _, f := range files {
		if ok, _ := filepath.Match(pattern, f.Name()); !f.IsDir() && ok {
			fullPath := filepath.Join(directory, f.Name())
			AddFile(fileType, realFile(fullPath))
		}
	}
}

func AddFilesFromAssets(fileType, directory, pattern string) {
	files, err := AssetDir(directory)
	if err != nil {
		return
	}
	for _, f := range files {
		if ok, _ := path.Match(pattern, f); ok {
			AddFile(fileType, assetFile(path.Join(directory, f)))
		}
	}
}

func FindExtensionFile(fileType, name string) ExtensionFile {
	for _, f := range ListExtensionFiles(fileType) {
		if f.Name() == name {
			return f
		}
	}
	return nil
}

func ListExtensionFiles(fileType string) []ExtensionFile {
	if files, ok := allFiles[fileType]; ok {
		return files
	} else {
		return []ExtensionFile{}
	}
}

func InitExtensionFiles() {
	add := func(fileType, dir, pattern string) {
		AddFilesFromDirectory(fileType, filepath.Join(configDir, dir), pattern)
		AddFilesFromAssets(fileType, path.Join("runtime", dir), pattern)
	}

	add(FILE_ColorScheme, "colorschemes", "*.micro")
	add(FILE_Syntax, "syntax", "*.micro")
	add(FILE_Help, "help", "*.md")
}
