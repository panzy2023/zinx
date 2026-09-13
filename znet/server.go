package znet

import (
	"fmt"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

// iServer的接口实现，定义一个Server的服务器模块
type Server struct {
	//服务器名称
	Name string
	//服务器绑定的IP版本
	IPVersion string
	//服务器监听的IP
	IP string
	//服务器监听的端口
	Port int
	//当前server的消息管理模块，用来绑定MsgID和对应的处理业务API关系
	MsgHandle ziface.IMsgHandle
	//该server的连接管理器
	ConnMgr ziface.IConnManager

	//该server创建连接之后自动调用Hook函数OnConnStart
	OnConnStart func(conn ziface.IConnection)
	//该server销毁连接之前自动调用Hook函数OnConnStop
	OnConnStop func(conn ziface.IConnection)
}

func (s *Server) Start() {
	fmt.Printf("[zinx] server name : %s, listenner at IP : %s, port:%d is starting\n", utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)
	fmt.Printf("[zinx] version %s, MaxConn:%d, MaxPackageSize:%d\n", utils.GlobalObject.Version, utils.GlobalObject.MaxConn, utils.GlobalObject.MaxPackageSize)
	go func() {
		//开启消息队列和Worker工作池
		s.MsgHandle.StartWorkerPool()
		//获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr error: ", err)
			return
		}
		//监听服务器的地址
		listenner, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen ", s.IPVersion, " err ", err)
			return
		}
		fmt.Println("start zinx server success ", s.Name, ", Listenning...")
		var cid uint32
		cid = 0
		//阻塞等待客户端链接，处理客户端链接业务
		for {
			//如果有客户端连接过来，阻塞会返回
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err: ", err)
				continue
			}

			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				fmt.Println("too many connections")
				conn.Close()
				continue
			}

			//将处理新连接的业务方法和conn进行绑定 得到链接模块
			dealConn := NewConnection(s, conn, cid, s.MsgHandle)
			cid++

			//启动当前的连接业务处理
			go dealConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("[stop] zinx server name ", s.Name)
	s.ConnMgr.ClearConn()
}

func (s *Server) Server() {
	//启动server的服务功能
	s.Start()
	//阻塞状态
	select {}
}

func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {
	s.MsgHandle.AddRouter(msgID, router)
	fmt.Println("Add Router Success!")
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

func NewServer(name string) ziface.IServer {
	return &Server{
		Name:      utils.GlobalObject.Name,
		IPVersion: "tcp4",
		IP:        utils.GlobalObject.Host,
		Port:      utils.GlobalObject.TcpPort,
		MsgHandle: NewMsgHandle(),
		ConnMgr:   NewConnManager(),
	}
}
func (s *Server) SetOnConnStart(hookFunc func(conn ziface.IConnection)) {
	s.OnConnStart = hookFunc
}
func (s *Server) SetOnConnStop(hookFunc func(conn ziface.IConnection)) {
	s.OnConnStop = hookFunc
}
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("----->Call OnConnStart()...")
		s.OnConnStart(conn)
	}
}
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("----->Call OnConnStop()...")
		s.OnConnStop(conn)
	}
}
