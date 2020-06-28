package imt

import (
	"encoding/json"
	"reflect"
	"strings"
)

type socket struct {
	data string
	header map[string]string
	response map[string]string
	responseLineOne string
	controllerReturn *ControllerReturn
	Conn *Connect
}

func (this *socket) Init(connect *Connect) *socket {
	this.Conn = connect
	return this
}

func (this *socket) SetHeader(key string,value string) {
	this.header[key] = value
}

func (this *socket) GetHeaderAll() map[string]string {
	return this.header
}

func (this *socket) GetHeader(key string) string {
	return this.header[key]
}

func (this *socket) GetBody() *MessageHeader {
	var message = new(MessageHeader)

	if err := json.Unmarshal([]byte(this.data),message); err != nil {
		return nil
	}
	return message
}

func (this *socket) SetBody(s string) {
	this.data = strings.Trim(s,"")
}

func (this *socket) SetResponse(key string,value string) {
	this.response[key] = value
}

func (this *socket) GetResponse() string {
	var responses = []string{
		this.responseLineOne,
	}
	for k,v := range this.response {
		responses = append(responses, k+": "+v)
	}

	return strings.Join(responses,"\r\n")+"\r\n\r\n"
}

func (this *socket) GetOneSelfName() string {
	return reflect.TypeOf(this).Name()
}

func (this *socket) SendMessage(s string) error {
	var err error
	if err = this.Conn.SendString(s+string(MESSAGE_CHANNEL_END_BYTE)); err != nil {
		return err
	}
	return nil
}

func (this *socket) ReadMessage() (string,error) {
	var data string
	var err error
	//获取资源数据
	if  data,err = this.Conn.ReadByteEnd(MESSAGE_CHANNEL_END_BYTE); err != nil {
		return "",err
	}
	return data,nil
}

func (this *socket) SendStruct(m Message) error {
	var data []byte
	var err error

	if data,err = json.Marshal(m); err != nil {
		return err
	}

	if err = this.Conn.SendString(string(append(data, MESSAGE_CHANNEL_END_BYTE))); err != nil {
		return err
	}
	return nil
}

func (this *socket) ReadStruct(v interface{}) (interface{},error) {
	var data string
	var err error

	//获取资源数据
	if  data,err = this.Conn.ReadByteEnd(MESSAGE_CHANNEL_END_BYTE); err != nil {
		return "",err
	}

	if err = json.Unmarshal([]byte(data)[:len(data)-1],v); err != nil {
		return nil,err
	}

	return v,nil
}

func (this *socket) SetReturn(controllerReturn *ControllerReturn) Protocol {
	this.controllerReturn = controllerReturn
	this.SendMessage(this.controllerReturn.Data)
	return this
}

func (this *socket) GetReturn() *ControllerReturn {
	return this.controllerReturn
}

func (this *socket) GetConnect() *Connect {
	return this.Conn
}