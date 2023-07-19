# go-websocket

#### 介绍
go语言WebSocket连接框架

#### 启动
```
func main() {
	defer utils.CustomError()
	// 初始化配置
	config.InitConf("")
	//初始化mysql
	db.InitDb()
	//初始化redis
	redis.InitRedis()
	utils.InitGlobalConf()
	//创建server句柄，使用api
	s := service.NewServer("websocket")
	// 心跳
	s.AddRouter(101, &router.PingRouter{})   //添加对应路由
	//注册连接的Hook钩子函数
	s.SetConnStart(DoConnectionBegin)
	s.SetConnStop(DoConnectionLost)

	// Start heartbeating detection. (启动心跳检测)
	s.StartHeartBeatWithOption(5*time.Second, &impl.HeartBeatOption{
		MakeMsg:          myHeartBeatMsg,
		OnRemoteNotAlive: myOnRemoteNotAlive,
		Router:           &myHeartBeatRouter{},
		HeadBeatMsgID:    uint32(100),
	})
	//启动Server
	s.Server()
}
```
### 配置
复制config/conf.ini.example为conf.ini到当前目录

#### linux 脚本其他
```api
./script.sh start
```


#### 参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request

