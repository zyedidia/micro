package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zyedidia/micro/cmd/micro/highlight"
)

// LargeFileThreshold is the number of bytes when fastdirty is forced
// because hashing is too slow
const LargeFileThreshold = 50000

// The BufType defines what kind of buffer this is
type BufType struct {
	Kind     int
	Readonly bool // The file cannot be edited
	Scratch  bool // The file cannot be saved
}

var (
	btDefault = BufType{0, false, false}
	btHelp    = BufType{1, true, true}
	btLog     = BufType{2, true, true}
	btScratch = BufType{3, false, true}
	btRaw     = BufType{4, true, true}
)

type Buffer struct {
	*LineArray
	*EventHandler

	// Path to the file on disk
	Path string
	// Absolute path to the file on disk
	AbsPath string
	// Name of the buffer on the status line
	name string

	// Whether or not the buffer has been modified since it was opened
	isModified bool

	// Stores the last modification time of the file the buffer is pointing to
	ModTime time.Time

	syntaxDef   *highlight.Def
	highlighter *highlight.Highlighter

	// Hash of the original buffer -- empty if fastdirty is on
	origHash [md5.Size]byte

	// Settings customized by the user
	Settings map[string]interface{}

	// Type of the buffer (e.g. help, raw, scratch etc..)
	Type BufType
}

// NewBufferFromFile opens a new buffer using the given path
// It will also automatically handle `~`, and line/column with filename:l:c
// It will return an empty buffer if the path does not exist
// and an error if the file is a directory
func NewBufferFromFile(path string) (*Buffer, error) {
	var err error
	filename, cursorPosition := GetPathAndCursorPosition(path)
	filename, err = ReplaceHome(filename)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filename)
	fileInfo, _ := os.Stat(filename)

	if err == nil && fileInfo.IsDir() {
		return nil, errors.New(filename + " is a directory")
	}

	defer file.Close()

	var buf *Buffer
	if err != nil {
		// File does not exist -- create an empty buffer with that name
		buf = NewBufferFromString("", filename)
	} else {
		buf = NewBuffer(file, FSize(file), filename, cursorPosition)
	}

	return buf, nil
}

// NewBufferFromString creates a new buffer containing the given string
func NewBufferFromString(text, path string) *Buffer {
	return NewBuffer(strings.NewReader(text), int64(len(text)), path, nil)
}

// NewBuffer creates a new buffer from a given reader with a given path
// Ensure that ReadSettings and InitGlobalSettings have been called before creating
// a new buffer
func NewBuffer(reader io.Reader, size int64, path string, cursorPosition []string) *Buffer {
	b := new(Buffer)

	b.Settings = DefaultLocalSettings()
	for k, v := range globalSettings {
		if _, ok := b.Settings[k]; ok {
			b.Settings[k] = v
		}
	}
	InitLocalSettings(b)

	b.LineArray = NewLineArray(uint64(size), FFAuto, reader)

	absPath, _ := filepath.Abs(path)

	b.Path = path
	b.AbsPath = absPath

	// The last time this file was modified
	b.ModTime, _ = GetModTime(b.Path)

	b.EventHandler = NewEventHandler(b)

	b.UpdateRules()
	log.Println("Filetype detected: ", b.Settings["filetype"])

	if !b.Settings["fastdirty"].(bool) {
		if size > LargeFileThreshold {
			// If the file is larger than LargeFileThreshold fastdirty needs to be on
			b.Settings["fastdirty"] = true
		} else {
			calcHash(b, &b.origHash)
		}
	}

	return b
}

// GetName returns the name that should be displayed in the statusline
// for this buffer
func (b *Buffer) GetName() string {
	if b.name == "" {
		if b.Path == "" {
			return "No name"
		}
		return b.Path
	}
	return b.name
}

// FileType returns the buffer's filetype
func (b *Buffer) FileType() string {
	return b.Settings["filetype"].(string)
}

