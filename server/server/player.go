package server

type Player struct {
	Name string
	Uid  uint32
}

func CreatePlayer(name string, uid uint32) *Player {
	return &Player{
		Name: name,
		Uid:  uid,
	}
}

func (p *Player) GetName() string {
	return p.Name
}

func (p *Player) GetUid() uint32 {
	return p.Uid
}

func (p *Player) SendMessage(msg *Message) {
	theWorld := GetTheWorld()
	theWorld.Send(p.GetUid(), msg.Data)
}
