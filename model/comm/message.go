package comm

import (
	"encoding/json"
	"websocket/impl"
)

type BaseReqMsg struct {
	MsgId uint32
	Data  map[string]interface{}
}

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

func SendMsg(conn impl.Iconnection, msgId uint32, resp ResponseMsg) {
	resp.Code = msgId
	marshal, err := json.Marshal(resp)
	if err != nil {
		return
	}
	conn.SendMsg(msgId, marshal)
}
