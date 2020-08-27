package server

import (
	"database/sql"
	"errors"
	"fmt"
	"hzhgagaga/hiface"
	"hzhgagaga/hnet"
	"hzhgagaga/server/core"
	"hzhgagaga/server/msgwork"
	"hzhgagaga/server/pb"
	"hzhgagaga/server/siface"
	"reflect"
	"runtime"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

const (
	DB_USER_NAME = "hzh"
	DB_PASS_WORD = "hzh"
	DB_HOST      = "172.16.29.167"
	DB_PORT      = "3306"
	DB_DATABASE  = "chat_server"
	DB_CHARSET   = "utf8"
)

type TheWorld struct {
	Roles            map[uint32]siface.IRole
	UsersConns       map[uint32]hiface.IConnection
	MessageStructMap []interface{}
	HandleMap        map[uint32]reflect.Value
	AsyncPool        *hnet.AsyncThreadPool
	Proto            *core.ServerProto
	DB               *sql.DB
}

var theWorld *TheWorld

func GetTheWorld() *TheWorld {
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

func (w *TheWorld) AddRole(role siface.IRole) {
	w.Roles[role.GetUid()] = role
}

func (w *TheWorld) GetRole(conn hiface.IConnection) (siface.IRole, error) {
	if role, ok := w.Roles[conn.GetConnID()]; ok {
		return role, nil
	}
	return nil, errors.New("Role nil, connID:" + strconv.Itoa(int(conn.GetConnID())))
}

func (w *TheWorld) GetAsyncPool() *hnet.AsyncThreadPool {
	return w.AsyncPool
}

func (w *TheWorld) GetProto() *core.ServerProto {
	return w.Proto
}

func (w *TheWorld) GetDB() *sql.DB {
	return w.DB
}

func MsgHandle(conn hiface.IConnection, msg hiface.IMessage) {
	theWorld := GetTheWorld()
	msgID, message, err := theWorld.Proto.Decode(msg)
	role, err := theWorld.GetRole(conn)
	if err != nil {
		theWorld.UsersConns[conn.GetConnID()] = conn
		role = CreatePlayer(conn.GetConnID(), theWorld)
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
	theWorld = &TheWorld{
		Roles:      make(map[uint32]siface.IRole),
		UsersConns: make(map[uint32]hiface.IConnection),
		HandleMap:  make(map[uint32]reflect.Value),
		AsyncPool:  hnet.NewAsyncThreadPool(runtime.NumCPU()),
		Proto:      core.CreateServerProto(),
	}

	theWorld.AddMsgStruct(&msgwork.ChatMessage{})
	theWorld.AddMsgStruct(&msgwork.PlayerMessage{})

	theWorld.InitProtocol()
	theWorld.AsyncPool.Start()

	dbConfig := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", DB_USER_NAME, DB_PASS_WORD, DB_HOST, DB_PORT, DB_DATABASE, DB_CHARSET)
	//	dbConfig := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", dbConf.User, dbConf.Pwd, dbConf.Host, dbConf.Port, dbConf.Db, dbConf.Char)
	db, err := sql.Open("mysql", dbConfig)
	if err != nil {
		panic("DB err:" + err.Error())
	}

	if err = db.Ping(); err != nil {
		panic("DB connect err:" + err.Error())
	}

	fmt.Println("DB connected")
	theWorld.DB = db
}
