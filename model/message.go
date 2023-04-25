package model

type ReceiveMsg struct {
	MsgId  uint32
	UserId uint32
	Data   map[string]string
}

type SendStringMsg struct {
	MsgId uint32
	Data  string
}
