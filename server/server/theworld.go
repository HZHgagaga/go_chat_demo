package server

import (
	"errors"
	"fmt"
	"hzhgagaga/hiface"
	"hzhgagaga/server/pb"
	"reflect"
	"strconv"
	"sync"
)

type TheWorld struct {
	Users            map[uint32]*Player
	UsersConns       map[uint32]hiface.IConnection
	MessageStructMap []interface{}
	HandleMap        map[uint32]reflect.Value
}

var theWorld *TheWorld
var once sync.Once

func GetTheWorld() *TheWorld {
	once.Do(func() {
		theWorld = &TheWorld{
			Users:      make(map[uint32]*Player),
			UsersConns: make(map[uint32]hiface.IConnection),
			HandleMap:  make(map[uint32]reflect.Value),
		}

		theWorld.AddMsgStruct(&ChatMessage{})
		theWorld.AddMsgStruct(&PlayerMessage{})
		theWorld.InitProtocol()
	})

	return theWorld
}

func (w *TheWorld) AddMsgStruct(ms interface{}) {
	w.MessageStructMap = append(w.MessageStructMap, ms)
	fmt.Println("Add Msg Struct: ", reflect.ValueOf(ms).Type())
}

func getValues(param ...interface{}) []reflect.Value {
	vals := make([]reflect.Value, 0, len(param))
	for i := range param {
		vals = append(vals, reflect.ValueOf(param[i]))
	}
	return vals
}

func (w *TheWorld) InitProtocol() {
	for _, ms := range w.MessageStructMap {
		v := reflect.ValueOf(ms)
		t := reflect.TypeOf(ms)
		for i := 0; i < v.NumMethod(); i++ {
			protocolName := t.Method(i).Name[2:]
			protocolName = "M_" + protocolName
			if num, ok := pb.MSG_value[protocolName]; ok {
				w.HandleMap[uint32(num)] = v.Method(i)
				fmt.Println("Init protocol func :", t.Method(i).Name)
				//plr := CreatePlayer("test", 1)
				//msg := NewMessage(nil)
				//w.HandleMap[uint32(num)].Call(getValues(plr, msg))
			}
		}
	}
}

func (w *TheWorld) CallProtocolFunc(id uint32, plr *Player, msg *Message) {
	if handle, ok := w.HandleMap[id]; ok {
		handle.Call(getValues(plr, msg))
	} else {
		fmt.Println("CallProtocolFunc err nil, msgID: ", id)
	}
}

func (w *TheWorld) CreateAndAddPlayer(conn hiface.IConnection, msg *Message) {
	newPlayer := CreatePlayer(string(msg.Data), conn.GetConnID())
	fmt.Println("----------TheWorld---------AddPlayer")
	w.Users[newPlayer.GetUid()] = newPlayer
	w.UsersConns[newPlayer.GetUid()] = conn
}

func (w *TheWorld) GetPlayer(conn hiface.IConnection) (*Player, error) {
	if plr, ok := w.Users[conn.GetConnID()]; ok {
		return plr, nil
	}
	return nil, errors.New("Player nil, connID:" + strconv.Itoa(int(conn.GetConnID())))
}

func MsgHandle(conn hiface.IConnection, msg hiface.IMessage) {
	theWorld := GetTheWorld()
	msgID := msg.GetID()
	message := &Message{
		Data: msg.GetData(),
	}
	plr, err := theWorld.GetPlayer(conn)
	if err != nil {
		//	fmt.Println("GetPlayer nil: ", err)
		//	return
		theWorld.CreateAndAddPlayer(conn, message)
	}

	theWorld.CallProtocolFunc(msgID, plr, message)
}

func (w *TheWorld) Send(uid uint32, data []byte) {
	conn, ok := w.UsersConns[uid]
	if !ok {
		fmt.Println("SendMessage err: player's connection is nil")
		return
	}
	conn.SendMessage(data)
}

func (w *TheWorld) Broadcast(msg *Message) {
	for _, player := range theWorld.Users {
		fmt.Println(msg.Data)
		w.Send(player.GetUid(), msg.Data)
	}
}

func init() {
	_ = GetTheWorld()
}
