package config

import (
	"github.com/go-ini/ini"
	"log"
	"time"
)

var ConfFile *ini.File

func InitConf() {
	var err error
	ConfFile, err = ini.Load("config/conf.ini")
	if err != nil {
		panic("配置加载失败")
	}
	log.Println("config 初始化成功,")
}

type Conf struct {
	Port       string
	MaxConnect uint32
	OssUrl     string
}

type Database struct {
	Type     string
	User     string
	Password string
	Host     string
	Name     string
}

type Redis struct {
	Host     string
	Password string
	Select   int
	PoolSize int
}

type Token struct {
	Type   string
	Key    string
	Expire int
}

type Base struct {
	MapLocationExpire time.Duration
	HeartbeatMax      time.Duration
}
