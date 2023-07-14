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
	onlinePlayers := []game.Player{{Id: 5}, {Id: 6}}

	for i, _ := range onlinePlayers {
		//onlinePlayer.Id = 1
		onlinePlayers[i].Id = 6
	}
	fmt.Printf("result:%+v", onlinePlayers)
	//var winPlayers []game.Player
	//// 获胜狼的列表
	//db.Db.Table("fa_game_player p").
	//	Select("p.user_id,p.room_id,p.role,p.status,count(*) cn").
	//	Joins("left join fa_game_vote v on v.room_id = p.room_id and v.user_id=p.user_id and v.status=1").
	//	Where("p.room_id = ? AND p.role = ? AND p.status = ?", 6051, 1, 1).
	//	Group("user_id").
	//	Order("cn desc").
	//	Limit(3).
	//	Find(&winPlayers)
	//log.Printf("players:%+v", winPlayers)

}
