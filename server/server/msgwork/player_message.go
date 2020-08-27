package msgwork

import (
	"fmt"
	"hzhgagaga/server/core"
	"hzhgagaga/server/pb"
	"hzhgagaga/server/siface"

	"google.golang.org/protobuf/proto"
)

type PlayerMessage struct {
}

//CMCreatePlayer协议会进入到这个函数进行业务处理
func (p *PlayerMessage) OnCMCreatePlayer(role siface.IRole, msg *core.Message) {
	theWorld := role.GetTheWorld()
	cdata := &pb.CMCreatePlayer{}
	err := proto.Unmarshal(msg.GetData(), cdata)
	if err != nil {
		fmt.Println("proto.Unmarshal err:", err)
	}

	role.SetName(cdata.GetName())
	theWorld.AddRole(role)
	fmt.Println("----------TheWorld---------AddPlayer Name:", cdata.GetName())

	okPb := &pb.SMCreatePlayer{}
	okData, err := proto.Marshal(okPb)
	if err != nil {
		fmt.Println("proto.Marshal err:", err)
	}

	req, err := theWorld.GetProto().Encode("SMCreatePlayer", okData)
	if err != nil {
		fmt.Println("Encode err:", err)
	}
	role.SendMessage(req)
}
