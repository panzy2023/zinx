package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"zinx/ziface"
)

type GlobalObj struct {
	TcpServer ziface.IServer
	Host      string
	TcpPort   int
	Name      string

	Version          string //当前zinx版本号
	MaxConn          int    //当前服务器主机允许的最大链接数
	MaxPackageSize   uint32 //当前zinx框架数据包的最大值
	WorkerPoolSize   uint32 //当前业务工作Worker池的Goroutine数量
	MaxWorkerTaskLen uint32 //zinx框架允许用户最多开辟多少个
}

var GlobalObject *GlobalObj

func (g *GlobalObj) Reload() {
	data, err := os.ReadFile("S:/goproject/zinx/conf/zinx.json")
	if err != nil {
		fmt.Println("警告：未读取到conf/zinx.json，使用默认配置, err:", err)
		return
	}
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		fmt.Println("警告：json解析失败，使用默认配置, err:", err)
		return
	}
}

func init() {
	GlobalObject = &GlobalObj{
		Name:             "ZinxServerApp",
		Version:          "V0.4",
		TcpPort:          8999,
		Host:             "0.0.0.0",
		MaxConn:          1000,
		MaxPackageSize:   4096,
		WorkerPoolSize:   10,   //Worker工作池的队列的个数
		MaxWorkerTaskLen: 1024, //每个worker对应的消息队列的任务的数量最大值
	}
	GlobalObject.Reload()
}
