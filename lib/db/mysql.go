package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"websocket/config"
)

var Db *gorm.DB

func InitDb() {
	var conf = &config.Database{}
	err := config.ConfFile.Section("database").MapTo(conf)
	if err != nil {
		panic("获取配置参数不正确")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", conf.User, conf.Password, conf.Host, conf.Name)
	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败")
	}
	log.Println("db 初始化成功,")
}
