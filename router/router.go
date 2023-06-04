package router

import (
	"encoding/json"
	"log"
	"strconv"
	"websocket/lib/mylog"
	"websocket/model"
	"websocket/ziface"
	"websocket/znet"
)

type HolleRouter struct {
	znet.BaseRouter
}

// ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}
type LocationRouter struct {
	znet.BaseRouter
}
type ChangeGroupRouter struct {
	znet.BaseRouter
}

func (this *HolleRouter) Handle(request ziface.IRequest) {
	m := model.ReceiveMsg{}
	err := json.Unmarshal(request.GetData(), &m)
	if err != nil {
		mylog.Error("Message parsing JSON error:" + err.Error())
		return
	}
	if m.MsgId == 0 {
		mylog.Error("Incorrect message parameters:" + err.Error())
		return
	}
	request.GetConnection().GetTCPConnection().PongHandler()
	//from_id, ok := m.Data["uid"]
	//if !ok {
	//	mylog.Error("发送参数不正确")
	//}
	//s.GetConnMgr().GetTotalConnections()
	//err = znet.Managers.Connections[m.UserId].SendMsg(200, request.GetData())
	//if err != nil {
	//	mylog.Error("Send message:" + err.Error())
	//	return
	//}
}

// Handle MsgId=100  心跳
func (this *PingRouter) Handle(request ziface.IRequest) {
	//先读取客户端数据再回写
	//log.Println("recv from client msgID=", request.GetMsgId(), ",data=", string(request.GetData()))
	msg := model.SendStringMsg{
		MsgId: 200,
		Data:  "pong",
	}
	marshal, err := json.Marshal(msg)
	if err != nil {
		mylog.Error("Marshal msg err:" + err.Error())
		return
	}
	err = request.GetConnection().SendMsg(200, marshal)
	if err != nil {
		mylog.Error("Send msg err:" + err.Error())
	}
	log.Printf("【心跳】ID:%d", request.GetConnection().GetConnID())
}

// Handle MsgId=101 上传定位
func (this *LocationRouter) Handle(request ziface.IRequest) {
	// 获取定位信息并存入redis
	uid := request.GetConnection().GetConnID()
	if uid == 0 {
		return
	}
	msg := model.ReceiveMsg{}
	err := json.Unmarshal(request.GetData(), &msg)
	if err != nil {
		mylog.Error("Unmarshal msg err:" + err.Error())
		return
	}
	longitude, ok := msg.Data["longitude"]
	if !ok {
		mylog.Error("get longitude empty")
		return
	}
	latitude, ok := msg.Data["latitude"]
	if !ok {
		mylog.Error("get latitude empty")
		return
	}
	electricity := msg.Data["electricity"]
	user := model.User{
		Id:          uid,
		Longitude:   longitude,
		Latitude:    latitude,
		Electricity: electricity,
	}
	model.SetUserLocation(user)
	log.Printf("【上传定位】ID:%d,Longitude:%s,Latitude:%s", user.Id, user.Longitude, user.Latitude)
}

// Handle MsgId=102 切换频道
func (this *ChangeGroupRouter) Handle(request ziface.IRequest) {
	// 获取定位信息并存入redis
	uid := request.GetConnection().GetConnID()
	if uid == 0 {
		return
	}
	msg := model.ReceiveMsg{}
	err := json.Unmarshal(request.GetData(), &msg)
	if err != nil {
		mylog.Error("Unmarshal msg err:" + err.Error())
		return
	}
	roomType, ok := msg.Data["type"]
	if !ok {
		mylog.Error("get room type empty")
		return
	}
	if roomType != "0" && roomType != "1" && roomType != "2" && roomType != "3" {
		// roomType值错误
		return
	}
	request.GetConnection().SetProperty("type", roomType)
	if roomType == "2" || roomType == "3" {
		if roomId, ok := msg.Data["roomId"]; ok {
			request.GetConnection().SetProperty("roomId", roomId)
			var message model.SendLocationMsg
			message.MsgId = 201
			message.Type = roomType
			message.UserId = uid
			intRoomId, err := strconv.Atoi(roomId)
			if err != nil {
				return
			}
			message.RoomId = intRoomId
			if roomType == "2" {
				message.Users = model.GetActivityMemberLocation(intRoomId, uid)
			} else {
				message.Users = model.GetClubMemberLocation(intRoomId, uid)
			}
			marshal, err := json.Marshal(message)
			if err != nil {
				return
			}
			request.GetConnection().SendMsg(201, marshal)
			log.Printf("【切换频道】ID:%d,Type:%s,RoomId:%s", uid, roomType, roomId)
		}
	} else {
		log.Printf("【切换频道】ID:%d,Type:%s", uid, roomType)
	}
}
