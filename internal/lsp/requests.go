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

type RPCCompletionAlternate struct {
	RPCVersion string               `json:"jsonrpc"`
	ID         int                  `json:"id"`
	Result     []lsp.CompletionItem `json:"result"`
}

type RPCHover struct {
	RPCVersion string    `json:"jsonrpc"`
	ID         int       `json:"id"`
	Result     lsp.Hover `json:"result"`
}

type RPCFormat struct {
	RPCVersion string         `json:"jsonrpc"`
	ID         int            `json:"id"`
	Result     []lsp.TextEdit `json:"result"`
}

type hoverAlternate struct {
	// Contents is the hover's content
	Contents []interface{} `json:"contents"`

	// Range an optional range is a range inside a text document
	// that is used to visualize a hover, e.g. by changing the background color.
	Range lsp.Range `json:"range,omitempty"`
}

type RPCHoverAlternate struct {
	RPCVersion string         `json:"jsonrpc"`
	ID         int            `json:"id"`
	Result     hoverAlternate `json:"result"`
}

func Position(x, y int) lsp.Position {
	return lsp.Position{
		Line:      float64(y),
		Character: float64(x),
	}
}

func (s *Server) DocumentFormat(filename string, options lsp.FormattingOptions) ([]lsp.TextEdit, error) {
	doc := lsp.TextDocumentIdentifier{
		URI: uri.File(filename),
	}

	params := lsp.DocumentFormattingParams{
		Options:      options,
		TextDocument: doc,
	}

	resp, err := s.sendRequest(lsp.MethodTextDocumentFormatting, params)
	if err != nil {
		return nil, err
	}

	var r RPCFormat
	err = json.Unmarshal(resp, &r)
	if err != nil {
		return nil, err
	}

	return r.Result, nil
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
	resp, err := s.sendRequest(lsp.MethodTextDocumentCompletion, params)
	if err != nil {
		return nil, err
	}

	var r RPCCompletion
	err = json.Unmarshal(resp, &r)
	if err == nil {
		return r.Result.Items, nil
	}
	var ra RPCCompletionAlternate
	err = json.Unmarshal(resp, &ra)
	if err != nil {
		return nil, err
	}
	return ra.Result, nil
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

	resp, err := s.sendRequest(lsp.MethodTextDocumentHover, params)
	if err != nil {
		return "", err
	}

	var r RPCHover
	err = json.Unmarshal(resp, &r)
	if err == nil {
		return r.Result.Contents.Value, nil
	}

	var ra RPCHoverAlternate
	err = json.Unmarshal(resp, &ra)
	if err != nil {
		return "", err
	}

	for _, c := range ra.Result.Contents {
		switch t := c.(type) {
		case string:
			return t, nil
		case map[string]string:
			return t["value"], nil
		}
	}
	return "", nil
}
