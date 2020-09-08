package server

import (
	"errors"
	"hzhgagaga/hiface"
	"hzhgagaga/server/siface"
)

//玩家的抽象
type Player struct {
	Conn     hiface.IConnection
	theWorld siface.ITheWorld
	Name     string
	Uid      uint32
	Status   int8
}

func CreatePlayer(conn hiface.IConnection, world siface.ITheWorld) *Player {
	return &Player{
		Conn:     conn,
		theWorld: world,
		Uid:      conn.GetConnID(),
		Status:   -1,
	}
}

func (p *Player) GetConn() hiface.IConnection {
	return p.Conn
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
	return p.Status == value
}

func (p *Player) SetStatus(value int8) {
	p.Status = value
}

func (p *Player) GetTheWorld() siface.ITheWorld {
	return p.theWorld
}

//发送数据，将传递到网络层的发送协程
func (p *Player) SendMessage(msg hiface.IMessage) error {
	if p.Conn.IsClose() {
		return errors.New("Conn closed")
	}

	p.Conn.SendMessage(msg)
	return nil
}
