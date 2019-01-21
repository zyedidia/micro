package buffer

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/zyedidia/micro/cmd/micro/util"
)

type Completer func(*Buffer) (string, []string)

func (b *Buffer) GetSuggestions() {

}

func (b *Buffer) Autocomplete(c Completer) {

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
func FileComplete(b *Buffer) (string, []string) {
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

	var suggestions []string
	if err != nil {
		return "", suggestions
	}
	for _, f := range files {
		name := f.Name()
		if f.IsDir() {
			name += sep
		}
		if strings.HasPrefix(name, dirs[len(dirs)-1]) {
			suggestions = append(suggestions, name)
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		if len(dirs) > 1 {
			chosen = strings.Join(dirs[:len(dirs)-1], sep) + sep + suggestions[0]
		} else {
			chosen = suggestions[0]
		}
	} else {
		if len(dirs) > 1 {
			chosen = strings.Join(dirs[:len(dirs)-1], sep) + sep
		}
	}
	chosen = util.SliceEndStr(chosen, c.X-argstart)

	return chosen, suggestions
}
