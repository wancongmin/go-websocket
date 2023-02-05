package znet

import "websocket/ziface"

type Request struct {
	//已经和客户端建立好的链接
	conn ziface.Iconnection
	//客户端请求的数据
	msg Message
}

//得到当前链接
func (r *Request) GetConnection() ziface.Iconnection {
	return r.conn
}

//得到当前数据
func (r *Request) GetData() []byte {
	return r.msg.GetData()
}

func (r *Request) GetMsgId() uint32 {
	return r.msg.GetMsgId()
}
