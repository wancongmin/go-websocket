package game

import (
	"encoding/json"
	"errors"
	"time"
	"websocket/core"
	"websocket/lib/db"
	"websocket/lib/redis"
	"websocket/model/comm"
)

type GamePlayer struct {
	Id         int       `gorm:"id"`
	UserId     uint32    `gorm:"user_id"`
	RoomId     string    `gorm:"room_id"`
	Role       int       `gorm:"room_id"`
	Status     int       `gorm:"status"`
	CreateTime int64     `gorm:"create_time"`
	User       comm.User `gorm:"-"`
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

// 获取玩家信息
func GetPlayerByUid(uid uint32, roomId string) (GamePlayer, error) {
	players := GetPlayersByRoomId(roomId)
	for _, player := range players {
		if player.UserId == uid {
			return player, nil
		}
	}
	return GamePlayer{}, errors.New("empty")
}

// 获取房间内玩家列表
func GetPlayersByRoomId(roomId string) []GamePlayer {
	key := "gameRoomPlayers:roomId_" + roomId
	var players []GamePlayer
	result, err := redis.Redis.Get(key).Result()
	if err == nil {
		_ = json.Unmarshal([]byte(result), &players)
		return players
	}
	db.Db.Table("fa_game_player").
		Where("room_id = ?", roomId).
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

// 获取有定位信息的用户数据
func GetOlinePlayers(roomId string) []GamePlayer {
	var resPlayers []GamePlayer
	players := GetPlayersByRoomId(roomId)
	for _, player := range players {
		if player.Status != 0 {
			continue
		}
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
