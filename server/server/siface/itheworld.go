package siface

import (
	"hzhgagaga/hiface"
	"hzhgagaga/server/core"
)

type ITheWorld interface {
	Send(uid uint32, msg hiface.IMessage)
	AddRole(role IRole)
	GetProto() *core.ServerProto
	Broadcast(hiface.IMessage)
}
