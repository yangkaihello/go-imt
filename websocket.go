package imt

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

const (
	WEBSOCKET_FRAME_FINNOT = 0x00	//0 描述符0表示未完结 1
	WEBSOCKET_FRAME_FINEND = 0x01 //0 描述符1表示结尾 1
	WEBSOCKET_RSV1 = 0x00	//1 1
	WEBSOCKET_RSV2 = 0x00 //2 1
	WEBSOCKET_RSV3 = 0x00 //3 1
	WEBSOCKET_OPCODE_APPEND = 0x00 //4-7 附加数据帧 1
	WEBSOCKET_OPCODE_TEXT = 0x01 //4-7 文本数据帧 1
	WEBSOCKET_OPCODE_BINARY = 0x02 //4-7 二进制数据帧 1
	WEBSOCKET_OPCODE_TROUGH_3 = 0x03 //4-7 预留 1
	WEBSOCKET_OPCODE_TROUGH_4 = 0x04 //4-7 预留 1
	WEBSOCKET_OPCODE_TROUGH_5 = 0x05 //4-7 预留 1
	WEBSOCKET_OPCODE_TROUGH_6 = 0x06 //4-7 预留 1
	WEBSOCKET_OPCODE_TROUGH_7 = 0x07 //4-7 预留 1
	WEBSOCKET_OPCODE_CLOSE = 0x08 //4-7 连接关闭 1
	WEBSOCKET_OPCODE_PING = 0x09 //4-7 ping 1
	WEBSOCKET_OPCODE_PONG = 0x0A //4-7 pong 1
	WEBSOCKET_OPCODE_TROUGH_B = 0x0B //4-7 预留 1
	WEBSOCKET_OPCODE_TROUGH_C = 0x0C //4-7 预留 1
	WEBSOCKET_OPCODE_TROUGH_D = 0x0D //4-7 预留 1
	WEBSOCKET_OPCODE_TROUGH_E = 0x0E //4-7 预留 1
	WEBSOCKET_OPCODE_TROUGH_F = 0x0F //4-7 预留 1
	WEBSOCKET_MASK_NOT = 0x00 //8 掩码关闭 2
	WEBSOCKET_MASK_OPEN = 0x01 //8 掩码开启 2
	WEBSOCKET_PAYLEN7 = 0x7D //9 实际传输数据长度len <= 125 2
	WEBSOCKET_PAYLEN16 = 0x7E //9 实际传输数据长度125 < len <= 65535 2
	WEBSOCKET_PAYLEN64 = 0x7F //9 实际传输数据长度65535 < len 2
)

type webSocket struct {
	data string
	header map[string]string
	response map[string]string
	responseLineOne string
	controllerReturn *ControllerReturn
	Conn *Connect
}

func (this *webSocket) Init(connect *Connect) *webSocket {
	this.Conn = connect
	this.header = map[string]string{}

	//websocket头部固定协议
	this.responseLineOne = "HTTP/1.1 101 Switching Protocols"
	this.response = map[string]string{
		"Connection":"Upgrade",
		"Upgrade":"websocket",
		"Sec-WebSocket-Accept":"websocket",
	}
	return this
}

func (this *webSocket) SetHeader(key string,value string) {
	this.header[key] = value
}

func (this *webSocket) GetHeaderAll() map[string]string {
	return this.header
}

func (this *webSocket) GetHeader(key string) string {
	return this.header[key]
}

func (this *webSocket) GetBody() *MessageHeader {
	var message = new(MessageHeader)
	if err := json.Unmarshal([]byte(this.data),message); err != nil {
		return nil
	}
	return message
}

func (this *webSocket) SetBody(s string) {
	this.data = strings.Trim(s,"")
}

func (this *webSocket) SetResponse(key string,value string) {
	this.response[key] = value
}

func (this *webSocket) GetResponse() string {
	var responses = []string{
		this.responseLineOne,
	}
	for k,v := range this.response {
		responses = append(responses, k+": "+v)
	}

	return strings.Join(responses,"\r\n")+"\r\n\r\n"
}

func (this *webSocket) GetOneSelfName() string {
	return reflect.TypeOf(this).Name()
}

func (this *webSocket) SendMessage(s string) error {
	var err error

	var data = this.FrameEncode([]byte(s))

	if err = this.Conn.SendString(string(data)); err != nil {
		return err
	}

	return nil
}

func (this *webSocket) ReadMessage() (string,error) {
	var data string
	var err error
	var frame []byte
	//获取资源数据
	if  data,err = this.Conn.ReadString(); err != nil {
		return "",err
	}
	//websocket数据解码
	if frame,err = this.FrameDecode([]byte(data)); err != nil {
		return "",err
	}
	data = string(frame)

	return data,nil
}

