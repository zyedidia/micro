package buffer

import (
	"encoding/gob"
	"errors"
	"io"
	"os"

	"github.com/zyedidia/micro/cmd/micro/config"
	. "github.com/zyedidia/micro/cmd/micro/util"
)

func init() {
	gob.Register(TextEvent{})
	gob.Register(SerializedBuffer{})
}

// Serialize serializes the buffer to config.ConfigDir/buffers
func (b *Buffer) Serialize() error {
	if !b.Settings["savecursor"].(bool) && !b.Settings["saveundo"].(bool) {
		return nil
	}

	name := config.ConfigDir + "/buffers/" + EscapePath(b.AbsPath)

	return overwriteFile(name, func(file io.Writer) error {
		return gob.NewEncoder(file).Encode(SerializedBuffer{
			b.EventHandler,
			b.GetActiveCursor().Loc,
			b.ModTime,
		})
	})
}

func (b *Buffer) Unserialize() error {
	// If either savecursor or saveundo is turned on, we need to load the serialized information
	// from ~/.config/micro/buffers
	file, err := os.Open(config.ConfigDir + "/buffers/" + EscapePath(b.AbsPath))
	defer file.Close()
	if err == nil {
		var buffer SerializedBuffer
		decoder := gob.NewDecoder(file)
		gob.Register(TextEvent{})
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
				b.EventHandler.buf = b
			}
		}
	}
	return err
}
