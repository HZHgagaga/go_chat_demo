package hnet

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type MessageHead struct {
	Id  uint32
	Len uint32
}

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

type Proto struct {
}

func CreateProto() *Proto {
	return &Proto{}
}

func (p *Proto) GetMsgHeadLen() uint32 {
	return uint32(binary.Size(MessageHead{}))
}

func (p *Proto) Encode(m *Message) []byte {
	data := bytes.NewBuffer([]byte{})
	return data.Bytes()
}

func (p *Proto) Decode(buf []byte) (*Message, error) {
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
