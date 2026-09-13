package znet

import (
	"fmt"
	"zinx/utils"
	"zinx/ziface"
)

/*
	消息处理模块
*/

type MsgHandle struct {
	//存放每个msgID对应的处理方法
	Apis map[uint32]ziface.IRouter
	//负责worker取任务的消息队列
	TaskQueue []chan ziface.IRequest
	//业务工作worker池的worker数量
	WorkerPoolSize uint32
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		TaskQueue:      make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

// 调度执行对应的Router消息处理方法
func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgID = ", request.GetMsgID(), " is not found need registe")
	}
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgID uint32, router ziface.IRouter) {
	if _, ok := mh.Apis[msgID]; ok {
		fmt.Println("repeat api , msgID = ", msgID)
		return
	}
	mh.Apis[msgID] = router
	fmt.Println("Add api  msgID = ", msgID)

}

// 启动一个Worker工作池
func (mh *MsgHandle) StartWorkerPool() {
	//根据WorkerPoolSize分别开启Worker
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//启动一个worker
		//当前worker对应channel消息队列 开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		//启动当前的worker，阻塞等待消息从channel传进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// 启动一个Worker工作流程
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("workerID = ", workerID, " is started...")
	//不断阻塞等待对应消息队列的消息
	for {
		select {
		//如果有消息过来，出列的就是第一个客户端的Request，执行当前Request绑定的业务
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// 将消息交给taskqueue
func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	//将消息平均分配给不通过的worker
	//根据客户端建立的ConnID分配
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	fmt.Println("Add ConnID = ", request.GetConnection().GetConnID(),
		" request MsgID = ", request.GetMsgID(), " to WorkerID = ", workerID)

	//将消息发送给对应的worker的对应TaskQueue
	mh.TaskQueue[workerID] <- request
}