func (this *webSocket) SendStruct(m Message) error {
	var data []byte
	var err error

	if data,err = json.Marshal(m); err != nil {
		return err
	}

	data = this.FrameEncode(data)

	if err = this.Conn.SendString(string(data)); err != nil {
		return err
	}
	return nil
}

func (this *webSocket) ReadStruct(v interface{}) (interface{},error) {
	var data string
	var err error
	var frame []byte

	//获取资源数据
	if  data,err = this.Conn.ReadString(); err != nil {
		return "",err
	}

	//websocket数据解码
	if frame,err = this.FrameDecode([]byte(data)); err != nil {
		return "",err
	}
	data = string(frame)

	if err = json.Unmarshal([]byte(data)[:len(data)-1],v); err != nil {
		return nil,err
	}

	return v,nil
}

func (this *webSocket) SetReturn(controllerReturn *ControllerReturn) Protocol {
	this.controllerReturn = controllerReturn
	this.SendMessage(this.controllerReturn.Data)
	return this
}

func (this *webSocket) GetReturn() *ControllerReturn {
	return this.controllerReturn
}

func (this *webSocket) FrameDecode(content []byte) ([]byte,error) {
	var lenByteNumber int

	if len(content) <= 2 {
		return nil,errors.New("解码数据异常确认是websocket帧数据")
	}

	//头部数据
	_ = content[0]

	//step one
	//验证字符数据长度
	var b byte = content[1]
	var Paylen = b - 0x80
	if Paylen <= WEBSOCKET_PAYLEN7 {
		lenByteNumber = 0
	}else if Paylen == WEBSOCKET_PAYLEN16 {
		lenByteNumber = 2
	}else if Paylen == WEBSOCKET_PAYLEN64 {
		lenByteNumber = 8
	}

	//step two
	//数据字节统计长度 paylen+bytelen【暂时无用后期可以用作验证数据的完整性】
	var ReaderLength = content[1:lenByteNumber+2]
	//数据长度
	var DataLength int64
	if len(ReaderLength) == 1 {
		DataLength = int64(Paylen)
	}else{
		var product = map[int]int64{
			0:256,
			1:256*256,
			2:256*256*256,
			3:256*256*256*256,
			4:256*256*256*256*256,
			5:256*256*256*256*256*256,
			6:256*256*256*256*256*256*256,
		}
		//统计数据长度
		for i:=len(ReaderLength)-2 ; i <= 0 ; i-- {
			DataLength += product[len(ReaderLength)-2-i]*int64(ReaderLength[i+1])
		}
		DataLength += int64(ReaderLength[len(ReaderLength)-1])
	}
	var MaskData = content[lenByteNumber+2:]
	//step three
	//是否解码，数据最终获取
	var data []byte
	b = content[1]
	//是否需要解码
	if b >> 7 == 1 {
		var mask = MaskData[:4]
		//解码
		for k,v := range MaskData[4:] {
			data = append(data, v^mask[k%4])
		}
	}else{
		data = MaskData[4:]
	}
	return data,nil

}

func (this *webSocket) FrameEncode(content []byte) []byte {
	var header = []byte{
		WEBSOCKET_FRAME_FINEND<<7+WEBSOCKET_OPCODE_TEXT, //header中的第一位,FIN+RSV+OPCODE
	}

	//计算数据长度 MASK+Payload len
	var ContentLength = len(content)
	//开启mask验证
	var PayLen = []byte{WEBSOCKET_MASK_NOT<<7}
	//计算Payload len的长度
	if ContentLength <= WEBSOCKET_PAYLEN7 {
		PayLen[0] += byte(ContentLength)
	} else if ContentLength <= 0xFFFF {
		PayLen[0] += WEBSOCKET_PAYLEN16
		PayLen = append(PayLen, byte((ContentLength&0xFF00)>>8),byte(ContentLength&0xFF))
	} else {
		PayLen[0] += WEBSOCKET_PAYLEN64
		PayLen = append(PayLen, byte((ContentLength&0xFF000000)>>24),byte((ContentLength&0xFF0000)>>16),byte((ContentLength&0xFF00)>>8),byte(ContentLength&0xFF))
	}
	header = append(header, PayLen...)

	//最后用来加密的4字节,服务器返回无需加密
	/*var key = []byte{
		0x01,0x02,0x03,0x04,
	}
	header = append(header,key...)

	//内容数据加密
	for k,v := range content {
		header = append(header, v^key[k%4])
	}*/
	header = append(header, content...)
	return header
}

func (this *webSocket) GetConnect() *Connect {
	return this.Conn
}