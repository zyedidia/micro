package buffer

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"unicode"

	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/micro/v2/internal/util"
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
func overwriteFile(name string, enc encoding.Encoding, fn func(io.Writer) error, withSudo bool) (err error) {
	var writeCloser io.WriteCloser

	if withSudo {
		cmd := exec.Command(config.GlobalSettings["sucmd"].(string), "dd", "bs=4k", "of="+name)

		if writeCloser, err = cmd.StdinPipe(); err != nil {
			return
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			cmd.Process.Kill()
		}()

		defer func() {
			screenb := screen.TempFini()
			if e := cmd.Run(); e != nil && err == nil {
				err = e
			}
			screen.TempStart(screenb)
		}()
	} else if writeCloser, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return
	}

	w := bufio.NewWriter(transform.NewWriter(writeCloser, enc.NewEncoder()))
	err = fn(w)
	w.Flush()

	if e := writeCloser.Close(); e != nil && err == nil {
		err = e
	}

	return
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
	if withSudo && runtime.GOOS == "windows" {
		return errors.New("Save with sudo not supported on Windows")
	}

	if b.Settings["rmtrailingws"].(bool) {
		for i, l := range b.lines {
			leftover := util.CharacterCount(bytes.TrimRightFunc(l.data, unicode.IsSpace))

			linelen := util.CharacterCount(l.data)
			b.Remove(Loc{leftover, i}, Loc{linelen, i})
		}

		b.RelocateCursors()
	}

	if b.Settings["eofnewline"].(bool) {
		end := b.End()
		if b.RuneAt(Loc{end.X, end.Y}) != '\n' {
			b.insert(end, []byte{'\n'})
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

	if err = overwriteFile(absFilename, enc, fwriter, withSudo); err != nil {
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
	b.UpdateRules()
	return err
}
