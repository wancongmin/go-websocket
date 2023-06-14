package service

import (
	"log"
	"time"
	"websocket/impl"
)

type HeartbeatChecker struct {
	interval         time.Duration         //  Heartbeat detection interval(心跳检测时间间隔)
	quitChan         chan bool             // Quit signal(退出信号)
	makeMsg          impl.HeartBeatMsgFunc //User-defined heartbeat message processing method(用户自定义的心跳检测消息处理方法)
	onRemoteNotAlive impl.OnRemoteNotAlive //  User-defined method for handling remote connections that are not alive (用户自定义的远程连接不存活时的处理方法)
	msgID            uint32                // Heartbeat message ID(心跳的消息ID)
	router           impl.IRouter          // User-defined heartbeat message business processing router(用户自定义的心跳检测消息业务处理路由)
	conn             impl.Iconnection      // Bound connection(绑定的链接)

	beatFunc impl.HeartBeatFunc // // User-defined heartbeat sending function(用户自定义心跳发送函数)
}

/*
Default callback routing business for receiving remote heartbeat messages
(收到remote心跳消息的默认回调路由业务)
*/
type HeatBeatDefaultRouter struct {
	BaseRouter
}

func (r *HeatBeatDefaultRouter) Handle(req impl.IRequest) {
	//zlog.Ins().InfoF("Recv Heartbeat from %s, MsgID = %+v, Data = %s",
	//	req.GetConnection().RemoteAddr(), req.GetMsgID(), string(req.GetData()))
	log.Println("HeatBeatDefaultRouter")
}

func HeatBeatDefaultHandle(req impl.IRequest) {
	//zlog.Ins().InfoF("Recv Heartbeat from %s, MsgID = %+v, Data = %s",
	//	req.GetConnection().RemoteAddr(), req.GetMsgID(), string(req.GetData()))
	log.Println("HeatBeatDefaultHandle")
}

func makeDefaultMsg(conn impl.Iconnection) []byte {
	//msg := fmt.Sprintf("heartbeat [%s->%s]", conn.LocalAddr(), conn.RemoteAddr())
	log.Println("makeDefaultMsg")
	return []byte("makeDefaultMsg")
}

func notAliveDefaultFunc(conn impl.Iconnection) {
	//zlog.Ins().InfoF("Remote connection %s is not alive, stop it", conn.RemoteAddr())
	log.Println("notAliveDefaultFunc")
	conn.Stop()
}

func NewHeartbeatChecker(interval time.Duration) impl.IHeartbeatChecker {
	heartbeat := &HeartbeatChecker{
		interval: interval,
		quitChan: make(chan bool),
		// Use default heartbeat message generation function and remote connection not alive handling method
		// (均使用默认的心跳消息生成函数和远程连接不存活时的处理方法)
		makeMsg:          makeDefaultMsg,
		onRemoteNotAlive: notAliveDefaultFunc,
		msgID:            impl.HeartBeatDefaultMsgID,
		router:           &HeatBeatDefaultRouter{},
		beatFunc:         nil,
	}

	return heartbeat
}

func (h *HeartbeatChecker) SetOnRemoteNotAlive(f impl.OnRemoteNotAlive) {
	if f != nil {
		h.onRemoteNotAlive = f
	}
}

func (h *HeartbeatChecker) SetHeartbeatMsgFunc(f impl.HeartBeatMsgFunc) {
	if f != nil {
		h.makeMsg = f
	}
}

func (h *HeartbeatChecker) SetHeartbeatFunc(beatFunc impl.HeartBeatFunc) {
	if beatFunc != nil {
		h.beatFunc = beatFunc
	}
}

func (h *HeartbeatChecker) BindRouter(msgID uint32, router impl.IRouter) {
	if router != nil {
		h.msgID = msgID
		h.router = router
	}
}

// 开启心跳检测定时器
func (h *HeartbeatChecker) start() {
	ticker := time.NewTicker(h.interval)
	for {
		select {
		case <-ticker.C:
			h.check()
		case <-h.quitChan:
			ticker.Stop()
			return
		}
	}
}

// 心跳检测
func (h *HeartbeatChecker) check() (err error) {
	if h.conn == nil {
		return nil
	}
	if !h.conn.IsAlive() {
		h.onRemoteNotAlive(h.conn)
	} else {
		if h.beatFunc != nil {
			err = h.beatFunc(h.conn)
		} else {
			err = h.SendHeartBeatMsg()
		}
	}

	return err
}

func (h *HeartbeatChecker) Start() {
	go h.start()
}

func (h *HeartbeatChecker) Stop() {
	go func() {
		h.quitChan <- true
	}()
}

func (h *HeartbeatChecker) SendHeartBeatMsg() error {

	msg := h.makeMsg(h.conn)

	err := h.conn.SendMsg(h.msgID, msg)
	if err != nil {
		//zlog.Ins().ErrorF("send heartbeat msg error: %v, msgId=%+v msg=%+v", err, h.msgID, msg)
		return err
	}

	return nil
}

func (h *HeartbeatChecker) BindConn(conn impl.Iconnection) {
	h.conn = conn
	conn.SetHeartBeat(h)
}

// Clone clones to a specified connection
// (克隆到一个指定的链接上)
func (h *HeartbeatChecker) Clone() impl.IHeartbeatChecker {

	heartbeat := &HeartbeatChecker{
		interval:         h.interval,
		quitChan:         make(chan bool),
		beatFunc:         h.beatFunc,
		makeMsg:          h.makeMsg,
		onRemoteNotAlive: h.onRemoteNotAlive,
		msgID:            h.msgID,
		router:           h.router,
		conn:             nil, // The bound connection needs to be reassigned
	}

	return heartbeat
}

func (h *HeartbeatChecker) MsgID() uint32 {
	return h.msgID
}

func (h *HeartbeatChecker) Router() impl.IRouter {
	return h.router
}
