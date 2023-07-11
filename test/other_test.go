package test

import (
	"fmt"
	"github.com/go-ini/ini"
	"log"
	"math"
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
	fmt.Println(math.Ceil(5 / 2))
}
