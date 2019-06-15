package buffer

import (
	"encoding/gob"
	"errors"
	"io"
	"os"
	"time"

	"golang.org/x/text/encoding"

	"github.com/zyedidia/micro/internal/config"
	. "github.com/zyedidia/micro/internal/util"
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

	name := config.ConfigDir + "/buffers/" + EscapePath(b.AbsPath)

	return overwriteFile(name, encoding.Nop, func(file io.Writer) error {
		err := gob.NewEncoder(file).Encode(SerializedBuffer{
			b.EventHandler,
			b.GetActiveCursor().Loc,
			b.ModTime,
		})
		return err
	})
}

func (b *Buffer) Unserialize() error {
	// If either savecursor or saveundo is turned on, we need to load the serialized information
	// from ~/.config/micro/buffers
	if b.Path == "" {
		return nil
	}
	file, err := os.Open(config.ConfigDir + "/buffers/" + EscapePath(b.AbsPath))
	defer file.Close()
	if err == nil {
		var buffer SerializedBuffer
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(&buffer)
		if err != nil {
			return errors.New(err.Error() + "\nYou may want to remove the files in ~/.config/micro/buffers (these files store the information for the 'saveundo' and 'savecursor' options) if this problem persists.")
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
