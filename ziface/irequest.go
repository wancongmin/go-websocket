package ziface

/*
实际是吧客户请求的链接信息，和请求的数据包装到了一个Request中
*/

type IRequest interface {
	//得到当前链接
	GetConnection() Iconnection
	//得到当前的消息数据
	GetData() []byte

	GetMsgId() uint32
}