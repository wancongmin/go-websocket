package router

import (
	"encoding/json"
	"log"
	"websocket/impl"
	"websocket/lib/mylog"
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
