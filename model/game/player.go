package game

import (
	"encoding/json"
	"log"
	"time"
	"websocket/core"
	"websocket/lib/db"
	"websocket/lib/mylog"
	"websocket/lib/redis"
	"websocket/model/comm"
	"websocket/utils"
)

type Player struct {
	Id             int       `gorm:"id"`
	UserId         uint32    `gorm:"user_id"`
	RoomId         string    `gorm:"room_id"`
	Role           int       `gorm:"room_id"`
	Status         int       `gorm:"status"`
	CreateTime     int64     `gorm:"create_time"`
	ErrorMsg       string    `gorm:"error_msg"`
	User           comm.User `gorm:"-"`
	LastActiveTime int64     `gorm:"-" json:"-"`
	Distance       float64   `gorm:"-"`
}

type EnterRoomResMsg struct {
	MsgId int
	Data  map[string]string
}

type EnterRoomRespMsg struct {
	MsgId int
	Code  int
	Msg   string
}

// GetRunningPlayer 获取玩家信息
func GetRunningPlayer(uid uint32) []Player {
	var players []Player
	db.Db.Table("fa_game_player").
		Where("user_id = ? AND status = ? ", uid, 0).
		Find(&players)
	return players
}

// ExitRoom 退出房间
func ExitRoom(player Player) error {
	room, err := GetRoomCache(player.RoomId)
	if err != nil {
		return err
	}
	if room.MasterId == player.UserId {
		var newMaster Player
		db.Db.Table("fa_game_player").
			Where("room_id = ? AND status = ? AND user_id <> ?", player.RoomId, 0, player.UserId).
			First(&newMaster)
		db.Db.Table("fa_game_room").
			Where("id = ? ", room.Id).
			Updates(Room{MasterId: newMaster.UserId})
		ClearRoomCache(room.Id)
	}
	db.Db.Table("fa_game_player").
		Where("id = ?", player.Id).
		Updates(Player{Status: 3})
	ClearPlayersCache(player.RoomId)
	return nil
}

// GetRunningPlayersByRoomId 获取房间内所有玩家列表
func GetRunningPlayersByRoomId(roomId string) []Player {
	key := "gameRoomPlayers:roomId_" + roomId
	var players []Player
	result, err := redis.Redis.Get(key).Result()
	if err == nil {
		_ = json.Unmarshal([]byte(result), &players)
		return players
	}
	db.Db.Table("fa_game_player").
		Where("room_id = ? AND status = ?", roomId, 0).
		Find(&players)
	marshal, err := json.Marshal(players)
	if err != nil {
		return players
	}
	redis.Redis.Set(key, marshal, 30*time.Second)
	return players
}

func ClearPlayersCache(roomId string) {
	key := "gameRoomPlayers:roomId_" + roomId
	redis.Redis.Del(key)
}

// GetOlinePlayers 获取有定位信息的用户数据
func GetOlinePlayers(players []Player) []Player {
	var resPlayers []Player
	//players := GetRunningPlayersByRoomId(roomId)
	for _, player := range players {
		// 用户不在线
		olinePlayer := core.WorldMgrObj.GetPlayerByPID(player.UserId)
		if olinePlayer == nil {
			continue
		}
		// 未获取到用户定位信息
		user, err := comm.GetUserTempLocation(player.UserId)
		if err != nil {
			continue
		}
		user.Avatar = utils.RoundThumb(user.Avatar, "50")
		player.User = user
		resPlayers = append(resPlayers, player)
	}
	return resPlayers
}

// ErrorOutRoom 异常退出房间
func ErrorOutRoom(player Player, errorMsg string) {
	db.Db.Table("fa_game_player").
		Where("id = ? AND status = ?", player.Id, 0).Updates(Player{Status: 6, ErrorMsg: errorMsg})
	ClearPlayersCache(player.RoomId)
	msg := comm.ResponseMsg{
		Code: 1,
		Msg:  errorMsg,
	}
	SendMessage(player.UserId, 217, msg)
	log.Printf("【Game】用户异常退出房间:%d", player.UserId)
}

// SendMessage 发送消息
func SendMessage(uid, msgId uint32, msg comm.ResponseMsg) {
	corePlayer := core.WorldMgrObj.GetPlayerByPID(uid)
	if corePlayer == nil {
		return
	}
	comm.SendMsg(corePlayer.Conn, msgId, msg)
}

// SendMsgToPlayers 给玩家发消息
func SendMsgToPlayers(roomId string, role int, players []Player, msgId uint32, msg comm.ResponseMsg) {
	if len(players) == 0 {
		players = GetRunningPlayersByRoomId(roomId)
	}
	for _, player := range players {
		if role != 0 && player.Role != role {
			continue
		}
		SendMessage(player.UserId, msgId, msg)
	}

}

// GetRuleNum 获取角色数量
func GetRuleNum(players []Player) (ruleOneNum, ruleTowNum int) {
	var roleOneNum, roleTowNum int
	for _, player := range players {
		if player.Role == 1 {
			roleOneNum++
		}
		if player.Role == 2 {
			roleTowNum++
		}
	}
	return roleOneNum, roleTowNum
}

// CheckActive 获取玩家信息
func CheckActive(uid uint32) bool {
	//检查用户是否在线
	oline := core.WorldMgrObj.GetPlayerByPID(uid)
	if oline == nil {
		return false
	}
	//检查用户是否正确上传定位信息
	_, err := comm.GetUserTempLocation(uid)
	if err != nil {
		return false
	}
	return true
}

func PlayerDistance(uid uint32, roomStatus int, players []Player) (comm.User, []Player) {
	user, err := comm.GetUserTempLocation(uid)
	var respPlayer []Player
	if err != nil {
		mylog.Error("获取临时定为信息错误:" + err.Error())
		return user, respPlayer
	}
	// 计算玩家距离
	for _, player := range players {
		if uid != player.UserId {
			distance, _ := utils.EarthDistance(user.Latitude, user.Longitude, player.User.Latitude, player.User.Longitude)
			player.Distance = distance
		}
		if roomStatus == 1 && uid != player.UserId {
			continue
		}
		respPlayer = append(respPlayer, player)
	}
	return user, respPlayer
}
