package buffer

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/micro/v2/internal/util"
)

const BackupMsg = `A backup was detected for:

%s

This likely means that micro crashed while editing this file,
or another instance of micro is currently editing this file,
or an error occurred while saving this file so it may be corrupted.

The backup was created on %s and its path is:

%s

* 'recover' will apply the backup as unsaved changes to the current buffer.
  When the buffer is closed, the backup will be removed.
* 'ignore' will ignore the backup, discarding its changes. The backup file
  will be removed.
* 'abort' will abort the open operation, and instead open an empty buffer.

Options: [r]ecover, [i]gnore, [a]bort: `

const backupSeconds = 8

type backupRequestType int

const (
	backupCreate = iota
	backupRemove
)

type backupRequest struct {
	buf     *SharedBuffer
	reqType backupRequestType
}

var requestedBackups map[*SharedBuffer]bool

func init() {
	requestedBackups = make(map[*SharedBuffer]bool)
}

func (b *SharedBuffer) RequestBackup() {
	backupRequestChan <- backupRequest{buf: b, reqType: backupCreate}
}

func (b *SharedBuffer) CancelBackup() {
	backupRequestChan <- backupRequest{buf: b, reqType: backupRemove}
}

func handleBackupRequest(br backupRequest) {
	switch br.reqType {
	case backupCreate:
		// schedule periodic backup
		requestedBackups[br.buf] = true
	case backupRemove:
		br.buf.RemoveBackup()
		delete(requestedBackups, br.buf)
	}
}

func periodicBackup() {
	for buf := range requestedBackups {
		err := buf.Backup()
		if err == nil {
			delete(requestedBackups, buf)
		}
	}
}

func (b *SharedBuffer) backupDir() string {
	backupdir, err := util.ReplaceHome(b.Settings["backupdir"].(string))
	if backupdir == "" || err != nil {
		backupdir = filepath.Join(config.ConfigDir, "backups")
	}
	return backupdir
}

func (b *SharedBuffer) keepBackup() bool {
	return b.forceKeepBackup || b.Settings["permbackup"].(bool)
}

func (b *SharedBuffer) writeBackup(path string) (string, error) {
	backupdir := b.backupDir()
	if _, err := os.Stat(backupdir); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		if err = os.Mkdir(backupdir, os.ModePerm); err != nil {
			return "", err
		}
	}

	name := util.DetermineEscapePath(backupdir, path)

	// If no existing backup, just write the backup.
	if _, err := os.Stat(name); errors.Is(err, fs.ErrNotExist) {
		_, err = b.overwriteFile(name)
		if err != nil {
			os.Remove(name)
		}
		return name, err
	}

	// If a backup already exists, replace it atomically.
	tmp := util.AppendBackupSuffix(name)
	_, err := b.overwriteFile(tmp)
	if err != nil {
		os.Remove(tmp)
		return name, err
	}
	err = os.Rename(tmp, name)
	if err != nil {
		os.Remove(tmp)
		return name, err
	}

	return name, nil
}

// Backup saves the buffer to the backups directory
func (b *SharedBuffer) Backup() error {
	if !b.Settings["backup"].(bool) || b.Path == "" || b.Type != BTDefault {
		return nil
	}

	_, err := b.writeBackup(b.AbsPath)
	return err
}

// RemoveBackup removes any backup file associated with this buffer
func (b *SharedBuffer) RemoveBackup() {
	if b.keepBackup() || b.Path == "" || b.Type != BTDefault {
		return
	}
	f := util.DetermineEscapePath(b.backupDir(), b.AbsPath)
	os.Remove(f)
}

// ApplyBackup applies the corresponding backup file to this buffer (if one exists)
// Returns true if a backup was applied
func (b *SharedBuffer) ApplyBackup(fsize int64) (bool, bool) {
	if b.Settings["backup"].(bool) && !b.Settings["permbackup"].(bool) && len(b.Path) > 0 && b.Type == BTDefault {
		backupfile := util.DetermineEscapePath(b.backupDir(), b.AbsPath)
		if info, err := os.Stat(backupfile); err == nil {
			backup, err := os.Open(backupfile)
			if err == nil {
				defer backup.Close()
				t := info.ModTime()
				msg := fmt.Sprintf(BackupMsg, b.Path, t.Format("Mon Jan _2 at 15:04, 2006"), backupfile)
				choice := screen.TermPrompt(msg, []string{"r", "i", "a", "recover", "ignore", "abort"}, true)

				if choice%3 == 0 {
					// recover
					b.LineArray = NewLineArray(uint64(fsize), FFAuto, backup)
					b.setModified()
					return true, true
				} else if choice%3 == 1 {
					// delete
					os.Remove(backupfile)
				} else if choice%3 == 2 {
					return false, false
				}
			}
		}
	}

	return false, true
}
