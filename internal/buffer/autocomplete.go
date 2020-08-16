package buffer

import (
	"bytes"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/zyedidia/micro/v2/internal/lsp"
	"github.com/zyedidia/micro/v2/internal/util"
	"go.lsp.dev/protocol"
)

// A Completer is a function that takes a buffer and returns info
// describing what autocompletions should be inserted at the current
// cursor location
// It returns a list of string suggestions which will be inserted at
// the current cursor location if selected as well as a list of
// suggestion names which can be displayed in an autocomplete box or
// other UI element
type Completer func(*Buffer) []Completion

type Completion struct {
	Edits       []Delta
	Label       string
	CommitChars []rune
	Kind        string
	Filter      string
	Detail      string
	Doc         string
}

// Autocomplete starts the autocomplete process
func (b *Buffer) Autocomplete(c Completer) bool {
	b.Completions = c(b)
	if len(b.Completions) == 0 {
		return false
	}
	b.CurCompletion = -1
	b.CycleAutocomplete(true)
	return true
}

// CycleAutocomplete moves to the next suggestion
func (b *Buffer) CycleAutocomplete(forward bool) {
	prevCompletion := b.CurCompletion

	if forward {
		b.CurCompletion++
	} else {
		b.CurCompletion--
	}
	if b.CurCompletion >= len(b.Completions) {
		b.CurCompletion = 0
	} else if b.CurCompletion < 0 {
		b.CurCompletion = len(b.Completions) - 1
	}

	// undo prev completion
	if prevCompletion != -1 {
		prev := b.Completions[prevCompletion]
		for i := 0; i < len(prev.Edits); i++ {
			if len(prev.Edits[i].Text) != 0 {
				b.UndoOneEvent()
			}
			if !prev.Edits[i].Start.Equal(prev.Edits[i].End) {
				b.UndoOneEvent()
			}
		}
	}

	// apply current completion
	comp := b.Completions[b.CurCompletion]
	b.ApplyDeltas(comp.Edits)
	if len(b.Completions) > 1 {
		b.HasSuggestions = true
	}
}

// GetWord gets the most recent word separated by any separator
// (whitespace, punctuation, any non alphanumeric character)
func GetWord(b *Buffer) ([]byte, int) {
	c := b.GetActiveCursor()
	l := b.LineBytes(c.Y)
	l = util.SliceStart(l, c.X)

	if c.X == 0 || util.IsWhitespace(b.RuneAt(c.Loc)) {
		return []byte{}, -1
	}

	if util.IsNonAlphaNumeric(b.RuneAt(c.Loc)) {
		return []byte{}, c.X
	}

	args := bytes.FieldsFunc(l, util.IsNonAlphaNumeric)
	input := args[len(args)-1]
	return input, c.X - util.CharacterCount(input)
}

// GetArg gets the most recent word (separated by ' ' only)
func GetArg(b *Buffer) (string, int) {
	c := b.GetActiveCursor()
	l := b.LineBytes(c.Y)
	l = util.SliceStart(l, c.X)

	args := bytes.Split(l, []byte{' '})
	input := string(args[len(args)-1])
	argstart := 0
	for i, a := range args {
		if i == len(args)-1 {
			break
		}
		argstart += util.CharacterCount(a) + 1
	}

	return input, argstart
}

// FileComplete autocompletes filenames
func FileComplete(b *Buffer) []Completion {
	c := b.GetActiveCursor()
	input, argstart := GetArg(b)

	sep := string(os.PathSeparator)
	dirs := strings.Split(input, sep)

	var files []os.FileInfo
	var err error
	if len(dirs) > 1 {
		directories := strings.Join(dirs[:len(dirs)-1], sep) + sep

		directories, _ = util.ReplaceHome(directories)
		files, err = ioutil.ReadDir(directories)
	} else {
		files, err = ioutil.ReadDir(".")
	}

	if err != nil {
		return nil
	}

	var suggestions []string
	for _, f := range files {
		name := f.Name()
		if f.IsDir() {
			name += sep
		}
		if strings.HasPrefix(name, dirs[len(dirs)-1]) {
			suggestions = append(suggestions, name)
		}
	}

	sort.Strings(suggestions)
	completions := make([]string, len(suggestions))
	for i := range suggestions {
		var complete string
		if len(dirs) > 1 {
			complete = strings.Join(dirs[:len(dirs)-1], sep) + sep + suggestions[i]
		} else {
			complete = suggestions[i]
		}
		completions[i] = util.SliceEndStr(complete, c.X-argstart)
	}

	return ConvertCompletions(completions, suggestions, c)
}

