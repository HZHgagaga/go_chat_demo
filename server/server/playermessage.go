package server

import "fmt"

type PlayerMessage struct {
}

func (p *PlayerMessage) OnCreatePlayer(plr *Player, msg *Message) {
	fmt.Println("OnCreatePlayer")
}
