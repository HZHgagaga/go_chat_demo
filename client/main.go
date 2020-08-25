package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type MessageHead struct {
	Id  uint32
	Len uint32
}

type Message struct {
	MessageHead
	Data []byte
}

func ReadLoop(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 256)
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Read error: ", err)
			return
		}

		fmt.Println("Recv msg: ", string(buf))
	}
}

func Decode(m *Message) []byte {
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

func EnterChat(conn net.Conn) {
	fmt.Print("Please input your name: ")
	buf := make([]byte, 256)

	fmt.Scanln(&buf)
	msg := &Message{}
	msg.Id = uint32(0)
	msg.Len = uint32(binary.Size(buf))
	msg.Data = buf
	sendMsg := Decode(msg)
	_, err := conn.Write(sendMsg)
	if err != nil {
		fmt.Println("Write error", err)
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
	EnterChat(conn)
	buf := make([]byte, 256)
	for {
		fmt.Scanln(&buf)
		msg := &Message{}
		msg.Id = uint32(1)
		msg.Len = uint32(binary.Size(buf))
		msg.Data = buf
		sendMsg := Decode(msg)
		_, err := conn.Write(sendMsg)
		if err != nil {
			fmt.Println("Write error", err)
		}

		fmt.Println("Send msg: ", string(buf))
	}
}
