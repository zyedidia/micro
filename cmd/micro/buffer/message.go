package buffer

import (
	"log"

	"github.com/zyedidia/micro/cmd/micro/config"
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
	Owner      int
}

func NewMessage(owner int, msg string, start, end Loc, kind MsgType) *Message {
	return &Message{
		Msg:   msg,
		Start: start,
		End:   end,
		Kind:  kind,
		Owner: owner,
	}
}

func NewMessageAtLine(owner int, msg string, line int, kind MsgType) *Message {
	start := Loc{-1, line}
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
			log.Println("Return error")
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

func (b *Buffer) ClearMessages(owner int) {
	for i := len(b.Messages) - 1; i >= 0; i-- {
		if b.Messages[i].Owner == owner {
			b.removeMsg(i)
		}
	}
}

func (b *Buffer) ClearAllMessages() {
	b.Messages = make([]*Message, 0)
}
