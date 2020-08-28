package siface

import (
	"database/sql"
	"hzhgagaga/hiface"
	"hzhgagaga/hnet"
	"hzhgagaga/server/core"
)

type ITheWorld interface {
	Send(uid uint32, msg hiface.IMessage)
	AddRole(role IRole)
	GetAsyncPool() *hnet.AsyncThreadPool
	GetRole(hiface.IConnection) (IRole, error)
	GetRoleByName(string) (IRole, error)
	AddRoleByName(role IRole)
	GetAllRoles() map[string]IRole
	GetProto() *core.ServerProto
	GetDB() *sql.DB
	Broadcast(hiface.IMessage)
}
