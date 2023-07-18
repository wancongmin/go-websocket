package game

import (
	"fmt"
	"math"
	"time"
	"websocket/lib/db"
	"websocket/lib/mylog"
	"websocket/model/comm"
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
	StartTime  int64  `gorm:"start_time"`
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
		Where("room_id = ? AND status in ?", roomId, []int{-1, 0}).
		Find(&votes).Error
	if err != nil {
		mylog.Error("获取投票信息错误" + err.Error())
		return
	}
	if len(votes) == 0 {
		return
	}
	//log.Printf("获取votes:%+v", votes)
	players := GetRunningPlayersByRoomId(roomId)
	for _, vote := range votes {
		if vote.Status == -1 {
			CheckStartVote(vote, players)
		} else {
			CheckJudgeVote(vote, players)
		}
	}
}
func CheckStartVote(vote Vote, players []Player) {
	if vote.StartTime > time.Now().Unix() {
		return
	}
	db.Db.Table("fa_game_vote").
		Where("id = ?", vote.Id).
		Select("status").
		Updates(Vote{Status: 0})
	// 发送消息
	toUser := comm.GetUserById(vote.ToUserId)
	//发送通知消息
	msg := comm.ResponseMsg{
		Code:       1,
		FromUserId: "admin",
		Msg:        fmt.Sprintf("正在进行%s 的投票", toUser.Nickname),
		Data:       map[string]string{"VoteId": vote.Id},
	}
	SendMsgToPlayers(vote.RoomId, 0, players, 218, msg)
}

// CheckJudgeVote 裁决投票
func CheckJudgeVote(vote Vote, players []Player) {
	if vote.EndTime > time.Now().Unix() {
		return
	}
	// 获取投票人数
	var approveCount int64
	var voteStatus int
	db.Db.Table("fa_game_vote_log").Where("vote_id = ? AND status = ?", vote.Id, 1).Count(&approveCount)
	if (float64(approveCount) + 1) >= math.Ceil(float64(len(players))/2) {
		voteStatus = 1
		result := db.Db.Table("fa_game_player").
			Where("room_id = ? AND user_id = ? AND status = ?", vote.RoomId, vote.ToUserId, 0).Updates(Player{Role: 1})
		if result.RowsAffected > 0 {
			user := comm.GetUserById(vote.ToUserId)
			//发送通知消息
			msg := comm.ResponseMsg{
				Code:       1,
				FromUserId: "admin",
				Msg:        user.Nickname + " 被抓，成了狼",
			}
			SendMsgToPlayers(vote.RoomId, 0, players, 220, msg)
		}
		ClearPlayersCache(vote.RoomId)
	} else {
		voteStatus = 2
	}
	db.Db.Table("fa_game_vote").
		Where("id = ?", vote.Id).
		Updates(Vote{Status: voteStatus})
}

// GetFinishVoteNum 获取完成投票数量
func GetFinishVoteNum(roomId string) int64 {
	var count int64
	db.Db.Table("fa_game_vote").
		Where("room_id = ? AND status in ?", roomId, []int{1, 2}).
		Find(&count)
	return count
}
