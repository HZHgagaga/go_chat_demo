package main

import (
	"bytes"
	"chat_client/pb"
	"encoding/binary"
	"io"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/jeanphorn/log4go"
	"google.golang.org/protobuf/proto"
)

const Robot_num = 10000

var Test_string = []string{
	"你好",
	"好的好的",
	"收到了",
	"哈哈哈哈哈哈啊哈哈哈哈哈哈哈哈哈哈",
	"不不不不不不不不不不不不不不不不不不不不不不不不不不不不不不不不不不不",
	"没",
	"一二三四五六七八九十",
}

type Robot struct {
	okChan   chan bool
	exitChan chan bool
	Conn     net.Conn
	Name     string
	Mutex    sync.Mutex
}

type RobotManager struct {
	Robots       map[string]*Robot
	Mutex        sync.Mutex
	RobotsOnline map[string]*Robot
	Max          int
}

func (m *RobotManager) AddRobotOnline(r *Robot) {
	m.Mutex.Lock()
	m.RobotsOnline[r.Name] = r
	m.Mutex.Unlock()
}

func (m *RobotManager) RemoveRobotOnline(r *Robot) {
	m.Mutex.Lock()
	delete(m.RobotsOnline, r.Name)
	m.Mutex.Unlock()
}

func (m *RobotManager) GetOnlineNum() int {
	return len(m.RobotsOnline)
}

func (m *RobotManager) AddRobot(r *Robot) {
	m.Mutex.Lock()
	m.Robots[r.Name] = r
	m.Mutex.Unlock()
}

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

func (r *Robot) CreatePlayer() {
	msg := &Message{}
	msg.Id = uint32(pb.MSG_value["M_CMCreatePlayer"])
	cdata := &pb.CMCreatePlayer{}
	cdata.Name = string(r.Name)
	data, err := proto.Marshal(cdata)
	if err != nil {
		RoManager.RemoveRobotOnline(r)
		panic("proto.Marshal err:" + err.Error())
		return
	}

	msg.Len = uint32(binary.Size(data))
	msg.Data = data
	sendMsg := Encode(msg)
	_, err = r.Conn.Write(sendMsg)
	if err != nil {
		RoManager.RemoveRobotOnline(r)
		panic("Write err:" + err.Error())
	}
}

func (r *Robot) GetHistoryChat() {
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
	_, err = r.Conn.Write(sendMsg)
	if err != nil {
		RoManager.RemoveRobotOnline(r)
		panic("Write err:" + err.Error())
	}
	go r.Write()
}

func (r *Robot) ReadLoop() {
	defer r.Conn.Close()
	for {
		buf := make([]byte, 8)
		if _, err := io.ReadFull(r.Conn, buf); err != nil {
			if err == io.EOF {
				RoManager.RemoveRobotOnline(r)
				panic("服务器已关闭")
			}
			RoManager.RemoveRobotOnline(r)
			panic("Recv err:" + err.Error())
		}

		msg, err := Decode(buf)
		if err != nil {
			panic("Decode err: " + err.Error())
		}

		dataBuf := make([]byte, msg.GetLen())
		if _, err := io.ReadFull(r.Conn, dataBuf); err != nil {
			if err == io.EOF {
				panic("服务器已关闭")
			}
			RoManager.RemoveRobotOnline(r)
			panic("io.ReadFull err:" + err.Error())
		}

		switch int32(msg.GetID()) {
		case pb.MSG_value["M_SMEnterWorld"]:
			log4go.Debug(r.Name, "Enter World")
			r.CreatePlayer()
		case pb.MSG_value["M_SMCreatePlayer"]:
			data := &pb.SMCreatePlayer{}
			if err := proto.Unmarshal(dataBuf, data); err != nil {
				log4go.Debug("proto.Unmarshal err: ", err)
				r.CreatePlayer()
				break
			}

			if data.GetOk() {
				log4go.Debug(r.Name, "CreatePlayer OK!")
				r.GetHistoryChat()
			} else {
				log4go.Debug(data.GetMsg())
				r.CreatePlayer()
			}
		case pb.MSG_value["M_SMBroadcastChat"]:
			chat := &pb.SMBroadcastChat{}
			if err := proto.Unmarshal(dataBuf, chat); err != nil {
				log4go.Debug("proto.Unmarshal err: ", err)
				break
			}
			//log4go.Debug(chat.GetTime(), chat.GetName(), "say:", chat.GetChatdata())
		case pb.MSG_value["M_SMHistoryChat"]:
			chat := &pb.SMHistoryChat{}
			if err := proto.Unmarshal(dataBuf, chat); err != nil {
				log4go.Debug("proto.Unmarshal err: ", err)
				break
			}

			sort.Stable(msgs(chat.Msg))
			//log4go.Debug("[======================history=======================]")
			//for _, msg := range chat.GetMsg() {
			//	log4go.Debug(msg.GetTime(), msg.GetName(), "say:", msg.GetChatdata())
			//}
			//log4go.Debug("[========================end=========================]")
			//log4go.Debug("输入 show@ 可查看所有在线的用户名")
			//log4go.Debug("输入 private@[用户名]:[msg] 可进行私密聊天")
		case pb.MSG_value["M_SMAllPlayers"]:
			players := &pb.SMAllPlayers{}
			if err := proto.Unmarshal(dataBuf, players); err != nil {
				log4go.Debug("proto.Unmarshal err: ", err)
				break
			}

			choice := rand.Intn(len(players.GetNames()))
			msg := &Message{}
			msg.Id = uint32(pb.MSG_value["M_CMPrivateChat"])
			chat := &pb.CMPrivateChat{}
			chat.Name = players.GetNames()[choice]
			chatIndex := rand.Intn(len(Test_string))
			chat.Chat = Test_string[chatIndex]
			data, err := proto.Marshal(chat)
			if err != nil {
				log4go.Debug("proto.Marshal err: ", err)
				return
			}

			msg.Len = uint32(binary.Size(data))
			msg.Data = data
			sendMsg := Encode(msg)
			r.Mutex.Lock()
			_, err = r.Conn.Write(sendMsg)
			r.Mutex.Unlock()
			if err != nil {
				log4go.Debug("Write error", err)
			}
		case pb.MSG_value["M_SMPrivateChat"]:
			chat := &pb.SMPrivateChat{}
			if err := proto.Unmarshal(dataBuf, chat); err != nil {
				log4go.Debug("proto.Unmarshal err: ", err)
				break
			}
			//	log4go.Debug("(私密聊天)", chat.GetTime(), chat.GetName(), "say:", chat.GetChatdata())
		}
	}
}

