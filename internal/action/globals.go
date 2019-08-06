package action

import "github.com/zyedidia/micro/internal/buffer"

var InfoBar *InfoPane
var LogBufPane *BufPane

func InitGlobals() {
	InfoBar = NewInfoBar()
	buffer.LogBuf = buffer.NewBufferFromString("", "Log", buffer.BTLog)
}

func GetInfoBar() *InfoPane {
	return InfoBar
}

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

func OpenLogBuf(h *BufPane) {
	LogBufPane = h.HSplitBuf(buffer.LogBuf)
	LogBufPane.CursorEnd()

	v := LogBufPane.GetView()
	endY := buffer.LogBuf.End().Y

	if endY > v.StartLine+v.Height {
		v.StartLine = buffer.LogBuf.End().Y - v.Height + 2
		LogBufPane.SetView(v)
	}
}
