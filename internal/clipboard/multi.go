package clipboard

import (
	"bytes"
	"hash/fnv"
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
		buf.WriteByte('\n')
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

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// isValid checks if the text stored in this multi-clipboard is the same as the
// text stored in the system clipboard (provided as an argument), and therefore
// if it is safe to use the multi-clipboard for pasting instead of the system
// clipboard.
func (c multiClipboard) isValid(r Register, ncursors int, clipboard string) bool {
	content := c[r]
	if content == nil || len(content) != ncursors {
		return false
	}

	return hash(clipboard) == hash(c.getAllText(r))
}

func (c multiClipboard) writeText(text string, r Register, num int) {
	content := c[r]
	if content == nil || num >= cap(content) {
		content = make([]string, num+1, num+1)
	}

	content[num] = text
}

func init() {
	multi = make(multiClipboard)
}
