package buffer

import (
	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/tcell"
)

type MsgType int

const (
	MTInfo = iota
	MTWarning
	MTError
)

type Message struct {
	Msg        string
	Start, End Loc
	Kind       MsgType
	Owner      string
}

func NewMessage(owner string, msg string, start, end Loc, kind MsgType) *Message {
	return &Message{
		Msg:   msg,
		Start: start,
		End:   end,
		Kind:  kind,
		Owner: owner,
	}
}

func NewMessageAtLine(owner string, msg string, line int, kind MsgType) *Message {
	start := Loc{-1, line - 1}
	end := start
	return NewMessage(owner, msg, start, end, kind)
}

func (m *Message) Style() tcell.Style {
	switch m.Kind {
	case MTInfo:
		if style, ok := config.Colorscheme["gutter-info"]; ok {
			return style
		}
	case MTWarning:
		if style, ok := config.Colorscheme["gutter-warning"]; ok {
			return style
		}
	case MTError:
		if style, ok := config.Colorscheme["gutter-error"]; ok {
			return style
		}
	}
	return config.DefStyle
}

func (b *Buffer) AddMessage(m *Message) {
	b.Messages = append(b.Messages, m)
}

func (b *Buffer) removeMsg(i int) {
	copy(b.Messages[i:], b.Messages[i+1:])
	b.Messages[len(b.Messages)-1] = nil
	b.Messages = b.Messages[:len(b.Messages)-1]
}

func (b *Buffer) ClearMessages(owner string) {
	for i := len(b.Messages) - 1; i >= 0; i-- {
		if b.Messages[i].Owner == owner {
			b.removeMsg(i)
		}
	}
}

func (b *Buffer) ClearAllMessages() {
	b.Messages = make([]*Message, 0)
}
