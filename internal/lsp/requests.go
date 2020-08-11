package lsp

import (
	"encoding/json"

	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type RPCCompletion struct {
	RPCVersion string             `json:"jsonrpc"`
	ID         int                `json:"id"`
	Result     lsp.CompletionList `json:"result"`
}

type RPCHover struct {
	RPCVersion string    `json:"jsonrpc"`
	ID         int       `json:"id"`
	Result     lsp.Hover `json:"result"`
}

func Position(x, y int) lsp.Position {
	return lsp.Position{
		Line:      float64(y),
		Character: float64(x),
	}
}

func (s *Server) DocumentFormat() {

}

func (s *Server) DocumentRangeFormat() {

}

func (s *Server) Completion(filename string, pos lsp.Position) ([]lsp.CompletionItem, error) {
	cc := lsp.CompletionContext{
		TriggerKind: lsp.Invoked,
	}

	docpos := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: uri.File(filename),
		},
		Position: pos,
	}

	params := lsp.CompletionParams{
		TextDocumentPositionParams: docpos,
		Context:                    &cc,
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

func (s *Server) Hover(filename string, pos lsp.Position) (string, error) {
	params := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: uri.File(filename),
		},
		Position: pos,
	}

	resp, err := s.sendRequest("textDocument/hover", params)
	if err != nil {
		return "", err
	}

	var r RPCHover
	err = json.Unmarshal(resp, &r)
	if err != nil {
		return "", err
	}

	return r.Result.Contents.Value, nil
}
