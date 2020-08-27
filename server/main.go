package main

import (
	"hzhgagaga/hnet"
	"hzhgagaga/server"
)

func main() {
	//新建一个服务器
	s := hnet.NewServer("HZHChatServer", "0.0.0.0", "16666")
	//业务层消息处理函数初始化
	s.ServerInit(server.MsgHandle)
	s.Start()
}
