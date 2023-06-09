package znet

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
	"websocket/config"
	"websocket/core"
	"websocket/lib/db"
	"websocket/lib/mylog"
	"websocket/model"
	"websocket/utils"
	"websocket/utils/token"
	"websocket/ziface"
)

// iserver的接口实现，定义一个server的服务器模块
type Server struct {
	//服务器的名称
	Name string
	//服务器绑定的IP版本
	IPversion string
	//服务器监听的ip
	IP string
	//Port
	Port string
	//当前的server添加一个router
	//Router ziface.IRouter
	//当前server的消息管理模块，用来绑定MsgID和对应的处理业务api关系
	MsgHandle ziface.IMsgHandle
	//该server的连接管理器
	ConnMgr ziface.IConnManager
	//该Server创建连接之后自动调用Hook函数OnConnStart
	OnConnStart func(conn ziface.Iconnection)
	//该Server销毁连接只求自动调用Hook函数 OnConnStop
	OnConnStop func(conn ziface.Iconnection)
	// Heartbeat checker
	// (心跳检测器)
	hc ziface.IHeartbeatChecker
}

// 启动服务器
func (s *Server) Start() {
	go func() {
		defer utils.CustomError()
		//开启消息队列及worker工作池
		s.MsgHandle.StartWorkerPoll()
		// 获取一个tcp Addr
		//fmt.Println(s.IPversion,fmt.Sprintf("%s:%d",s.IP,s.Port))
		//addr,err:=net.ResolveTCPAddr(s.IPversion,fmt.Sprintf("%s:%d",s.IP,s.Port))
		http.HandleFunc("/", s.wsPage)
		err := http.ListenAndServe(":"+s.Port, nil)
		if err != nil {
			return
		}
		log.Println("Starting application success listen port:" + s.Port)
	}()
}

func (s *Server) wsPage(res http.ResponseWriter, req *http.Request) {
	defer utils.CustomError()
	var conn *websocket.Conn
	//defer conn.Close()
	//如果有客户端连接过来，阻塞会返回
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if err != nil {
		// mylog.Error("conn err:" + err.Error())
		return
	}
	tokenStr := req.Header.Get("token")
	if tokenStr == "" {
		tokenStr = req.FormValue("token")
	}
	userToken, err := token.Get(tokenStr)
	if err != nil {
		mylog.Error("token err:" + err.Error())
		_ = conn.Close()
		return
	}
	cid := userToken.UserId
	user := &model.User{}
	db.Db.Table("fa_user").Where(model.User{Id: cid}).First(user)
	if user.Id == 0 {
		mylog.Error("用户id不正确:" + fmt.Sprintf("%v", cid))
		_ = conn.Close()
		return
	}
	ConnMgr := s.ConnMgr.GetTotalConnections()
	if oldConn, ok := ConnMgr[cid]; ok {
		mylog.Error("当前id已存在，老连接下线:" + fmt.Sprintf("%v", cid))
		oldConn.Stop()
	}
	//设置最大连接个数的判断，如果超过最大连接，那么关闭此新的连接
	var conf = &config.Conf{}
	err = config.ConfFile.Section("conf").MapTo(conf)
	if err != nil {
		mylog.Error("获取配置参数不正确:" + err.Error())
		_ = conn.Close()
		return
	}
	if s.ConnMgr.Len() >= conf.MaxConnect {
		//TODO 给客户端相应一个超出最大连接的错误包
		err := conn.Close()
		if err != nil {
			mylog.Error("Close conn:" + err.Error())
			_ = conn.Close()
			return
		}
		mylog.Error("Connection max" + fmt.Sprintf("%v", s.ConnMgr.Len()))
		_ = conn.Close()
		return
	}
	dealConn := NewConnetion(s, conn, cid, s.MsgHandle)
	// HeartBeat check
	if s.hc != nil {
		// Clone a heart-beat checker from the server side
		heartBeatChecker := s.hc.Clone()
		// Bind current connection
		heartBeatChecker.BindConn(dealConn)
	}
	go dealConn.Start()

}

func (s *Server) Stop() {
	//将一些服务器的资源，状态或者一些已经开辟的链接信息，进行停止回收
	log.Println("[STOP Zinx server name]", s.Name)
	s.ConnMgr.ClearConn()
}

func (s *Server) Server() {
	//启动server 的服务器功能
	s.Start()
	// TODO 做一些启动服务器之后的额外工作
	s.LocationWork()
	//阻塞状态
	select {}
}

