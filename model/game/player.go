package game

import (
	"encoding/json"
	"log"
	"time"
	"websocket/core"
	"websocket/lib/db"
	"websocket/lib/redis"
	"websocket/model/comm"
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
func GetRunningPlayer(uid uint32) (Player, error) {
	var player Player
	err := db.Db.Table("fa_game_player").
		Where("user_id = ? AND status = ? ", uid, 0).
		First(&player).Error
	log.Println("数据库err", err)
	log.Printf("player:%+v", player)
	return player, nil
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
	key := "gameRoom:roomId_" + roomId
	redis.Redis.Del(key)
}

// ChangeRoleTwo 变羊
func ChangeRoleTwo(roomId string, userId uint32) {
	db.Db.Table("fa_game_player").
		Where("room_id = ? AND user_id = ? AND status = ?", roomId, userId, 0).Updates(Player{Role: 2})
	ClearPlayersCache(roomId)
}

// GetOlinePlayers 获取有定位信息的用户数据
func GetOlinePlayers(roomId string) []Player {
	var resPlayers []Player
	players := GetRunningPlayersByRoomId(roomId)
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
	SendMessage(player.UserId, 207, msg)
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