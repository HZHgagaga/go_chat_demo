package hiface

import "net"

type IConnection interface {
	GetConnID() uint32
	WriteLoop()
	ReadLoop()
	Start()
	GetTCPConn() *net.TCPConn
	IsClose() bool
	Stop()
	SendMessage(msg IMessage)
}
