package model

import (
	"hzhgagaga/server/core"
	"hzhgagaga/server/siface"
)

type Player struct {
	theWorld siface.ITheWorld
	Name     string
	Uid      uint32
}

func CreatePlayer(name string, uid uint32, world siface.ITheWorld) *Player {
	return &Player{
		theWorld: world,
		Name:     name,
		Uid:      uid,
	}
}

func (p *Player) GetName() string {
	return p.Name
}

func (p *Player) GetUid() uint32 {
	return p.Uid
}

func (p *Player) GetTheWorld() siface.ITheWorld {
	return p.theWorld
}

func (p *Player) SendMessage(msg *core.Message) {
	p.theWorld.Send(p.GetUid(), msg.Data)
}
