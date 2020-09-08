package main

import (
	"hzhgagaga/hnet"
	"hzhgagaga/server"

	"github.com/jeanphorn/log4go"
	"github.com/spf13/viper"
)

func main() {
	defer log4go.Close()
	//新建一个服务器
	s := hnet.NewServer("HZHChatServer", viper.GetString(`server.ip`), viper.GetString(`server.port`))
	//业务层消息处理函数初始化
	s.ServerInit(server.MsgHandle)
	s.Start()
}
