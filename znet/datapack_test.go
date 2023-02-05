package znet

import (
	"testing"
	"net"
	"fmt"
	"io"
	"time"
)

//只负责测试datapack拆包 封包的单元测试
func TestDataPack(t *testing.T)  {
	//模拟的服务器
	listenner,err:=net.Listen("tcp","127.0.0.1:7777")
	if err!=nil{
		fmt.Println("server listen err:",err)
		return
	}
	//创建一个go 承载负责从客户端处理业务
	go func() {
		//从客户端读取数据，拆包处理
		for{
			conn,err:=listenner.Accept()
			if err!=nil{
				fmt.Println("server accept error",err)
			}

			go func(conn net.Conn) {
				//处理客户端的请求
				//---->拆包的过程
				//定义一个拆包的对象
				dp:=NewDataPack()
				for{
					//第一次从conn读，把包的head读出来
					headData:=make([]byte,dp.GetHeadLen())
					_,err:=io.ReadFull(conn,headData)
					if err!=nil{
						fmt.Println("read head error",err)
						return
					}
					msgHead,err:=dp.Unpack(headData)
					if err!=nil{
						fmt.Println("server unpacke err",err)
						return
					}

					//第二次从nonn读，根据head中的datalen 再读取data的内容
					if msgHead.GetMsgLen() >0{
						//msg 是有数据的，需要进行第二次读取
						//2 第二次从comm读，根据head中的datalen 再读取data内容
						msg:=msgHead.(*Message)
						msg.Data=make([]byte,msg.GetMsgLen())
						//根据datalen的长度再次从io流中读取
						_,err:=io.ReadFull(conn,msg.Data)
						if err!=nil{
							fmt.Println("server unpack data err:",err)
							return
						}
						//完整的一个消息已经读取完毕
						fmt.Println("-->Recv MsgID:",msg.Id,"msgLen:",msg.DataLen,"data:",string(msg.Data))
					}
				}
			}(conn)
		}
	}()

	//模拟客户端
	conn,err:=net.Dial("tcp","127.0.0.1:7777")
	if err!=nil{
		fmt.Println("client dial err",err)
		return
	}
	//创建一个dp
	dp:=NewDataPack()

	//模拟粘包过程，封装两个msg一同发送
	//封装第一个msg包
	msg1:=&Message{
		Id:1,
		DataLen:2,
		Data:[]byte{'1','a'},
	}
	sendData1,err:=dp.Pack(msg1)
	if err!=nil{
		fmt.Println("client pack msg1 error",err)
		return
	}
	//封装第二个msg包
	msg2:=&Message{
		Id:2,
		DataLen:4,
		Data:[]byte{'n','a','b','x'},
	}
	sendData2,err:=dp.Pack(msg2)
	if err!=nil{
		fmt.Println("client pack msg2 error",err)
		return
	}
	//将两个包粘在一起
	sendData1=append(sendData1,sendData2...)
	//一次性发给服务器
	conn.Write(sendData1)
	//客户端阻塞
	//select {}
	time.Sleep(5*time.Second)
}
