package server

type Player struct {
	Name string
}

func CreatePlayer(name string) *Player {
	return &Player{
		Name: name,
	}
}
