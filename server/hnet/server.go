package hnet

import (
	"errors"
	"hzhgagaga/hiface"
	"net"

	"github.com/jeanphorn/log4go"
)

type MsgHandle = func(hiface.IConnection, hiface.IMessage)

//服务器的抽象
type Server struct {
	name      string
	iP        string
	port      string
	msgHandle MsgHandle
}

func NewServer(name string, ip string, port string) hiface.IServer {
	s := &Server{
		name: name,
		iP:   ip,
		port: port,
	}

	return s
}

//一系列socket函数调用
func (s *Server) Start() {
	//log4go.AddFilter("stdout", log4go.DEBUG, log4go.NewConsoleLogWriter())
	//log4go.AddFilter("file", log4go.INFO, log4go.NewFileLogWriter("server.log", true, true))

	log4go.Debug(s.name + " start " + s.iP + ":" + s.port)
	addr, err := net.ResolveTCPAddr("tcp4", s.iP+":"+s.port)
	if err != nil {
		panic("ResolveTCPAddr err:" + err.Error())
	}

	log4go.Debug("ResolveTCPAddr succeeded...")

	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic("ListenTCP err:" + err.Error())
	}

	log4go.Debug("ListenTCP succeeded...")
	var connID uint32 = 1
	proto := CreateProto()
	WorkPool.Start()

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			panic("AcceptTCP err:" + err.Error())
		}
		log4go.Debug("one user connect: ", conn.RemoteAddr().String())
		//一个客户端创建一个Connection
		c := NewConnection(connID, conn, s, proto)
		c.Start()
		connID++
	}
}

func (s *Server) ServerInit(handle MsgHandle) {
	if s.msgHandle != nil {
		panic("MsgHandle is registered")
	}

	s.msgHandle = handle
	log4go.Debug("MsgHandle is registered")
}

//获取消息处理函数
func (s *Server) GetMsgHandle() (MsgHandle, error) {
	if s.msgHandle == nil {
		return nil, errors.New("MsgHandle is unregistered")
	}

	return s.msgHandle, nil
}