// BufferComplete autocompletes based on previous words in the buffer
func BufferComplete(b *Buffer) []Completion {
	c := b.GetActiveCursor()
	input, argstart := GetWord(b)

	if argstart == -1 {
		return nil
	}

	inputLen := util.CharacterCount(input)

	suggestionsSet := make(map[string]struct{})

	var suggestions []string
	for i := c.Y; i >= 0; i-- {
		l := b.LineBytes(i)
		words := bytes.FieldsFunc(l, util.IsNonAlphaNumeric)
		for _, w := range words {
			if bytes.HasPrefix(w, input) && util.CharacterCount(w) > inputLen {
				strw := string(w)
				if _, ok := suggestionsSet[strw]; !ok {
					suggestionsSet[strw] = struct{}{}
					suggestions = append(suggestions, strw)
				}
			}
		}
	}
	for i := c.Y + 1; i < b.LinesNum(); i++ {
		l := b.LineBytes(i)
		words := bytes.FieldsFunc(l, util.IsNonAlphaNumeric)
		for _, w := range words {
			if bytes.HasPrefix(w, input) && util.CharacterCount(w) > inputLen {
				strw := string(w)
				if _, ok := suggestionsSet[strw]; !ok {
					suggestionsSet[strw] = struct{}{}
					suggestions = append(suggestions, strw)
				}
			}
		}
	}
	if len(suggestions) > 1 {
		suggestions = append(suggestions, string(input))
	}

	completions := make([]string, len(suggestions))
	for i := range suggestions {
		completions[i] = util.SliceEndStr(suggestions[i], c.X-argstart)
	}

	return ConvertCompletions(completions, suggestions, c)
}

func LSPComplete(b *Buffer) []Completion {
	if !b.HasLSP() {
		return nil
	}

	c := b.GetActiveCursor()
	pos := lsp.Position(c.X, c.Y)
	items, err := b.Server.Completion(b.AbsPath, pos)
	if err != nil {
		return nil
	}

	completions := make([]Completion, len(items))

	for i, item := range items {
		completions[i] = Completion{
			Label:  item.Label,
			Detail: item.Detail,
			Kind:   toKindStr(item.Kind),
			Doc:    getDoc(item.Documentation),
		}

		if item.TextEdit != nil && len(item.TextEdit.NewText) > 0 {
			completions[i].Edits = []Delta{Delta{
				Text:  []byte(item.TextEdit.NewText),
				Start: toLoc(item.TextEdit.Range.Start),
				End:   toLoc(item.TextEdit.Range.End),
			}}
			for _, e := range item.AdditionalTextEdits {
				d := Delta{
					Text:  []byte(e.NewText),
					Start: toLoc(e.Range.Start),
					End:   toLoc(e.Range.End),
				}
				completions[i].Edits = append(completions[i].Edits, d)
			}
		} else {
			var t string
			if len(item.InsertText) > 0 {
				t = item.InsertText
			} else {
				t = item.Label
			}
			_, argstart := GetWord(b)
			str := util.SliceEnd([]byte(t), c.X-argstart)
			completions[i].Edits = []Delta{Delta{
				Text:  str,
				Start: Loc{c.X, c.Y},
				End:   Loc{c.X, c.Y},
			}}
		}
	}

	return completions
}

// ConvertCompletions converts a list of insert text with suggestion labels
// to an array of completion objects ready for autocompletion
func ConvertCompletions(completions, suggestions []string, c *Cursor) []Completion {
	comp := make([]Completion, len(completions))

	for i := 0; i < len(completions); i++ {
		comp[i] = Completion{
			Label: suggestions[i],
		}
		comp[i].Edits = []Delta{Delta{
			Text:  []byte(completions[i]),
			Start: Loc{c.X, c.Y},
			End:   Loc{c.X, c.Y},
		}}
	}
	return comp
}

func toKindStr(k protocol.CompletionItemKind) string {
	s := k.String()
	return strings.ToLower(string(s[0]))
}

// returns documentation from a string | MarkupContent item
func getDoc(documentation interface{}) string {
	var doc string
	switch s := documentation.(type) {
	case string:
		doc = s
	case protocol.MarkupContent:
		doc = s.Value
	}

	return strings.Split(doc, "\n")[0]
}
