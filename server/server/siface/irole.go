package siface

import "hzhgagaga/hiface"

const (
	ENTER = iota
	ONLINE
)

type IRole interface {
	GetConn() hiface.IConnection
	GetUid() uint32
	GetName() string
	SetName(string)
	GetTheWorld() ITheWorld
	SetStatus(int8)
	IsStatus(int8) bool
	SendMessage(msg hiface.IMessage) error
}
