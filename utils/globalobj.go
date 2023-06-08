package utils

import (
	"time"
	"websocket/config"
	"websocket/lib/mylog"
)

type GlobalConf struct {
	HeartbeatMax time.Duration
}

var GlobalObject *GlobalConf

func InitGlobalConf() {
	var conf = &config.Base{}
	err := config.ConfFile.Section("base").MapTo(conf)
	if err != nil {
		mylog.Error("获取配置参数不正确:" + err.Error())
		panic("获取配置参数不正确" + err.Error())
	}
	GlobalObject = &GlobalConf{
		HeartbeatMax: conf.HeartbeatMax,
	}
}
