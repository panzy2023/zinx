package znet

import (
	"fmt"
	"io"
	"net"
	"testing"
)

func TestDataPack(t *testing.T) {
	/*
		模拟的服务器
	*/
	//创建socketTCP
	listenner, err := net.Listen("tcp", "127.0.0.1:8777")
	if err != nil {
		fmt.Println("server listen err: ", err)
		return
	}
	go func() {
		//从客户端读取数据，拆包处理
		for {
			conn, err := listenner.Accept()
			if err != nil {
				fmt.Println("server accept error", err)
			}
			go func(conn net.Conn) {
				//处理客户端请求
				//拆包过程
				dp := NewDataPack()
				for {
					//第一次从conn读，读包的head
					headData := make([]byte, dp.GetHeadLen())
					if _, err := io.ReadFull(conn, headData); err != nil {
						fmt.Println("read head error")
						break
					}
					msgHead, err := dp.Unpack(headData)
					if err != nil {
						fmt.Println("server unpack err ", err)
						return
					}
					if msgHead.GetMsgLen() > 0 {
						//第二次根据head中的datalen 读data内容
						msg := msgHead.(*Message)
						msg.Data = make([]byte, msg.GetMsgLen())
						_, err := io.ReadFull(conn, msg.Data)
						if err != nil {
							fmt.Println("server unpack data err: ", err)
							return
						}
						fmt.Println("---> Receive MsgID: ", msg.Id, ", datalen= ", msg.DataLen, ",data = ", string(msg.Data))
					}

				}
			}(conn)
		}
	}()

	//模拟客户端
	conn, err := net.Dial("tcp", "127.0.0.1:8777")
	if err != nil {
		fmt.Println("client dial err:", err)
		return
	}

	//创建一个封包对象
	dp := NewDataPack()
	//封包1
	msg1 := &Message{
		Id:      1,
		DataLen: 4,
		Data:    []byte{'z', 'i', 'n', 'x'},
	}
	sendData1, err := dp.Pack(msg1)
	if err != nil {
		fmt.Println("client pack msg1 error ", err)
		return
	}
	//封包2
	msg2 := &Message{
		Id:      1,
		DataLen: 4,
		Data:    []byte{'h', 'e', 'l', 'o'},
	}
	sendData2, err := dp.Pack(msg2)
	if err != nil {
		fmt.Println("client pack msg2 error ", err)
		return
	}
	//两个包粘到一起
	sendData1 = append(sendData1, sendData2...)
	//一次发送给服务器
	conn.Write(sendData1)

	select {}
}
