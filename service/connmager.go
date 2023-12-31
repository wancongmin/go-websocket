package service

import (
	"errors"
	"log"
	"sync"
	"websocket/impl"
)

// 连接管理模块
type ConnManager struct {
	Connections map[uint32]impl.Iconnection //管理的连接集合
	connLock    sync.RWMutex                //保护连接集合的读写锁
}

// 创建当前连接的方法
//var Managers = ConnManager{}

func NewConnMamager() *ConnManager {
	//Managers.Connections = make(map[uint32]impl.Iconnection)
	//return &Managers
	return &ConnManager{
		Connections: make(map[uint32]impl.Iconnection),
	}
}

// 添加连接
func (c *ConnManager) Add(conn impl.Iconnection) {
	//保护共享资源map，加写锁
	c.connLock.Lock()
	defer c.connLock.Unlock()
	c.Connections[conn.GetConnID()] = conn
	log.Printf("【上线信息】用户:%d,用户数:%d,远程地址:%s,总用户详情:%v", conn.GetConnID(), c.Len(), conn.RemoteAddr(), c.GetTotalConnections())
}

// 删除连接
func (c *ConnManager) Remove(conn impl.Iconnection) {
	//保护共享资源map，加写锁
	c.connLock.Lock()
	defer c.connLock.Unlock()
	delete(c.Connections, conn.GetConnID())
	log.Printf("【下线信息】用户:%d,用户数:%d,远程地址:%s,总用户详情:%v", conn.GetConnID(), c.Len(), conn.RemoteAddr(), c.GetTotalConnections())
}

// 根据connID获取连接
func (c *ConnManager) Get(connID uint32) (impl.Iconnection, error) {
	//保护共享资源map，加读锁
	c.connLock.RLock()
	defer c.connLock.RUnlock()
	if conn, ok := c.Connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connections is not found")
	}
}

// 得到当前链接总数
func (c *ConnManager) Len() uint32 {
	//c.connLock.RLock()
	//defer c.connLock.RUnlock()
	return uint32(len(c.Connections))
}

// 清楚并终止所有的连接
func (c *ConnManager) ClearConn() {
	//保护共享资源map，加写锁
	c.connLock.Lock()
	defer c.connLock.Unlock()
	//删除conn并停止conn的工作
	for connID, conn := range c.Connections {
		//停止
		conn.Stop()
		//删除
		delete(c.Connections, connID)
	}
	log.Println("clear all connectins succ conn nun=", c.Len())
}

func (c *ConnManager) GetTotalConnections() map[uint32]impl.Iconnection {
	//c.connLock.RLock()
	//defer c.connLock.RUnlock()
	return c.Connections
}
