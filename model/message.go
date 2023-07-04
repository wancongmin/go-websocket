package model

import (
	"encoding/json"
	"websocket/impl"
)

type ReceiveMsg struct {
	MsgId  uint32
	UserId uint32
	Data   map[string]string
}
type ResponseMsg struct {
	MsgId uint32
	Code  uint32
	Msg   string
	Data  interface{}
}

type SendStringMsg struct {
	MsgId uint32
	Data  string
}

type SendLocationMsg struct {
	MsgId  uint32
	Type   string
	UserId uint32
	RoomId int
	Users  []User
}

func SendMsg(request impl.IRequest, msgId uint32, resp ResponseMsg) {
	marshal, err := json.Marshal(resp)
	if err != nil {
		return
	}
	request.GetConnection().SendMsg(msgId, marshal)
}
