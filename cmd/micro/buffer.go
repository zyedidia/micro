package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/zyedidia/micro/cmd/micro/highlight"
)

const LargeFileThreshold = 50000

var (
	// 0 - no line type detected
	// 1 - lf detected
	// 2 - crlf detected
	fileformat = 0
)

// Buffer stores the text for files that are loaded into the text editor
// It uses a rope to efficiently store the string and contains some
// simple functions for saving and wrapper functions for modifying the rope
type Buffer struct {
	// The eventhandler for undo/redo
	*EventHandler
	// This stores all the text in the buffer as an array of lines
	*LineArray

	Cursor    Cursor
	cursors   []*Cursor // for multiple cursors
	curCursor int       // the current cursor

	// Path to the file on disk
	Path string
	// Absolute path to the file on disk
	AbsPath string
	// Name of the buffer on the status line
	name string

	// Whether or not the buffer has been modified since it was opened
	IsModified bool

	// Stores the last modification time of the file the buffer is pointing to
	ModTime time.Time

	// NumLines is the number of lines in the buffer
	NumLines int

	syntaxDef   *highlight.Def
	highlighter *highlight.Highlighter

	// Hash of the original buffer -- empty if fastdirty is on
	origHash [md5.Size]byte

	// Buffer local settings
	Settings map[string]interface{}
}

// The SerializedBuffer holds the types that get serialized when a buffer is saved
// These are used for the savecursor and saveundo options
type SerializedBuffer struct {
	EventHandler *EventHandler
	Cursor       Cursor
	ModTime      time.Time
}

