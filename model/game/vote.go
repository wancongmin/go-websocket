package game

import (
	"math"
	"time"
	"websocket/lib/db"
)

// Vote 房间结构
type Vote struct {
	Id         string `gorm:"id"`
	RoomId     string `gorm:"room_id"`
	UserId     uint32 `gorm:"user_id"`
	ToUserId   uint32 `gorm:"to_user_id"`
	Status     int    `gorm:"status"`
	Image      string `gorm:"image"`
	CreateTime int64  `gorm:"create_time"`
	EndTime    int64  `gorm:"end_time"`
}

type VoteLog struct {
	Id         int    `gorm:"id"`
	UserId     uint32 `gorm:"user_id"`
	VoteId     int    `gorm:"vote_id"`
	RoomId     string `gorm:"room_id"`
	Status     int    `gorm:"status"`
	CreateTime int64  `gorm:"create_time"`
}

// CheckVoteByRoomId 检查投票
func CheckVoteByRoomId(roomId string) {
	var votes []Vote
	err := db.Db.Table("fa_game_vote").
		Where("room_id = ? AND status = ?", roomId, 0).
		Find(&votes).Error
	if err != nil {
		return
	}
	if len(votes) == 0 {
		return
	}
	players := GetRunningPlayersByRoomId(roomId)
	for _, vote := range votes {
		CheckVote(vote, players)
	}
}

func CheckVote(vote Vote, players []Player) {
	if vote.EndTime > time.Now().Unix() {
		return
	}
	// 获取投票人数
	var approveCount int64
	db.Db.Table("fa_game_vote_log").Where("vote_id = ? AND status = ?", vote.Id, 1).Count(&approveCount)
	if float64(approveCount) > math.Ceil(float64(len(players))/2) {
		ChangeRoleOne(vote.RoomId, vote.ToUserId)
	}
}
