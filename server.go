package main

import (
	"fmt"
	"log"
	"websocket/config"
	"websocket/lib/db"
	"websocket/lib/mylog"
	"websocket/lib/redis"
	"websocket/router"
	"websocket/utils"
	"websocket/ziface"
	"websocket/znet"
)

// 创建链接之后执行的钩子函数
func DoConnectionBegin(conn ziface.Iconnection) {
	fmt.Println("====>DoConnection is Call")
	if err := conn.SendMsg(202, []byte("DoConnection Beagin")); err != nil {
		mylog.Error("Send message:" + err.Error())
	}
	//链接之前设置一些属性
	//conn.SetProperty("name", "wancongmin")
	//conn.SetProperty("home", "wuhan")
}

// 链接断开执行的钩子函数
func DoConnectionLost(conn ziface.Iconnection) {
	log.Println("====>DoConnectionLost is Call")
	log.Println("====>conn ID =", conn.GetConnID())
	//获取链接属性
	if val, err := conn.GetProperty("name"); err == nil {
		log.Println("name", val)
	}
	if val, err := conn.GetProperty("home"); err == nil {
		log.Println("home", val)
	}
}

func main() {
	defer utils.CustomError()
	config.InitConf()
	db.InitDb()
	redis.InitRedis()
	//创建server句柄，使用zinx的api
	s := znet.NewServer("funParty")
	s.AddRouter(100, &router.PingRouter{})
	s.AddRouter(101, &router.LocationRouter{})
	s.AddRouter(102, &router.ChangeGroupRouter{})

	//注册连接的Hook钩子函数
	s.SetConnStart(DoConnectionBegin)
	s.SetConnStop(DoConnectionLost)
	//启动Server
	s.Server()
}
