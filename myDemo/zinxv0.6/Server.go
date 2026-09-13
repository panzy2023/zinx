package main

import (
	"fmt"
	"zinx/ziface"
	"zinx/znet"
)

/*
	基于zinx框架开发的服务器端应用程序
*/

// ping test
type PingRouter struct {
	znet.BaseRouter
}

func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call Router Handler")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("receive from client: msgID = ", request.GetMsgID(),
		", data = ", string(request.GetData()))
	err := request.GetConnection().SendMsg(200, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println(err)
	}
}

// hello
type HelloZinxRouter struct {
	znet.BaseRouter
}

func (this *HelloZinxRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call Router Handler")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("receive from client: msgID = ", request.GetMsgID(),
		", data = ", string(request.GetData()))
	err := request.GetConnection().SendMsg(201, []byte("hello zinx!!!"))
	if err != nil {
		fmt.Println(err)
	}
}
func main() {
	//创建一个server句柄，使用zinx的api
	s := znet.NewServer("[zinx V0.3]")
	//添加router
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloZinxRouter{})
	//启动server
	s.Server()
}
