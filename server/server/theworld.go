package server

import (
	"fmt"
	"hzhgagaga/hiface"
	"sync"
)

type TheWorld struct {
	Users map[uint32]hiface.IConnection
	Mutex *sync.RWMutex
}

var theWorld *TheWorld
var once sync.Once

func GetTheWorld() *TheWorld {
	once.Do(func() {
		theWorld = &TheWorld{
			Users: make(map[uint32]hiface.IConnection),
			Mutex: &sync.RWMutex{},
		}
	})

	return theWorld
}

func (m *TheWorld) Lock() {
	m.Mutex.Lock()
}

func (m *TheWorld) Unlock() {
	m.Mutex.Unlock()
}

func MsgHandle(conn hiface.IConnection, msg hiface.IMessage) {
	theWorld := GetTheWorld()

	theWorld.Lock()
	if theWorld.Users[conn.GetConnID()] == nil {
		theWorld.Users[conn.GetConnID()] = conn
	}
	theWorld.Unlock()

	for _, conn := range theWorld.Users {
		fmt.Println(msg.GetData())
		conn.SendMessage(msg.GetData())
	}
}
