package main

import (
	"fmt"
	"io"
	"net"
	"time"
	"zinx/znet"
)

/*
模拟客户端
*/
func main() {
	fmt.Println("client start ...")
	time.Sleep(1 * time.Second)
	//直接连接远程服务器，得到一个conn连接
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("client start err, exit!")
		return
	}

	for {
		//封包
		dp := znet.NewDataPack()
		binaryMsg, err := dp.Pack(znet.NewMsgPackage(0, []byte("zinxv0.5 client test message")))
		if err != nil {
			fmt.Println("Pack error: ", err)
			return
		}
		if _, err := conn.Write(binaryMsg); err != nil {
			fmt.Println("write error ", err)
			return
		}
		time.Sleep(1 * time.Second)

		//拆包
		binaryHead := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(conn, binaryHead); err != nil {
			fmt.Println("read head error ", err)
			break
		}
		msgHead, err := dp.Unpack(binaryHead)
		if err != nil {
			fmt.Println("client unpack msgHead error ", err)
			break
		}
		var data []byte
		if msgHead.GetMsgLen() > 0 {
			data = make([]byte, msgHead.GetMsgLen())
			if _, err := io.ReadFull(conn, data); err != nil {
				fmt.Println("client read msg data error ", err)
				return
			}
			msgHead.SetData(data)
			fmt.Println("--->receive server msg : ID = ", msgHead.GetMsgId(), ",data len = ", msgHead.GetMsgLen(), " ,data = ", string(msgHead.GetData()))
		}
	}
}
