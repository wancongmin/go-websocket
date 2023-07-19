# go-websocket

#### 介绍
go语言WebSocket连接框架
框架开发初衷，由于公司业务需求，需要开发一款小型MMO游戏，考虑过用市面上成熟的框架，但是框架太庞大，学习和使用成本过高,
所以自己弄了这款websocket框架去开发公司的小游戏项目
本框架适用范围
1）小型websocket游戏
2）会话聊天业务


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

### 鸣谢，本项目参考zinx，做了很多简化，更适合新人上手

#### 参与贡献
1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request

