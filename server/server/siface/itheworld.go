package siface

import "hzhgagaga/server/core"

type ITheWorld interface {
	Send(uid uint32, msg []byte)
	Broadcast(msg *core.Message)
}
