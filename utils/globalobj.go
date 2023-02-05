package utils

import "websocket/ziface"

type GlobalObj struct {
	//server
	TcpServer ziface.Iserver //当前全局Server对象
	Host      string
	TcpPost   int
	Name      string
	//zinx
}
