package server

import (
	"fmt"
	"hzhgagaga/hiface"
	"hzhgagaga/server/siface"
)

//玩家的抽象
type Player struct {
	theWorld siface.ITheWorld
	Name     string
	Uid      uint32
	Status   int8
}

func CreatePlayer(uid uint32, world siface.ITheWorld) *Player {
	return &Player{
		theWorld: world,
		Uid:      uid,
		Status:   -1,
	}
}

func (p *Player) GetName() string {
	return p.Name
}

func (p *Player) GetUid() uint32 {
	return p.Uid
}

func (p *Player) SetName(name string) {
	p.Name = name
}

func (p *Player) IsStatus(value int8) bool {
	fmt.Println(p.Status, value)
	return p.Status == value
}

func (p *Player) SetStatus(value int8) {
	p.Status = value
}

func (p *Player) GetTheWorld() siface.ITheWorld {
	return p.theWorld
}

//发送数据，将传递到网络层的发送协程
func (p *Player) SendMessage(msg hiface.IMessage) {
	p.theWorld.Send(p.GetUid(), msg)
}
