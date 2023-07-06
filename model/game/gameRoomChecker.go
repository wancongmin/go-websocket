package game

import (
	"time"
	"websocket/core"
	"websocket/lib/mylog"
	"websocket/model/comm"
)

type GameRoomChecker struct {
	interval      time.Duration //  Heartbeat detection interval(检测时间间隔)
	quitChan      chan bool     // Quit signal(退出信号)
	GameRoom      GameRoom
	lastErrorTime int //上次异常时间
	EndHandle     func(roomId string)
}

func NewGameRoomChecker(interval time.Duration, room GameRoom) *GameRoomChecker {
	roomChecker := &GameRoomChecker{
		interval: interval,
		quitChan: make(chan bool),
		GameRoom: room,
	}

	return roomChecker
}

// 游戏检测器
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
func (h *GameRoomChecker) check() (err error) {
	// 获取最新房间信息
	room, err := GetRoomCache(h.GameRoom.Id)
	if err != nil {
		h.Stop()
	}
	//log.Printf("【房间信息】:%+v", room)
	//// 获取玩家信息
	//players := GetPlayersByRoomId(h.GameRoom.Id)
	//log.Printf("【玩家信息】%+v \n", players)
	h.sendPlayersMsg(room.Id)
	return err
}

func (h *GameRoomChecker) Start() {
	go h.start()
}

func (h *GameRoomChecker) Stop() {
	go func() {
		h.quitChan <- true
	}()
}

// 检查房间用户
func (h *GameRoomChecker) CheckPlayers(roomId string) {
	players := GetPlayersByRoomId(roomId)
	if len(players) == 0 {
		h.Stop()
		return
	}
	var number int
	for _, player := range players {
		if player.Status == 0 {
			number++
		}
	}
	if number == 0 {
		h.Stop()
		return
	}
}

// 给所有玩家发送房间消息
func (h *GameRoomChecker) sendPlayersMsg(roomId string) {
	room, err := GetRoomCache(roomId) //获取房间信息
	if err != nil {
		mylog.Error("获取房间信息错误,roomId:" + roomId)
	}
	data := make(map[string]interface{})
	data["Room"] = room
	data["OnlinePlayers"] = GetOlinePlayers(roomId) //在线玩家
	players := GetPlayersByRoomId(roomId)           //所有玩家
	for _, player := range players {
		corePlayer := core.WorldMgrObj.GetPlayerByPID(player.UserId)
		if corePlayer == nil {
			continue
		}
		data["Self"] = player
		msg := comm.ResponseMsg{
			Code: 1,
			Msg:  "success",
			Data: data,
		}
		comm.SendMsg(corePlayer.Conn, 206, msg)
	}
}
