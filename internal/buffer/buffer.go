package buffer

import (
	"bytes"
	"crypto/md5"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	luar "layeh.com/gopher-luar"

	"github.com/zyedidia/micro/internal/config"
	ulua "github.com/zyedidia/micro/internal/lua"
	"github.com/zyedidia/micro/internal/screen"
	"github.com/zyedidia/micro/internal/util"
	"github.com/zyedidia/micro/pkg/highlight"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const backupTime = 8000

var (
	// OpenBuffers is a list of the currently open buffers
	OpenBuffers []*Buffer
	// LogBuf is a reference to the log buffer which can be opened with the
	// `> log` command
	LogBuf *Buffer
)

// The BufType defines what kind of buffer this is
type BufType struct {
	Kind     int
	Readonly bool // The buffer cannot be edited
	Scratch  bool // The buffer cannot be saved
	Syntax   bool // Syntax highlighting is enabled
}

var (
	// BTDefault is a default buffer
	BTDefault = BufType{0, false, false, true}
	// BTHelp is a help buffer
	BTHelp = BufType{1, true, true, true}
	// BTLog is a log buffer
	BTLog = BufType{2, true, true, false}
	// BTScratch is a buffer that cannot be saved (for scratch work)
	BTScratch = BufType{3, false, true, false}
	// BTRaw is is a buffer that shows raw terminal events
	BTRaw = BufType{4, false, true, false}
	// BTInfo is a buffer for inputting information
	BTInfo = BufType{5, false, true, false}

	// ErrFileTooLarge is returned when the file is too large to hash
	// (fastdirty is automatically enabled)
	ErrFileTooLarge = errors.New("File is too large to hash")
)

// SharedBuffer is a struct containing info that is shared among buffers
// that have the same file open
type SharedBuffer struct {
	*LineArray
	// Stores the last modification time of the file the buffer is pointing to
	ModTime time.Time
	// Type of the buffer (e.g. help, raw, scratch etc..)
	Type BufType

	isModified bool
	// Whether or not suggestions can be autocompleted must be shared because
	// it changes based on how the buffer has changed
	HasSuggestions bool
}

func (b *SharedBuffer) insert(pos Loc, value []byte) {
	b.isModified = true
	b.HasSuggestions = false
	b.LineArray.insert(pos, value)
}
func (b *SharedBuffer) remove(start, end Loc) []byte {
	b.isModified = true
	b.HasSuggestions = false
	return b.LineArray.remove(start, end)
}

// Buffer stores the main information about a currently open file including
// the actual text (in a LineArray), the undo/redo stack (in an EventHandler)
// all the cursors, the syntax highlighting info, the settings for the buffer
// and some misc info about modification time and path location.
// The syntax highlighting info must be stored with the buffer because the syntax
// highlighter attaches information to each line of the buffer for optimization
// purposes so it doesn't have to rehighlight everything on every update.
type Buffer struct {
	*EventHandler
	*SharedBuffer

	cursors     []*Cursor
	curCursor   int
	StartCursor Loc

	// Path to the file on disk
	Path string
	// Absolute path to the file on disk
	AbsPath string
	// Name of the buffer on the status line
	name string

	// SyntaxDef represents the syntax highlighting definition being used
	// This stores the highlighting rules and filetype detection info
	SyntaxDef *highlight.Def
	// The Highlighter struct actually performs the highlighting
	Highlighter highlight.Highlighter
	// Modifications is the list of modified regions for syntax highlighting
	Modifications []Loc

	// Hash of the original buffer -- empty if fastdirty is on
	origHash [md5.Size]byte

	// Settings customized by the user
	Settings map[string]interface{}

	Suggestions   []string
	Completions   []string
	CurSuggestion int

	Messages []*Message

	// counts the number of edits
	// resets every backupTime edits
	lastbackup time.Time
}

// NewBufferFromFile opens a new buffer using the given path
// It will also automatically handle `~`, and line/column with filename:l:c
// It will return an empty buffer if the path does not exist
// and an error if the file is a directory
func NewBufferFromFile(path string, btype BufType) (*Buffer, error) {
	var err error
	filename, cursorPos := util.GetPathAndCursorPosition(path)
	filename, err = util.ReplaceHome(filename)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filename)
	fileInfo, _ := os.Stat(filename)

	if err == nil && fileInfo.IsDir() {
		return nil, errors.New("Error: " + filename + " is a directory and cannot be opened")
	}

	defer file.Close()

	cursorLoc, cursorerr := ParseCursorLocation(cursorPos)
	if cursorerr != nil {
		cursorLoc = Loc{-1, -1}
	}

	var buf *Buffer
	if err != nil {
		// File does not exist -- create an empty buffer with that name
		buf = NewBufferFromString("", filename, btype)
	} else {
		buf = NewBuffer(file, util.FSize(file), filename, cursorLoc, btype)
	}

	return buf, nil
}

