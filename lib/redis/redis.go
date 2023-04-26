package redis

import (
	"github.com/go-redis/redis/v7"
	"log"
	"websocket/config"
)

var Redis *redis.Client

func InitRedis() {
	var conf = &config.Redis{}
	err := config.ConfFile.Section("redis").MapTo(conf)
	if err != nil {
		panic("获取配置参数不正确")
	}
	Redis = redis.NewClient(&redis.Options{
		Addr:     conf.Host,
		Password: conf.Password, // password set
		DB:       conf.Select,   // use default DB
		PoolSize: conf.PoolSize,
	})
	pong, err := Redis.Ping().Result()
	if err != nil {
		panic("连接redis失败")
	}
	log.Println("Redis 初始化成功,", pong)
}