// NewBufferFromFile opens a new buffer using the given path
// It will also automatically handle `~`, and line/column with filename:l:c
// It will return an empty buffer if the path does not exist
// and an error if the file is a directory
func NewBufferFromFile(path string) (*Buffer, error) {
	filename, cursorPosition := GetPathAndCursorPosition(path)
	filename = ReplaceHome(filename)
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
func NewBuffer(reader io.Reader, size int64, path string, cursorPosition []string) *Buffer {
	// check if the file is already open in a tab. If it's open return the buffer to that tab
	if path != "" {
		for _, tab := range tabs {
			for _, view := range tab.Views {
				if view.Buf.Path == path {
					return view.Buf
				}
			}
		}
	}

	b := new(Buffer)
	b.LineArray = NewLineArray(size, reader)

	b.Settings = DefaultLocalSettings()
	for k, v := range globalSettings {
		if _, ok := b.Settings[k]; ok {
			b.Settings[k] = v
		}
	}

	if fileformat == 1 {
		b.Settings["fileformat"] = "unix"
	} else if fileformat == 2 {
		b.Settings["fileformat"] = "dos"
	}

	absPath, _ := filepath.Abs(path)

	b.Path = path
	b.AbsPath = absPath

	// The last time this file was modified
	b.ModTime, _ = GetModTime(b.Path)

	b.EventHandler = NewEventHandler(b)

	b.Update()
	b.UpdateRules()

	if _, err := os.Stat(configDir + "/buffers/"); os.IsNotExist(err) {
		os.Mkdir(configDir+"/buffers/", os.ModePerm)
	}

	cursorLocation, cursorLocationError := GetBufferCursorLocation(cursorPosition, b)
	b.Cursor = Cursor{
		Loc: cursorLocation,
		buf: b,
	}

	InitLocalSettings(b)

	if cursorLocationError != nil && len(*flagStartPos) == 0 && (b.Settings["savecursor"].(bool) || b.Settings["saveundo"].(bool)) {
		// If either savecursor or saveundo is turned on, we need to load the serialized information
		// from ~/.config/micro/buffers
		file, err := os.Open(configDir + "/buffers/" + EscapePath(b.AbsPath))
		defer file.Close()
		if err == nil {
			var buffer SerializedBuffer
			decoder := gob.NewDecoder(file)
			gob.Register(TextEvent{})
			err = decoder.Decode(&buffer)
			if err != nil {
				TermMessage(err.Error(), "\n", "You may want to remove the files in ~/.config/micro/buffers (these files store the information for the 'saveundo' and 'savecursor' options) if this problem persists.")
			}
			if b.Settings["savecursor"].(bool) {
				b.Cursor = buffer.Cursor
				b.Cursor.buf = b
				b.Cursor.Relocate()
			}

			if b.Settings["saveundo"].(bool) {
				// We should only use last time's eventhandler if the file wasn't modified by someone else in the meantime
				if b.ModTime == buffer.ModTime {
					b.EventHandler = buffer.EventHandler
					b.EventHandler.buf = b
				}
			}
		}
	}

	if !b.Settings["fastdirty"].(bool) {
		if size > LargeFileThreshold {
			// If the file is larger than a megabyte fastdirty needs to be on
			b.Settings["fastdirty"] = true
		} else {
			calcHash(b, &b.origHash)
		}
	}

	b.cursors = []*Cursor{&b.Cursor}

	return b
}

func GetBufferCursorLocation(cursorPosition []string, b *Buffer) (Loc, error) {
	// parse the cursor position. The cursor location is ALWAYS initialised to 0, 0 even when
	// an error occurs due to lack of arguments or because the arguments are not numbers
	cursorLocation, cursorLocationError := ParseCursorLocation(cursorPosition)

	// Put the cursor at the first spot. In the logic for cursor position the -startpos
	// flag is processed first and will overwrite any line/col parameters with colons after the filename
	if len(*flagStartPos) > 0 || cursorLocationError == nil {
		var lineNum, colNum int
		var errPos1, errPos2 error

		positions := strings.Split(*flagStartPos, ",")

		// if the -startpos flag contains enough args use them for the cursor location.
		// In this case args passed at the end of the filename will be ignored
		if len(positions) == 2 {
			lineNum, errPos1 = strconv.Atoi(positions[0])
			colNum, errPos2 = strconv.Atoi(positions[1])
		}
		// if -startpos has invalid arguments, use the arguments from filename.
		// This will have a default value (0, 0) even when the filename arguments are invalid
		if errPos1 != nil || errPos2 != nil {
			// otherwise check if there are any arguments after the filename and use them
			lineNum = cursorLocation.Y
			colNum = cursorLocation.X
		}

		// if some arguments were found make sure they don't go outside the file and cause overflows
		cursorLocation.Y = lineNum
		cursorLocation.X = colNum
		// Check to avoid line overflow
		if cursorLocation.Y > b.NumLines-1 {
			cursorLocation.Y = b.NumLines - 1
		} else if cursorLocation.Y < 0 {
			cursorLocation.Y = 0
		}
		// Check to avoid column overflow
		if cursorLocation.X > len(b.Line(cursorLocation.Y)) {
			cursorLocation.X = len(b.Line(cursorLocation.Y))
		} else if cursorLocation.X < 0 {
			cursorLocation.X = 0
		}
	}
	return cursorLocation, cursorLocationError
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

// FileType returns the buffer's filetype
func (b *Buffer) FileType() string {
	return b.Settings["filetype"].(string)
}

// IndentString returns a string representing one level of indentation
func (b *Buffer) IndentString() string {
	if b.Settings["tabstospaces"].(bool) {
		return Spaces(int(b.Settings["tabsize"].(float64)))
	}
	return "\t"
}

// CheckModTime makes sure that the file this buffer points to hasn't been updated
// by an external program since it was last read
// If it has, we ask the user if they would like to reload the file
func (b *Buffer) CheckModTime() {
	modTime, ok := GetModTime(b.Path)
	if ok {
		if modTime != b.ModTime {
			choice, canceled := messenger.YesNoPrompt("The file has changed since it was last read. Reload file? (y,n)")
			messenger.Reset()
			messenger.Clear()
			if !choice || canceled {
				// Don't load new changes -- do nothing
				b.ModTime, _ = GetModTime(b.Path)
			} else {
				// Load new changes
				b.ReOpen()
			}
		}
	}
}

// ReOpen reloads the current buffer from disk
func (b *Buffer) ReOpen() {
	data, err := ioutil.ReadFile(b.Path)
	txt := string(data)

	if err != nil {
		messenger.Error(err.Error())
		return
	}
	b.EventHandler.ApplyDiff(txt)

	b.ModTime, _ = GetModTime(b.Path)
	b.IsModified = false
	b.Update()
	b.Cursor.Relocate()
}

// Update fetches the string from the rope and updates the `text` and `lines` in the buffer
func (b *Buffer) Update() {
	b.NumLines = len(b.lines)
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

// Save saves the buffer to its default path
func (b *Buffer) Save() error {
	return b.SaveAs(b.Path)
}

// SaveWithSudo saves the buffer to the default path with sudo
func (b *Buffer) SaveWithSudo() error {
	return b.SaveAsWithSudo(b.Path)
}

// Serialize serializes the buffer to configDir/buffers
func (b *Buffer) Serialize() error {
	if !b.Settings["savecursor"].(bool) && !b.Settings["saveundo"].(bool) {
		return nil
	}

	name := configDir + "/buffers/" + EscapePath(b.AbsPath)

	return overwriteFile(name, func(file io.Writer) error {
		return gob.NewEncoder(file).Encode(SerializedBuffer{
			b.EventHandler,
			b.Cursor,
			b.ModTime,
		})
	})
}

func init() {
	gob.Register(TextEvent{})
	gob.Register(SerializedBuffer{})
}

// SaveAs saves the buffer to a specified path (filename), creating the file if it does not exist
func (b *Buffer) SaveAs(filename string) error {
	b.UpdateRules()
	if b.Settings["rmtrailingws"].(bool) {
		for i, l := range b.lines {
			pos := len(bytes.TrimRightFunc(l.data, unicode.IsSpace))

			if pos < len(l.data) {
				b.deleteToEnd(Loc{pos, i})
			}
		}

		b.Cursor.Relocate()
	}

	if b.Settings["eofnewline"].(bool) {
		end := b.End()
		if b.RuneAt(Loc{end.X - 1, end.Y}) != '\n' {
			b.Insert(end, "\n")
		}
	}

	defer func() {
		b.ModTime, _ = GetModTime(filename)
	}()

	// Removes any tilde and replaces with the absolute path to home
	absFilename := ReplaceHome(filename)

	// Get the leading path to the file | "." is returned if there's no leading path provided
	if dirname := filepath.Dir(absFilename); dirname != "." {
		// Check if the parent dirs don't exist
		if _, statErr := os.Stat(dirname); os.IsNotExist(statErr) {
			// Prompt to make sure they want to create the dirs that are missing
			if yes, canceled := messenger.YesNoPrompt("Parent folders \"" + dirname + "\" do not exist. Create them? (y,n)"); yes && !canceled {
				// Create all leading dir(s) since they don't exist
				if mkdirallErr := os.MkdirAll(dirname, os.ModePerm); mkdirallErr != nil {
					// If there was an error creating the dirs
					return mkdirallErr
				}
			} else {
				// If they canceled the creation of leading dirs
				return errors.New("Save aborted")
			}
		}
	}

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
	b.IsModified = false
	return b.Serialize()
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

// SaveAsWithSudo is the same as SaveAs except it uses a neat trick
// with tee to use sudo so the user doesn't have to reopen micro with sudo
func (b *Buffer) SaveAsWithSudo(filename string) error {
	b.UpdateRules()
	b.Path = filename

	// Shut down the screen because we're going to interact directly with the shell
	screen.Fini()
	screen = nil

	// Set up everything for the command
	cmd := exec.Command(globalSettings["sucmd"].(string), "tee", filename)
	cmd.Stdin = bytes.NewBufferString(b.SaveString(b.Settings["fileformat"] == "dos"))

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

	// Start the screen back up
	InitScreen()
	if err == nil {
		b.IsModified = false
		b.ModTime, _ = GetModTime(filename)
		b.Serialize()
	}
	return err
}

// Modified returns if this buffer has been modified since
// being opened
func (b *Buffer) Modified() bool {
	if b.Settings["fastdirty"].(bool) {
		return b.IsModified
	}

	var buff [md5.Size]byte

	calcHash(b, &buff)
	return buff != b.origHash
}

func (b *Buffer) insert(pos Loc, value []byte) {
	b.IsModified = true
	b.LineArray.insert(pos, value)
	b.Update()
}
func (b *Buffer) remove(start, end Loc) string {
	b.IsModified = true
	sub := b.LineArray.remove(start, end)
	b.Update()
	return sub
}
func (b *Buffer) deleteToEnd(start Loc) {
	b.IsModified = true
	b.LineArray.DeleteToEnd(start)
	b.Update()
}

// Start returns the location of the first character in the buffer
func (b *Buffer) Start() Loc {
	return Loc{0, 0}
}

// End returns the location of the last character in the buffer
func (b *Buffer) End() Loc {
	return Loc{utf8.RuneCount(b.lines[b.NumLines-1].data), b.NumLines - 1}
}

// RuneAt returns the rune at a given location in the buffer
func (b *Buffer) RuneAt(loc Loc) rune {
	line := b.LineRunes(loc.Y)
	if len(line) > 0 {
		return line[loc.X]
	}
	return '\n'
}

// LineBytes returns a single line as an array of runes
func (b *Buffer) LineBytes(n int) []byte {
	if n >= len(b.lines) {
		return []byte{}
	}
	return b.lines[n].data
}

// LineRunes returns a single line as an array of runes
func (b *Buffer) LineRunes(n int) []rune {
	if n >= len(b.lines) {
		return []rune{}
	}
	return toRunes(b.lines[n].data)
}

// Line returns a single line
func (b *Buffer) Line(n int) string {
	if n >= len(b.lines) {
		return ""
	}
	return string(b.lines[n].data)
}

// LinesNum returns the number of lines in the buffer
func (b *Buffer) LinesNum() int {
	return len(b.lines)
}

// Lines returns an array of strings containing the lines from start to end
func (b *Buffer) Lines(start, end int) []string {
	lines := b.lines[start:end]
	var slice []string
	for _, line := range lines {
		slice = append(slice, string(line.data))
	}
	return slice
}

// Len gives the length of the buffer
func (b *Buffer) Len() (n int) {
	for _, l := range b.lines {
		n += utf8.RuneCount(l.data)
	}

	if len(b.lines) > 1 {
		n += len(b.lines) - 1 // account for newlines
	}

	return
}

// MoveLinesUp moves the range of lines up one row
func (b *Buffer) MoveLinesUp(start int, end int) {
	// 0 < start < end <= len(b.lines)
	if start < 1 || start >= end || end > len(b.lines) {
		return // what to do? FIXME
	}
	if end == len(b.lines) {
		b.Insert(
			Loc{
				utf8.RuneCount(b.lines[end-1].data),
				end - 1,
			},
			"\n"+b.Line(start-1),
		)
	} else {
		b.Insert(
			Loc{0, end},
			b.Line(start-1)+"\n",
		)
	}
	b.Remove(
		Loc{0, start - 1},
		Loc{0, start},
	)
}

// MoveLinesDown moves the range of lines down one row
func (b *Buffer) MoveLinesDown(start int, end int) {
	// 0 <= start < end < len(b.lines)
	// if end == len(b.lines), we can't do anything here because the
	// last line is unaccessible, FIXME
	if start < 0 || start >= end || end >= len(b.lines)-1 {
		return // what to do? FIXME
	}
	b.Insert(
		Loc{0, start},
		b.Line(end)+"\n",
	)
	end++
	b.Remove(
		Loc{0, end},
		Loc{0, end + 1},
	)
}

// ClearMatches clears all of the syntax highlighting for this buffer
func (b *Buffer) ClearMatches() {
	for i := range b.lines {
		b.SetMatch(i, nil)
		b.SetState(i, nil)
	}
}

func (b *Buffer) clearCursors() {
	for i := 1; i < len(b.cursors); i++ {
		b.cursors[i] = nil
	}
	b.cursors = b.cursors[:1]
	b.UpdateCursors()
	b.Cursor.ResetSelection()
}

var bracePairs = [][2]rune{
	{'(', ')'},
	{'{', '}'},
	{'[', ']'},
}

// FindMatchingBrace returns the location in the buffer of the matching bracket
// It is given a brace type containing the open and closing character, (for example
// '{' and '}') as well as the location to match from
func (b *Buffer) FindMatchingBrace(braceType [2]rune, start Loc) Loc {
	curLine := b.LineRunes(start.Y)
	startChar := curLine[start.X]
	var i int
	if startChar == braceType[0] {
		for y := start.Y; y < b.NumLines; y++ {
			l := b.LineRunes(y)
			xInit := 0
			if y == start.Y {
				xInit = start.X
			}
			for x := xInit; x < len(l); x++ {
				r := l[x]
				if r == braceType[0] {
					i++
				} else if r == braceType[1] {
					i--
					if i == 0 {
						return Loc{x, y}
					}
				}
			}
		}
	} else if startChar == braceType[1] {
		for y := start.Y; y >= 0; y-- {
			l := []rune(string(b.lines[y].data))
			xInit := len(l) - 1
			if y == start.Y {
				xInit = start.X
			}
			for x := xInit; x >= 0; x-- {
				r := l[x]
				if r == braceType[0] {
					i--
					if i == 0 {
						return Loc{x, y}
					}
				} else if r == braceType[1] {
					i++
				}
			}
		}
	}
	return start
}
