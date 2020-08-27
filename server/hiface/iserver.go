package hiface

type MsgHandle = func(IConnection, IMessage)

type IServer interface {
	Start()
	ServerInit(handle MsgHandle)
}
