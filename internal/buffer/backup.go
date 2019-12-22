package buffer

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/micro/internal/util"
	"golang.org/x/text/encoding"
)

const backupMsg = `A backup was detected for this file. This likely means that micro
crashed while editing this file, or another instance of micro is currently
editing this file.

The backup was created at %s.

* 'recover' will apply the backup as unsaved changes to the current buffer.
  When the buffer is closed, the backup will be removed.
* 'ignore' will ignore the backup, discarding its changes. The backup file
  will be removed.

Options: [r]ecover, [i]gnore: `

// Backup saves the current buffer to ConfigDir/backups
func (b *Buffer) Backup(checkTime bool) error {
	if !b.Settings["backup"].(bool) {
		return nil
	}

	if checkTime {
		sub := time.Now().Sub(b.lastbackup)
		if sub < time.Duration(backup_time)*time.Millisecond {
			log.Println("Backup event but not enough time has passed", sub)
			return nil
		}
	}

	b.lastbackup = time.Now()

	backupdir := config.ConfigDir + "/backups/"
	if _, err := os.Stat(backupdir); os.IsNotExist(err) {
		os.Mkdir(backupdir, os.ModePerm)
		log.Println("Creating backup dir")
	}

	name := backupdir + util.EscapePath(b.AbsPath)

	log.Println("Backing up to", name)

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

// RemoveBackup removes any backup file associated with this buffer
func (b *Buffer) RemoveBackup() {
	if !b.Settings["backup"].(bool) {
		return
	}
	f := config.ConfigDir + "/backups/" + util.EscapePath(b.AbsPath)
	os.Remove(f)
}

// ApplyBackup applies the corresponding backup file to this buffer (if one exists)
func (b *Buffer) ApplyBackup() error {
	return nil
}
