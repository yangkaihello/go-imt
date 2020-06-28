package imt

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"
)

type Imt struct {

}

func (this *Imt) Init() *Imt {
	//首次登录的持久通信管道注入到路由中
	ControllerLoader.Add("/Login",Login)
	ControllerLoader.Add("/OnlineUsers",OnlineUsers)
	ControllerLoader.Add("/BothAdd",BothAdd)
	ControllerLoader.Add("/BothSend",BothSend)
	ControllerLoader.Add("/Leave",Leave)
	ControllerLoader.Add("/SystemCtl",SystemCtl)

	return this
}

func (this *Imt) Start(port int)  {
	var err error
	var listen net.Listener

	if listen,err = net.Listen("tcp",":"+strconv.Itoa(port)); err != nil{
		panic(err.Error())
	}

	for  {
		var conn net.Conn
		var err error
		var data string
		if conn,err = listen.Accept(); err != nil{
			continue
		}

		var router = new(Router).Init()
		var Connect = new(Connect).Init(&conn)

		//获取协议头首次握手
		if data,err = Connect.readHeader(); err != nil {
			Connect.Close();continue
		}
		//首次读取的内容表示协议
		var loaderImt = new(loaderImt).Init(Connect)
		if loaderImt.Create(data) == nil {
			Connect.Close();continue
		}

		//用户其他操作控制器,只能是一次性的
		router.AddAll(ControllerLoader.Copy())

		//路由验证
		if router.Has(loaderImt.Path) {
			var protocol = router.Run(loaderImt.Path,RouterParams{Data:loaderImt.GetData(),Connect:loaderImt.Protocol})
			var Return = protocol.GetReturn()

			//终结除了聊天通信之外的所有通信
			if Return.Types != RETURN_KEYWORD_GOROUTINE {
				Connect.Close();continue
			}
		} else {
			loaderImt.Protocol.SetReturn(&ControllerReturn{
				Types: "string",
				Data:  "close",
			});Connect.Close();continue
		}

	}

}

const (
	PROTOCOL_TYPE_HTTP = "http"
	PROTOCOL_TYPE_WEBSOCKET = "websocket"
	PROTOCOL_TYPE_SOCKET = "socket"
)

var PROTOCOL_TYPES = map[string]string{
	"HEAD":PROTOCOL_TYPE_HTTP,
	"POST":PROTOCOL_TYPE_HTTP,
	"PUT":PROTOCOL_TYPE_HTTP,
	"DELETE":PROTOCOL_TYPE_HTTP,
	"TRACE":PROTOCOL_TYPE_HTTP,
	"OPTIONS":PROTOCOL_TYPE_HTTP,
	"CONNECT":PROTOCOL_TYPE_HTTP,
	"HTTP/1.1":PROTOCOL_TYPE_HTTP,
	"GET":PROTOCOL_TYPE_WEBSOCKET,
	"SOCKET":PROTOCOL_TYPE_SOCKET,
	"TCP":PROTOCOL_TYPE_SOCKET,
}

type loaderImt struct {
	Types string
	Path string
	Conn *Connect
	Protocol Protocol
}

func (this *loaderImt) Init(conn *Connect) *loaderImt {
	this.Conn = conn
	return this
}

func (this *loaderImt) Create(header string) Protocol {
	var lines = strings.Split(header,"\r\n")
	//根据http协议头第一行表明协议类型
	var lineOne = lines[0]
	var lineOnes = strings.Split(lineOne," ")
	var ResultType string //最终确定的协议

	if len(lineOnes) < 3{
		return nil
	}

	//如果没有协议就表示空
	if t,ok := PROTOCOL_TYPES[lineOnes[0]]; ok {
		ResultType = t
	}

	if t,ok := PROTOCOL_TYPES[lineOnes[2]]; ok && ResultType == "" {
		ResultType = t
	}

	if ResultType == "" {
		return nil
	}

	this.Path = lineOnes[1]
	this.Types = ResultType

	//创建对于的协议通信
	switch this.Types {
	case PROTOCOL_TYPE_HTTP:
		this.Protocol = new(httpSocket).Init(this.Conn)
		this.setHeader(lines[1:],&this.Protocol)
		this.Protocol.SetBody(lines[len(lines)-1])
		break
	case PROTOCOL_TYPE_SOCKET:
		this.Protocol = new(socket).Init(this.Conn)
		this.setHeader(lines[1:],&this.Protocol)
		this.Protocol.SetBody(lines[len(lines)-1])
		break
	case PROTOCOL_TYPE_WEBSOCKET:
		this.Protocol = new(webSocket).Init(this.Conn)
		this.setHeader(lines[1:],&this.Protocol)
		this.Protocol.SetResponse("Sec-WebSocket-Accept",new(Hash).Sha1WevSocketAccept(this.Protocol.GetHeader("Sec-WebSocket-Key")))
		_ = this.Conn.SendString(this.Protocol.GetResponse())

		var d,_ = this.Protocol.ReadMessage()
		this.Protocol.SetBody(d)
		break
	}

	return this.Protocol
}

func (this *loaderImt) setHeader(lines []string,Protocol *Protocol)  {
	//其余协议添加
	for _,line := range lines {
		if len(strings.Trim(line,"\r\n")) != 0 {
			both := strings.Split(line,":")
			if len(both) == 2 {
				(*Protocol).SetHeader(strings.Trim(both[0]," "),strings.Trim(both[1]," "))
			}
		}
	}
}

func (this *loaderImt) GetData() string {
	var message *MessageHeader
	if message = this.Protocol.GetBody(); message == nil {
		return ""
	}
	var data,_ = json.Marshal(*message)
	return string(data)
}
