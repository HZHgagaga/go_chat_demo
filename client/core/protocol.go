package core

import (
	"errors"
	"hzhgagaga/hiface"
)

type Message struct {
	ID   uint32
	Data []byte
}

func (m *Message) GetID() uint32 {
	return m.ID
}

func (m *Message) GetData() []byte {
	return m.Data
}

type ServerProto struct {
	NametoIDMap map[string]uint32
}

func CreateServerProto() *ServerProto {
	return &ServerProto{
		NametoIDMap: make(map[string]uint32),
	}
}

func (p *ServerProto) AddNametoIDMap(name string, ID uint32) {
	p.NametoIDMap[name] = ID
}

func (p *ServerProto) Encode(name string, msg []byte) (hiface.IMessage, error) {
	msgID, ok := p.NametoIDMap[name]
	if !ok {
		return nil, errors.New("Encode err name: " + name)
	}
	req := &Message{
		ID:   msgID,
		Data: msg,
	}
	return req, nil
}

func (p *ServerProto) Decode(msg hiface.IMessage) (ID uint32, req *Message, err error) {
	req = &Message{
		ID:   msg.GetID(),
		Data: msg.GetData(),
	}
	return req.GetID(), req, nil
}

//var Protocol = map[string]uint32{
//	"CreatePlayer":  1,
//	"BroadcastChat": 2,
//}
