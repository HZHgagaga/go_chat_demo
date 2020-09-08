package main

import (
	"bytes"
	"chat_client/pb"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sort"
	"strings"

	"github.com/jeanphorn/log4go"
	"google.golang.org/protobuf/proto"
)

var okChan chan bool = make(chan bool)
var exitChan chan bool = make(chan bool)

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

func Decode(buf []byte) (*Message, error) {
	bbuf := bytes.NewBuffer(buf)
	m := &Message{}
	if err := binary.Read(bbuf, binary.LittleEndian, &m.Id); err != nil {
		log4go.Debug("binary.Read Id err: ", err)
		return nil, err
	}

	if err := binary.Read(bbuf, binary.LittleEndian, &m.Len); err != nil {
		log4go.Debug("binary.Read Len err: ", err)
		return nil, err
	}

	return m, nil
}

func Encode(m *Message) []byte {
	bbuf := bytes.NewBuffer([]byte{})
	if err := binary.Write(bbuf, binary.LittleEndian, m.Id); err != nil {
		log4go.Debug("binary.Read id err: ", err)
	}

	if err := binary.Write(bbuf, binary.LittleEndian, m.Len); err != nil {
		log4go.Debug("binary.Read len err: ", err)
	}

	if err := binary.Write(bbuf, binary.LittleEndian, m.Data); err != nil {
		log4go.Debug("binary.Read data err: ", err)
	}

	return bbuf.Bytes()
}

type msgs []*pb.SMBroadcastChat

func (s msgs) Len() int           { return len(s) }
func (s msgs) Less(i, j int) bool { return s[i].GetId() < s[j].GetId() }
func (s msgs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func CreatePlayer(conn net.Conn) {
	buf := make([]byte, 256)
	log4go.Debug("Please input your name:")

	fmt.Scanln(&buf)
	msg := &Message{}
	msg.Id = uint32(pb.MSG_value["M_CMCreatePlayer"])
	cdata := &pb.CMCreatePlayer{}
	cdata.Name = string(buf)
	data, err := proto.Marshal(cdata)
	if err != nil {
		panic("proto.Marshal err:" + err.Error())

	}

	msg.Len = uint32(binary.Size(data))
	msg.Data = data
	sendMsg := Encode(msg)
	_, err = conn.Write(sendMsg)
	if err != nil {
		panic("Write err:" + err.Error())
	}
}

func GetHistoryChat(conn net.Conn) {
	msg := &Message{}
	msg.Id = uint32(pb.MSG_value["M_CMHistoryChat"])
	hdata := &pb.CMHistoryChat{}
	data, err := proto.Marshal(hdata)
	if err != nil {
		panic("proto.Marshal err:" + err.Error())

	}

	msg.Len = uint32(binary.Size(data))
	msg.Data = data
	sendMsg := Encode(msg)
	_, err = conn.Write(sendMsg)
	if err != nil {
		panic("Write err:" + err.Error())
	}
	okChan <- true
}

func ReadLoop(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 8)
		if _, err := io.ReadFull(conn, buf); err != nil {
			if err == io.EOF {
				log4go.Debug("服务器已关闭")
				return
			}
			panic("Recv err:" + err.Error())
		}

		msg, err := Decode(buf)
		if err != nil {
			log4go.Debug("Decode err: ", err)
			return
		}

		dataBuf := make([]byte, msg.GetLen())
		if _, err := io.ReadFull(conn, dataBuf); err != nil {
			if err == io.EOF {
				log4go.Debug("服务器已关闭")
				return
			}
			panic("io.ReadFull err:" + err.Error())
		}

		switch int32(msg.GetID()) {
		case pb.MSG_value["M_SMEnterWorld"]:
			log4go.Debug("Enter World")
			CreatePlayer(conn)
		case pb.MSG_value["M_SMCreatePlayer"]:
			data := &pb.SMCreatePlayer{}
			if err := proto.Unmarshal(dataBuf, data); err != nil {
				log4go.Debug("proto.Unmarshal err: ", err)
				CreatePlayer(conn)
				break
			}

			if data.GetOk() {
				log4go.Debug("CreatePlayer OK!")
				GetHistoryChat(conn)
			} else {
				log4go.Debug(data.GetMsg())
				CreatePlayer(conn)
			}
		case pb.MSG_value["M_SMBroadcastChat"]:
			chat := &pb.SMBroadcastChat{}
			if err := proto.Unmarshal(dataBuf, chat); err != nil {
				log4go.Debug("proto.Unmarshal err: ", err)
				break
			}
			log4go.Debug(chat.GetTime(), chat.GetName(), "say:", chat.GetChatdata())
		case pb.MSG_value["M_SMHistoryChat"]:
			chat := &pb.SMHistoryChat{}
			if err := proto.Unmarshal(dataBuf, chat); err != nil {
				log4go.Debug("proto.Unmarshal err: ", err)
				break
			}

			sort.Stable(msgs(chat.Msg))
			log4go.Debug("[======================history=======================]")
			for _, msg := range chat.GetMsg() {
				log4go.Debug(msg.GetTime(), msg.GetName(), "say:", msg.GetChatdata())
			}
			log4go.Debug("[========================end=========================]")
			log4go.Debug("输入 show@ 可查看所有在线的用户名")
			log4go.Debug("输入 private@[用户名]:[msg] 可进行私密聊天")
		case pb.MSG_value["M_SMAllPlayers"]:
			players := &pb.SMAllPlayers{}
			if err := proto.Unmarshal(dataBuf, players); err != nil {
				log4go.Debug("proto.Unmarshal err: ", err)
				break
			}

			log4go.Debug("------All player------")
			for _, name := range players.GetNames() {
				log4go.Debug(name)
			}
			log4go.Debug("---------end----------")
			log4go.Debug("总共:", len(players.GetNames()), "人")
			log4go.Debug("输入 show@ 可查看所有在线的用户名")
			log4go.Debug("输入 private@[用户名]:[msg] 可进行私密聊天")
		case pb.MSG_value["M_SMPrivateChat"]:
			chat := &pb.SMPrivateChat{}
			if err := proto.Unmarshal(dataBuf, chat); err != nil {
				log4go.Debug("proto.Unmarshal err: ", err)
				break
			}
			log4go.Debug("(私密聊天)", chat.GetTime(), chat.GetName(), "say:", chat.GetChatdata())
		}
	}
}

