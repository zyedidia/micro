package lsp

// mappings for when micro filetypes don't match LSP language identifiers
var languages = map[string]string{
	"batch":           "bat",
	"c++":             "cpp",
	"git-rebase-todo": "git-rebase",
	"html4":           "html",
	"html5":           "html",
	"python2":         "python",
	"shell":           "shellscript",
	// "tex": "latex",
}

func Filetype(ft string) string {
	if l, ok := languages[ft]; ok {
		return l
	}
	return ft
}
