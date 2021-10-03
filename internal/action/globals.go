package action

import "github.com/zyedidia/micro/v2/internal/buffer"

// InfoBar is the global info bar.
var InfoBar *InfoPane

// LogBufPane is a global log buffer.
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
	}
}

// OpenLogBuf opens the log buffer from the current bufpane
// If the current bufpane is a log buffer nothing happens,
// otherwise the log buffer is opened in a horizontal split
func (h *BufPane) OpenLogBuf() {
	LogBufPane = h.HSplitBuf(buffer.LogBuf)
	LogBufPane.CursorEnd()
}
