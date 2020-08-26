package msgwork

import (
	"fmt"
	"hzhgagaga/server/core"
	"hzhgagaga/server/pb"
	"hzhgagaga/server/siface"
	"time"

	"google.golang.org/protobuf/proto"
)

type ChatMessage struct {
}

func (c *ChatMessage) OnCMBroadcastChat(role siface.IRole, msg *core.Message) {
	chat := &pb.CSBroadcastChat{}
	err := proto.Unmarshal(msg.GetData(), chat)
	if err != nil {
		fmt.Println("proto.Unmarshal err: ", err)
	}

	reqChat := &pb.SMBroadcastChat{}
	reqChat.Time = time.Now().Format("2006-01-02 15:04:05")
	reqChat.Name = role.GetName()
	reqChat.Chatdata = chat.GetChatdata()
	fmt.Println(reqChat.Time, reqChat.Name, reqChat.Chatdata)
	reqData, err := proto.Marshal(reqChat)
	if err != nil {
		fmt.Println("proto.Marshal err: ", err)
	}

	req, err := role.GetTheWorld().GetProto().Encode("SMBroadcastChat", reqData)
	if err != nil {
		fmt.Println("Encode err: ", err)
	}
	role.GetTheWorld().Broadcast(req)
}
