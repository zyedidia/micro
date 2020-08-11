package lsp

import (
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (s *Server) DidOpen(filename, language, text string, version *uint64) {
	doc := lsp.TextDocumentItem{
		URI:        uri.File(filename),
		LanguageID: lsp.LanguageIdentifier(language),
		Version:    float64(*version), // not sure why this is a float on go.lsp.dev
		Text:       text,
	}

	params := lsp.DidOpenTextDocumentParams{
		TextDocument: doc,
	}

	go s.sendNotification("textDocument/didOpen", params)
}

func (s *Server) DidSave(filename string) {
	doc := lsp.TextDocumentIdentifier{
		URI: uri.File(filename),
	}

	params := lsp.DidSaveTextDocumentParams{
		TextDocument: doc,
	}
	go s.sendNotification("textDocument/didSave", params)
}

func (s *Server) DidChange(filename string, version *uint64, changes []lsp.TextDocumentContentChangeEvent) {
	doc := lsp.VersionedTextDocumentIdentifier{
		TextDocumentIdentifier: lsp.TextDocumentIdentifier{
			URI: uri.File(filename),
		},
		Version: version,
	}

	params := lsp.DidChangeTextDocumentParams{
		TextDocument:   doc,
		ContentChanges: changes,
	}
	go s.sendNotification("textDocument/didChange", params)
}

func (s *Server) DidClose(filename string) {
	doc := lsp.TextDocumentIdentifier{
		URI: uri.File(filename),
	}

	params := lsp.DidCloseTextDocumentParams{
		TextDocument: doc,
	}
	go s.sendNotification("textDocument/didClose", params)
}
