package lsp

var servers = `[language.rust]
command = "rls"
install = [
  ["rustup", "update"],
  ["rustup", "component", "add", "rls", "rust-analysis", "rust-src"],
]

[language.javascript]
command = "typescript-language-server"
args = ["--stdio"]
install = [["npm", "install", "-g", "typescript-language-server"]]

[language.typescript]
command = "typescript-language-server"
args = ["--stdio"]
install = [["npm", "install", "-g", "typescript-language-server"]]

[language.html]
command = "html-languageserver"
args = ["--stdio"]
install = [["npm", "install", "-g", "vscode-html-languageserver-bin"]]

[language.ocaml]
command = "ocaml-language-server"
args = ["--stdio"]
install = [["npm", "install", "-g", "ocaml-language-server"]]

[language.python]
command = "pyls"
install = [["pip", "install", "python-language-server"]]

[language.c]
command = "ccls"
args = []

[language.cpp]
command = "ccls"
args = []

[language.haskell]
command = "hie"
args = ["--lsp"]

[language.go]
command = "gopls"
args = ["serve"]
install = [["go", "get", "-u", "golang.org/x/tools/gopls"]]

[language.dart]
command = "dart_language_server"
install = [["pub", "global", "activate", "dart_language_server"]]

[language.ruby]
command = "solargraph"
args = ["stdio"]
install = [["gem", "install", "solargraph"]]

[language.css]
command = "css-languageserver"
args = ["--stdio"]
install = [["npm", "install", "-g", "vscode-css-languageserver-bin"]]

[language.scss]
command = "css-languageserver"
args = ["--stdio"]
install = [["npm", "install", "-g", "vscode-css-languageserver-bin"]]

[language.viml]
command = "vim-language-server"
args = ["--stdio"]
install = [["npm", "install", "-g", "vim-language-server"]]

[language.purescript]
command = "purescript-language-server"
args = ["--stdio"]
install = [["npm", "install", "-g", "purescript-language-server"]]
`
