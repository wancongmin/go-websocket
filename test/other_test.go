package test

import (
	"github.com/go-ini/ini"
	"log"
	"testing"
)

var ConfFile *ini.File

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
	var roleOneNum, roleTowNum int
	roleOneNum++
	roleOneNum++
	roleTowNum++
	log.Println(roleOneNum)
	log.Println(roleTowNum)
}
