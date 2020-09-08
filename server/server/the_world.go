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
	"log"
	"reflect"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jeanphorn/log4go"
	"github.com/spf13/viper"
)

//业务世界的抽象
type TheWorld struct {
	Roles            map[uint32]siface.IRole //所有角色
	RolesByName      map[string]siface.IRole
	MessageStructMap []interface{}            //协议处理结构体数组
	HandleMap        map[uint32]reflect.Value //协议处理方法集合
	Proto            *core.ServerProto        //解封包对象
	DB               *sql.DB                  //数据库对象
}

var theWorld *TheWorld

func GetTheWorld() *TheWorld {
	return theWorld
}

//协议处理对象添加
func (w *TheWorld) AddMsgStruct(ms interface{}) {
	w.MessageStructMap = append(w.MessageStructMap, ms)
	log4go.Debug("Add Msg Struct: ", reflect.ValueOf(ms).Type())
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
				log4go.Debug("Init protocol func :", t.Method(i).Name)
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
		log4go.Error("CallProtocolFunc err nil, msgID: ", id)
	}
}

//添加角色
func (w *TheWorld) AddRole(role siface.IRole) {
	w.Roles[role.GetUid()] = role
}

func (w *TheWorld) AddRoleByName(role siface.IRole) {
	w.RolesByName[role.GetName()] = role
}

func (w *TheWorld) LeaveUser(conn hiface.IConnection) {
	id := conn.GetConnID()

	if role, ok := w.Roles[id]; ok {
		delete(w.Roles, id)
		if role.IsStatus(siface.ONLINE) {
			delete(w.RolesByName, role.GetName())
		}
	}
}

//获取角色
func (w *TheWorld) GetRole(conn hiface.IConnection) (siface.IRole, error) {
	if role, ok := w.Roles[conn.GetConnID()]; ok {
		return role, nil
	}
	return nil, errors.New("Role nil, connID:" + strconv.Itoa(int(conn.GetConnID())))
}

func (w *TheWorld) GetRoleByName(name string) (siface.IRole, error) {
	if role, ok := w.RolesByName[name]; ok {
		return role, nil
	}
	return nil, errors.New("Role nil, name:" + name)
}

func (w *TheWorld) GetAllRoles() map[string]siface.IRole {
	return w.RolesByName
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
		//log4go.Debug("Recv msg:", pb.MSG_name[int32(msgID)])
		role, err := theWorld.GetRole(conn)
		if err != nil {
			role = CreatePlayer(conn, theWorld)
			log4go.Info("[TheWorld] Now role num:%d", len(theWorld.Roles))
		}

		theWorld.CallProtocolFunc(msgID, role, message)
	} else {
		theWorld.LeaveUser(conn)
	}
}

//广播发送
func (w *TheWorld) Broadcast(msg hiface.IMessage) {
	for _, role := range theWorld.Roles {
		role.GetConn().SendMessage(msg)
	}
}

func init() {
	//日志初始化
	log4go.AddFilter("file", log4go.DEBUG, log4go.NewFileLogWriter("server.log", true, true))

	theWorld = &TheWorld{
		Roles:       make(map[uint32]siface.IRole),
		RolesByName: make(map[string]siface.IRole),
		HandleMap:   make(map[uint32]reflect.Value),
		Proto:       core.CreateServerProto(),
	}

	//添加需要的协议处理结构体
	theWorld.AddMsgStruct(&msgwork.ChatMessage{})
	theWorld.AddMsgStruct(&msgwork.PlayerMessage{})
	//end

	theWorld.InitProtocol()
	//异步任务协程池启动
	hnet.AsyncPool.Start()

	//配置读取
	viper.SetConfigName("config")
	viper.AddConfigPath(".")    // 设置配置文件和可执行二进制文件在用一个目录
	err := viper.ReadInConfig() // 根据以上配置读取加载配置文件
	if err != nil {
		log4go.Error("viper.ReadInConfig err:", err)
		log.Fatal(err)
	}

	theWorld.DB = dbInit()
}

func dbInit() *sql.DB {
	//DB初始化
	dbConfig := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", viper.GetString(`mysql.username`), viper.GetString(`mysql.password`), viper.GetString(`mysql.host`), viper.GetString(`mysql.port`), viper.GetString(`mysql.database`), viper.GetString(`mysql.chatset`))
	log4go.Debug(dbConfig)
	db, err := sql.Open("mysql", dbConfig)
	if err != nil {
		panic("DB err:" + err.Error())
	}

	if err = db.Ping(); err != nil {
		panic("DB connect err:" + err.Error())
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	log4go.Debug("DB connected")

	return db
}
