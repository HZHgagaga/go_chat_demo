package msgwork

import (
	"fmt"
	"hzhgagaga/server/core"
	"hzhgagaga/server/model"
)

type ChatMessage struct {
}

func (c *ChatMessage) OnBroadcastChat(plr *model.Player, msg *core.Message) {
	fmt.Println("--------------------------")
	plr.GetTheWorld().Broadcast(msg)
}
