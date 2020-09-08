package hnet

import (
	"hzhgagaga/hiface"
	"io"
	"net"

	"github.com/jeanphorn/log4go"
)

//客户端连接的抽象
type Connection struct {
	server   *Server
	connID   uint32
	conn     *net.TCPConn
	proto    *Proto
	sendChan chan hiface.IMessage
	exitChan chan bool
	close    bool
}

func NewConnection(uid uint32, conn *net.TCPConn, server *Server, pro *Proto) hiface.IConnection {
	c := &Connection{
		server:   server,
		connID:   uid,
		conn:     conn,
		proto:    pro,
		sendChan: make(chan hiface.IMessage, 5000),
		exitChan: make(chan bool, 1),
	}

	return c
}

func (c *Connection) IsClose() bool {
	return c.close
}

func (c *Connection) GetConnID() uint32 {
	return c.connID
}

//每个客户端一个写协程
func (c *Connection) WriteLoop() {
	go func() {
		for {
			select {
			case <-c.exitChan:
				close(c.sendChan)
				c.conn.Close()
				log4go.Info("[Connection] ConnID:%d ExitChan Read SendChan:%d", c.connID, len(c.sendChan))
				return
			case msg := <-c.sendChan:
				data := c.proto.Encode(msg)
				if _, err := c.conn.Write(data); err != nil {
					log4go.Info(c.connID, " write err: ", err)
					c.close = true
					return
				}
			}
		}
	}()
}

//每个客户端一个读协程
func (c *Connection) ReadLoop() {
	go func() {
		defer c.Stop()
		log4go.Debug("readLoop start")
		for {
			buf := make([]byte, c.proto.GetMsgHeadLen())
			log4go.Debug("make []byte")
			if _, err := io.ReadFull(c.conn, buf); err != nil {
				log4go.Info(c.connID, " read err: ", err)
				return
			}

			msg, err := c.proto.Decode(buf)
			if err != nil {
				log4go.Info("proto.Decode err:", err)
				return
			}

			if msg.GetLen() > 0 {
				dataBuf := make([]byte, msg.GetLen())
				if _, err := io.ReadFull(c.conn, dataBuf); err != nil {
					log4go.Info(c.connID, "read err: ", err)
					return
				}
				msg.SetData(dataBuf)
			}
			log4go.Debug("%+v", msg)
			//解包出来的消息放入业务处理协程
			WorkPool.AddTask(
				func() {
					switch msg.GetID() {
					default:
						handle, err := c.server.GetMsgHandle()
						if err != nil {
							log4go.Info("Get MsgHandle err: ", err)
							return
						}
						handle(c, msg)
					}
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
	return c.conn
}

func (c *Connection) Stop() {
	c.exitChan <- true
	c.close = true
	msg := &Message{}
	WorkPool.AddTask(
		func() {
			switch msg.GetID() {
			default:
				handle, err := c.server.GetMsgHandle()
				if err != nil {
					return
				}
				handle(c, msg)
			}
		},
	)
}

//提供业务层发送数据
func (c *Connection) SendMessage(msg hiface.IMessage) {
	if !c.IsClose() {
		c.sendChan <- msg
	}
}