func WriteLoop(conn net.Conn) {
	for {
		buf := make([]byte, 256)
		fmt.Scanln(&buf)
		para := strings.Split(string(buf), "@")
		switch strings.ToLower(para[0]) {
		case "show":
			msg := &Message{}
			msg.Id = uint32(pb.MSG_value["M_CMAllPlayers"])
			data := &pb.CMAllPlayers{}
			req, err := proto.Marshal(data)
			if err != nil {
				log4go.Debug("proto.Marshal err: ", err)
				return
			}

			msg.Len = uint32(binary.Size(req))
			msg.Data = req
			sendMsg := Encode(msg)
			_, err = conn.Write(sendMsg)
			if err != nil {
				log4go.Debug("Write error", err)
			}
		case "private":
			msgs := strings.Split(para[1], ":")
			msg := &Message{}
			msg.Id = uint32(pb.MSG_value["M_CMPrivateChat"])
			chat := &pb.CMPrivateChat{}
			chat.Name = msgs[0]
			chat.Chat = msgs[1]
			data, err := proto.Marshal(chat)
			if err != nil {
				log4go.Debug("proto.Marshal err: ", err)
				return
			}

			msg.Len = uint32(binary.Size(data))
			msg.Data = data
			sendMsg := Encode(msg)
			_, err = conn.Write(sendMsg)
			if err != nil {
				log4go.Debug("Write error", err)
			}
			log4go.Debug("OK")
		default:
			msg := &Message{}
			msg.Id = uint32(pb.MSG_value["M_CMBroadcastChat"])
			chat := &pb.CMBroadcastChat{}
			chat.Name = My_name
			chat.Chatdata = string(buf)
			data, err := proto.Marshal(chat)
			if err != nil {
				log4go.Debug("proto.Marshal err: ", err)
				return
			}

			msg.Len = uint32(binary.Size(data))
			msg.Data = data
			sendMsg := Encode(msg)
			_, err = conn.Write(sendMsg)
			if err != nil {
				log4go.Debug("Write error", err)
			}
		}
	}
}

func EnterWorld(conn net.Conn) {
	msg := &Message{}
	msg.Id = uint32(pb.MSG_value["M_CMEnterWorld"])
	data := &pb.CMEnterWorld{}
	req, err := proto.Marshal(data)
	if err != nil {
		log4go.Debug("proto.Marshal err: ", err)
		return
	}

	msg.Len = uint32(binary.Size(req))
	msg.Data = req
	sendMsg := Encode(msg)
	_, err = conn.Write(sendMsg)
	if err != nil {
		log4go.Debug("Write error", err)
	}
}

func main() {
	log4go.Debug("Client start test...")
	conn, err := net.Dial("tcp4", "127.0.0.1:16666")
	if err != nil {
		log4go.Debug("Dial error: ", err)
		return
	} else {
		log4go.Debug("Connect server succeed")
	}

	go ReadLoop(conn)
	EnterWorld(conn)
	select {
	case <-okChan:
		go WriteLoop(conn)
	}
	select {
	case <-exitChan:
	}
}
