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
	"websocket/lib/db"
	"websocket/lib/mylog"
	"websocket/model"
	"websocket/utils"
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
	Port int
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
		var conf = &config.Conf{}
		err := config.ConfFile.Section("conf").MapTo(conf)
		if err != nil {
			mylog.Error("获取配置参数不正确:" + err.Error())
			return
		}
		http.HandleFunc("/", s.wsPage)
		err = http.ListenAndServe(":"+conf.Port, nil)
		if err != nil {
			return
		}
		log.Println("Starting application success listen port:" + conf.Port)
	}()
}

var cid uint32

func (s *Server) wsPage(res http.ResponseWriter, req *http.Request) {
	defer utils.CustomError()
	var conn *websocket.Conn
	//如果有客户端连接过来，阻塞会返回
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if err != nil {
		mylog.Error("Accept err:" + err.Error())
		_ = conn.Close()
		return
	}
	uid := req.Header.Get("uid")
	if uid == "" {
		uid = req.FormValue("uid")
	}
	parseInt, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		mylog.Error("Get uid err:" + err.Error())
		_ = conn.Close()
		return
	}
	cid = uint32(parseInt)
	user := &model.User{}
	db.Db.Table("fa_user").Where(model.User{Id: cid}).First(user)
	if user.Id == 0 {
		mylog.Error("用户id不正确:" + fmt.Sprintf("%v", cid))
		_ = conn.Close()
		return
	}
	ConnMgr := s.ConnMgr.GetTotalConnections()
	if _, ok := ConnMgr[cid]; ok {
		mylog.Error("当前id已存在，不可重复登录:" + fmt.Sprintf("%v", cid))
		_ = conn.Close()
		return
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
	cid++
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
		for _, conn := range s.ConnMgr.GetTotalConnections() {
			typeVal, err := conn.GetProperty("type")
			if err != nil {
				continue
			}
			roomType := typeVal.(string)
			userId := conn.GetConnID()
			var message model.SendLocationMsg
			message.MsgId = 201
			message.Type = roomType
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
				log.Println(roomId)
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
	s := &Server{
		Name:      name,
		IPversion: "tcp4",
		IP:        "0.0.0.0",
		Port:      8091,
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
		log.Println("--->Cal OnConnStart()...")
		s.OnConnStart(conn)
	}
}

// 调用OnConnStop 钩子函数的方法
func (s *Server) CallConnStop(conn ziface.Iconnection) {
	if s.OnConnStop != nil {
		log.Println("--->Cal OnConnStop()...")
		s.OnConnStop(conn)
	}
}
