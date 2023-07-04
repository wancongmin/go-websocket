package model

import (
	"errors"
	"websocket/lib/db"
)

type GamePlayer struct {
	Id         int
	UserId     int
	RoomId     string
	Role       int
	Status     int
	CreateTime int
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
func GetPlayerById(uid uint32, playerId string) (GamePlayer, error) {
	var player GamePlayer
	db.Db.Table("fa_game_player").
		Where("id = ? and user_id=?", playerId, uid).
		First(&player)
	if player.Id == 0 {
		return player, errors.New("数据不处存在")
	}
	return player, nil
}

// 获取房间内玩家列表
func getPlayersByRoomId(roomId string) []GamePlayer {
	var players []GamePlayer
	db.Db.Table("fa_game_player").
		Where("room_id = ?", roomId).
		Find(&players)
	return players
}
