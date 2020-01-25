package highlight

import (
	"bytes"
	"fmt"
	"regexp"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/rust"
)

var captureNameToScopeName = map[string]string{
	"attribute": "special",
	//"charset": "",
	"comment":          "comment",
	"constant":         "constant",
	"constant.builtin": "constant",
	"constructor":      "type",
	//"delimiter": "",
	//"embedded": "",
	"escape":                  "constant.specialChar",
	"function":                "identifier",
	"function.builtin":        "identifier",
	"function.macro":          "identifier",
	"function.method":         "identifier",
	"function.method.builtin": "identifier",
	"function.special":        "identifier",
	//"import": "",
	//"keyframes": "",
	"keyword": "statement",
	//"label": "",
	//"media": "",
	//"namespace": "",
	"number":   "constant.number",
	"operator": "symbol.operator",
	//"property": "",
	"punctuation.bracket": "symbol.brackets",
	//"punctuation.delimiter": "",
	//"punctuation.special": "",
	"string":                "constant.string",
	"string.special":        "constant.string",
	"string.special.regex":  "constant.string",
	"string.special.symbol": "constant.string",
	//"supports": "",
	"tag":          "symbol.tag",
	"tag.error":    "error",
	"type":         "type",
	"type.builtin": "type",
	//"variable": "",
	"variable.builtin": "type",
	//"variable.parameter": "",
}

type queryPredicate func([]sitter.QueryCapture, []byte) bool

type treeSitterHighlighter struct {
	parser       *sitter.Parser
	query        *sitter.Query
	captureNames []string
	predicates   [][]queryPredicate
}

func NewTreeSitterHighlighter(name string) Highlighter {
	var language *sitter.Language
	var queryString string

	switch name {
	case "go":
		language = golang.GetLanguage()
		queryString = goQuery
	case "rust":
		language = rust.GetLanguage()
		queryString = rustQuery
	default:
		return nil
	}

	// Make sure all required scope names are registered
	for _, scopeName := range captureNameToScopeName {
		if _, ok := Groups[scopeName]; !ok {
			numGroups++
			Groups[scopeName] = numGroups
		}
	}

	highlighter := new(treeSitterHighlighter)

	highlighter.parser = sitter.NewParser()
	highlighter.parser.SetLanguage(language)

	query, err := sitter.NewQuery([]byte(queryString), language)
	if err != nil {
		panic(err)
	}
	highlighter.query = query

	captureCount := query.CaptureCount()
	for i := uint32(0); i < captureCount; i++ {
		captureName := query.CaptureNameForId(i)
		highlighter.captureNames = append(highlighter.captureNames, captureName)
	}

	var stringValues []string
	stringCount := query.StringCount()
	for i := uint32(0); i < stringCount; i++ {
		stringValue := query.StringValueForId(i)
		stringValues = append(stringValues, stringValue)
	}

	patternCount := query.PatternCount()
	for i := uint32(0); i < patternCount; i++ {
		highlighter.predicates = append(highlighter.predicates, []queryPredicate{})

		var predicateSteps []sitter.QueryPredicateStep
		for _, predicateStep := range query.PredicatesForPattern(i) {
			if predicateStep.Type == sitter.QueryPredicateStepTypeDone {
				predicate, err := makePredicate(predicateSteps, stringValues)
				if err != nil {
					panic(err)
				}
				highlighter.predicates[i] = append(highlighter.predicates[i], predicate)
				predicateSteps = nil
			} else {
				predicateSteps = append(predicateSteps, predicateStep)
			}
		}
	}

	return highlighter
}

