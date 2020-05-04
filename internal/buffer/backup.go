package buffer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/micro/v2/internal/util"
	"golang.org/x/text/encoding"
)

const backupMsg = `A backup was detected for this file. This likely means that micro
crashed while editing this file, or another instance of micro is currently
editing this file.

The backup was created on %s, and the file is

%s

* 'recover' will apply the backup as unsaved changes to the current buffer.
  When the buffer is closed, the backup will be removed.
* 'ignore' will ignore the backup, discarding its changes. The backup file
  will be removed.

Options: [r]ecover, [i]gnore: `

// Backup saves the current buffer to ConfigDir/backups
func (b *Buffer) Backup(checkTime bool) error {
	if !b.Settings["backup"].(bool) || b.Path == "" || b.Type != BTDefault {
		return nil
	}

	if checkTime {
		sub := time.Now().Sub(b.lastbackup)
		if sub < time.Duration(backupTime)*time.Millisecond {
			return nil
		}
	}

	b.lastbackup = time.Now()

	backupdir := filepath.Join(config.ConfigDir, "backups")
	if _, err := os.Stat(backupdir); os.IsNotExist(err) {
		os.Mkdir(backupdir, os.ModePerm)
	}

	name := filepath.Join(backupdir, util.EscapePath(b.AbsPath))

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
	}, false)

	return err
}

// RemoveBackup removes any backup file associated with this buffer
func (b *Buffer) RemoveBackup() {
	if !b.Settings["backup"].(bool) || b.Path == "" || b.Type != BTDefault {
		return
	}
	f := filepath.Join(config.ConfigDir, "backups", util.EscapePath(b.AbsPath))
	os.Remove(f)
}

// ApplyBackup applies the corresponding backup file to this buffer (if one exists)
// Returns true if a backup was applied
func (b *Buffer) ApplyBackup(fsize int64) bool {
	if b.Settings["backup"].(bool) && len(b.Path) > 0 && b.Type == BTDefault {
		backupfile := filepath.Join(config.ConfigDir, "backups", util.EscapePath(b.AbsPath))
		if info, err := os.Stat(backupfile); err == nil {
			backup, err := os.Open(backupfile)
			if err == nil {
				defer backup.Close()
				t := info.ModTime()
				msg := fmt.Sprintf(backupMsg, t.Format("Mon Jan _2 at 15:04, 2006"), util.EscapePath(b.AbsPath))
				choice := screen.TermPrompt(msg, []string{"r", "i", "recover", "ignore"}, true)

				if choice%2 == 0 {
					// recover
					b.LineArray = NewLineArray(uint64(fsize), FFAuto, backup)
					b.isModified = true
					return true
				} else if choice%2 == 1 {
					// delete
					os.Remove(backupfile)
				}
			}
		}
	}

	return false
}
