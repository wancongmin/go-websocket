package impl

import "time"

type Iserver interface {
	//启动
	Start()
	//停止
	Stop()
	//运行
	Server()
	//路由功能，给当前的服务注册一个路由方法，提供客户端的链接处理使用
	AddRouter(msgID uint32, router IRouter)
	GetConnMgr() IConnManager

	//注册OnConnStart 钩子函数的方法
	SetConnStart(func(connection Iconnection))
	//注册OnConnStop 钩子函数的方法
	SetConnStop(func(connection Iconnection))
	//调用OnConnStart 钩子函数的方法
	CallConnStart(connection Iconnection)
	//调用OnConnStop 钩子函数的方法
	CallConnStop(connection Iconnection)
	// 启动心跳检测(自定义回调)
	StartHeartBeatWithOption(time.Duration, *HeartBeatOption)
	// Get the heartbeat checker
	// (获取心跳检测器)
	GetHeartBeat() IHeartbeatChecker
}
