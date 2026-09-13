package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"zinx/utils"
	"zinx/ziface"
)

/*
连接模块
*/
type Connection struct {
	//当前Conn属于哪个Server
	TcpServer ziface.IServer
	//当前连接的socket TCP套接字
	Conn *net.TCPConn
	//链接的ID
	ConnID uint32
	//当前的链接状态
	isClosed bool
	//告知当前连接已经退出的/停止channel
	ExitChan chan bool
	//无缓冲管道 读写共routine之间的消息通信
	msgChan chan []byte
	//该链接处理的方法
	MsgHandle ziface.IMsgHandle

	//链接属性集合
	property map[string]interface{}
	//保护链接属性的锁
	propertyLock sync.RWMutex
}

// 初始化链接模块
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer: server,
		Conn:      conn,
		ConnID:    connID,
		MsgHandle: msgHandler,
		isClosed:  false,
		msgChan:   make(chan []byte),
		ExitChan:  make(chan bool, 1),
		property:  make(map[string]interface{}),
	}
	//将conn加入ConnManager
	c.TcpServer.GetConnMgr().Add(c)

	return c
}

// 连接的读业务方法
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running...")
	defer fmt.Println("connID = ", c.ConnID, " Reader is exit, remote addr is ", c.RemoteAddr().String())
	defer c.Stop()

	for {
		//创建一个拆包对象
		dp := NewDataPack()
		//读取客户端的msg head 二进制流8个字节
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error", err)
			break
		}
		//拆包，得到msgID和msgDatalen放在msg消息中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			break
		}
		//根据datalen 再次读取data，放在msg.Data
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				break
			}
		}
		msg.SetData(data)
		//得到当前conn的Request请求数据
		req := Request{
			conn: c,
			msg:  msg,
		}

		if utils.GlobalObject.WorkerPoolSize > 0 {
			c.MsgHandle.SendMsgToTaskQueue(&req)
		} else {
			go c.MsgHandle.DoMsgHandler(&req)
		}

	}
}

/*
写消息goroutine，发送给客户端消息的模块
*/
func (c *Connection) StartWriter() {
	fmt.Println("[writer goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), " [conn writer exit]")

	//不断阻塞等待channel的消息，写给客户端
	for {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("send data error, ", err)
				return
			}
		case <-c.ExitChan:
			return
		}
	}
}

// 启动连接
func (c *Connection) Start() {
	fmt.Println("Conn Start()...ConnID = ", c.ConnID)
	//启动从当前连接的读数据的业务
	go c.StartReader()
	go c.StartWriter()

	//按照开发者传递进来的 创建连接之后需要调用的处理业务，执行对应Hook函数
	c.TcpServer.CallOnConnStart(c)
}

// 停止连接
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()...ConnID = ", c.ConnID)
	//如果当前连接已关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	//调用销毁链接之前需要执行的业务hook函数
	c.TcpServer.CallOnConnStop(c)

	//关闭socket链接
	c.Conn.Close()
	//告知writer关闭
	c.ExitChan <- true

	c.TcpServer.GetConnMgr().Remove(c)
	close(c.ExitChan)
	close(c.msgChan)
}

// 获取当前连接的绑定socket conn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// 获取当前连接模块的连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端的TCP状态 IP port
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 提供一个SendMsg方法 将我们要发送给客户端的数据，先进性封包，再发送
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("connection closed when send msg")
	}
	//将data进行封包
	dp := NewDataPack()
	binaryMsg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("pack error msg id = ", msgId)
		return errors.New("pack error msg")
	}
	//将数据发送给客户端
	c.msgChan <- binaryMsg
	return nil
}

func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
