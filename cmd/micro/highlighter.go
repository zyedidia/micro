package main

import "github.com/zyedidia/highlight"

var syntaxDefs []*highlight.Def

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

	highlight.ResolveIncludes(syntaxDefs)
}

func LoadSyntaxFile(text []byte, filename string) {
	def, err := highlight.ParseDef(text)

	if err != nil {
		TermMessage("Syntax file error: " + filename + ": " + err.Error())
		return
	}

	syntaxDefs = append(syntaxDefs, def)
}
