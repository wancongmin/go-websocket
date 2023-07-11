package comm

import (
	"encoding/json"
	"websocket/core"
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
	MsgId      uint32
	Code       uint32
	Msg        string
	FromUserId string
	Data       interface{}
}

type QueueMsg struct {
	MsgId      uint32      `json:"msg_id"`
	FromUserId string      `json:"from_user_id"`
	ToUserIds  []uint32    `json:"to_user_ids"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data"`
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

// SendPlayerMessage 给玩家发消息
func SendPlayerMessage(uid, msgId uint32, msg ResponseMsg) {
	corePlayer := core.WorldMgrObj.GetPlayerByPID(uid)
	if corePlayer == nil {
		return
	}
	SendMsg(corePlayer.Conn, msgId, msg)
}

func SendMsg(conn impl.Iconnection, msgId uint32, resp ResponseMsg) {
	resp.MsgId = msgId
	marshal, err := json.Marshal(resp)
	if err != nil {
		return
	}
	conn.SendMsg(msgId, marshal)
}
