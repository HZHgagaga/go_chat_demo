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

//进入世界的协议
func (p *PlayerMessage) OnCMEnterWorld(role siface.IRole, msg *core.Message) {
	fmt.Println("One user enter the world")
	theWorld := role.GetTheWorld()
	eData := &pb.SMEnterWorld{}
	okData, err := proto.Marshal(eData)
	if err != nil {
		fmt.Println("proto.Marshal err:", err)
		return
	}

	req, err := theWorld.GetProto().Encode("SMEnterWorld", okData)
	if err != nil {
		fmt.Println("Encode err:", err)
		return
	}
	role.SendMessage(req)
	theWorld.AddRole(role)
	role.SetStatus(siface.ENTER)
}

//CMCreatePlayer协议会进入到这个函数进行业务处理
func (p *PlayerMessage) OnCMCreatePlayer(role siface.IRole, msg *core.Message) {
	if !role.IsStatus(siface.ENTER) {
		return
	}

	theWorld := role.GetTheWorld()
	cdata := &pb.CMCreatePlayer{}
	err := proto.Unmarshal(msg.GetData(), cdata)
	if err != nil {
		fmt.Println("proto.Unmarshal err:", err)
		return
	}

	okPb := &pb.SMCreatePlayer{}

	if res, _ := theWorld.GetRoleByName(cdata.GetName()); res != nil {
		okPb.Ok = false
		okPb.Msg = "名字已经存在"
	} else {
		role.SetName(cdata.GetName())
		fmt.Println("----------TheWorld---------AddPlayer Name:", cdata.GetName())
		okPb.Ok = true
		role.SetStatus(siface.ONLINE)
		theWorld.AddRoleByName(role)
	}

	okData, err := proto.Marshal(okPb)
	if err != nil {
		fmt.Println("proto.Marshal err:", err)
		return
	}

	req, err := theWorld.GetProto().Encode("SMCreatePlayer", okData)
	if err != nil {
		fmt.Println("Encode err:", err)
		return
	}
	role.SendMessage(req)
}

//获取所有玩家的协议
func (p *PlayerMessage) OnCMAllPlayers(role siface.IRole, msg *core.Message) {
	if !role.IsStatus(siface.ONLINE) {
		return
	}

	theWorld := role.GetTheWorld()
	pData := &pb.SMAllPlayers{}
	roles := theWorld.GetAllRoles()
	for name, _ := range roles {
		pData.Names = append(pData.Names, name)
	}

	Data, err := proto.Marshal(pData)
	if err != nil {
		fmt.Println("proto.Marshal err:", err)
		return
	}

	req, err := theWorld.GetProto().Encode("SMAllPlayers", Data)
	if err != nil {
		fmt.Println("Encode err:", err)
		return
	}
	role.SendMessage(req)
}
