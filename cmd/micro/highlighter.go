package main

import "github.com/zyedidia/micro/cmd/micro/highlight"

var syntaxFiles []*highlight.File

func LoadSyntaxFiles() {
	InitColorscheme()
	for _, f := range ListRuntimeFiles(RTSyntax) {
		data, err := f.Data()
		if err != nil {
			TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
		} else {
			LoadSyntaxFile(data, f.Name())
		}
	}
}

func LoadSyntaxFile(text []byte, filename string) {
	f, err := highlight.ParseFile(text)

	if err != nil {
		TermMessage("Syntax file error: " + filename + ": " + err.Error())
		return
	}

	syntaxFiles = append(syntaxFiles, f)
}
