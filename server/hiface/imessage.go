package hiface

type IMessage interface {
	GetID() uint32
	GetLen() uint32
	GetData() []byte
}
