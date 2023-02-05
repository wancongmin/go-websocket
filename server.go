package main

import (
	"fmt"
	"websocket/config"
	"websocket/lib/db"
	"websocket/lib/redis"
	"websocket/utils"
	"websocket/ziface"
	"websocket/znet"
)

//ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}

func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call PingRouter Handle..")
	//先读取客户端数据再回写
	fmt.Println("recv from client msgID=", request.GetMsgId(), ",data=", string(request.GetData()))
	err := request.GetConnection().SendMsg(200, request.GetData())
	if err != nil {
		fmt.Println("Handle err", err)
	}
}

type HolleRouter struct {
	znet.BaseRouter
}

func (this *HolleRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call HolleRouter Handle..")
	//先读取客户端数据再回写
	fmt.Println("recv from client msgID=", request.GetMsgId(), ",data=", string(request.GetData()))
	//err:=request.GetConnection().SendMsg(201,request.GetData())
}

//创建链接之后执行的钩子函数
func DoConnectionBegin(conn ziface.Iconnection) {
	fmt.Println("====>DoConnection is Call")
	if err := conn.SendMsg(202, []byte("DoConnection Beagin")); err != nil {
		fmt.Println(err)
	}
	//链接之前设置一些属性
	fmt.Println("Set Property....")
	conn.SetProperty("name", "wancongmin")
	conn.SetProperty("home", "wuhan")
}

//链接断开执行的钩子函数
func DoConnectionLost(conn ziface.Iconnection) {
	fmt.Println("====>DoConnectionLost is Call")
	fmt.Println("====>conn ID =", conn.GetConnID())
	//获取链接属性
	if val, err := conn.GetProperty("name"); err == nil {
		fmt.Println("name", val)
	}
	if val, err := conn.GetProperty("home"); err == nil {
		fmt.Println("home", val)
	}
}

func main() {
	utils.CustomError()
	config.InitConf()
	db.InitDb()
	redis.InitRedis()
	//创建server句柄，使用zinx的api
	s := znet.NewServer("mysocket-01")
	s.AddRouter(1, &PingRouter{})
	//s.AddRouter(1,&HolleRouter{})

	//注册连接的Hook钩子函数
	s.SetConnStart(DoConnectionBegin)
	s.SetConnStop(DoConnectionLost)
	//启动Server
	s.Server()
}
