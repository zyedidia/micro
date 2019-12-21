package buffer

import (
	"io"

	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/micro/internal/util"
	"golang.org/x/text/encoding"
)

// Backup saves the current buffer to ConfigDir/backups
func (b *Buffer) Backup() error {
	if !b.Settings["backup"].(bool) {
		return nil
	}

	name := config.ConfigDir + "/backups" + util.EscapePath(b.AbsPath)

	err := overwriteFile(name, encoding.Nop, func(file io.Writer) (e error) {
		if len(b.lines) == 0 {
			return
		}

		// end of line
		eol := []byte{'\n'}

		// write lines
		if _, e = file.Write(b.lines[0].data); e != nil {
			return
		}

		for _, l := range b.lines[1:] {
			if _, e = file.Write(eol); e != nil {
				return
			}
			if _, e = file.Write(l.data); e != nil {
				return
			}
		}
		return
	})

	return err
}

// ApplyBackup applies the corresponding backup file to this buffer (if one exists)
func (b *Buffer) ApplyBackup() error {
	return nil
}
