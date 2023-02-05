package config

import (
	"github.com/go-ini/ini"
	"log"
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
