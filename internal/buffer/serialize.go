package buffer

import (
	"encoding/gob"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/text/encoding"

	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/util"
)

// The SerializedBuffer holds the types that get serialized when a buffer is saved
// These are used for the savecursor and saveundo options
type SerializedBuffer struct {
	EventHandler *EventHandler
	Cursor       Loc
	ModTime      time.Time
}

// Serialize serializes the buffer to config.ConfigDir/buffers
func (b *Buffer) Serialize() error {
	if !b.Settings["savecursor"].(bool) && !b.Settings["saveundo"].(bool) {
		return nil
	}
	if b.Path == "" {
		return nil
	}

	name := filepath.Join(config.ConfigDir, "buffers", util.EscapePath(b.AbsPath))

	return overwriteFile(name, encoding.Nop, func(file io.Writer) error {
		err := gob.NewEncoder(file).Encode(SerializedBuffer{
			b.EventHandler,
			b.GetActiveCursor().Loc,
			b.ModTime,
		})
		return err
	}, false)
}

// Unserialize loads the buffer info from config.ConfigDir/buffers
func (b *Buffer) Unserialize() error {
	// If either savecursor or saveundo is turned on, we need to load the serialized information
	// from ~/.config/micro/buffers
	if b.Path == "" {
		return nil
	}
	file, err := os.Open(filepath.Join(config.ConfigDir, "buffers", util.EscapePath(b.AbsPath)))
	defer file.Close()
	if err == nil {
		var buffer SerializedBuffer
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(&buffer)
		if err != nil {
			return errors.New(err.Error() + "\nYou may want to remove the files in ~/.config/micro/buffers (these files\nstore the information for the 'saveundo' and 'savecursor' options) if\nthis problem persists.\nThis may be caused by upgrading to version 2.0, and removing the 'buffers'\ndirectory will reset the cursor and undo history and solve the problem.")
		}
		if b.Settings["savecursor"].(bool) {
			b.StartCursor = buffer.Cursor
		}

		if b.Settings["saveundo"].(bool) {
			// We should only use last time's eventhandler if the file wasn't modified by someone else in the meantime
			if b.ModTime == buffer.ModTime {
				b.EventHandler = buffer.EventHandler
				b.EventHandler.cursors = b.cursors
				b.EventHandler.buf = b.SharedBuffer
			}
		}
	}
	return nil
}