// ReOpen reloads the current buffer from disk
func (b *Buffer) ReOpen() error {
	data, err := ioutil.ReadFile(b.Path)
	txt := string(data)

	if err != nil {
		return err
	}
	b.EventHandler.ApplyDiff(txt)

	b.ModTime, err = GetModTime(b.Path)
	b.isModified = false
	return err
	// TODO: buffer cursor
	// b.Cursor.Relocate()
}

// Saving

// Save saves the buffer to its default path
func (b *Buffer) Save() error {
	return b.SaveAs(b.Path)
}

// SaveAs saves the buffer to a specified path (filename), creating the file if it does not exist
func (b *Buffer) SaveAs(filename string) error {
	// TODO: rmtrailingws and updaterules
	b.UpdateRules()
	// if b.Settings["rmtrailingws"].(bool) {
	// 	for i, l := range b.lines {
	// 		pos := len(bytes.TrimRightFunc(l.data, unicode.IsSpace))
	//
	// 		if pos < len(l.data) {
	// 			b.deleteToEnd(Loc{pos, i})
	// 		}
	// 	}
	//
	// 	b.Cursor.Relocate()
	// }

	if b.Settings["eofnewline"].(bool) {
		end := b.End()
		if b.RuneAt(Loc{end.X - 1, end.Y}) != '\n' {
			b.Insert(end, "\n")
		}
	}

	// Update the last time this file was updated after saving
	defer func() {
		b.ModTime, _ = GetModTime(filename)
	}()

	// Removes any tilde and replaces with the absolute path to home
	absFilename, _ := ReplaceHome(filename)

	// TODO: save creates parent dirs
	// // Get the leading path to the file | "." is returned if there's no leading path provided
	// if dirname := filepath.Dir(absFilename); dirname != "." {
	// 	// Check if the parent dirs don't exist
	// 	if _, statErr := os.Stat(dirname); os.IsNotExist(statErr) {
	// 		// Prompt to make sure they want to create the dirs that are missing
	// 		if yes, canceled := messenger.YesNoPrompt("Parent folders \"" + dirname + "\" do not exist. Create them? (y,n)"); yes && !canceled {
	// 			// Create all leading dir(s) since they don't exist
	// 			if mkdirallErr := os.MkdirAll(dirname, os.ModePerm); mkdirallErr != nil {
	// 				// If there was an error creating the dirs
	// 				return mkdirallErr
	// 			}
	// 		} else {
	// 			// If they canceled the creation of leading dirs
	// 			return errors.New("Save aborted")
	// 		}
	// 	}
	// }

	var fileSize int

	err := overwriteFile(absFilename, func(file io.Writer) (e error) {
		if len(b.lines) == 0 {
			return
		}

		// end of line
		var eol []byte
		if b.Settings["fileformat"] == "dos" {
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
	})

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
	// TODO: serialize
	// return b.Serialize()
	return nil
}

// overwriteFile opens the given file for writing, truncating if one exists, and then calls
// the supplied function with the file as io.Writer object, also making sure the file is
// closed afterwards.
func overwriteFile(name string, fn func(io.Writer) error) (err error) {
	var file *os.File

	if file, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return
	}

	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = e
		}
	}()

	w := bufio.NewWriter(file)

	if err = fn(w); err != nil {
		return
	}

	err = w.Flush()
	return
}

// SaveWithSudo saves the buffer to the default path with sudo
func (b *Buffer) SaveWithSudo() error {
	return b.SaveAsWithSudo(b.Path)
}

// SaveAsWithSudo is the same as SaveAs except it uses a neat trick
// with tee to use sudo so the user doesn't have to reopen micro with sudo
func (b *Buffer) SaveAsWithSudo(filename string) error {
	b.UpdateRules()
	b.Path = filename
	absPath, _ := filepath.Abs(filename)
	b.AbsPath = absPath

	// Set up everything for the command
	cmd := exec.Command(globalSettings["sucmd"].(string), "tee", filename)
	cmd.Stdin = bytes.NewBuffer(b.Bytes())

	// This is a trap for Ctrl-C so that it doesn't kill micro
	// Instead we trap Ctrl-C to kill the program we're running
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cmd.Process.Kill()
		}
	}()

	// Start the command
	cmd.Start()
	err := cmd.Wait()

	if err == nil {
		b.isModified = false
		b.ModTime, _ = GetModTime(filename)
		// TODO: serialize
	}
	return err
}

