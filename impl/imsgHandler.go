package impl

//消息管理抽象层

type IMsgHandle interface {
	//执行对应的Router消息处理方法
	DoMsgHandler(request IRequest)
	AddRouter(msgID uint32, router IRouter)
	//启动worker工作池
	StartWorkerPoll()
	//将消息发送给消息任务队列
	SendMsgToTaskQueue(request IRequest)
}
