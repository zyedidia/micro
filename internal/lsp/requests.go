package lsp

import (
	"log"

	"github.com/sourcegraph/go-lsp"
)

func (s *Server) DidOpen(filename, language, text string, version int) error {
	doc := lsp.TextDocumentItem{
		URI:        lsp.DocumentURI("file://" + filename),
		LanguageID: language,
		Version:    version,
		Text:       text,
	}

	params := lsp.DidOpenTextDocumentParams{
		TextDocument: doc,
	}

	resp, err := s.SendMessage("textDocument/didOpen", params)
	if err != nil {
		return err
	}

	log.Println("Received", string(resp))

	return nil
}