func (r *Robot) Write() {
	for i := 1; i <= 10; i++ {
		choice := rand.Intn(2)
		//choice := 1
		time.Sleep(1 * time.Second)
		switch choice {
		case 1:
			msg := &Message{}
			msg.Id = uint32(pb.MSG_value["M_CMAllPlayers"])
			data := &pb.CMAllPlayers{}
			req, err := proto.Marshal(data)
			if err != nil {
				RoManager.RemoveRobotOnline(r)
				log4go.Debug("proto.Marshal err: ", err)
				return
			}

			msg.Len = uint32(binary.Size(req))
			msg.Data = req
			sendMsg := Encode(msg)
			r.Mutex.Lock()
			_, err = r.Conn.Write(sendMsg)
			r.Mutex.Unlock()
			if err != nil {
				RoManager.RemoveRobotOnline(r)
				panic("Write error" + err.Error())
			}
		case 0:
			index := rand.Intn(len(Test_string))
			buf := Test_string[index]
			msg := &Message{}
			msg.Id = uint32(pb.MSG_value["M_CMBroadcastChat"])
			chat := &pb.CMBroadcastChat{}
			chat.Name = My_name
			chat.Chatdata = string(buf)
			data, err := proto.Marshal(chat)
			if err != nil {
				RoManager.RemoveRobotOnline(r)
				log4go.Debug("proto.Marshal err: ", err)
				return
			}

			msg.Len = uint32(binary.Size(data))
			msg.Data = data
			sendMsg := Encode(msg)
			r.Mutex.Lock()
			_, err = r.Conn.Write(sendMsg)
			r.Mutex.Unlock()
			if err != nil {
				RoManager.RemoveRobotOnline(r)
				panic("Write error" + err.Error())
			}
		}
	}

	select {}
	r.exitChan <- true
}

func (r *Robot) EnterWorld() {
	msg := &Message{}
	msg.Id = uint32(pb.MSG_value["M_CMEnterWorld"])
	data := &pb.CMEnterWorld{}
	req, err := proto.Marshal(data)
	if err != nil {
		RoManager.RemoveRobotOnline(r)
		log4go.Debug("proto.Marshal err: ", err)
		return
	}

	msg.Len = uint32(binary.Size(req))
	msg.Data = req
	sendMsg := Encode(msg)
	_, err = r.Conn.Write(sendMsg)
	if err != nil {
		RoManager.RemoveRobotOnline(r)
		panic("Write error" + err.Error())
	}
}

var RoManager *RobotManager

func RobotStart(robot *Robot) {
	RoManager.AddRobotOnline(robot)
	RoManager.Max++
	go robot.ReadLoop()
	robot.EnterWorld()
	select {
	case <-robot.exitChan:
		robot.Conn.Close()
	}
}

func main() {
	RoManager = &RobotManager{
		Robots:       make(map[string]*Robot),
		RobotsOnline: make(map[string]*Robot),
	}

	rand.Seed(time.Now().UnixNano())
	for i := 1; i <= Robot_num; i++ {
		name := "hzh_" + strconv.Itoa(i)
		log4go.Debug("Robot", name, "start")
		robot := &Robot{
			okChan:   make(chan bool, 1),
			exitChan: make(chan bool, 1),
			Name:     name,
		}
		RoManager.AddRobot(robot)
		conn, err := net.Dial("tcp4", "127.0.0.1:16666")
		if err != nil {
			log4go.Debug("Dial error: ", err)
			return
		} else {
			log4go.Debug("Connect server succeed")
		}
		robot.Conn = conn
		go RobotStart(robot)
	}

	for {
		time.Sleep(1 * time.Second)
		log4go.Debug("robot num:", Robot_num, " now online:", RoManager.GetOnlineNum(), " max num:", RoManager.Max)
	}
}
