package main

import (
	"hzhgagaga/hnet"
	"hzhgagaga/server"
)

func main() {
	s := hnet.NewServer("HZHChatServer", "0.0.0.0", "16666")
	s.ServerInit(server.MsgHandle)
	s.Start()
}
