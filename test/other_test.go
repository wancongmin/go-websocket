package test

import (
	"fmt"
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
	m := map[string]*string{}
	a := "aaa"
	b := "bbb"
	m["a"] = &a
	m["b"] = &b
	fmt.Println(m["a"])
	fmt.Println(m["b"])
	fmt.Println(m["c"])
	fmt.Printf("结果1：%+v \n", m["a"])
	fmt.Printf("结果2：%+v \n", m["b"])
	fmt.Printf("结果3：%+v \n", m["c"])
}
