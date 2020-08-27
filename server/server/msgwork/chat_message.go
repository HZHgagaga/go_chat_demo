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

func (c *ChatMessage) OnCMHistoryChat(role siface.IRole, msg *core.Message) {
	theWorld := role.GetTheWorld()
	theWorld.GetAsyncPool().AsyncRun(
		func() {
			rows, _ := theWorld.GetDB().Query("select chat_id, chat_name, chat_time, chat_data from chat_msg order by chat_id desc LIMIT 50")
			chatArr := &pb.SMHistoryChat{}
			for rows.Next() {
				chat := &pb.SMBroadcastChat{}
				rows.Scan(&chat.Id, &chat.Name, &chat.Time, &chat.Chatdata)
				chatArr.Msg = append(chatArr.Msg, chat)
			}

			chatArrData, err := proto.Marshal(chatArr)
			if err != nil {
				fmt.Println("proto.Marshal err:", err)
			}

			req, err := theWorld.GetProto().Encode("SMHistoryChat", chatArrData)
			if err != nil {
				fmt.Println("Encode err:", err)
			}
			role.SendMessage(req)
		},
	)
}

func (c *ChatMessage) OnCMBroadcastChat(role siface.IRole, msg *core.Message) {
	chat := &pb.CMBroadcastChat{}
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

	world := role.GetTheWorld()
	world.Broadcast(req)

	db := world.GetDB()
	world.GetAsyncPool().AsyncRun(
		func() {
			_, err := db.Exec("insert into chat_msg (chat_name, chat_time, chat_data) values(?, ?, ?)", reqChat.Name, reqChat.Time, reqChat.Chatdata)
			if err != nil {
				fmt.Println("insert db err:", err)
				panic("DB err:" + err.Error())
			}
		},
	)
}
