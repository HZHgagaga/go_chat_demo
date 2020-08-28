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
	DB_HOST      = "172.16.30.189"
	DB_PORT      = "3306"
	DB_DATABASE  = "chat_server"
	DB_CHARSET   = "utf8"
)

//业务世界的抽象
type TheWorld struct {
	Roles            map[uint32]siface.IRole       //所有角色
	UsersConns       map[uint32]hiface.IConnection //该世界的所有连接
	MessageStructMap []interface{}                 //协议处理结构体数组
	HandleMap        map[uint32]reflect.Value      //协议处理方法集合
	AsyncPool        *hnet.AsyncThreadPool         //异步协程任务池
	Proto            *core.ServerProto             //解封包对象
	DB               *sql.DB                       //数据库对象
}

var theWorld *TheWorld

func GetTheWorld() *TheWorld {
	return theWorld
}

//协议处理对象添加
func (w *TheWorld) AddMsgStruct(ms interface{}) {
	w.MessageStructMap = append(w.MessageStructMap, ms)
	fmt.Println("Add Msg Struct: ", reflect.ValueOf(ms).Type())
}

//调用HandleMap中存储的方法时用到
func getValues(param ...interface{}) []reflect.Value {
	vals := make([]reflect.Value, 0, len(param))
	for i := range param {
		vals = append(vals, reflect.ValueOf(param[i]))
	}
	return vals
}

//HandleMap初始化，通过反射将需要的协议处理对象的方法存入其中
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

	//初始化ProtoBuf枚举的协议
	w.Proto.InitProtocol()
}

//调用HandleMap中存储的协议方法
func (w *TheWorld) CallProtocolFunc(id uint32, role siface.IRole, msg *core.Message) {
	if handle, ok := w.HandleMap[id]; ok {
		handle.Call(getValues(role, msg))
	} else {
		fmt.Println("CallProtocolFunc err nil, msgID: ", id)
	}
}

//添加角色
func (w *TheWorld) AddRole(role siface.IRole) {
	w.Roles[role.GetUid()] = role
}

func (w *TheWorld) LeaveUser(conn hiface.IConnection) {
	id := conn.GetConnID()
	if _, ok := w.UsersConns[id]; ok {
		delete(w.UsersConns, id)
	}

	if role, ok := w.Roles[id]; ok {
		delete(w.Roles, id)
		fmt.Println("Role:", role.GetName(), "leave the world")
	}

}

//获取角色
func (w *TheWorld) GetRole(conn hiface.IConnection) (siface.IRole, error) {
	if role, ok := w.Roles[conn.GetConnID()]; ok {
		return role, nil
	}
	return nil, errors.New("Role nil, connID:" + strconv.Itoa(int(conn.GetConnID())))
}

//获取异步池对象
func (w *TheWorld) GetAsyncPool() *hnet.AsyncThreadPool {
	return w.AsyncPool
}

//获取解封包对象
func (w *TheWorld) GetProto() *core.ServerProto {
	return w.Proto
}

//获取数据库对象
func (w *TheWorld) GetDB() *sql.DB {
	return w.DB
}

//业务层的入口，网络层传来的消息处理函数
func MsgHandle(conn hiface.IConnection, msg hiface.IMessage) {
	theWorld := GetTheWorld()
	if !conn.IsClose() {
		msgID, message, err := theWorld.Proto.Decode(msg)
		role, err := theWorld.GetRole(conn)
		if err != nil {
			theWorld.UsersConns[conn.GetConnID()] = conn
			role = CreatePlayer(conn.GetConnID(), theWorld)
		}

		theWorld.CallProtocolFunc(msgID, role, message)
	} else {
		theWorld.LeaveUser(conn)
	}
}

//将消息发送至网络层的发送协程
func (w *TheWorld) Send(uid uint32, msg hiface.IMessage) {
	conn, ok := w.UsersConns[uid]
	if !ok {
		fmt.Println("SendMessage err: player's connection is nil")
		return
	}
	conn.SendMessage(msg)
}

//广播发送
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

	//添加需要的协议处理结构体
	theWorld.AddMsgStruct(&msgwork.ChatMessage{})
	theWorld.AddMsgStruct(&msgwork.PlayerMessage{})
	//end

	theWorld.InitProtocol()
	//异步任务协程池启动
	theWorld.AsyncPool.Start()

	//DB初始化
	dbConfig := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", DB_USER_NAME, DB_PASS_WORD, DB_HOST, DB_PORT, DB_DATABASE, DB_CHARSET)
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