// NewBufferFromString creates a new buffer containing the given string
func NewBufferFromString(text, path string, btype BufType) *Buffer {
	return NewBuffer(strings.NewReader(text), int64(len(text)), path, Loc{-1, -1}, btype)
}

// NewBuffer creates a new buffer from a given reader with a given path
// Ensure that ReadSettings and InitGlobalSettings have been called before creating
// a new buffer
// Places the cursor at startcursor. If startcursor is -1, -1 places the
// cursor at an autodetected location (based on savecursor or :LINE:COL)
func NewBuffer(r io.Reader, size int64, path string, startcursor Loc, btype BufType) *Buffer {
	absPath, _ := filepath.Abs(path)

	b := new(Buffer)

	b.Settings = config.DefaultCommonSettings()
	for k, v := range config.GlobalSettings {
		if _, ok := b.Settings[k]; ok {
			b.Settings[k] = v
		}
	}
	config.InitLocalSettings(b.Settings, path)

	enc, err := htmlindex.Get(b.Settings["encoding"].(string))
	if err != nil {
		enc = unicode.UTF8
		b.Settings["encoding"] = "utf-8"
	}

	reader := transform.NewReader(r, enc.NewDecoder())

	found := false
	if len(path) > 0 {
		for _, buf := range OpenBuffers {
			if buf.AbsPath == absPath && buf.Type != BTInfo {
				found = true
				b.SharedBuffer = buf.SharedBuffer
				b.EventHandler = buf.EventHandler
			}
		}
	}

	b.Path = path
	b.AbsPath = absPath

	if !found {
		b.SharedBuffer = new(SharedBuffer)
		b.Type = btype

		hasBackup := b.ApplyBackup(size)

		if !hasBackup {
			b.LineArray = NewLineArray(uint64(size), FFAuto, reader)
		}
		b.EventHandler = NewEventHandler(b.SharedBuffer, b.cursors)
	}

	if b.Settings["readonly"].(bool) && b.Type == BTDefault {
		b.Type.Readonly = true
	}

	// The last time this file was modified
	b.UpdateModTime()

	switch b.Endings {
	case FFUnix:
		b.Settings["fileformat"] = "unix"
	case FFDos:
		b.Settings["fileformat"] = "dos"
	}

	b.UpdateRules()
	config.InitLocalSettings(b.Settings, b.Path)

	if _, err := os.Stat(config.ConfigDir + "/buffers/"); os.IsNotExist(err) {
		os.Mkdir(config.ConfigDir+"/buffers/", os.ModePerm)
	}

	if startcursor.X != -1 && startcursor.Y != -1 {
		b.StartCursor = startcursor
	} else {
		if b.Settings["savecursor"].(bool) || b.Settings["saveundo"].(bool) {
			err := b.Unserialize()
			if err != nil {
				screen.TermMessage(err)
			}
		}
	}

	b.AddCursor(NewCursor(b, b.StartCursor))
	b.GetActiveCursor().Relocate()

	if !b.Settings["fastdirty"].(bool) {
		if size > LargeFileThreshold {
			// If the file is larger than LargeFileThreshold fastdirty needs to be on
			b.Settings["fastdirty"] = true
		} else {
			calcHash(b, &b.origHash)
		}
	}

	err = config.RunPluginFn("onBufferOpen", luar.New(ulua.L, b))
	if err != nil {
		screen.TermMessage(err)
	}

	b.Modifications = make([]Loc, 10)

	OpenBuffers = append(OpenBuffers, b)

	return b
}

// Close removes this buffer from the list of open buffers
func (b *Buffer) Close() {
	for i, buf := range OpenBuffers {
		if b == buf {
			b.Fini()
			copy(OpenBuffers[i:], OpenBuffers[i+1:])
			OpenBuffers[len(OpenBuffers)-1] = nil
			OpenBuffers = OpenBuffers[:len(OpenBuffers)-1]
			return
		}
	}
}

