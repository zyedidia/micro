package buffer

import (
	"crypto/md5"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zyedidia/micro/cmd/micro/config"
	"github.com/zyedidia/micro/cmd/micro/highlight"
	"github.com/zyedidia/micro/cmd/micro/screen"
	. "github.com/zyedidia/micro/cmd/micro/util"
)

var OpenBuffers []*Buffer

// The BufType defines what kind of buffer this is
type BufType struct {
	Kind     int
	Readonly bool // The file cannot be edited
	Scratch  bool // The file cannot be saved
	Syntax   bool // Syntax highlighting is enabled
}

var (
	BTDefault = BufType{0, false, false, true}
	BTHelp    = BufType{1, true, true, true}
	BTLog     = BufType{2, true, true, false}
	BTScratch = BufType{3, false, true, false}
	BTRaw     = BufType{4, true, true, false}
	BTInfo    = BufType{5, false, true, false}

	ErrFileTooLarge = errors.New("File is too large to hash")
)

// Buffer stores the main information about a currently open file including
// the actual text (in a LineArray), the undo/redo stack (in an EventHandler)
// all the cursors, the syntax highlighting info, the settings for the buffer
// and some misc info about modification time and path location.
// The syntax highlighting info must be stored with the buffer because the syntax
// highlighter attaches information to each line of the buffer for optimization
// purposes so it doesn't have to rehighlight everything on every update.
type Buffer struct {
	*LineArray
	*EventHandler

	cursors     []*Cursor
	curCursor   int
	StartCursor Loc

	// Path to the file on disk
	Path string
	// Absolute path to the file on disk
	AbsPath string
	// Name of the buffer on the status line
	name string

	// Stores the last modification time of the file the buffer is pointing to
	ModTime *time.Time

	SyntaxDef   *highlight.Def
	Highlighter *highlight.Highlighter

	// Hash of the original buffer -- empty if fastdirty is on
	origHash [md5.Size]byte

	// Settings customized by the user
	Settings map[string]interface{}

	// Type of the buffer (e.g. help, raw, scratch etc..)
	Type BufType

	Messages []*Message
}

// NewBufferFromFile opens a new buffer using the given path
// It will also automatically handle `~`, and line/column with filename:l:c
// It will return an empty buffer if the path does not exist
// and an error if the file is a directory
func NewBufferFromFile(path string, btype BufType) (*Buffer, error) {
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
		buf = NewBufferFromString("", filename, btype)
	} else {
		buf = NewBuffer(file, FSize(file), filename, cursorPosition, btype)
	}

	return buf, nil
}

// NewBufferFromString creates a new buffer containing the given string
func NewBufferFromString(text, path string, btype BufType) *Buffer {
	return NewBuffer(strings.NewReader(text), int64(len(text)), path, nil, btype)
}

// NewBuffer creates a new buffer from a given reader with a given path
// Ensure that ReadSettings and InitGlobalSettings have been called before creating
// a new buffer
func NewBuffer(reader io.Reader, size int64, path string, cursorPosition []string, btype BufType) *Buffer {
	absPath, _ := filepath.Abs(path)

	b := new(Buffer)
	b.Type = btype

	b.Settings = config.DefaultLocalSettings()
	for k, v := range config.GlobalSettings {
		if _, ok := b.Settings[k]; ok {
			b.Settings[k] = v
		}
	}
	config.InitLocalSettings(b.Settings, b.Path)

	found := false
	for _, buf := range OpenBuffers {
		if buf.AbsPath == absPath {
			found = true
			b.LineArray = buf.LineArray
			b.EventHandler = buf.EventHandler
			b.ModTime = buf.ModTime
			b.isModified = buf.isModified
		}
	}

	if !found {
		b.LineArray = NewLineArray(uint64(size), FFAuto, reader)
		b.EventHandler = NewEventHandler(b.LineArray, b.cursors)
		b.ModTime = new(time.Time)
		b.isModified = new(bool)
		*b.isModified = false
		*b.ModTime = time.Time{}
	}

	b.Path = path
	b.AbsPath = absPath

	// The last time this file was modified
	*b.ModTime, _ = GetModTime(b.Path)

	b.UpdateRules()

	if _, err := os.Stat(config.ConfigDir + "/buffers/"); os.IsNotExist(err) {
		os.Mkdir(config.ConfigDir+"/buffers/", os.ModePerm)
	}

	// cursorLocation, err := GetBufferCursorLocation(cursorPosition, b)
	// b.startcursor = Cursor{
	// 	Loc: cursorLocation,
	// 	buf: b,
	// }
	// TODO flagstartpos
	if b.Settings["savecursor"].(bool) || b.Settings["saveundo"].(bool) {
		err := b.Unserialize()
		if err != nil {
			screen.TermMessage(err)
		}
	}

	b.AddCursor(NewCursor(b, b.StartCursor))

	if !b.Settings["fastdirty"].(bool) {
		if size > LargeFileThreshold {
			// If the file is larger than LargeFileThreshold fastdirty needs to be on
			b.Settings["fastdirty"] = true
		} else {
			calcHash(b, &b.origHash)
		}
	}

	OpenBuffers = append(OpenBuffers, b)

	return b
}

// Close removes this buffer from the list of open buffers
func (b *Buffer) Close() {
	for i, buf := range OpenBuffers {
		if b == buf {
			copy(OpenBuffers[i:], OpenBuffers[i+1:])
			OpenBuffers[len(OpenBuffers)-1] = nil
			OpenBuffers = OpenBuffers[:len(OpenBuffers)-1]
			return
		}
	}
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

	*b.ModTime, err = GetModTime(b.Path)
	*b.isModified = false
	for _, c := range b.cursors {
		c.Relocate()
	}
	return err
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
		return *b.isModified
	}

	var buff [md5.Size]byte

	calcHash(b, &buff)
	return buff != b.origHash
}

