package core

import (
	"hzhgagaga/hiface"
)

type Message struct {
	Data []byte
}

func NewMessage(im hiface.IMessage) *Message {
	return &Message{
		//	Data: im.GetData(),
	}
}

//var Protocol = map[string]uint32{
//	"CreatePlayer":  1,
//	"BroadcastChat": 2,
//}
