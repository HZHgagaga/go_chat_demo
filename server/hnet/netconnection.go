package hnet

import (
	"fmt"
	"hzhgagaga/hiface"
	"io"
	"net"
)

type Connection struct {
	Server   *Server
	ConnID   uint32
	Conn     *net.TCPConn
	proto    *Proto
	SendChan chan []byte
}

func NewConnection(uid uint32, conn *net.TCPConn, server *Server, pro *Proto) hiface.IConnection {
	c := &Connection{
		Server:   server,
		ConnID:   uid,
		Conn:     conn,
		proto:    pro,
		SendChan: make(chan []byte, 5000),
	}

	return c
}

func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

func (c *Connection) WriteLoop() {
	go func() {
		for {
			select {
			case msg := <-c.SendChan:
				c.Conn.Write(msg)
			}
		}
	}()
}

func (c *Connection) ReadLoop() {
	go func() {
		defer c.Stop()
		for {
			buf := make([]byte, c.proto.GetMsgHeadLen())
			if _, err := io.ReadFull(c.Conn, buf); err != nil {
				fmt.Println(c.ConnID, " read err: ", err)
				return
			}

			msg, err := c.proto.Decode(buf)
			if err != nil {
				return
			}

			if msg.GetLen() > 0 {
				dataBuf := make([]byte, msg.GetLen())
				if _, err := io.ReadFull(c.Conn, dataBuf); err != nil {
					fmt.Println(c.ConnID, "read err: ", err)
					return
				}
				msg.Data = dataBuf
			}
			c.Server.ThreadPool.AddTask(
				func() {
					handle, err := c.Server.GetMsgHandle()
					if err != nil {
						fmt.Println("Get MsgHandle err: ", err)
						return
					}

					handle(c, msg)
				},
			)
		}
	}()
}

func (c *Connection) Start() {
	c.ReadLoop()
	c.WriteLoop()
}

func (c *Connection) GetTCPConn() *net.TCPConn {
	return c.Conn
}

func (c *Connection) Stop() {
	c.Conn.Close()
}

func (c *Connection) SendMessage(msg []byte) {
	c.SendChan <- msg
}
