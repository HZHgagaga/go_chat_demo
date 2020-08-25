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
	Port       int
	WorkThread *WorkThread
	MsgHandle  MsgHandle
}

func NewServer(name string, ip string, port int) hiface.IServer {
	s := &Server{
		Name:       name,
		IP:         ip,
		Port:       port,
		WorkThread: NewWorkThread(),
		MsgHandle:  nil,
	}

	return s
}

func (s *Server) Start() {
	fmt.Println(s.Name, " start...")
	addr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:16666")
	if err != nil {
		fmt.Println("ResolveTCPAddr err: ", err)
		return
	}

	fmt.Println("ResolveTCPAddr succeeded...")

	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println("ListenTCP err: ", err)
		return
	}

	fmt.Println("ListenTCP succeeded...")
	var connID uint32 = 1
	proto := CreateProto()
	s.WorkThread.Start()

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			fmt.Println("AcceptTCP err: ", err)
			return
		}
		fmt.Println("one user connect: ", conn.RemoteAddr().String())
		c := NewConnection(connID, conn, s, proto)
		c.Start()
		connID++
	}
}

func (s *Server) ServerInit(handle MsgHandle) error {
	if s.MsgHandle != nil {
		return errors.New("MsgHandle is registered")
	}

	s.MsgHandle = handle
	fmt.Println("MsgHandle is registered")
	return nil
}

func (s *Server) GetMsgHandle() (MsgHandle, error) {
	if s.MsgHandle == nil {
		return nil, errors.New("MsgHandle is unregistered")
	}

	return s.MsgHandle, nil
}
