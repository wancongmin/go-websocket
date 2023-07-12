package test

import (
	"github.com/go-ini/ini"
	"log"
	"testing"
	"websocket/config"
	"websocket/lib/db"
	"websocket/lib/redis"
	"websocket/model/comm"
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
	user := comm.GetUserById(588)
	log.Printf("user:%+v", user)
}
