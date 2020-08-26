package siface

import "hzhgagaga/hiface"

type IRole interface {
	GetUid() uint32
	GetName() string
	SetName(string)
	GetTheWorld() ITheWorld
	SendMessage(msg hiface.IMessage)
}
