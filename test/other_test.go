package test

import (
	"errors"
	"fmt"
	"github.com/go-ini/ini"
	"math"
	"strconv"
	"testing"
	"websocket/config"
	"websocket/lib/db"
	"websocket/lib/redis"
	"websocket/model/game"
	"websocket/utils"
)

var ConfFile *ini.File

func init() {
	config.InitConf("../config/conf.ini")
	db.InitDb()
	redis.InitRedis()
}

// 使用代理请求
func TestOther(t *testing.T) {
	lat1, lng1 := 32.060255, 118.796877
	lat2, lng2 := 39.904211, 116.407395
	distance, _ := EarthDistance(lat1, lng1, lat2, lng2)
	fmt.Printf("%fkm", distance)
}

func EarthDistance(lat1, lng1, lat2, lng2 float64) (float64, error) {
	if lat1 == 0 || lng1 == 0 || lat2 == 0 || lng2 == 0 {
		return 0, errors.New("参数错误")
	}
	radius := 6378.137
	rad := math.Pi / 180.0
	lat1 = lat1 * rad
	lng1 = lng1 * rad
	lat2 = lat2 * rad
	lng2 = lng2 * rad
	theta := lng2 - lng1
	dist := math.Acos(math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))
	return Decimal(dist * radius * 1000), nil
}

func Decimal(num float64) float64 {
	num, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", num), 64)
	return num
}

func TestGetPlayersByRoomId(t *testing.T) {
	res := game.GetFinishVoteNum("9610")
	fmt.Println(res)
	return

	var allPlayerNum int64
	db.Db.Table("fa_game_player").Where("room_id = ? AND status = ?", 1554, 0).Count(&allPlayerNum)

	val := utils.GetConfVal("game_hunter_rate")
	float, err := strconv.ParseFloat(val, 64)
	var roleOne float64 = 1
	if err == nil {
		num := math.Floor(float64(allPlayerNum) * float / 100)
		roleOne = math.Max(num, 1)
	}
	//分配角色
	//if err = db.Db.Debug().Table("fa_game_player").
	//	Where("room_id = ? AND status = ?", 1554, 0).
	//	Order("rand()").
	//	Limit(int(roleOne)).
	//	Updates(game.Player{Role: 1}).Error; err != nil {
	//	fmt.Println(err)
	//}
	//
	//return
	//开始抓捕
	tx := db.Db.Begin()
	if err = tx.Table("fa_game_room").Where("id = ?", 1554).Updates(game.Room{Status: 2}).Error; err != nil {
		tx.Rollback()
		return
	}
	//分配角色
	if err = tx.Table("fa_game_player").
		Where("room_id = ? AND status = ?", 1554, 0).
		Order("rand()").
		Limit(int(roleOne)).
		Updates(game.Player{Role: 1}).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Table("fa_game_player").
		Where("room_id = ? AND status = ?", 1554, 0).
		Updates(game.Player{Role: 2}).Error; err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}
