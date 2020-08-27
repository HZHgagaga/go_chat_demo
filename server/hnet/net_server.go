package hnet

import (
	"errors"
	"fmt"
	"hzhgagaga/hiface"
	"net"
)

type MsgHandle = func(hiface.IConnection, hiface.IMessage)

type Server struct {
	Name       string
	IP         string
	Port       string
	WorkThread *WorkThread
	MsgHandle  MsgHandle
}

func NewServer(name string, ip string, port string) hiface.IServer {
	s := &Server{
		Name:       name,
		IP:         ip,
		Port:       port,
		WorkThread: NewWorkThread(),
	}

	return s
}

func (s *Server) Start() {
	fmt.Println(s.Name, "start", s.IP+":"+s.Port)
	addr, err := net.ResolveTCPAddr("tcp4", s.IP+":"+s.Port)
	if err != nil {
		panic("ResolveTCPAddr err:" + err.Error())
	}

	fmt.Println("ResolveTCPAddr succeeded...")

	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic("ListenTCP err:" + err.Error())
	}

	fmt.Println("ListenTCP succeeded...")
	var connID uint32 = 1
	proto := CreateProto()
	s.WorkThread.Start()

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			panic("AcceptTCP err:" + err.Error())
		}
		fmt.Println("one user connect: ", conn.RemoteAddr().String())
		c := NewConnection(connID, conn, s, proto)
		c.Start()
		connID++
	}
}

func (s *Server) ServerInit(handle MsgHandle) {
	if s.MsgHandle != nil {
		panic("MsgHandle is registered")
	}

	s.MsgHandle = handle
	fmt.Println("MsgHandle is registered")
}

func (s *Server) GetMsgHandle() (MsgHandle, error) {
	if s.MsgHandle == nil {
		return nil, errors.New("MsgHandle is unregistered")
	}

	return s.MsgHandle, nil
}
