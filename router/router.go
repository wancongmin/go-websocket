package router

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"websocket/impl"
	"websocket/lib/mylog"
	"websocket/model/club"
	"websocket/model/comm"
	"websocket/service"
)

type HolleRouter struct {
	service.BaseRouter
}

// ping test 自定义路由
type PingRouter struct {
	service.BaseRouter
}
type LocationRouter struct {
	service.BaseRouter
}
type ChangeGroupRouter struct {
	service.BaseRouter
}

// Handle MsgId=100  心跳
func (this *PingRouter) Handle(request impl.IRequest) {
	//先读取客户端数据再回写
	//log.Println("recv from client msgID=", request.GetMsgId(), ",data=", string(request.GetData()))
	msg := comm.SendStringMsg{
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
func (this *LocationRouter) Handle(request impl.IRequest) {
	// 获取定位信息并存入redis
	uid := request.GetConnection().GetConnID()
	if uid == 0 {
		return
	}
	location := comm.LocationReq{}
	err := json.Unmarshal(request.GetData(), &location)
	if err != nil {
		mylog.Error("Unmarshal msg err:" + err.Error())
		return
	}
	if strings.TrimSpace(location.Longitude) == "" {
		mylog.Error("get longitude empty")
		return
	}
	if strings.TrimSpace(location.Latitude) == "" {
		mylog.Error("get latitude empty")
		return
	}
	user := comm.User{
		Id:          uid,
		Longitude:   location.Longitude,
		Latitude:    location.Latitude,
		Electricity: location.Electricity,
	}
	comm.SetUserLocation(user)
	log.Printf("【上传定位】ID:%d,Longitude:%s,Latitude:%s", user.Id, user.Longitude, user.Latitude)
}

// Handle MsgId=102 切换频道
func (this *ChangeGroupRouter) Handle(request impl.IRequest) {
	// 获取定位信息并存入redis
	uid := request.GetConnection().GetConnID()
	if uid == 0 {
		return
	}
	userType := comm.UserType{}
	err := json.Unmarshal(request.GetData(), &userType)
	if err != nil {
		mylog.Error("Unmarshal msg err:" + err.Error())
		return
	}
	if userType.Type != "0" && userType.Type != "1" && userType.Type != "2" && userType.Type != "3" {
		// roomType值错误
		mylog.Error("get room type error,type:" + userType.Type)
		return
	}
	var message comm.SendLocationMsg
	message.MsgId = 201
	message.Type = userType.Type
	message.UserId = uid
	if userType.Type == "2" || userType.Type == "3" {
		if userType.RoomId != "" {
			comm.SetUserType(request, userType)
			intRoomId, err := strconv.Atoi(userType.RoomId)
			if err != nil {
				return
			}
			message.RoomId = intRoomId
			if userType.Type == "2" {
				message.Users = club.GetActivityMemberLocation(intRoomId, uid)
			} else {
				message.Users = club.GetClubMemberLocation(intRoomId, uid)
			}
			marshal, err := json.Marshal(message)
			if err != nil {
				return
			}
			request.GetConnection().SendMsg(201, marshal)
			log.Printf("【切换频道】ID:%d,Type:%s,RoomId:%s", uid, userType.Type, userType.RoomId)
		}
	} else {
		if userType.Type == "1" {
			message.Users = comm.GetFriendLocation(uid)
			marshal, err := json.Marshal(message)
			if err != nil {
				return
			}
			request.GetConnection().SendMsg(201, marshal)
		}
		comm.SetUserType(request, userType)
		log.Printf("【切换频道】ID:%d,Type:%s", uid, userType.Type)
	}
}
