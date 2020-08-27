package core

import (
	"encoding/binary"
	"errors"
	"hzhgagaga/hiface"
	"hzhgagaga/server/pb"
)

//业务层消息抽象
type Message struct {
	ID   uint32
	Data []byte
}

func (m *Message) GetID() uint32 {
	return m.ID
}

func (m *Message) GetLen() uint32 {
	return uint32(binary.Size(m.Data))
}

func (m *Message) GetData() []byte {
	return m.Data
}

func (m *Message) SetData(data []byte) {
	m.Data = data
}

//含有消息名称和ID的map
type ServerProto struct {
	NametoIDMap map[string]uint32
}

func CreateServerProto() *ServerProto {
	return &ServerProto{
		NametoIDMap: make(map[string]uint32),
	}
}

func (p *ServerProto) InitProtocol() {
	for k, v := range pb.MSG_value {
		protocolName := k[2:]
		p.AddNametoIDMap(protocolName, uint32(v))
	}
}

func (p *ServerProto) AddNametoIDMap(name string, ID uint32) {
	p.NametoIDMap[name] = ID
}

//通过消息名称封包
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
