package game

import (
	"fmt"
	"log"
	"sync"
	"time"
	"websocket/core"
	"websocket/lib/mylog"
	"websocket/model/comm"
)

type GameRoomChecker struct {
	interval        time.Duration //  Heartbeat detection interval(检测时间间隔)
	quitChan        chan bool     // Quit signal(退出信号)
	GameRoom        GameRoom
	lastErrorTime   int                    //上次异常时间
	Players         map[uint32]*GamePlayer //当前在线的玩家集合
	pLock           sync.RWMutex           //保护Players的互斥读写机制
	CloseRoomHandle func(roomId string)
	EndRoomHandle   func(roomId string)
}

func NewGameRoomChecker(interval time.Duration, room GameRoom) *GameRoomChecker {
	roomChecker := &GameRoomChecker{
		interval: interval,
		quitChan: make(chan bool),
		GameRoom: room,
		Players:  make(map[uint32]*GamePlayer),
	}

	return roomChecker
}

// 游戏开始，开启检测器
func (r *GameRoomChecker) start() {
	ticker := time.NewTicker(r.interval)
	for {
		select {
		case <-ticker.C:
			r.check()
		case <-r.quitChan:
			ticker.Stop()
			return
		}
	}
}

// 房间轮训
func (h *GameRoomChecker) check() {
	// 获取最新房间信息
	room, err := GetRoomCache(h.GameRoom.Id)
	if err != nil {
		log.Printf("【game】获取房间信息异常,退出,roomId:%s", h.GameRoom.Id)
		mylog.Error("获取房间信息异常,退出,roomId:" + fmt.Sprintf("%s", h.GameRoom.Id))
		h.Stop()
		return
	}
	//游戏结束
	if (room.Status == 2 && room.EndTime < time.Now().Unix()) || room.Status == 3 {
		log.Printf("【game】游戏结束,roomId:%s", h.GameRoom.Id)
		h.EndRoomHandle(h.GameRoom.Id)
		h.Stop()
		return
	}
	//玩家信息检测
	if !h.CheckRoomPlayers(h.GameRoom.Id) {
		return
	}
	//定时推送房间信息
	h.sendRoomInfoToPlayers(room.Id)
}

func (h *GameRoomChecker) Start() {
	go h.start()
}

func (h *GameRoomChecker) Stop() {
	go func() {
		h.quitChan <- true
	}()
}

// SuperPlayer 提供添加一个玩家的的功能，将玩家添加进玩家信息表Players
func (h *GameRoomChecker) SuperPlayer(player *GamePlayer) {
	h.pLock.Lock()
	h.Players[player.UserId] = player
	h.pLock.Unlock()
}

// GetPlayerByUid 通过玩家ID 获取对应玩家信息
func (h *GameRoomChecker) GetPlayerByUid(uid uint32) *GamePlayer {
	h.pLock.RLock()
	defer h.pLock.RUnlock()

	return h.Players[uid]
}

// CheckRoomPlayers 检查房间用户
func (h *GameRoomChecker) CheckRoomPlayers(roomId string) bool {
	players := GetRunningPlayersByRoomId(roomId)
	if len(players) == 0 {
		log.Printf("【Game】房间没有玩家，游戏结束,roomId:%s", roomId)
		h.CloseRoomHandle(roomId)
		h.Stop()
		return false
	}
	for _, player := range players {
		//检查用户异常时长
		currentPlayer := h.GetPlayerByUid(player.UserId)
		if currentPlayer == nil {
			player.LastActiveTime = time.Now().Unix()
			h.SuperPlayer(&player)
			continue
		}
		if (currentPlayer.LastActiveTime + 60) < time.Now().Unix() {
			ErrorOutRoom(player, "当前连接异常或长时间未上传定位信息")
		}
		//检查用户是否在线
		oline := core.WorldMgrObj.GetPlayerByPID(player.UserId)
		if oline == nil {
			continue
		}
		//检查用户是否正确上传定位信息
		user, err := comm.GetUserTempLocation(player.UserId)
		if err != nil {
			continue
		}
		player.User = user
		player.LastActiveTime = time.Now().Unix()
		h.SuperPlayer(&player)
	}
	return true
}

// 给所有玩家发送房间消息
func (h *GameRoomChecker) sendRoomInfoToPlayers(roomId string) {
	room, err := GetRoomCache(roomId) //获取房间信息
	if err != nil {
		mylog.Error("获取房间信息错误,roomId:" + roomId)
	}
	data := make(map[string]interface{})
	data["Room"] = room
	data["OnlinePlayers"] = GetOlinePlayers(roomId) //在线玩家
	players := GetRunningPlayersByRoomId(roomId)    //所有玩家
	var successNum = 0
	for _, player := range players {
		data["Self"] = player
		msg := comm.ResponseMsg{
			Code: 1,
			Msg:  "success",
			Data: data,
		}
		SendMessage(player.UserId, 206, msg)
		successNum++
	}
	log.Printf("【Game】定时发送房间信息,roomId:%s,allNum:%d,sucessNum:%d", roomId, len(players), successNum)
}
