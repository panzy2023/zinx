package ziface

/*
	消息管理抽象层
*/

type IMsgHandle interface {
	DoMsgHandler(request IRequest)

	AddRouter(msgID uint32, router IRouter)

	//启动Worker工作池
	StartWorkerPool()

	SendMsgToTaskQueue(request IRequest)
}
