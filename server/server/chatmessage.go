package server

import "fmt"

type ChatMessage struct {
}

func (c *ChatMessage) OnBroadcastChat(plr *Player, msg *Message) {
	fmt.Println("--------------------------")
	theWorld := GetTheWorld()
	theWorld.Broadcast(msg)
}
