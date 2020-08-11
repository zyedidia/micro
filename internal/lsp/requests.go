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
	resp, err := s.sendRequest("textDocument/completion", params)
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
