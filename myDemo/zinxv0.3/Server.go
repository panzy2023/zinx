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

func (this *PingRouter) PreHandle(request ziface.IRequest) {
	fmt.Println("Call Router PreHandler")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("befer ping"))
	if err != nil {
		fmt.Println("call back before ping error")
	}
}
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call Router Handler")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte(" ping ping "))
	if err != nil {
		fmt.Println("call back ping ping error")
	}
}
func (this *PingRouter) PostHandle(request ziface.IRequest) {
	fmt.Println("Call PostRouter Handler")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("after ping"))
	if err != nil {
		fmt.Println("call back after ping error")
	}
}

func main() {
	//创建一个server句柄，使用zinx的api
	s := znet.NewServer("[zinx V0.3]")
	//添加router
	s.AddRouter(&PingRouter{})
	//启动server
	s.Server()
}
