package main

import (
	"log"
	"websocket/config"
	"websocket/core"
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
	//fmt.Println("====>DoConnection is Call")
	if err := conn.SendMsg(202, []byte("DoConnection Beagin")); err != nil {
		mylog.Error("Send message:" + err.Error())
	}
	//创建一个玩家
	//player := core.NewPlayer(conn)
	////将当前新上线玩家添加到worldManager中
	//core.WorldMgrObj.AddPlayer(player)
	log.Printf("【上线成功】ID:%d,plays:%v", conn.GetConnID(), core.WorldMgrObj.GetAllPlayers())
	//链接之前设置一些属性
	//conn.SetProperty("name", "")
	//conn.SetProperty("home", "")
}

// 链接断开执行的钩子函数
func DoConnectionLost(conn ziface.Iconnection) {
	//log.Println("====>DoConnectionLost is Call")
	//log.Println("====>conn ID =", conn.GetConnID())

	//根据pID获取对应的玩家对象
	//player := core.WorldMgrObj.GetPlayerByPID(conn.GetConnID())
	//
	////触发玩家下线业务
	//if player != nil {
	//	player.LostConnection()
	//}
	log.Printf("【下线成功】ID:%d，plays:%v", conn.GetConnID(), core.WorldMgrObj.GetAllPlayers())
	//获取链接属性
	//if val, err := conn.GetProperty("name"); err == nil {
	//	log.Println("name", val)
	//}
	//if val, err := conn.GetProperty("home"); err == nil {
	//	log.Println("home", val)
	//}
}

func main() {
	defer utils.CustomError()
	config.InitConf()
	db.InitDb()
	redis.InitRedis()
	//创建server句柄，使用zinx的api
	s := znet.NewServer("funParty")
	s.AddRouter(199, &router.PingRouter{})
	s.AddRouter(100, &router.PingRouter{})
	s.AddRouter(101, &router.LocationRouter{})
	s.AddRouter(102, &router.ChangeGroupRouter{})

	//注册连接的Hook钩子函数
	s.SetConnStart(DoConnectionBegin)
	s.SetConnStop(DoConnectionLost)
	//启动Server
	s.Server()
}
