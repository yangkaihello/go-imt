package imt

import (
	"sync"
)

//发送消息的结构体
const (
	MESSAGE_CHANNEL_END_BYTE byte = 0x03 //结尾符

	MESSAGE_TYPES_ERROR = "error" //系统异常
	MESSAGE_TYPES_SYSTEM = "system" //系统类型的通信，不作为用户展示信息
	MESSAGE_TYPES_SYSTEM_HEART = "system_heart" //系统类型的通信，不作为用户展示信息（客户端心跳检测信息）
	MESSAGE_TYPES_SYSTEM_USER_JOIN = "system_user_join" //系统类型的通信，不作为用户展示信息（有用户加入系统）
	MESSAGE_TYPES_SYSTEM_USER_LEAVE = "system_user_leave" //系统类型的通信，不作为用户展示信息（用户的离开）
	MESSAGE_TYPES_SYSTEM_USER_ID = "system_user_id" //系统类型的通信，不作为用户展示信息
	MESSAGE_TYPES_SYSTEM_GROUP_ID = "system_group_id" //系统类型的通信，不作为用户展示信息

	MESSAGE_TYPES_STRING = "string" //正常消息内容
	MESSAGE_TYPES_IMAGE = "image"	//图片消息
	MESSAGE_TYPES_EMOJI = "emoji"	//系统图片

)

//内部控制器return关键词
const (
	RETURN_KEYWORD_GOROUTINE = "GOROUTINE"
	RETURN_KEYWORD_STRING = "STRING"
)

//聊天组
var groupLock sync.Mutex
var GroupId int
var ConstGroups = new(groups).Init()

//双人通信
var bothLock sync.Mutex
var ConstBothWay = new(bothWay).Init()

//用户组
var userLock sync.Mutex
var ConstUsers = new(users).Init()

//用户地区分组
var addressLock sync.Mutex
var ConstAddress = new(address).Init()

//开发注入的路由控制器
var ControllerLoader = controllers{}

//对已经登录的用户进行消息发送
var messageChannelsLock sync.Mutex
var ConstMessageChannels = new(messageChannels).Init()

//用户离开需要释放的资源
//顺便告诉所在地区用户有人离开了
func LeaveClose(user *User)  {
	//获取相关区域用户
	var users = ConstAddress.Get(user)
	//告知在线用户有人离开了
	for _,u := range users {
		//获取管道以及发送消息的内容
		var C = ConstMessageChannels.Get(u)
		//发送
		if C != nil {
			*C <- SendMessage{
				Users:   []UserUnique{u.Unique},
				Message: Message{
					Types: MESSAGE_TYPES_SYSTEM_USER_LEAVE,
					Data:  string(user.Unique),
				},
			}
		}
	}

	ConstBothWay.Delete(user)
	ConstMessageChannels.Delete(user)
	ConstAddress.Delete(user)
	user.Send.GetConnect().Close()
	ConstUsers.Delete(user)

}

//定义接口类
type Protocol interface {
	SetHeader(key string,value string)
	GetHeaderAll() map[string]string
	GetHeader(key string) string
	GetBody() *MessageHeader
	SetBody(s string)
	SetResponse(key string,value string)
	GetResponse() string
	GetOneSelfName() string
	SendMessage(s string) error
	ReadMessage() (string,error)
	SendStruct(m Message) error
	ReadStruct(v interface{}) (interface{},error)
	SetReturn(controllerReturn *ControllerReturn) Protocol
	GetReturn() *ControllerReturn
	GetConnect() *Connect
}
