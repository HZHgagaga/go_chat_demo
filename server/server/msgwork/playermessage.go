package msgwork

import (
	"fmt"
	"hzhgagaga/server/core"
	"hzhgagaga/server/model"
)

type PlayerMessage struct {
}

func (p *PlayerMessage) OnCreatePlayer(plr *model.Player, msg *core.Message) {
	fmt.Println("OnCreatePlayer")
}
