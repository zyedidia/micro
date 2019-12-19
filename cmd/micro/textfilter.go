package main

import (
	"bytes"
	"os/exec"
	"strings"
)

// TextFilter command filters the selection through the command.
// Selection goes to the command input.
// On successfull run command output replaces the current selection.
func TextFilter(args []string) {
	if len(args) == 0 {
		messenger.Error("usage: textfilter arguments")
		return
	}
	v := CurView()
	sel := v.Cursor.GetSelection()
	if sel == "" {
		v.Cursor.SelectWord()
		sel = v.Cursor.GetSelection()
	}
	var bout, berr bytes.Buffer
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = strings.NewReader(sel)
	cmd.Stderr = &berr
	cmd.Stdout = &bout
	err := cmd.Run()
	if err != nil {
		messenger.Error(err.Error() + " " + berr.String())
		return
	}
	v.Cursor.DeleteSelection()
	v.Buf.Insert(v.Cursor.Loc, bout.String())
}
