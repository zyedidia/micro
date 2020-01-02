// +build linux darwin dragonfly openbsd_amd64 freebsd

package action

import (
	shellquote "github.com/kballard/go-shellquote"
	"github.com/zyedidia/micro/internal/shell"
)

const TermEmuSupported = true

func RunTermEmulator(h *BufPane, input string, wait bool, getOutput bool, callback string, userargs []interface{}) error {
	args, err := shellquote.Split(input)
	if err != nil {
		return err
	}

	t := new(shell.Terminal)
	t.Start(args, getOutput, wait, callback, userargs)

	h.AddTab()
	id := MainTab().Panes[0].ID()

	v := h.GetView()
	MainTab().Panes[0] = NewTermPane(v.X, v.Y, v.Width, v.Height, t, id)
	MainTab().SetActive(0)

	return nil
}