// calcHash calculates md5 hash of all lines in the buffer
func calcHash(b *Buffer, out *[md5.Size]byte) error {
	h := md5.New()

	size := 0
	if len(b.lines) > 0 {
		n, e := h.Write(b.lines[0].data)
		if e != nil {
			return e
		}
		size += n

		for _, l := range b.lines[1:] {
			n, e = h.Write([]byte{'\n'})
			if e != nil {
				return e
			}
			size += n
			n, e = h.Write(l.data)
			if e != nil {
				return e
			}
			size += n
		}
	}

	if size > LargeFileThreshold {
		return ErrFileTooLarge
	}

	h.Sum((*out)[:0])
	return nil
}

// UpdateRules updates the syntax rules and filetype for this buffer
// This is called when the colorscheme changes
func (b *Buffer) UpdateRules() {
	if !b.Type.Syntax {
		return
	}
	rehighlight := false
	var files []*highlight.File
	for _, f := range config.ListRuntimeFiles(config.RTSyntax) {
		data, err := f.Data()
		if err != nil {
			screen.TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
		} else {
			file, err := highlight.ParseFile(data)
			if err != nil {
				screen.TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
				continue
			}
			ftdetect, err := highlight.ParseFtDetect(file)
			if err != nil {
				screen.TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
				continue
			}

			ft := b.Settings["filetype"].(string)
			if (ft == "Unknown" || ft == "") && !rehighlight {
				if highlight.MatchFiletype(ftdetect, b.Path, b.lines[0].data) {
					header := new(highlight.Header)
					header.FileType = file.FileType
					header.FtDetect = ftdetect
					b.SyntaxDef, err = highlight.ParseDef(file, header)
					if err != nil {
						screen.TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
						continue
					}
					rehighlight = true
				}
			} else {
				if file.FileType == ft && !rehighlight {
					header := new(highlight.Header)
					header.FileType = file.FileType
					header.FtDetect = ftdetect
					b.SyntaxDef, err = highlight.ParseDef(file, header)
					if err != nil {
						screen.TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
						continue
					}
					rehighlight = true
				}
			}
			files = append(files, file)
		}
	}

	if b.SyntaxDef != nil {
		highlight.ResolveIncludes(b.SyntaxDef, files)
	}

	if b.Highlighter == nil || rehighlight {
		if b.SyntaxDef != nil {
			b.Settings["filetype"] = b.SyntaxDef.FileType
			b.Highlighter = highlight.NewHighlighter(b.SyntaxDef)
			if b.Settings["syntax"].(bool) {
				b.Highlighter.HighlightStates(b)
			}
		}
	}
}

// ClearMatches clears all of the syntax highlighting for the buffer
func (b *Buffer) ClearMatches() {
	for i := range b.lines {
		b.SetMatch(i, nil)
		b.SetState(i, nil)
	}
}

// IndentString returns this buffer's indent method (a tabstop or n spaces
// depending on the settings)
func (b *Buffer) IndentString(tabsize int) string {
	if b.Settings["tabstospaces"].(bool) {
		return Spaces(tabsize)
	}
	return "\t"
}

// SetCursors resets this buffer's cursors to a new list
func (b *Buffer) SetCursors(c []*Cursor) {
	b.cursors = c
}

// AddCursor adds a new cursor to the list
func (b *Buffer) AddCursor(c *Cursor) {
	b.cursors = append(b.cursors, c)
	b.EventHandler.cursors = b.cursors
	b.UpdateCursors()
}

// SetCurCursor sets the current cursor
func (b *Buffer) SetCurCursor(n int) {
	b.curCursor = n
}

// GetActiveCursor returns the main cursor in this buffer
func (b *Buffer) GetActiveCursor() *Cursor {
	return b.cursors[b.curCursor]
}

// GetCursor returns the nth cursor
func (b *Buffer) GetCursor(n int) *Cursor {
	return b.cursors[n]
}

// GetCursors returns the list of cursors in this buffer
func (b *Buffer) GetCursors() []*Cursor {
	return b.cursors
}

// NumCursors returns the number of cursors
func (b *Buffer) NumCursors() int {
	return len(b.cursors)
}

// MergeCursors merges any cursors that are at the same position
// into one cursor
func (b *Buffer) MergeCursors() {
	var cursors []*Cursor
	for i := 0; i < len(b.cursors); i++ {
		c1 := b.cursors[i]
		if c1 != nil {
			for j := 0; j < len(b.cursors); j++ {
				c2 := b.cursors[j]
				if c2 != nil && i != j && c1.Loc == c2.Loc {
					b.cursors[j] = nil
				}
			}
			cursors = append(cursors, c1)
		}
	}

	b.cursors = cursors

	for i := range b.cursors {
		b.cursors[i].Num = i
	}

	if b.curCursor >= len(b.cursors) {
		b.curCursor = len(b.cursors) - 1
	}
}

// UpdateCursors updates all the cursors indicies
func (b *Buffer) UpdateCursors() {
	for i, c := range b.cursors {
		c.Num = i
	}
}

func (b *Buffer) RemoveCursor(i int) {
	copy(b.cursors[i:], b.cursors[i+1:])
	b.cursors[len(b.cursors)-1] = nil
	b.cursors = b.cursors[:len(b.cursors)-1]
}

// ClearCursors removes all extra cursors
func (b *Buffer) ClearCursors() {
	for i := 1; i < len(b.cursors); i++ {
		b.cursors[i] = nil
	}
	b.cursors = b.cursors[:1]
	b.UpdateCursors()
	b.curCursor = 0
	b.GetActiveCursor().ResetSelection()
}
