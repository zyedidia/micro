package buffer

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"unicode"
	"unicode/utf8"

	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/micro/internal/screen"
	"github.com/zyedidia/micro/internal/util"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"
)

// LargeFileThreshold is the number of bytes when fastdirty is forced
// because hashing is too slow
const LargeFileThreshold = 50000

// overwriteFile opens the given file for writing, truncating if one exists, and then calls
// the supplied function with the file as io.Writer object, also making sure the file is
// closed afterwards.
func overwriteFile(name string, enc encoding.Encoding, fn func(io.Writer) error) (err error) {
	var file *os.File

	if file, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return
	}

	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = e
		}
	}()

	w := transform.NewWriter(file, enc.NewEncoder())
	// w := bufio.NewWriter(file)

	if err = fn(w); err != nil {
		return
	}

	// err = w.Flush()
	return
}

// overwriteFileAsRoot executes dd as root and then calls the supplied function
// with dd's standard input as an io.Writer object. Dd opens the given file for writing,
// truncating it if it exists, and writes what it receives on its standard input to the file.
func overwriteFileAsRoot(name string, enc encoding.Encoding, fn func(io.Writer) error) (err error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		return errors.New("Save with sudo not supported on Windows")
	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command(config.GlobalSettings["sucmd"].(string), "dd", "bs=4k", "of="+name)
	} else {
		cmd = exec.Command(config.GlobalSettings["sucmd"].(string), "dd", "status=none", "bs=4K", "of="+name)
	}
	var stdin io.WriteCloser

	screenb := screen.TempFini()
	defer screen.TempStart(screenb)

	// This is a trap for Ctrl-C so that it doesn't kill micro
	// Instead we trap Ctrl-C to kill the program we're running
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cmd.Process.Kill()
		}
	}()

	if stdin, err = cmd.StdinPipe(); err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		return
	}

	e := fn(stdin)

	if err = stdin.Close(); err != nil {
		return
	}

	if err = cmd.Wait(); err != nil {
		return
	}

	return e
}

// Save saves the buffer to its default path
func (b *Buffer) Save() error {
	return b.SaveAs(b.Path)
}

// SaveAs saves the buffer to a specified path (filename), creating the file if it does not exist
func (b *Buffer) SaveAs(filename string) error {
	return b.saveToFile(filename, false)
}

func (b *Buffer) SaveWithSudo() error {
	return b.SaveAsWithSudo(b.Path)
}

func (b *Buffer) SaveAsWithSudo(filename string) error {
	return b.saveToFile(filename, true)
}

func (b *Buffer) saveToFile(filename string, withSudo bool) error {
	var err error
	if b.Type.Readonly {
		return errors.New("Cannot save readonly buffer")
	}
	if b.Type.Scratch {
		return errors.New("Cannot save scratch buffer")
	}

	b.UpdateRules()
	if b.Settings["rmtrailingws"].(bool) {
		for i, l := range b.lines {
			leftover := utf8.RuneCount(bytes.TrimRightFunc(l.data, unicode.IsSpace))

			linelen := utf8.RuneCount(l.data)
			b.Remove(Loc{leftover, i}, Loc{linelen, i})
		}

		b.RelocateCursors()
	}

	if b.Settings["eofnewline"].(bool) {
		end := b.End()
		if b.RuneAt(Loc{end.X - 1, end.Y}) != '\n' {
			b.Insert(end, "\n")
		}
	}

	// Update the last time this file was updated after saving
	defer func() {
		b.ModTime, _ = util.GetModTime(filename)
		err = b.Serialize()
	}()

	// Removes any tilde and replaces with the absolute path to home
	absFilename, _ := util.ReplaceHome(filename)

	// Get the leading path to the file | "." is returned if there's no leading path provided
	if dirname := filepath.Dir(absFilename); dirname != "." {
		// Check if the parent dirs don't exist
		if _, statErr := os.Stat(dirname); os.IsNotExist(statErr) {
			// Prompt to make sure they want to create the dirs that are missing
			if b.Settings["mkparents"].(bool) {
				// Create all leading dir(s) since they don't exist
				if mkdirallErr := os.MkdirAll(dirname, os.ModePerm); mkdirallErr != nil {
					// If there was an error creating the dirs
					return mkdirallErr
				}
			} else {
				return errors.New("Parent dirs don't exist, enable 'mkparents' for auto creation")
			}
		}
	}

	var fileSize int

	enc, err := htmlindex.Get(b.Settings["encoding"].(string))
	if err != nil {
		return err
	}

	fwriter := func(file io.Writer) (e error) {
		if len(b.lines) == 0 {
			return
		}

		// end of line
		var eol []byte
		if b.Endings == FFDos {
			eol = []byte{'\r', '\n'}
		} else {
			eol = []byte{'\n'}
		}

		// write lines
		if fileSize, e = file.Write(b.lines[0].data); e != nil {
			return
		}

		for _, l := range b.lines[1:] {
			if _, e = file.Write(eol); e != nil {
				return
			}
			if _, e = file.Write(l.data); e != nil {
				return
			}
			fileSize += len(eol) + len(l.data)
		}
		return
	}

	if withSudo {
		err = overwriteFileAsRoot(absFilename, enc, fwriter)
	} else {
		err = overwriteFile(absFilename, enc, fwriter)
	}

	if err != nil {
		return err
	}

	if !b.Settings["fastdirty"].(bool) {
		if fileSize > LargeFileThreshold {
			// For large files 'fastdirty' needs to be on
			b.Settings["fastdirty"] = true
		} else {
			calcHash(b, &b.origHash)
		}
	}

	b.Path = filename
	absPath, _ := filepath.Abs(filename)
	b.AbsPath = absPath
	b.isModified = false
	return err
}
