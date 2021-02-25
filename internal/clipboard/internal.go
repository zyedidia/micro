package clipboard

type internalClipboard map[Register]string

var internal internalClipboard

func init() {
	internal = make(internalClipboard)
}

func (c internalClipboard) read(r Register) string {
	return c[r]
}

func (c internalClipboard) write(text string, r Register) {
	c[r] = text
}
