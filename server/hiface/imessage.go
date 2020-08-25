package hiface

type IMessage interface {
	GetID() uint32
	GetData() []byte
}
