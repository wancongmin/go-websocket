package ziface

// 连接管理模块抽象层
type IConnManager interface {
	// 添加连接
	Add(conn Iconnection)
	// 删除连接
	Remove(conn Iconnection)
	// 根据connID获取连接
	Get(connID uint32) (Iconnection, error)
	// 得到当前链接总数
	Len() uint32
	// 清楚并终止所有d连接
	ClearConn()
	// 获取所有连接
	GetTotalConnections() map[uint32]Iconnection
}
