package router

import (
	"encoding/json"
	"log"
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
	log.Println("Call HolleRouter Handle..")
	//先读取客户端数据再回写
	log.Println("recv from client msgID=", request.GetMsgId(), ",data=", string(request.GetData()))
	//err:=request.GetConnection().SendMsg(201,request.GetData())
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
	//err = znet.Managers.Connections[m.UserId].SendMsg(200, request.GetData())
	//if err != nil {
	//	mylog.Error("Send message:" + err.Error())
	//	return
	//}
}

// Handle MsgId=100  心跳
func (this *PingRouter) Handle(request ziface.IRequest) {
	//先读取客户端数据再回写
	log.Println("recv from client msgID=", request.GetMsgId(), ",data=", string(request.GetData()))
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
			log.Println("change group type success ", "type:"+roomType, "roomId:"+roomId)
		}
	} else {
		log.Println("change group type success ", "type:"+roomType)
	}
}
