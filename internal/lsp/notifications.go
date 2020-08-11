package lsp

import "github.com/sourcegraph/go-lsp"

func (s *Server) DidOpen(filename, language, text string, version int) {
	doc := lsp.TextDocumentItem{
		URI:        lsp.DocumentURI("file://" + filename),
		LanguageID: language,
		Version:    version,
		Text:       text,
	}

	params := lsp.DidOpenTextDocumentParams{
		TextDocument: doc,
	}

	s.sendNotification("textDocument/didOpen", params)
}

func (s *Server) DidSave(filename string) {
	doc := lsp.TextDocumentIdentifier{
		URI: lsp.DocumentURI("file://" + filename),
	}

	params := lsp.DidSaveTextDocumentParams{
		TextDocument: doc,
	}
	s.sendNotification("textDocument/didSave", params)
}

func (s *Server) DidChange(filename string, version int, changes []lsp.TextDocumentContentChangeEvent) {
	doc := lsp.VersionedTextDocumentIdentifier{
		TextDocumentIdentifier: lsp.TextDocumentIdentifier{
			URI: lsp.DocumentURI("file://" + filename),
		},
		Version: version,
	}

	params := lsp.DidChangeTextDocumentParams{
		TextDocument:   doc,
		ContentChanges: changes,
	}
	s.sendNotification("textDocument/didChange", params)
}

func (s *Server) DidClose(filename string) {
	doc := lsp.TextDocumentIdentifier{
		URI: lsp.DocumentURI("file://" + filename),
	}

	params := lsp.DidCloseTextDocumentParams{
		TextDocument: doc,
	}
	s.sendNotification("textDocument/didClose", params)
}