// Fini should be called when a buffer is closed and performs
// some cleanup
func (b *Buffer) Fini() {
	if !b.Modified() {
		b.Serialize()
	}
	b.RemoveBackup()
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

//SetName changes the name for this buffer
func (b *Buffer) SetName(s string) {
	b.name = s
}

// Insert inserts the given string of text at the start location
func (b *Buffer) Insert(start Loc, text string) {
	if !b.Type.Readonly {
		b.EventHandler.cursors = b.cursors
		b.EventHandler.active = b.curCursor
		b.EventHandler.Insert(start, text)

		// b.Modifications is cleared every screen redraw so it's
		// ok to append duplicates
		b.Modifications = append(b.Modifications, Loc{start.Y, start.Y + strings.Count(text, "\n")})

		go b.Backup(true)
	}
}

// Remove removes the characters between the start and end locations
func (b *Buffer) Remove(start, end Loc) {
	if !b.Type.Readonly {
		b.EventHandler.cursors = b.cursors
		b.EventHandler.active = b.curCursor
		b.EventHandler.Remove(start, end)

		b.Modifications = append(b.Modifications, Loc{start.Y, start.Y})

		go b.Backup(true)
	}
}

// ClearModifications clears the list of modified lines in this buffer
// The list of modified lines is used for syntax highlighting so that
// we can selectively highlight only the necessary lines
// This function should be called every time this buffer is drawn to
// the screen
func (b *Buffer) ClearModifications() {
	// clear slice without resetting the cap
	b.Modifications = b.Modifications[:0]
}

// FileType returns the buffer's filetype
func (b *Buffer) FileType() string {
	return b.Settings["filetype"].(string)
}

// ExternallyModified returns whether the file being edited has
// been modified by some external process
func (b *Buffer) ExternallyModified() bool {
	modTime, err := util.GetModTime(b.Path)
	if err == nil {
		return modTime != b.ModTime
	}
	return false
}

// UpdateModTime updates the modtime of this file
func (b *Buffer) UpdateModTime() (err error) {
	b.ModTime, err = util.GetModTime(b.Path)
	return
}

// ReOpen reloads the current buffer from disk
func (b *Buffer) ReOpen() error {
	file, err := os.Open(b.Path)
	if err != nil {
		return err
	}

	enc, err := htmlindex.Get(b.Settings["encoding"].(string))
	if err != nil {
		return err
	}

	reader := transform.NewReader(file, enc.NewDecoder())
	data, err := ioutil.ReadAll(reader)
	txt := string(data)

	if err != nil {
		return err
	}
	b.EventHandler.ApplyDiff(txt)

	err = b.UpdateModTime()
	b.isModified = false
	b.RelocateCursors()
	return err
}

// RelocateCursors relocates all cursors (makes sure they are in the buffer)
func (b *Buffer) RelocateCursors() {
	for _, c := range b.cursors {
		c.Relocate()
	}
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
	if b.Type.Scratch {
		return false
	}

	if b.Settings["fastdirty"].(bool) {
		return b.isModified
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
	ft := b.Settings["filetype"].(string)
	if ft == "off" {
		return
	}
	syntaxFile := ""
	var header *highlight.Header
	for _, f := range config.ListRuntimeFiles(config.RTSyntaxHeader) {
		data, err := f.Data()
		if err != nil {
			screen.TermMessage("Error loading syntax header file " + f.Name() + ": " + err.Error())
			continue
		}

		header, err = highlight.MakeHeader(data)
		if err != nil {
			screen.TermMessage("Error reading syntax header file", f.Name(), err)
			continue
		}

		if ft == "unknown" || ft == "" {
			if highlight.MatchFiletype(header.FtDetect, b.Path, b.lines[0].data) {
				syntaxFile = f.Name()
				break
			}
		} else if header.FileType == ft {
			syntaxFile = f.Name()
			break
		}
	}

	if syntaxFile == "" {
		// search for the syntax file in the user's custom syntax files
		for _, f := range config.ListRealRuntimeFiles(config.RTSyntax) {
			data, err := f.Data()
			if err != nil {
				screen.TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
				continue
			}

			header, err = highlight.MakeHeaderYaml(data)
			file, err := highlight.ParseFile(data)
			if err != nil {
				screen.TermMessage("Error parsing syntax file " + f.Name() + ": " + err.Error())
				continue
			}

			if (ft == "unknown" || ft == "" && highlight.MatchFiletype(header.FtDetect, b.Path, b.lines[0].data)) || header.FileType == ft {
				syndef, err := highlight.ParseDef(file, header)
				if err != nil {
					screen.TermMessage("Error parsing syntax file " + f.Name() + ": " + err.Error())
					continue
				}
				b.SyntaxDef = syndef
				break
			}
		}
	} else {
		for _, f := range config.ListRuntimeFiles(config.RTSyntax) {
			if f.Name() == syntaxFile {
				data, err := f.Data()
				if err != nil {
					screen.TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
					continue
				}

				file, err := highlight.ParseFile(data)
				if err != nil {
					screen.TermMessage("Error parsing syntax file " + f.Name() + ": " + err.Error())
					continue
				}

				syndef, err := highlight.ParseDef(file, header)
				if err != nil {
					screen.TermMessage("Error parsing syntax file " + f.Name() + ": " + err.Error())
					continue
				}
				b.SyntaxDef = syndef
				break
			}
		}
	}

	if b.SyntaxDef != nil && highlight.HasIncludes(b.SyntaxDef) {
		includes := highlight.GetIncludes(b.SyntaxDef)

		var files []*highlight.File
		for _, f := range config.ListRuntimeFiles(config.RTSyntax) {
			data, err := f.Data()
			if err != nil {
				screen.TermMessage("Error parsing syntax file " + f.Name() + ": " + err.Error())
				continue
			}
			header, err := highlight.MakeHeaderYaml(data)
			if err != nil {
				screen.TermMessage("Error parsing syntax file " + f.Name() + ": " + err.Error())
				continue
			}

			for _, i := range includes {
				if header.FileType == i {
					file, err := highlight.ParseFile(data)
					if err != nil {
						screen.TermMessage("Error parsing syntax file " + f.Name() + ": " + err.Error())
						continue
					}
					files = append(files, file)
					break
				}
			}
			if len(files) >= len(includes) {
				break
			}
		}

		highlight.ResolveIncludes(b.SyntaxDef, files)
	}

	if b.Highlighter == nil || syntaxFile != "" {
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
		return util.Spaces(tabsize)
	}
	return "\t"
}

// SetCursors resets this buffer's cursors to a new list
func (b *Buffer) SetCursors(c []*Cursor) {
	b.cursors = c
	b.EventHandler.cursors = b.cursors
	b.EventHandler.active = b.curCursor
}

// AddCursor adds a new cursor to the list
func (b *Buffer) AddCursor(c *Cursor) {
	b.cursors = append(b.cursors, c)
	b.EventHandler.cursors = b.cursors
	b.EventHandler.active = b.curCursor
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
	b.EventHandler.cursors = b.cursors
	b.EventHandler.active = b.curCursor
}

// UpdateCursors updates all the cursors indicies
func (b *Buffer) UpdateCursors() {
	b.EventHandler.cursors = b.cursors
	b.EventHandler.active = b.curCursor
	for i, c := range b.cursors {
		c.Num = i
	}
}

func (b *Buffer) RemoveCursor(i int) {
	copy(b.cursors[i:], b.cursors[i+1:])
	b.cursors[len(b.cursors)-1] = nil
	b.cursors = b.cursors[:len(b.cursors)-1]
	b.curCursor = util.Clamp(b.curCursor, 0, len(b.cursors)-1)
	b.UpdateCursors()
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

// MoveLinesUp moves the range of lines up one row
func (b *Buffer) MoveLinesUp(start int, end int) {
	if start < 1 || start >= end || end > len(b.lines) {
		return
	}
	l := string(b.LineBytes(start - 1))
	if end == len(b.lines) {
		b.Insert(
			Loc{
				utf8.RuneCount(b.lines[end-1].data),
				end - 1,
			},
			"\n"+l,
		)
	} else {
		b.Insert(
			Loc{0, end},
			l+"\n",
		)
	}
	b.Remove(
		Loc{0, start - 1},
		Loc{0, start},
	)
}

// MoveLinesDown moves the range of lines down one row
func (b *Buffer) MoveLinesDown(start int, end int) {
	if start < 0 || start >= end || end >= len(b.lines)-1 {
		return
	}
	l := string(b.LineBytes(end))
	b.Insert(
		Loc{0, start},
		l+"\n",
	)
	end++
	b.Remove(
		Loc{0, end},
		Loc{0, end + 1},
	)
}

var BracePairs = [][2]rune{
	{'(', ')'},
	{'{', '}'},
	{'[', ']'},
}

// FindMatchingBrace returns the location in the buffer of the matching bracket
// It is given a brace type containing the open and closing character, (for example
// '{' and '}') as well as the location to match from
// TODO: maybe can be more efficient with utf8 package
// returns the location of the matching brace
// if the boolean returned is true then the original matching brace is one character left
// of the starting location
func (b *Buffer) FindMatchingBrace(braceType [2]rune, start Loc) (Loc, bool) {
	curLine := []rune(string(b.LineBytes(start.Y)))
	startChar := ' '
	if start.X >= 0 && start.X < len(curLine) {
		startChar = curLine[start.X]
	}
	leftChar := ' '
	if start.X-1 >= 0 && start.X-1 < len(curLine) {
		leftChar = curLine[start.X-1]
	}
	var i int
	if startChar == braceType[0] || leftChar == braceType[0] {
		for y := start.Y; y < b.LinesNum(); y++ {
			l := []rune(string(b.LineBytes(y)))
			xInit := 0
			if y == start.Y {
				if startChar == braceType[0] {
					xInit = start.X
				} else {
					xInit = start.X - 1
				}
			}
			for x := xInit; x < len(l); x++ {
				r := l[x]
				if r == braceType[0] {
					i++
				} else if r == braceType[1] {
					i--
					if i == 0 {
						if startChar == braceType[0] {
							return Loc{x, y}, false
						}
						return Loc{x, y}, true
					}
				}
			}
		}
	} else if startChar == braceType[1] || leftChar == braceType[1] {
		for y := start.Y; y >= 0; y-- {
			l := []rune(string(b.lines[y].data))
			xInit := len(l) - 1
			if y == start.Y {
				if leftChar == braceType[1] {
					xInit = start.X - 1
				} else {
					xInit = start.X
				}
			}
			for x := xInit; x >= 0; x-- {
				r := l[x]
				if r == braceType[0] {
					i--
					if i == 0 {
						if leftChar == braceType[1] {
							return Loc{x, y}, true
						}
						return Loc{x, y}, false
					}
				} else if r == braceType[1] {
					i++
				}
			}
		}
	}
	return start, true
}

// Retab changes all tabs to spaces or vice versa
func (b *Buffer) Retab() {
	toSpaces := b.Settings["tabstospaces"].(bool)
	tabsize := util.IntOpt(b.Settings["tabsize"])
	dirty := false

	for i := 0; i < b.LinesNum(); i++ {
		l := b.LineBytes(i)

		ws := util.GetLeadingWhitespace(l)
		if len(ws) != 0 {
			if toSpaces {
				ws = bytes.Replace(ws, []byte{'\t'}, bytes.Repeat([]byte{' '}, tabsize), -1)
			} else {
				ws = bytes.Replace(ws, bytes.Repeat([]byte{' '}, tabsize), []byte{'\t'}, -1)
			}
		}

		l = bytes.TrimLeft(l, " \t")
		b.lines[i].data = append(ws, l...)
		dirty = true
	}

	b.isModified = dirty
}

// ParseCursorLocation turns a cursor location like 10:5 (LINE:COL)
// into a loc
func ParseCursorLocation(cursorPositions []string) (Loc, error) {
	startpos := Loc{0, 0}
	var err error

	// if no positions are available exit early
	if cursorPositions == nil {
		return startpos, errors.New("No cursor positions were provided.")
	}

	startpos.Y, err = strconv.Atoi(cursorPositions[0])
	startpos.Y -= 1
	if err == nil {
		if len(cursorPositions) > 1 {
			startpos.X, err = strconv.Atoi(cursorPositions[1])
			if startpos.X > 0 {
				startpos.X -= 1
			}
		}
	}

	return startpos, err
}

// Line returns the string representation of the given line number
func (b *Buffer) Line(i int) string {
	return string(b.LineBytes(i))
}

// WriteLog writes a string to the log buffer
func WriteLog(s string) {
	LogBuf.EventHandler.Insert(LogBuf.End(), s)
}

// GetLogBuf returns the log buffer
func GetLogBuf() *Buffer {
	return LogBuf
}
