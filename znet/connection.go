package znet

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"sync"
	"websocket/lib/mylog"
	"websocket/model"
	"websocket/utils"
	"websocket/ziface"
)

// 链接模块
type Connection struct {
	//当前Conn属于哪个sever
	TcpSever ziface.Iserver
	//当前链接的socket TCP 套接字
	Conn *websocket.Conn
	//链接ID
	ConnID uint32
	//当前链接状态
	isClose bool
	//当前链接所绑定的处理业务方法API
	//handleAPI ziface.HandleFunc
	//高中当前链接已经退出的/停止的channel
	ExitChan chan bool
	//无缓冲管道，用于读写Goroutime之前的消息通行
	smgChan chan []byte
	//消息的管理MsgID 和对应的处理业务API
	MsgHandler ziface.IMsgHandle
	//链接属性的集合
	property map[string]interface{}
	//保护连接属性的锁
	propertyLock sync.RWMutex
}

//初始化链接模块的方法

func NewConnetion(sever ziface.Iserver, conn *websocket.Conn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpSever:   sever,
		Conn:       conn,
		ConnID:     connID,
		isClose:    false,
		MsgHandler: msgHandler,
		ExitChan:   make(chan bool, 1),
		smgChan:    make(chan []byte),
		property:   make(map[string]interface{}),
	}
	//将conn加入到ConnMananger中
	c.TcpSever.GetConnMgr().Add(c)
	return c
}

func (c *Connection) StartReader() {
	//fmt.Println("[Reader Goruntine is runing]")
	defer utils.CustomError()
	defer c.Stop()
	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			// mylog.Error("read msg error:" + err.Error())
			c.Conn.Close()
			return
		}
		m := model.ReceiveMsg{}
		err = json.Unmarshal(data, &m)
		if err != nil {
			mylog.Error("消息解析json错误:" + err.Error() + "data:" + string(data))
			continue
		}
		if m.MsgId == 0 {
			mylog.Error("获取MsgId不正确")
		}
		// TODO  处理错误
		msg := Message{}
		msg.SetMsgId(m.MsgId)
		msg.SetData(data)
		req := Request{
			conn: c,
			msg:  msg,
		}
		c.MsgHandler.SendMsgToTaskQueue(&req)
		//根据绑定好的MsgID找到对应处理api业务执行
		//go c.MsgHandler.DoMsgHandler(&req)
	}
}

func (c *Connection) Start() {
	//fmt.Println("Connet Start()...ConnID=", c.ConnID)
	//启动当前链接的读数据的业务
	go c.StartReader()
	//TODO 启动从当前链接写数据的业务
	go c.StartWriter()

	//按照开发者传递进来的创建连接之后需要调用的处理业务，执行对应的Hook函数
	c.TcpSever.CallConnStart(c)
}

// 写消息的Goroutime
func (c *Connection) StartWriter() {
	//log.Println("[Writer Gortime is running]")
	defer utils.CustomError()
	//defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")
	//不断的阻塞的等待channel的消息，进行写给客户端
	for {
		select {
		case data := <-c.smgChan:
			//有数据要写给客户端
			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println("Send data error", err)
				return
			}
		case <-c.ExitChan:
			//reader已经退出，此时writer也要退出
			return
		}
	}
}

// 停止链接 结束当前的链接工作
func (c *Connection) Stop() {
	//log.Println("Conn Stop()...ConnID=", c.ConnID)
	//如果当前链接已经关闭
	if c.isClose == true {
		return
	}
	c.isClose = true

	//调用开发者注册的 销毁连接只求，执行对应的Hook函数
	c.TcpSever.CallConnStop(c)
	//关闭socket链接
	c.Conn.Close()
	c.ExitChan <- true
	//将当前连接从ConnMgr中删除
	c.TcpSever.GetConnMgr().Remove(c)
	//回收资源
	close(c.ExitChan)
	close(c.smgChan)
}

// 获取当前链接的绑定 socket conn
func (c *Connection) GetTCPConnection() *websocket.Conn {
	return c.Conn
}

// 获取当前链接模块的链接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端的TCP  状态 IP port
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 发送数据，将数据发送给远程客户端
func (c *Connection) Send(data []byte) error {
	return nil
}

// 提供一个SendMsg方法 将我们要发送给客户端的数据，先进行封包，在发送
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	defer utils.CustomError()
	if c.isClose == true {
		return errors.New("Connection close when send msg")
	}
	//将data进行封包
	//dp:=NewDataPack()
	//binaryMsg,err:=dp.Pack(NewMsgPackage(msgId,data))
	//
	//if err!=nil{
	//	fmt.Println("Pack error msg id=",err)
	//}
	//将数据发送客户端
	//if _,err:=c.Conn.Write(binaryMsg);err!=nil {
	//	fmt.Println("Write msg id=",msgId," erros:",err)
	//	return err
	//}
	//m := model.SendMsg{}
	//m.MsgId = msgId
	//m.Data = data
	//marshal, err := json.Marshal(m)
	//if err != nil {
	//	mylog.Error("发送参数转换json错误:" + err.Error())
	//	return err
	//}
	c.smgChan <- data
	return nil
}

// 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	c.property[key] = value
}

// 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if val, ok := c.property[key]; ok {
		return val, nil
	}
	return nil, errors.New("no property found")
}

// 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	delete(c.property, key)
}
