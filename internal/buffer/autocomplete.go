package buffer

import (
	"bytes"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/zyedidia/micro/internal/util"
)

// A Completer is a function that takes a buffer and returns info
// describing what autocompletions should be inserted at the current
// cursor location
// It returns a list of string suggestions which will be inserted at
// the current cursor location if selected as well as a list of
// suggestion names which can be displayed in a autocomplete box or
// other UI element
type Completer func(*Buffer) ([]string, []string)

func (b *Buffer) GetSuggestions() {

}

func (b *Buffer) Autocomplete(c Completer) {
	b.Completions, b.Suggestions = c(b)
	if len(b.Completions) != len(b.Suggestions) || len(b.Completions) == 0 {
		return
	}
	b.CurSuggestion = -1
	b.CycleAutocomplete(true)
}

func (b *Buffer) CycleAutocomplete(forward bool) {
	prevSuggestion := b.CurSuggestion

	if forward {
		b.CurSuggestion++
	} else {
		b.CurSuggestion--
	}
	if b.CurSuggestion >= len(b.Suggestions) {
		b.CurSuggestion = 0
	} else if b.CurSuggestion < 0 {
		b.CurSuggestion = len(b.Suggestions) - 1
	}

	c := b.GetActiveCursor()
	start := c.Loc
	end := c.Loc
	if prevSuggestion < len(b.Suggestions) && prevSuggestion >= 0 {
		start = end.Move(-utf8.RuneCountInString(b.Completions[prevSuggestion]), b)
	} else {
		end = start.Move(1, b)
	}

	b.Replace(start, end, b.Completions[b.CurSuggestion])
	b.HasSuggestions = true
}

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
		argstart += utf8.RuneCount(a) + 1
	}

	return input, argstart
}

// FileComplete autocompletes filenames
func FileComplete(b *Buffer) ([]string, []string) {
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
		return nil, nil
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

	return completions, suggestions
}
