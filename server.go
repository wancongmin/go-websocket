package main

import (
	"encoding/json"
	"log"
	"time"
	"websocket/config"
	"websocket/core"
	"websocket/impl"
	"websocket/lib/db"
	"websocket/lib/mylog"
	"websocket/lib/redis"
	"websocket/model/comm"
	"websocket/router"
	"websocket/service"
	"websocket/utils"
)

// 创建链接之后执行的钩子函数
func DoConnectionBegin(conn impl.Iconnection) {
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
func DoConnectionLost(conn impl.Iconnection) {
	//根据pID获取对应的玩家对象
	player := core.WorldMgrObj.GetPlayerByPID(conn.GetConnID())

	//触发玩家下线业务
	if player != nil {
		player.LostConnection()
	}
	log.Printf("【下线成功】ID:%d，plays:%v", conn.GetConnID(), core.WorldMgrObj.GetAllPlayerIds())
}

// User-defined heartbeat message processing method
// 用户自定义的心跳检测消息处理方法
func myHeartBeatMsg(conn impl.Iconnection) []byte {
	msg := comm.SendStringMsg{
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
func myOnRemoteNotAlive(conn impl.Iconnection) {
	//关闭链接
	conn.Stop()
}

type myHeartBeatRouter struct {
	service.BaseRouter
}

func (r *myHeartBeatRouter) Handle(request impl.IRequest) {
	log.Printf("【心跳】ID:%d", request.GetConnection().GetConnID())
}

func main() {
	defer utils.CustomError()
	// 初始化配置
	config.InitConf("")
	//初始化mysql
	db.InitDb()
	//初始化redis
	redis.InitRedis()
	utils.InitGlobalConf()
	//创建server句柄，使用api
	s := service.NewServer("websocket")
	// 心跳
	s.AddRouter(101, &router.PingRouter{})
	//注册连接的Hook钩子函数
	s.SetConnStart(DoConnectionBegin)
	s.SetConnStop(DoConnectionLost)

	// Start heartbeating detection. (启动心跳检测)
	s.StartHeartBeatWithOption(5*time.Second, &impl.HeartBeatOption{
		MakeMsg:          myHeartBeatMsg,
		OnRemoteNotAlive: myOnRemoteNotAlive,
		Router:           &myHeartBeatRouter{},
		HeadBeatMsgID:    uint32(100),
	})
	//启动Server
	s.Server()
}
