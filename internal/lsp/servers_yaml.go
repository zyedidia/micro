package lsp

var servers_internal = []byte(`language:
  rust:
    command: rls
    install: [["rustup", "update"], ["rustup", "component", "add", "rls", "rust-analysis", "rust-src"]]
  javascript:
    command: typescript-language-server
    args: ["--stdio"]
    install: [["npm", "install", "-g", "typescript-language-server"]]
  typescript:
    command: typescript-language-server
    args: ["--stdio"]
    install: [["npm", "install", "-g", "typescript-language-server"]]
  html:
    command: html-languageserver
    args: ["--stdio"]
    install: [["npm", "install", "-g", "vscode-html-languageserver-bin"]]
  ocaml:
    command: ocaml-language-server
    args: ["--stdio"]
    install: [["npm", "install", "-g", "ocaml-language-server"]]
  python:
    command: pyls
    install: [["pip", "install", "python-language-server"]]
  c:
    command: clangd
    args: []
  cpp:
    command: clangd
    args: []
  haskell:
    command: hie
    args: ["--lsp"]
  go:
    command: gopls
    args: ["serve"]
    install: [["go", "get", "-u", "golang.org/x/tools/gopls"]]
  dart:
    command: dart_language_server
    install: [["pub", "global", "activate", "dart_language_server"]]
  ruby:
    command: solargraph
    args: ["stdio"]
    install: [["gem", "install", "solargraph"]]
  css:
    command: css-languageserver
    args: ["--stdio"]
    install: [["npm", "install", "-g", "vscode-css-languageserver-bin"]]
  scss:
    command: css-languageserver
    args: ["--stdio"]
    install: [["npm", "install", "-g", "vscode-css-languageserver-bin"]]
  viml:
    command: vim-language-server
    args: ["--stdio"]
    install: [["npm", "install", "-g", "vim-language-server"]]
  purescript:
    command: purescript-language-server
    args: ["--stdio"]
    install: [["npm", "install", "-g", "purescript-language-server"]]
  verilog:
    command: svls
    install: [["cargo", "install", "svls"]]
  d:
    command: serve-d
`)
