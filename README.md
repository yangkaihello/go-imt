# 客户端调用案例
> 自研聊天系统 http://www.openmessage.cn
```
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

const (
	MESSAGE_CHANNEL_END_BYTE byte = 0x03 //结尾符
	MESSAGE_TYPES_SYSTEM_USER_ID = "system_user_id" //系统类型的通信，不作为用户展示信息
)

var UserUnique string

//用户请求资源内容
type MessageHeader struct {
	UserUnique string `json:"userUnique"`
	Data string `json:"data"`
}

func main ()  {
	var Conn net.Conn

	Conn,_ = net.Dial("tcp","127.0.0.1:1234")

	var header,_ = json.Marshal(MessageHeader{
		UserUnique: "",
		Data:       "杨凯",
	})
	header = append(header)
	Conn.Write(append([]byte("TCP /Login SOCKET\r\n\r\n"),header...))

	go func() {
		for  {

			var buf = bufio.NewReader(Conn)
			var message,err = buf.ReadString(MESSAGE_CHANNEL_END_BYTE)
			message = strings.Trim(message,string(MESSAGE_CHANNEL_END_BYTE))

			//解析读取数据结构
			var messageS = struct {
				Types string `json:"types"`
				Data string `json:"data"`
			}{}

			json.Unmarshal([]byte(message),&messageS)

			if err != nil || message == io.EOF.Error() {
				Conn.Close()
				break
			}else{
				switch messageS.Types {
				case MESSAGE_TYPES_SYSTEM_USER_ID:UserUnique = messageS.Data;break
				}
				fmt.Println(messageS)
			}

			time.Sleep(time.Second*1)

		}
	}()

	for  {
		var scan string
		fmt.Scan(&scan)

		switch scan {
		case "callback":Callback();continue
		case "getusers":GetUsers();continue
		case "leave":Leave();continue
		case "buildmessage":
			fmt.Scan(&scan)
			BuildMessage(scan);continue
		case "sendmessage":
			fmt.Scan(&scan)
			SendMessage(scan);continue
		default:
			Conn,_ = net.Dial("tcp","127.0.0.1:1234")
			var header,_ = json.Marshal(MessageHeader{
				UserUnique: "",
				Data:       scan,
			})
			Conn.Write([]byte("TCP /SystemCtl SOCKET\r\n"))
			Conn.Write(header)

			var buf = bufio.NewReader(Conn)
			var b = make([]byte,2048)
			var n,_ = buf.Read(b)
			fmt.Println(string(b[:n]))
		}

	}

}

func SendMessage(message string)  {
	var Conn net.Conn
	Conn,_ = net.Dial("tcp","127.0.0.1:1234")

	var header,_ = json.Marshal(MessageHeader{
		UserUnique: UserUnique,
		Data:       message,
	})
	header = append(header)
	Conn.Write(append([]byte("TCP /BothSend SOCKET\r\n\r\n"),header...))

	var buf = bufio.NewReader(Conn)
	var b = make([]byte,2048)
	var n,_ = buf.Read(b)
	fmt.Println(string(b[:n]))
}

func BuildMessage(userid string)  {
	var Conn net.Conn
	Conn,_ = net.Dial("tcp","127.0.0.1:1234")

	var header,_ = json.Marshal(MessageHeader{
		UserUnique: UserUnique,
		Data:       userid,
	})
	header = append(header)
	Conn.Write(append([]byte("TCP /BothAdd SOCKET\r\n\r\n"),header...))

	var buf = bufio.NewReader(Conn)
	var b = make([]byte,2048)
	var n,_ = buf.Read(b)
	fmt.Println(string(b[:n]))
}

func GetUsers()  {
	var Conn net.Conn
	Conn,_ = net.Dial("tcp","127.0.0.1:1234")

	var header,_ = json.Marshal(MessageHeader{
		UserUnique: "",
		Data:       "",
	})
	header = append(header)
	Conn.Write(append([]byte("TCP /OnlineUsers SOCKET\r\n\r\n"),header...))

	var buf = bufio.NewReader(Conn)
	var b = make([]byte,2048)
	var n,_ = buf.Read(b)
	fmt.Println(string(b[:n]))
}

func Leave()  {
	var Conn net.Conn
	Conn,_ = net.Dial("tcp","127.0.0.1:1234")

	var header,_ = json.Marshal(MessageHeader{
		UserUnique: UserUnique,
		Data:       "",
	})
	header = append(header)
	Conn.Write(append([]byte("TCP /Leave SOCKET\r\n\r\n"),header...))

	var buf = bufio.NewReader(Conn)
	var b = make([]byte,2048)
	var n,_ = buf.Read(b)
	fmt.Println(string(b[:n]))
}

func Callback()  {
	var Conn net.Conn
	Conn,_ = net.Dial("tcp","127.0.0.1:1234")

	var header,_ = json.Marshal(MessageHeader{
		UserUnique: UserUnique,
		Data:       "测试发送的消息",
	})
	header = append(header)
	Conn.Write(append([]byte("TCP /test SOCKET\r\n\r\n"),header...))

	var buf = bufio.NewReader(Conn)
	var b = make([]byte,2048)
	var n,_ = buf.Read(b)
	fmt.Println(string(b[:n]))
}
```