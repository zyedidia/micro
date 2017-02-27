package main

import (
	"os"
	"path/filepath"
)

type Files map[string]File

type File struct {
	X    int
	Y    int
	Path string
	Name string
}

func getFilesInCurrentDir() (files Files) {
	files = make(map[string]File)

	err := filepath.Walk(".", func(path string, f os.FileInfo, err error) error {
		// Only bother with files
		if !f.IsDir() {
			if path[:1] == "." {
				return nil
			}
			_, ok := files[path]
			if !ok {
				file := File{X: 0, Y: 0, Name: f.Name(), Path: path}
				files[path] = file
			}
		}
		return nil
	})

	if err != nil {
		TermMessage("Error", err)
	}
	return files
}
