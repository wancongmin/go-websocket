package core

import (
	"websocket/impl"
)

// 玩家对象
type Player struct {
	PID  uint32           //玩家ID
	Conn impl.Iconnection //当前玩家的连接
}

// 创建一个玩家对象
func NewPlayer(conn impl.Iconnection) *Player {
	p := &Player{
		PID:  conn.GetConnID(),
		Conn: conn, //角度为0，尚未实现
	}

	return p
}

// 玩家下线
func (p *Player) LostConnection() {
	WorldMgrObj.RemovePlayerByPID(p.PID)
}
