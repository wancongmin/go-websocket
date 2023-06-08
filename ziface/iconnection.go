package ziface

import (
	"github.com/gorilla/websocket"
	"net"
	"time"
)

type Iconnection interface {
	// Start 启动连接，让当前的链接装备工作
	Start()
	// Stop 停止链接 结束当前的链接工作
	Stop()
	//获取当前链接的绑定 socket conn
	GetTCPConnection() *websocket.Conn
	//获取当前链接模块的链接ID
	GetConnID() uint32
	//获取远程客户端的TCP  状态 IP port
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	//发送数据，将数据发送给远程客户端
	SendMsg(msgId uint32, data []byte) error
	//设置链接属性
	SetProperty(key string, value interface{})
	//获取链接属性
	GetProperty(key string) (interface{}, error)
	//移除链接属性
	RemoveProperty(key string)
	//检测是否活跃
	IsAlive() bool
	SetHeartBeat(checker IHeartbeatChecker) // Set the heartbeat detector (设置心跳检测器)
	GetActivityTime() time.Time             // 获取最后活跃时间
}

//定义一个处理链接业务的方法

type HandleFunc func(*net.TCPConn, []byte, int) error
