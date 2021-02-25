package action

import "github.com/zyedidia/micro/v2/internal/buffer"

var InfoBar *InfoPane
var LogBufPane *BufPane

// InitGlobals initializes the log buffer and the info bar
func InitGlobals() {
	InfoBar = NewInfoBar()
	buffer.LogBuf = buffer.NewBufferFromString("", "Log", buffer.BTLog)
}

// GetInfoBar returns the infobar pane
func GetInfoBar() *InfoPane {
	return InfoBar
}

// WriteLog writes a string to the log buffer
func WriteLog(s string) {
	buffer.WriteLog(s)
	if LogBufPane != nil {
		LogBufPane.CursorEnd()
		v := LogBufPane.GetView()
		endY := buffer.LogBuf.End().Y

		if endY > v.StartLine+v.Height {
			v.StartLine = buffer.LogBuf.End().Y - v.Height + 2
			LogBufPane.SetView(v)
		}
	}
}

// OpenLogBuf opens the log buffer from the current bufpane
// If the current bufpane is a log buffer nothing happens,
// otherwise the log buffer is opened in a horizontal split
func (h *BufPane) OpenLogBuf() {
	LogBufPane = h.HSplitBuf(buffer.LogBuf)
	LogBufPane.CursorEnd()

	v := LogBufPane.GetView()
	endY := buffer.LogBuf.End().Y

	if endY > v.StartLine+v.Height {
		v.StartLine = buffer.LogBuf.End().Y - v.Height + 2
		LogBufPane.SetView(v)
	}
}
