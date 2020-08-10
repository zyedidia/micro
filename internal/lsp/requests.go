package lsp

import (
	"encoding/json"

	"github.com/sourcegraph/go-lsp"
)

type RPCCompletion struct {
	RPCVersion string             `json:"jsonrpc"`
	ID         int                `json:"id"`
	Result     lsp.CompletionList `json:"result"`
}

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

	s.SendMessage("textDocument/didOpen", params)
}

func (s *Server) DidSave(filename string) {
	doc := lsp.TextDocumentIdentifier{
		URI: lsp.DocumentURI("file://" + filename),
	}

	params := lsp.DidSaveTextDocumentParams{
		TextDocument: doc,
	}
	s.SendMessage("textDocument/didSave", params)
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
	s.SendMessage("textDocument/didChange", params)
}

func (s *Server) DidClose(filename string) {
	doc := lsp.TextDocumentIdentifier{
		URI: lsp.DocumentURI("file://" + filename),
	}

	params := lsp.DidCloseTextDocumentParams{
		TextDocument: doc,
	}
	s.SendMessage("textDocument/didClose", params)
}

func (s *Server) DocumentFormat() {

}

func (s *Server) DocumentRangeFormat() {

}

func (s *Server) Completion(filename string, pos lsp.Position) ([]lsp.CompletionItem, error) {
	cc := lsp.CompletionContext{
		TriggerKind: lsp.CTKInvoked,
	}

	docpos := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.DocumentURI("file://" + filename),
		},
		Position: pos,
	}

	params := lsp.CompletionParams{
		TextDocumentPositionParams: docpos,
		Context:                    cc,
	}
	resp, err := s.SendMessageGetResponse("textDocument/completion", params)
	if err != nil {
		return nil, err
	}

	var r RPCCompletion
	err = json.Unmarshal(resp, &r)
	if err != nil {
		return nil, err
	}

	return r.Result.Items, nil
}

func (s *Server) CompletionResolve() {

}