// 发送定位消息给用户
func (s *Server) LocationWork() {
	for {
		//for _, conn := range s.GetConnMgr().GetTotalConnections() {
		for _, player := range core.WorldMgrObj.GetAllPlayers() {
			conn := player.Conn
			typeVal, err := conn.GetProperty("type")
			if err != nil {
				continue
			}
			roomType := typeVal.(string)
			userId := conn.GetConnID()
			var message model.SendLocationMsg
			message.MsgId = 201
			message.Type = roomType
			message.UserId = userId
			switch roomType {
			case "1":
				// TODO 获取密友定位
				message.Users = model.GetFriendLocation(userId)
			case "2", "3":
				// TODO 获取活动成员定位
				roomIdVal, err := conn.GetProperty("roomId")
				if err != nil {
					continue
				}
				roomId, err := strconv.Atoi(roomIdVal.(string))
				if err != nil {
					continue
				}
				if roomId == 0 && len(roomIdVal.(string)) > 0 {
					f, err := strconv.ParseFloat(roomIdVal.(string), 64)
					if err != nil {
						log.Println("get roomId err")
						continue
					}
					roomId = int(f)
				}
				if roomType == "2" {
					message.Users = model.GetActivityMemberLocation(roomId, userId)
				} else {
					message.Users = model.GetClubMemberLocation(roomId, userId)
				}
				message.RoomId = roomId
			default:
				time.Sleep(3 * time.Second)
				continue
			}
			marshal, err := json.Marshal(message)
			if err != nil {
				continue
			}
			//if conn, ok := s.GetConnMgr().GetTotalConnections()[userId]; ok {
			//	conn.SendMsg(201, marshal)
			//}
			conn.SendMsg(201, marshal)
		}
		time.Sleep(3 * time.Second)
	}
}

// 路由功能，给当前的服务注册一个路由方法，提供客户端的链接处理使用
func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {
	s.MsgHandle.AddRouter(msgID, router)
	fmt.Println("Add Router Success!")
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

// 初始化Server模块方法
func NewServer(name string) ziface.Iserver {
	var conf = &config.Conf{}
	err := config.ConfFile.Section("conf").MapTo(conf)
	if err != nil {
		mylog.Error("获取配置参数不正确:" + err.Error())
		panic("获取配置参数不正确" + err.Error())
	}
	s := &Server{
		Name:      name,
		IPversion: "tcp4",
		IP:        "0.0.0.0",
		Port:      conf.Port,
		MsgHandle: NewMsgHandle(),
		ConnMgr:   NewConnMamager(),
	}
	return s
}

// 注册OnConnStart 钩子函数的方法
func (s *Server) SetConnStart(hookFunc func(connection ziface.Iconnection)) {
	s.OnConnStart = hookFunc
}

// 注册OnConnStop 钩子函数的方法
func (s *Server) SetConnStop(hookFunc func(connection ziface.Iconnection)) {
	s.OnConnStop = hookFunc
}

// 调用OnConnStart 钩子函数的方法
func (s *Server) CallConnStart(conn ziface.Iconnection) {
	if s.OnConnStart != nil {
		//log.Println("--->Cal OnConnStart()...")
		s.OnConnStart(conn)
	}
}

// 调用OnConnStop 钩子函数的方法
func (s *Server) CallConnStop(conn ziface.Iconnection) {
	if s.OnConnStop != nil {
		//log.Println("--->Cal OnConnStop()...")
		s.OnConnStop(conn)
	}
}

// 启动心跳检测
// (option 心跳检测的配置)
func (s *Server) StartHeartBeatWithOption(interval time.Duration, option *ziface.HeartBeatOption) {
	checker := NewHeartbeatChecker(interval)
	// Configure the heartbeat checker with the provided options
	if option != nil {
		checker.SetHeartbeatMsgFunc(option.MakeMsg)
		checker.SetOnRemoteNotAlive(option.OnRemoteNotAlive)
		checker.BindRouter(option.HeadBeatMsgID, option.Router)
	}
	// Add the heartbeat checker's router to the server's router (添加心跳检测的路由)
	s.AddRouter(checker.MsgID(), checker.Router())
	// Bind the server with the heartbeat checker (server绑定心跳检测器)
	s.hc = checker
}

func (s *Server) GetHeartBeat() ziface.IHeartbeatChecker {
	return s.hc
}
