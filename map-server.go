package main

import (
	"encoding/json"
	"log"
	"time"
	"websocket/config"
	"websocket/core"
	"websocket/lib/db"
	"websocket/lib/mylog"
	"websocket/lib/redis"
	"websocket/model"
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
	player := core.NewPlayer(conn)
	//将当前新上线玩家添加到worldManager中
	core.WorldMgrObj.AddPlayer(player)
	log.Printf("【上线成功】ID:%d,plays:%v", conn.GetConnID(), core.WorldMgrObj.GetAllPlayerIds())
	//链接之前设置一些属性
	//conn.SetProperty("name", "")
	//conn.SetProperty("home", "")
}

// 链接断开执行的钩子函数
func DoConnectionLost(conn ziface.Iconnection) {
	//log.Println("====>DoConnectionLost is Call")
	//log.Println("====>conn ID =", conn.GetConnID())

	//根据pID获取对应的玩家对象
	player := core.WorldMgrObj.GetPlayerByPID(conn.GetConnID())

	//触发玩家下线业务
	if player != nil {
		player.LostConnection()
	}
	log.Printf("【下线成功】ID:%d，plays:%v", conn.GetConnID(), core.WorldMgrObj.GetAllPlayerIds())
	//获取链接属性
	//if val, err := conn.GetProperty("name"); err == nil {
	//	log.Println("name", val)
	//}
	//if val, err := conn.GetProperty("home"); err == nil {
	//	log.Println("home", val)
	//}
}

// User-defined heartbeat message processing method
// 用户自定义的心跳检测消息处理方法
func myHeartBeatMsg(conn ziface.Iconnection) []byte {
	msg := model.SendStringMsg{
		MsgId: 200,
		Data:  "pong",
	}
	marshal, err := json.Marshal(msg)
	if err != nil {
		mylog.Error("Marshal msg err:" + err.Error())
		return []byte("")
	}
	return marshal
}

// User-defined handling method for remote connection not alive.
// 用户自定义的远程连接不存活时的处理方法
func myOnRemoteNotAlive(conn ziface.Iconnection) {
	//关闭链接
	conn.Stop()
}

type myHeartBeatRouter struct {
	znet.BaseRouter
}

func (r *myHeartBeatRouter) Handle(request ziface.IRequest) {
	log.Printf("【心跳】ID:%d", request.GetMsgId())
}

func main() {
	defer utils.CustomError()
	config.InitConf()
	db.InitDb()
	redis.InitRedis()
	utils.InitGlobalConf()
	//创建server句柄，使用zinx的api
	s := znet.NewServer("funParty")
	//s.AddRouter(100, &router.PingRouter{})
	s.AddRouter(101, &router.LocationRouter{})
	s.AddRouter(102, &router.ChangeGroupRouter{})

	//注册连接的Hook钩子函数
	s.SetConnStart(DoConnectionBegin)
	s.SetConnStop(DoConnectionLost)

	// Start heartbeating detection. (启动心跳检测)
	s.StartHeartBeatWithOption(3*time.Second, &ziface.HeartBeatOption{
		MakeMsg:          myHeartBeatMsg,
		OnRemoteNotAlive: myOnRemoteNotAlive,
		Router:           &myHeartBeatRouter{},
		HeadBeatMsgID:    uint32(100),
	})
	//启动Server
	s.Server()
}
