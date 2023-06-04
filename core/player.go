package core

import (
	"websocket/ziface"
)

//玩家对象
type Player struct {
	PID  int32              //玩家ID
	Conn ziface.Iconnection //当前玩家的连接
}

//创建一个玩家对象
func NewPlayer(conn ziface.Iconnection) *Player {
	p := &Player{
		Conn: conn, //角度为0，尚未实现
	}

	return p
}
