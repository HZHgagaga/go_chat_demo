package siface

import (
	"database/sql"
	"hzhgagaga/hiface"
	"hzhgagaga/server/core"
)

type ITheWorld interface {
	AddRole(role IRole)
	GetRole(hiface.IConnection) (IRole, error)
	GetRoleByName(string) (IRole, error)
	AddRoleByName(role IRole)
	GetAllRoles() map[string]IRole
	GetProto() *core.ServerProto
	GetDB() *sql.DB
	Broadcast(hiface.IMessage)
}
