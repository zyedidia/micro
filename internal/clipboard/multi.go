package clipboard

import (
	"bytes"
)

// For storing multi cursor clipboard contents
type multiClipboard map[Register][]string

var multi multiClipboard

func (c multiClipboard) getAllText(r Register) string {
	content := c[r]
	if content == nil {
		return ""
	}

	buf := &bytes.Buffer{}
	for _, s := range content {
		buf.WriteString(s)
	}
	return buf.String()
}

func (c multiClipboard) getText(r Register, num int) string {
	content := c[r]
	if content == nil || len(content) <= num {
		return ""
	}

	return content[num]
}

// isValid checks if the text stored in this multi-clipboard is the same as the
// text stored in the system clipboard (provided as an argument), and therefore
// if it is safe to use the multi-clipboard for pasting instead of the system
// clipboard.
func (c multiClipboard) isValid(r Register, clipboard string, ncursors int) bool {
	content := c[r]
	if content == nil || len(content) != ncursors {
		return false
	}

	return clipboard == c.getAllText(r)
}

func (c multiClipboard) writeText(text string, r Register, num int, ncursors int) {
	content := c[r]
	if content == nil || len(content) != ncursors {
		content = make([]string, ncursors, ncursors)
		c[r] = content
	}

	if num >= ncursors {
		return
	}

	content[num] = text
}

func init() {
	multi = make(multiClipboard)
}
