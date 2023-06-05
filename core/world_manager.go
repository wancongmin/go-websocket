package core

import (
	"sync"
)

/*
当前游戏世界的总管理模块
*/
type WorldManager struct {
	Players map[uint32]*Player //当前在线的玩家集合
	pLock   sync.RWMutex       //保护Players的互斥读写机制
}

// 提供一个对外的世界管理模块句柄
var WorldMgrObj *WorldManager

// 提供WorldManager 初始化方法
func init() {
	WorldMgrObj = &WorldManager{
		Players: make(map[uint32]*Player),
	}
}

// 提供添加一个玩家的的功能，将玩家添加进玩家信息表Players
func (wm *WorldManager) AddPlayer(player *Player) {
	//将player添加到 世界管理器中
	wm.pLock.Lock()
	wm.Players[player.PID] = player
	wm.pLock.Unlock()
}

// 从玩家信息表中移除一个玩家
func (wm *WorldManager) RemovePlayerByPID(pID uint32) {
	wm.pLock.Lock()
	delete(wm.Players, pID)
	wm.pLock.Unlock()
}

// 通过玩家ID 获取对应玩家信息
func (wm *WorldManager) GetPlayerByPID(pID uint32) *Player {
	wm.pLock.RLock()
	defer wm.pLock.RUnlock()

	return wm.Players[pID]
}

// 获取所有玩家的信息
func (wm *WorldManager) GetAllPlayers() []*Player {
	wm.pLock.RLock()
	defer wm.pLock.RUnlock()

	//创建返回的player集合切片
	players := make([]*Player, 0)

	//添加切片
	for _, v := range wm.Players {
		players = append(players, v)
	}

	//返回
	return players
}
