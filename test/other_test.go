package test

import (
	"github.com/go-ini/ini"
	"log"
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
	var err error
	ConfFile, err = ini.Load("../config/conf.ini")
	if err != nil {
		panic(err)
	}
	log.Println("config 初始化成功,")
	//roomId, err := game.CreateChatRoom("http://quparty.local/app/other/createGroup")
	//fmt.Printf("roomId:%s,err:%s", roomId, err)
}

func TestGetPlayersByRoomId(t *testing.T) {
	var winPlayers []game.Player
	// 获胜狼的列表
	db.Db.Table("fa_game_player p").
		Select("p.user_id,p.room_id,p.role,p.status,count(*) cn").
		Joins("left join fa_game_vote v on v.room_id = p.room_id and v.user_id=p.user_id and v.status=1").
		Where("p.room_id = ? AND p.role = ? AND p.status = ?", 6051, 1, 1).
		Group("user_id").
		Order("cn desc").
		Limit(3).
		Find(&winPlayers)
	log.Printf("players:%+v", winPlayers)
	//fmt.Printf("result:%+v", votes)
}