func (b *Buffer) GetActiveCursor() *Cursor {
	return nil
}

func (b *Buffer) GetCursor(n int) *Cursor {
	return nil
}

func (b *Buffer) GetCursors() []*Cursor {
	return nil
}

func (b *Buffer) NumCursors() int {
	return 0
}

func (b *Buffer) LineBytes(n int) []byte {
	if n >= len(b.lines) || n < 0 {
		return []byte{}
	}
	return b.lines[n].data
}

func (b *Buffer) LinesNum() int {
	return len(b.lines)
}

func (b *Buffer) Start() Loc {
	return Loc{0, 0}
}

// End returns the location of the last character in the buffer
func (b *Buffer) End() Loc {
	numlines := len(b.lines)
	return Loc{utf8.RuneCount(b.lines[numlines-1].data), numlines - 1}
}

// RuneAt returns the rune at a given location in the buffer
func (b *Buffer) RuneAt(loc Loc) rune {
	line := b.LineBytes(loc.Y)
	if len(line) > 0 {
		i := 0
		for len(line) > 0 {
			r, size := utf8.DecodeRune(line)
			line = line[size:]
			i++

			if i == loc.X {
				return r
			}
		}
	}
	return '\n'
}

// Modified returns if this buffer has been modified since
// being opened
func (b *Buffer) Modified() bool {
	if b.Settings["fastdirty"].(bool) {
		return b.isModified
	}

	var buff [md5.Size]byte

	calcHash(b, &buff)
	return buff != b.origHash
}

// calcHash calculates md5 hash of all lines in the buffer
func calcHash(b *Buffer, out *[md5.Size]byte) {
	h := md5.New()

	if len(b.lines) > 0 {
		h.Write(b.lines[0].data)

		for _, l := range b.lines[1:] {
			h.Write([]byte{'\n'})
			h.Write(l.data)
		}
	}

	h.Sum((*out)[:0])
}

// UpdateRules updates the syntax rules and filetype for this buffer
// This is called when the colorscheme changes
func (b *Buffer) UpdateRules() {
	rehighlight := false
	var files []*highlight.File
	for _, f := range ListRuntimeFiles(RTSyntax) {
		data, err := f.Data()
		if err != nil {
			TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
		} else {
			file, err := highlight.ParseFile(data)
			if err != nil {
				TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
				continue
			}
			ftdetect, err := highlight.ParseFtDetect(file)
			if err != nil {
				TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
				continue
			}

			ft := b.Settings["filetype"].(string)
			if (ft == "Unknown" || ft == "") && !rehighlight {
				if highlight.MatchFiletype(ftdetect, b.Path, b.lines[0].data) {
					header := new(highlight.Header)
					header.FileType = file.FileType
					header.FtDetect = ftdetect
					b.syntaxDef, err = highlight.ParseDef(file, header)
					if err != nil {
						TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
						continue
					}
					rehighlight = true
				}
			} else {
				if file.FileType == ft && !rehighlight {
					header := new(highlight.Header)
					header.FileType = file.FileType
					header.FtDetect = ftdetect
					b.syntaxDef, err = highlight.ParseDef(file, header)
					if err != nil {
						TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
						continue
					}
					rehighlight = true
				}
			}
			files = append(files, file)
		}
	}

	if b.syntaxDef != nil {
		highlight.ResolveIncludes(b.syntaxDef, files)
	}

	if b.highlighter == nil || rehighlight {
		if b.syntaxDef != nil {
			b.Settings["filetype"] = b.syntaxDef.FileType
			b.highlighter = highlight.NewHighlighter(b.syntaxDef)
			if b.Settings["syntax"].(bool) {
				b.highlighter.HighlightStates(b)
			}
		}
	}
}
