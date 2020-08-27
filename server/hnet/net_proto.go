package hnet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hzhgagaga/hiface"
)

//消息包的头
type MessageHead struct {
	Id  uint32
	Len uint32
}

//消息包的抽象
type Message struct {
	MessageHead
	Data []byte
}

func (m *Message) GetID() uint32 {
	return m.Id
}

func (m *Message) GetLen() uint32 {
	return m.Len
}

func (m *Message) GetData() []byte {
	return m.Data
}

func (m *Message) SetData(data []byte) {
	m.Data = data
}

type Proto struct {
}

func CreateProto() *Proto {
	return &Proto{}
}

func (p *Proto) GetMsgHeadLen() uint32 {
	return uint32(binary.Size(MessageHead{}))
}

func (p *Proto) Encode(m hiface.IMessage) []byte {
	bbuf := bytes.NewBuffer([]byte{})
	if err := binary.Write(bbuf, binary.LittleEndian, m.GetID()); err != nil {
		fmt.Println("binary.Read id err: ", err)
	}

	if err := binary.Write(bbuf, binary.LittleEndian, m.GetLen()); err != nil {
		fmt.Println("binary.Read len err: ", err)
	}

	if err := binary.Write(bbuf, binary.LittleEndian, m.GetData()); err != nil {
		fmt.Println("binary.Read data err: ", err)
	}

	return bbuf.Bytes()
}

func (p *Proto) Decode(buf []byte) (hiface.IMessage, error) {
	bbuf := bytes.NewBuffer(buf)
	m := &Message{}
	if err := binary.Read(bbuf, binary.LittleEndian, &m.Id); err != nil {
		fmt.Println("binary.Read Id err: ", err)
		return nil, err
	}

	if err := binary.Read(bbuf, binary.LittleEndian, &m.Len); err != nil {
		fmt.Println("binary.Read Len err: ", err)
		return nil, err
	}

	return m, nil
}
