package main

import (
	"bytes"
	"chat_client/pb"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"google.golang.org/protobuf/proto"
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

func (m *Message) SetData(data []byte) {
	m.Data = data
}

var My_name string
var EnterChan chan bool = make(chan bool)

func Decode(buf []byte) (*Message, error) {
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

func Encode(m *Message) []byte {
	bbuf := bytes.NewBuffer([]byte{})
	if err := binary.Write(bbuf, binary.LittleEndian, m.Id); err != nil {
		fmt.Println("binary.Read id err: ", err)
	}

	if err := binary.Write(bbuf, binary.LittleEndian, m.Len); err != nil {
		fmt.Println("binary.Read len err: ", err)
	}

	if err := binary.Write(bbuf, binary.LittleEndian, m.Data); err != nil {
		fmt.Println("binary.Read data err: ", err)
	}

	return bbuf.Bytes()
}

func ReadLoop(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 8)
		if _, err := io.ReadFull(conn, buf); err != nil {
			fmt.Println(conn, " read err: ", err)
			return
		}

		msg, err := Decode(buf)
		if err != nil {
			fmt.Println("Decode err: ", err)
			return
		}

		dataBuf := make([]byte, msg.GetLen())
		if _, err := io.ReadFull(conn, dataBuf); err != nil {
			fmt.Println("io.ReadFull err: ", err)
			return
		}

		switch int32(msg.GetID()) {
		case pb.MSG_value["M_SMCreatePlayer"]:
			fmt.Println("CreatePlayer OK!")
			EnterChan <- true
		case pb.MSG_value["M_SMBroadcastChat"]:
			chat := &pb.SMBroadcastChat{}
			if err := proto.Unmarshal(dataBuf, chat); err != nil {
				fmt.Println("proto.Unmarshal err: ", err)
				break
			}
			fmt.Println(chat.GetTime(), chat.GetName(), "say:", chat.GetChatdata())
		}
	}
}

func WriteLoop(conn net.Conn) {
	for {
		buf := make([]byte, 256)
		fmt.Scanln(&buf)
		msg := &Message{}
		msg.Id = uint32(pb.MSG_value["M_CMBroadcastChat"])
		chat := &pb.CSBroadcastChat{}
		chat.Name = My_name
		chat.Chatdata = string(buf)
		data, err := proto.Marshal(chat)
		if err != nil {
			fmt.Println("proto.Marshal err: ", err)
			return
		}

		msg.Len = uint32(binary.Size(data))
		msg.Data = data
		sendMsg := Encode(msg)
		_, err = conn.Write(sendMsg)
		if err != nil {
			fmt.Println("Write error", err)
		}
	}
}

func main() {
	fmt.Println("Client start test...")
	conn, err := net.Dial("tcp4", "127.0.0.1:16666")
	if err != nil {
		fmt.Println("Dial error: ", err)
		return
	} else {
		fmt.Println("Connect server succeed")
	}

	defer conn.Close()
	go ReadLoop(conn)

	buf := make([]byte, 256)
	fmt.Print("Please input your name:")

	fmt.Scanln(&buf)
	msg := &Message{}
	msg.Id = uint32(pb.MSG_value["M_CMCreatePlayer"])
	cdata := &pb.CMCreatePlayer{}
	cdata.Name = string(buf)
	data, err := proto.Marshal(cdata)
	if err != nil {
		fmt.Println("proto.Marshal err: ", err)
		return
	}

	msg.Len = uint32(binary.Size(data))
	msg.Data = data
	sendMsg := Encode(msg)
	_, err = conn.Write(sendMsg)
	if err != nil {
		fmt.Println("Write error", err)
		return
	}
	select {
	case <-EnterChan:
	}

	go WriteLoop(conn)
	select {}
}
