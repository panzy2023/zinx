package ziface

/*
	封包 拆包
	面向TCP连接中的数据流，用于处理TCP粘包问题
*/
type IDataPack interface {
	//获取包头长度
	GetHeadLen() uint32
	//封包
	Pack(msg IMessage) ([]byte, error)
	//拆包
	Unpack([]byte) (IMessage, error)
}
