package server

import (
	"errors"
	"fmt"
	"hzhgagaga/hiface"
	"hzhgagaga/hnet"
	"hzhgagaga/server/core"
	"hzhgagaga/server/model"
	"hzhgagaga/server/msgwork"
	"hzhgagaga/server/pb"
	"hzhgagaga/server/siface"
	"reflect"
	"runtime"
	"strconv"
	"sync"
)

type TheWorld struct {
	Roles            map[uint32]siface.IRole
	UsersConns       map[uint32]hiface.IConnection
	MessageStructMap []interface{}
	HandleMap        map[uint32]reflect.Value
	SyncPool         *hnet.AsyncThreadPool
	Proto            *core.ServerProto
}

var theWorld *TheWorld
var once sync.Once

func GetTheWorld() *TheWorld {
	once.Do(func() {
		theWorld = &TheWorld{
			Roles:      make(map[uint32]siface.IRole),
			UsersConns: make(map[uint32]hiface.IConnection),
			HandleMap:  make(map[uint32]reflect.Value),
			SyncPool:   hnet.NewAsyncThreadPool(runtime.NumCPU()),
			Proto:      core.CreateServerProto(),
		}

		theWorld.AddMsgStruct(&msgwork.ChatMessage{})
		theWorld.AddMsgStruct(&msgwork.PlayerMessage{})

		theWorld.InitProtocol()
		theWorld.SyncPool.Start()
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
			if ID, ok := pb.MSG_value[protocolName]; ok {
				w.HandleMap[uint32(ID)] = v.Method(i)
				fmt.Println("Init protocol func :", t.Method(i).Name)
			}
		}
	}

	w.Proto.InitProtocol()
}

func (w *TheWorld) CallProtocolFunc(id uint32, role siface.IRole, msg *core.Message) {
	if handle, ok := w.HandleMap[id]; ok {
		handle.Call(getValues(role, msg))
	} else {
		fmt.Println("CallProtocolFunc err nil, msgID: ", id)
	}
}

//func (w *TheWorld) CreateAndAddPlayer(conn hiface.IConnection, msg *core.Message) {
//	theWorld := GetTheWorld()
//	newPlayer := model.CreatePlayer(conn.GetConnID(), theWorld)
//	fmt.Println("----------TheWorld---------AddPlayer")
//	w.Users[newPlayer.GetUid()] = newPlayer
//	w.UsersConns[newPlayer.GetUid()] = conn
//}

func (w *TheWorld) AddRole(role siface.IRole) {
	w.Roles[role.GetUid()] = role
}

func (w *TheWorld) GetRole(conn hiface.IConnection) (siface.IRole, error) {
	if role, ok := w.Roles[conn.GetConnID()]; ok {
		return role, nil
	}
	return nil, errors.New("Role nil, connID:" + strconv.Itoa(int(conn.GetConnID())))
}

func (w *TheWorld) GetProto() *core.ServerProto {
	return w.Proto
}

func MsgHandle(conn hiface.IConnection, msg hiface.IMessage) {
	theWorld := GetTheWorld()
	msgID, message, err := theWorld.Proto.Decode(msg)
	role, err := theWorld.GetRole(conn)
	if err != nil {
		theWorld.UsersConns[conn.GetConnID()] = conn
		role = model.CreatePlayer(conn.GetConnID(), theWorld)
		//theWorld.CreateAndAddPlayer(conn, message)
	}

	theWorld.CallProtocolFunc(msgID, role, message)
}

func (w *TheWorld) Send(uid uint32, msg hiface.IMessage) {
	conn, ok := w.UsersConns[uid]
	if !ok {
		fmt.Println("SendMessage err: player's connection is nil")
		return
	}
	conn.SendMessage(msg)
}

func (w *TheWorld) Broadcast(msg hiface.IMessage) {
	fmt.Println("----Broadcast----")
	for _, role := range theWorld.Roles {
		w.Send(role.GetUid(), msg)
	}
}

func init() {
	_ = GetTheWorld()
}
