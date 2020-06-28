package imt

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type httpSocket struct {
	data string
	header map[string]string
	response map[string]string
	responseLineOne string
	controllerReturn *ControllerReturn
	Conn *Connect
}

func (this *httpSocket) Init(connect *Connect) *httpSocket {
	this.Conn = connect
	this.header = map[string]string{}

	//websocket头部固定协议
	this.responseLineOne = "HTTP/1.1 200 OK"
	this.response = map[string]string{
		"Connection":"keep-alive",
		"Access-Control-Allow-Origin":"*",
		"Access-Control-Allow-Methods":"POST,PUT,OPTIONS,DELETE",
	}

	return this
}

func (this *httpSocket) SetHeader(key string,value string) {
	this.header[key] = value
}

func (this *httpSocket) GetHeaderAll() map[string]string {
	return this.header
}

func (this *httpSocket) GetHeader(key string) string {
	return this.header[key]
}

func (this *httpSocket) GetBody() *MessageHeader {
	var message = new(MessageHeader)
	var bodyKeyword = "MessageHeader"
	var data string

	if query,err := url.ParseQuery(this.data); err == nil {
		data = query.Get(bodyKeyword)
	}

	if data == "" {
		data = this.data
	}

	if err := json.Unmarshal([]byte(data),message); err != nil {
		return nil
	}
	return message
}
func (this *httpSocket) SetBody(s string) {
	this.data = strings.Trim(s,"")
}

func (this *httpSocket) SetResponse(key string,value string) {
	this.response[key] = value
}

func (this *httpSocket) GetResponse() string {
	var responses = []string{
		this.responseLineOne,
	}
	for k,v := range this.response {
		responses = append(responses, k+": "+v)
	}

	return strings.Join(responses,"\r\n")+"\r\n\r\n"
}

func (this *httpSocket) GetOneSelfName() string {
	return reflect.TypeOf(this).Name()
}

func (this *httpSocket) SendMessage(s string) error {
	var err error

	if err = this.Conn.SendString(s); err != nil {
		return err
	}

	return nil
}

func (this *httpSocket) ReadMessage() (string,error) {
	var data string
	var err error
	var frame []byte
	//获取资源数据
	if  data,err = this.Conn.ReadString(); err != nil {
		return "",err
	}

	data = string(frame)

	return data,nil
}

func (this *httpSocket) SendStruct(m Message) error {
	var data []byte
	var err error

	if data,err = json.Marshal(m); err != nil {
		return err
	}

	if err = this.Conn.SendString(string(data)); err != nil {
		return err
	}
	return nil
}

func (this *httpSocket) ReadStruct(v interface{}) (interface{},error) {
	var data string
	var err error

	//获取资源数据
	if  data,err = this.Conn.ReadString(); err != nil {
		return "",err
	}

	if err = json.Unmarshal([]byte(data)[:len(data)-1],v); err != nil {
		return nil,err
	}

	return v,nil
}

func (this *httpSocket) SetReturn(controllerReturn *ControllerReturn) Protocol {
	this.controllerReturn = controllerReturn
	this.SetResponse("Content-Length",strconv.Itoa(len(this.controllerReturn.Data)))

	//验证是否是json
	if json.Valid([]byte(this.controllerReturn.Data)) {
		this.SetResponse("Content-Type","application/json")
	}else{
		this.SetResponse("Content-Type","text/html")
	}

	this.SendMessage(this.GetResponse()+this.controllerReturn.Data)
	return this
}

func (this *httpSocket) GetReturn() *ControllerReturn {
	return this.controllerReturn
}

func (this *httpSocket) GetConnect() *Connect {
	return this.Conn
}