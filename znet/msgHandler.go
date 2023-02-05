package znet

import (
	"fmt"
	"strconv"
	"websocket/ziface"
)

//消息 处理模块的实现

type MsgHandle struct {
	//存放每个MsgId 所对应的处理方法
	Apis map[uint32]ziface.IRouter
	//负责Worker取任务的消息队列
	TaskQueue []chan ziface.IRequest
	//业务工作Worker池的worker数量
	WorkerPoolSize uint32
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: 10,                               //参数配置
		TaskQueue:      make([]chan ziface.IRequest, 10), //参数配置
	}
}

func (m *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	//1 从request中找到msgID
	handler, ok := m.Apis[request.GetMsgId()]
	if !ok {
		fmt.Println("api msgID=", request.GetMsgId(), "is not found")
	}
	//根据msgid 调动对应的routeryew
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

//为消息添加具体的处理逻辑
func (m *MsgHandle) AddRouter(msgID uint32, router ziface.IRouter) {
	//判断 当前msg绑定的api方法是否存在
	if _, ok := m.Apis[msgID]; ok {
		//id已经注册
		panic("repeat api msgID=" + strconv.Itoa(int(msgID)))
	}
	m.Apis[msgID] = router
	fmt.Println("Add api msgid succ! msgID=", msgID)
}

//启动一个Worker工作池（开启工作池的动作只有一次，一个框架只能有一个worker工作池）
func (m *MsgHandle) StartWorkerPoll() {
	for i := 0; i < int(m.WorkerPoolSize); i++ {
		//一个workerPoolSize分别启动Worker,每个Worker用一个go来承载
		m.TaskQueue[i] = make(chan ziface.IRequest, 1024)
		//启动worker，阻塞等待消息充channel传进来
		go m.StratOneWorker(i, m.TaskQueue[i])
	}
}

//启动一个Worker工作流程
func (m *MsgHandle) StratOneWorker(workID int, taskQueue chan ziface.IRequest) {
	fmt.Println("worker id=", workID, "is started...")
	for {
		select {
		case request := <-taskQueue:
			m.DoMsgHandler(request)
		}
	}
}

//将消息交给TaskQueue,由worker进行处理
func (m *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	//1 将消息平均分配给worker
	// 根据客户端建立的ConnID来进行分配
	workerID := request.GetConnection().GetConnID() % m.WorkerPoolSize
	fmt.Println("Add ConnID=", request.GetConnection().GetConnID(), " request MsgID=", request.GetMsgId(), "to workerID", workerID)
	//将消息发送给队友的worker的TaskQueue即可
	m.TaskQueue[workerID] <- request
	
}
