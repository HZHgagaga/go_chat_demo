package hiface

import "net"

type IConnection interface {
	GetConnID() uint32
	WriteLoop()
	ReadLoop()
	Start()
	GetTCPConn() *net.TCPConn
	Stop()
	SendMessage(msg []byte)
}
