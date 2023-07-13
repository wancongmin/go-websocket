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

type RoomChecker struct {
	interval        time.Duration //  Heartbeat detection interval(检测时间间隔)
	quitChan        chan bool     // Quit signal(退出信号)
	GameRoom        Room
	lastErrorTime   int                //上次异常时间
	Players         map[uint32]*Player //当前在线的玩家集合
	pLock           sync.RWMutex       //保护Players的互斥读写机制
	CloseRoomHandle func(room Room, errorMsg string)
	EndRoomHandle   func(room Room)
}

func NewGameRoomChecker(interval time.Duration, room Room) *RoomChecker {
	roomChecker := &RoomChecker{
		interval: interval,
		quitChan: make(chan bool),
		GameRoom: room,
		Players:  make(map[uint32]*Player),
	}

	return roomChecker
}

// 游戏开始，开启检测器
func (r *RoomChecker) start() {
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
func (h *RoomChecker) check() {
	// 获取最新房间信息
	room, err := GetRoomCache(h.GameRoom.Id)
	if err != nil {
		log.Printf("【game】获取房间信息异常,退出,roomId:%s", h.GameRoom.Id)
		mylog.Error("获取房间信息异常,退出,roomId:" + fmt.Sprintf("%s", h.GameRoom.Id))
		h.Stop()
		return
	}
	nowTime := time.Now().Unix()
	//游戏结束
	if room.Status == 2 && room.EndTime < nowTime {
		log.Printf("【game】游戏结束时间到,roomId:%s", h.GameRoom.Id)
		h.EndRoomHandle(room)
		h.Stop()
		return
	}
	//开始抓捕
	if room.Status == 1 && room.StartArrestTime < nowTime {
		StartArrest(room)
	}
	//玩家信息检测
	if !h.CheckRoomPlayers(room) {
		return
	}
	//定时检查投票信息

	//定时推送房间信息
	h.sendRoomInfoToPlayers(room)
}

func (h *RoomChecker) Start() {
	go h.start()
}

func (h *RoomChecker) Stop() {
	go func() {
		h.quitChan <- true
	}()
}

// SuperPlayer 提供添加一个玩家的的功能，将玩家添加进玩家信息表Players
func (h *RoomChecker) SuperPlayer(player *Player) {
	h.pLock.Lock()
	h.Players[player.UserId] = player
	h.pLock.Unlock()
}

// GetPlayerByUid 通过玩家ID 获取对应玩家信息
func (h *RoomChecker) GetPlayerByUid(uid uint32) *Player {
	h.pLock.RLock()
	defer h.pLock.RUnlock()

	return h.Players[uid]
}

// CheckRoomPlayers 检查房间用户
func (h *RoomChecker) CheckRoomPlayers(room Room) bool {
	players := GetRunningPlayersByRoomId(room.Id)
	if len(players) == 0 {
		log.Printf("【Game】房间没有玩家，游戏结束,roomId:%s", room.Id)
		h.CloseRoomHandle(room, "房间异常关闭")
		h.Stop()
		return false
	}
	var roleOneNum, roleTowNum int
	for _, player := range players {
		if player.Role == 1 {
			roleOneNum++
		}
		if player.Role == 2 {
			roleTowNum++
		}
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
	if room.Status == 1 || room.Status == 2 {
		// 角色2胜利
		if roleOneNum == 0 {
			log.Printf("【Game】游戏结束,房间没有在线猎人,roomId:%s", room.Id)
			Referee(room.Id, room, 2, players)
			h.Stop()
			return false
		}
		// 角色1胜利
		if roleTowNum == 0 {
			log.Printf("【Game】游戏结束,房间没有在线猎物,roomId:%s", room.Id)
			Referee(room.Id, room, 1, players)
			h.Stop()
			return false
		}
	}
	// 检查投票信息
	CheckVoteByRoomId(h.GameRoom.Id)
	return true
}

// 给所有玩家发送房间消息
func (h *RoomChecker) sendRoomInfoToPlayers(room Room) {
	data := make(map[string]interface{})
	data["Room"] = room
	data["OnlinePlayers"] = GetOlinePlayers(room.Id) //在线玩家
	players := GetRunningPlayersByRoomId(room.Id)    //所有玩家
	ruleOneNum, ruleTowNum := GetRuleNum(players)
	data["RuleOneNum"] = ruleOneNum
	data["RuleTowNum"] = ruleTowNum
	data["playerNum"] = len(players)
	data["finishVoteNum"] = GetFinishVoteNum(room.Id)
	var successNum = 0
	for _, player := range players {
		data["Self"] = player
		//msg := comm.ResponseMsg{
		//	Code: 1,
		//	Msg:  "success",
		//	Data: data,
		//}
		//SendMessage(player.UserId, 206, msg)
		successNum++
	}
	startArrestTime := time.Unix(room.StartArrestTime, 0).Format("2006-01-02 15:04:05")
	endTime := time.Unix(room.EndTime, 0).Format("2006-01-02 15:04:05")
	log.Printf("【Game】定时信息,roomId:%s,allNum:%d,sucessNum:%d,房间状态:%d,开始抓捕时间:%s,结束时间:%s", room.Id, len(players), successNum, room.Status, startArrestTime, endTime)
}
