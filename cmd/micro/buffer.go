package main

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"
	"unicode/utf8"
)

// Buffer stores the text for files that are loaded into the text editor
// It uses a rope to efficiently store the string and contains some
// simple functions for saving and wrapper functions for modifying the rope
type Buffer struct {
	// The eventhandler for undo/redo
	*EventHandler
	*LineArray

	Cursor Cursor

	// Path to the file on disk
	Path string
	// Name of the buffer on the status line
	Name string

	IsModified bool

	// Stores the last modification time of the file the buffer is pointing to
	ModTime time.Time

	NumLines int

	// Syntax highlighting rules
	rules []SyntaxRule
	// The buffer's filetype
	FileType string
}

// The SerializedBuffer holds the types that get serialized when a buffer is saved
type SerializedBuffer struct {
	EventHandler *EventHandler
	Cursor       Cursor
	ModTime      time.Time
}

// NewBuffer creates a new buffer from `txt` with path and name `path`
func NewBuffer(txt []byte, path string) *Buffer {
	b := new(Buffer)
	b.LineArray = NewLineArray(txt)
	b.Path = path
	b.Name = path

	b.ModTime, _ = GetModTime(b.Path)

	b.EventHandler = NewEventHandler(b)

	b.Update()
	b.UpdateRules()

	if _, err := os.Stat(configDir + "/buffers/"); os.IsNotExist(err) {
		os.Mkdir(configDir+"/buffers/", os.ModePerm)
	}

	// Put the cursor at the first spot
	b.Cursor = Cursor{
		Loc: Loc{
			X: 0,
			Y: 0,
		},
		buf: b,
	}

	if settings["savecursor"].(bool) || settings["saveundo"].(bool) {
		absPath, _ := filepath.Abs(b.Path)
		file, err := os.Open(configDir + "/buffers/" + EscapePath(absPath))
		if err == nil {
			var buffer SerializedBuffer
			decoder := gob.NewDecoder(file)
			gob.Register(TextEvent{})
			err = decoder.Decode(&buffer)
			if err != nil {
				TermMessage(err.Error(), "\n", "You may want to remove the files in ~/.config/micro/buffers (these files store the information for the 'saveundo' and 'savecursor' options) if this problem persists.")
			}
			if settings["savecursor"].(bool) {
				b.Cursor = buffer.Cursor
				b.Cursor.buf = b
				b.Cursor.Relocate()
			}

			if settings["saveundo"].(bool) {
				// We should only use last time's eventhandler if the file wasn't by someone else in the meantime
				if b.ModTime == buffer.ModTime {
					b.EventHandler = buffer.EventHandler
					b.EventHandler.buf = b
				}
			}
		}
		file.Close()
	}

	return b
}

// UpdateRules updates the syntax rules and filetype for this buffer
// This is called when the colorscheme changes
func (b *Buffer) UpdateRules() {
	b.rules, b.FileType = GetRules(b)
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
	if settings["savecursor"].(bool) || settings["saveundo"].(bool) {
		absPath, _ := filepath.Abs(b.Path)
		file, err := os.Create(configDir + "/buffers/" + EscapePath(absPath))
		if err == nil {
			enc := gob.NewEncoder(file)
			gob.Register(TextEvent{})
			err = enc.Encode(SerializedBuffer{
				b.EventHandler,
				b.Cursor,
				b.ModTime,
			})
			// err = enc.Encode(b.Cursor)
		}
		file.Close()
		return err
	}
	return nil
}

// SaveAs saves the buffer to a specified path (filename), creating the file if it does not exist
func (b *Buffer) SaveAs(filename string) error {
	b.UpdateRules()
	b.Name = filename
	b.Path = filename
	data := []byte(b.String())
	err := ioutil.WriteFile(filename, data, 0644)
	if err == nil {
		b.IsModified = false
		b.ModTime, _ = GetModTime(filename)
		return b.Serialize()
	}
	return err
}

// SaveAsWithSudo is the same as SaveAs except it uses a neat trick
// with tee to use sudo so the user doesn't have to reopen micro with sudo
func (b *Buffer) SaveAsWithSudo(filename string) error {
	b.UpdateRules()
	b.Name = filename
	b.Path = filename

	// The user may have already used sudo in which case we won't need the password
	// It's a bit nicer for them if they don't have to enter the password every time
	_, err := RunShellCommand("sudo -v")
	needPassword := err != nil

	// If we need the password, we have to close the screen and ask using the shell
	if needPassword {
		// Shut down the screen because we're going to interact directly with the shell
		screen.Fini()
		screen = nil
	}

	// Set up everything for the command
	cmd := exec.Command("sudo", "tee", filename)
	cmd.Stdin = bytes.NewBufferString(b.String())

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
	err = cmd.Wait()

	// If we needed the password, we closed the screen, so we have to initialize it again
	if needPassword {
		// Start the screen back up
		InitScreen()
	}
	if err == nil {
		b.IsModified = false
		b.ModTime, _ = GetModTime(filename)
		b.Serialize()
	}
	return err
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

// Start returns the location of the first character in the buffer
func (b *Buffer) Start() Loc {
	return Loc{0, 0}
}

// End returns the location of the last character in the buffer
func (b *Buffer) End() Loc {
	return Loc{utf8.RuneCount(b.lines[b.NumLines-1]), b.NumLines - 1}
}

// Line returns a single line
func (b *Buffer) Line(n int) string {
	return string(b.lines[n])
}

// Lines returns an array of strings containing the lines from start to end
func (b *Buffer) Lines(start, end int) []string {
	lines := b.lines[start:end]
	var slice []string
	for _, line := range lines {
		slice = append(slice, string(line))
	}
	return slice
}

// Len gives the length of the buffer
func (b *Buffer) Len() int {
	return Count(b.String())
}
