package znet

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"websocket/utils"
	"websocket/ziface"
)

//iserver的接口实现，定义一个server的服务器模块
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

//启动服务器
func (s *Server) Start() {
	//fmt.Printf("start server listenner at IP:%s,Port:%d\n",s.IP,s.Port)
	go func() {
		defer utils.CustomError()
		//开启消息队列及worker工作池
		s.MsgHandle.StartWorkerPoll()
		// 获取一个tcp Addr
		//fmt.Println(s.IPversion,fmt.Sprintf("%s:%d",s.IP,s.Port))
		//addr,err:=net.ResolveTCPAddr(s.IPversion,fmt.Sprintf("%s:%d",s.IP,s.Port))
		fmt.Println("Starting application...")
		http.HandleFunc("/ws", s.wsPage)
		err := http.ListenAndServe(":12345", nil)
		if err != nil {
			return
		}
	}()
}

var cid uint32

func (s *Server) wsPage(res http.ResponseWriter, req *http.Request) {
	defer utils.CustomError()
	//如果有客户端连接过来，阻塞会返回
	//conn,err:=listenner.AcceptTCP()
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if err != nil {
		fmt.Println("Accept err", err)
		return
	}
	uid := req.Header.Get("uid")
	parseInt, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		fmt.Println("get uid err", err)
	}
	cid = uint32(parseInt)
	//设置最大连接个数的判断，如果超过最大连接，那么关闭此新的连接
	if s.ConnMgr.Len() >= 100 {
		//TODO 给客户端相应一个超出最大连接的错误包
		conn.Close()
		fmt.Println("====================>>>>>>>>>>>>>>>>connection max")
		return
	}
	dealConn := NewConnetion(s, conn, cid, s.MsgHandle)
	cid++
	go dealConn.Start()

}

func (s *Server) Stop() {
	//将一些服务器的资源，状态或者一些已经开辟的链接信息，进行停止回收
	fmt.Println("[STOP Zinx server name]", s.Name)
	s.ConnMgr.ClearConn()
}

func (s *Server) Server() {
	//启动server 的服务器功能
	s.Start()
	//TODO 做一些启动服务器之后的额外工作
	//阻塞状态
	select {}
}

//路由功能，给当前的服务注册一个路由方法，提供客户端的链接处理使用
func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {
	s.MsgHandle.AddRouter(msgID, router)
	fmt.Println("Add Router Succ!")
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

//初始化Server模块方法
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

//注册OnConnStart 钩子函数的方法
func (s *Server) SetConnStart(hookFunc func(connection ziface.Iconnection)) {
	s.OnConnStart = hookFunc
}

//注册OnConnStop 钩子函数的方法
func (s *Server) SetConnStop(hookFunc func(connection ziface.Iconnection)) {
	s.OnConnStop = hookFunc
}

//调用OnConnStart 钩子函数的方法
func (s *Server) CallConnStart(conn ziface.Iconnection) {
	if s.OnConnStart != nil {
		fmt.Println("--->Cal OnConnStart()...")
		s.OnConnStart(conn)
	}
}

//调用OnConnStop 钩子函数的方法
func (s *Server) CallConnStop(conn ziface.Iconnection) {
	if s.OnConnStop != nil {
		fmt.Println("--->Cal OnConnStop()...")
		s.OnConnStop(conn)
	}
}