func makePredicate(steps []sitter.QueryPredicateStep, stringValues []string) (queryPredicate, error) {
	if len(steps) != 3 {
		return nil, fmt.Errorf("invalid number of predicate steps")
	}

	if steps[0].Type != sitter.QueryPredicateStepTypeString {
		return nil, fmt.Errorf("first predicate step must be a string")
	}

	if steps[1].Type != sitter.QueryPredicateStepTypeCapture {
		return nil, fmt.Errorf("second predicate step must be a capture")
	}

	operator := stringValues[steps[0].ValueId]

	// TODO: Implement "eq?", "is-not?" (used by some highlighting queries)
	switch operator {
	case "match?":
		if steps[2].Type != sitter.QueryPredicateStepTypeString {
			return nil, fmt.Errorf("third predicate step must be a string if the operator is 'match?'")
		}
		pattern, err := regexp.Compile(stringValues[steps[2].ValueId])
		if err != nil {
			return nil, err
		}
		predicate := func(captures []sitter.QueryCapture, source []byte) bool {
			for _, capture := range captures {
				if capture.Index == steps[1].ValueId {
					return pattern.MatchString(capture.Node.Content(source))
				}
			}
			return false
		}
		return predicate, nil
	default:
		return nil, fmt.Errorf("unsupported predicate operator: %v", operator)
	}
}

func (h *treeSitterHighlighter) HighlightString(input string) []LineMatch {
	var lineMatches []LineMatch
	return lineMatches
}

func (h *treeSitterHighlighter) HighlightStates(input LineStates) {
}

func (h *treeSitterHighlighter) HighlightMatches(input LineStates, startline, endline int) {
	if startline >= endline {
		return
	}

	// TODO: Use tree-sitter interface for reading buffer
	var buffer bytes.Buffer
	for i := 0; i < input.LinesNum(); i++ {
		buffer.Write(input.LineBytes(i))
		buffer.WriteString("\n")
	}

	source := buffer.Bytes()
	tree := h.parser.Parse(source)

	cursor := sitter.NewQueryCursor()
	cursor.SetPointRange(sitter.Point{uint32(startline), 0}, sitter.Point{uint32(endline), 0})
	cursor.Exec(h.query, tree.RootNode())

	defaultGroup := Groups["default"]

	var lineMatches []LineMatch
	for i := startline; i < endline; i++ {
		lineMatches = append(lineMatches, LineMatch{
			0: defaultGroup,
		})
	}

	var lastNode *sitter.Node
	var furthestScopeGroup Group
	var furthestScopeEnd uint32
	for {
		match, captureIndex, ok := cursor.NextCapture()
		if !ok {
			break
		}

		satisfiesPredicates := true
		for _, predicate := range h.predicates[match.PatternIndex] {
			satisfiesPredicates = satisfiesPredicates && predicate(match.Captures, source)
		}

		capture := match.Captures[captureIndex]

		// Only the first capture that matches a given node is considered
		// (this matches the behavior of the tree-sitter playground)
		if satisfiesPredicates && (lastNode == nil || !capture.Node.Equal(lastNode)) {
			lastNode = capture.Node

			captureName := h.captureNames[capture.Index]

			if scopeName, ok := captureNameToScopeName[captureName]; ok {
				scopeGroup := Groups[scopeName]

				start := capture.Node.StartPoint()
				end := capture.Node.EndPoint()

				startRow := int(start.Row)
				startColumn := int(start.Column)
				if startRow < startline {
					startRow = startline
					startColumn = 0
				}

				endRow := int(end.Row)
				endColumn := int(end.Column)
				if endRow >= endline {
					endRow = endline - 1
					endColumn = len(input.LineBytes(endRow))
				}

				lineMatches[startRow-startline][startColumn] = scopeGroup
				for i := startRow + 1; i <= endRow; i++ {
					lineMatches[i-startline][0] = scopeGroup
				}

				// The following is a simple and fast solution to the problem
				// of nested scopes, although it works only for one level of nesting
				endByte := capture.Node.EndByte()
				resetGroup := defaultGroup

				if endByte < furthestScopeEnd {
					resetGroup = furthestScopeGroup
				} else if endByte > furthestScopeEnd {
					furthestScopeGroup = scopeGroup
					furthestScopeEnd = endByte
				}

				lineMatches[endRow-startline][endColumn] = resetGroup
			}
		}
	}

	for i := startline; i < endline; i++ {
		if i >= input.LinesNum() {
			break
		}
		input.SetMatch(i, lineMatches[i-startline])
	}
}

func (h *treeSitterHighlighter) ReHighlightStates(input LineStates, startline int) {
}

func (h *treeSitterHighlighter) ReHighlightLine(input LineStates, lineN int) {
}
