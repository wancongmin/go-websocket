package impl

/*
封包，拆包模块
直接面向tcp链接中的数据流，用于处理tcp粘包的问题
*/
type IDataPack interface {
	//获取包的头的长度方法
	GetHeadLen() uint32
	//封包方法
	Pack(msg IMessage) ([]byte, error)
	//拆包方法
	Unpack([]byte) (IMessage, error)
}
