package router

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"websocket/lib/mylog"
	"websocket/lib/redis"
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
	err = znet.Managers.Connections[m.UserId].SendMsg(200, request.GetData())
	if err != nil {
		mylog.Error("Send message:" + err.Error())
		return
	}
}

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
	key := "mapLocation:uid_" + fmt.Sprintf("%v", uid)
	redis.Redis.Set("aaa", 15, 0)
	redis.Redis.HMSet(key, "longitude", longitude, "latitude", latitude)
	redis.Redis.Expire(key, 300*time.Second)
}

func (this *ChangeGroupRouter) Handle(request ziface.IRequest) {
	// 获取定位信息并存入redis
	uid := request.GetConnection().GetConnID()
	if uid == 0 {
		return
	}
}
