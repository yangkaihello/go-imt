package imt

import (
	"encoding/json"
	"runtime"
	"strings"
	"time"
)

//登录账户控制器
func Login(params RouterParams) Protocol  {
	var data interface{}
	var err error
	//返回一个指针类型的结构体
	if data,err = params.StructData(new(MessageHeader)); err != nil {
		return params.Connect.SetReturn(&ControllerReturn{
			Types: RETURN_KEYWORD_STRING,
			Data:  JsonEncode(Message{
				Types: MESSAGE_TYPES_ERROR,
				Data:  "结构体异常",
			}),
		})
	}

	//MessageHeader.Data = 用户的名称
	var MessageHeader = data.(*MessageHeader)
	var ips = strings.Split((*params.Connect.GetConnect().Conn).RemoteAddr().String(),":")
	var unique = UserUnique(ConstUsers.GetHash(ips[0]+":"+MessageHeader.Data))
	//验证用户是否已经创建
	if ConstUsers.Get(unique) != nil {
		return params.Connect.SetReturn(&ControllerReturn{
			Types: MESSAGE_TYPES_ERROR,
			Data:  JsonEncode(Message{
				Types: MESSAGE_TYPES_ERROR,
				Data:  "结构体异常",
			}),
		})
	}
	//解析ip来源地址
	var address,_ = LocationIP(ips[0])
	if len(address) < 3 {
		address = []string{"CN","局域网","局域网"}
	}
	//添加全局定义
	var user = ConstUsers.Add(&User{
		Username: MessageHeader.Data,
		Unique:   unique,
		Ip:       ips[0],
		Province: address[1],
		City: 	  address[2],
		Send:     params.Connect,
	})

	//加入地址组
	ConstAddress.Add(user)

	//用户通道创建
	ConstMessageChannels.Add(user)

	go func() {
		var C = ConstMessageChannels.Get(user)
		//通道监听向哪些用户进行消息发送
		for message := range *C {
			for _,u := range message.Users {
				//获取现有的通信进程
				var user = ConstUsers.Get(u)
				//确认通信是否正常建立
				if user != nil && user.Send.GetConnect().Has() {
					user.Send.SendStruct(message.Message)
				}else{
					continue
				}
			}
		}
	}()

	//对相同地区的用户进行广播告知已经登录
	go func() {
		for _,u := range ConstAddress.Get(user) {
			var u = ConstUsers.Get(u.Unique)
			if u.Unique == user.Unique {
				continue
			}	//不用返回自己的用户组

			//确认通信是否正常建立
			if u.Send.GetConnect().Has() {
				var info = struct {
					Username string `json:"username"`
					Unique string `json:"unique"`
					Sex   	 string `json:"sex"`
				}{user.Username,string(user.Unique),user.Sex}
				var infoByte,_ = json.Marshal(info)
				u.Send.SendStruct(Message{
					Types: MESSAGE_TYPES_SYSTEM_USER_JOIN,
					Data:  string(infoByte),
				})
			}
		}
	}()

	//心跳检测通过向客户端发包来证明这个连接是有效的
	go func() {
		for  {
			if err := params.Connect.SendStruct(Message{
				Types: MESSAGE_TYPES_SYSTEM_HEART,
				Data:  "close",
			}); err != nil {
				LeaveClose(user)
				break
			}
			time.Sleep(time.Second*1)
		}
	}()

	return params.Connect.SetReturn(&ControllerReturn{
		Types: RETURN_KEYWORD_GOROUTINE,
		Data:  JsonEncode(Message{
			Types: MESSAGE_TYPES_SYSTEM_USER_ID,
			Data:  string(user.Unique),
		}),
	})
}

//在线的所有用户
func OnlineUsers(params RouterParams) Protocol {
	var data interface{}
	var err error
	//返回一个指针类型的结构体
	if data,err = params.StructData(new(MessageHeader)); err != nil {
		return params.Connect.SetReturn(&ControllerReturn{
			Types: MESSAGE_TYPES_ERROR,
			Data:  "close",
		})
	}
	//MessageHeader.Data = 用户的名称
	var MessageHeader = data.(*MessageHeader)

	var AllUsers = map[string]struct {
		Username string `json:"username"`
		Unique string `json:"unique"`
		Sex   	 string `json:"sex"`
	}{}
	var user = ConstUsers.Get(MessageHeader.UserUnique)
	if user != nil {
		for _,u := range ConstAddress.Get(user) {
			if u.Unique == MessageHeader.UserUnique {
				continue
			}	//不用返回自己的用户组
			AllUsers[string(u.Unique)] = struct {
				Username string `json:"username"`
				Unique   string `json:"unique"`
				Sex   	 string `json:"sex"`
			}{Username: u.Username, Unique: string(u.Unique), Sex:u.Sex}
		}
	}

	var AllUsersString,_ = json.Marshal(AllUsers)

	return params.Connect.SetReturn(&ControllerReturn{
		Types: "string",
		Data:  string(AllUsersString),
	})
}

//建立普通聊天通道
func BothAdd(params RouterParams) Protocol {
	var data interface{}
	var err error
	//返回一个指针类型的结构体
	if data,err = params.StructData(new(MessageHeader)); err != nil {
		return params.Connect.SetReturn(&ControllerReturn{
			Types: MESSAGE_TYPES_ERROR,
			Data:  "close",
		})
	}
	//MessageHeader.Data = 用户的名称
	var MessageHeader = data.(*MessageHeader)

	var UserOne = ConstUsers.Get(MessageHeader.UserUnique)
	var UserTow = ConstUsers.Get(UserUnique(MessageHeader.Data))
	//联系人和被联系人需要都在线才可以建立通信
	if UserOne != nil && UserTow != nil {
		ConstBothWay.Touch(UserOne,UserTow)
	}else{
		return params.Connect.SetReturn(&ControllerReturn{
			Types: MESSAGE_TYPES_ERROR,
			Data:  "close",
		})
	}

	return params.Connect.SetReturn(&ControllerReturn{
		Types: "string",
		Data:  "ok",
	})
}

//角色发送聊天信息
func BothSend(params RouterParams) Protocol {
	var data interface{}
	var err error
	//返回一个指针类型的结构体
	if data,err = params.StructData(new(MessageHeader)); err != nil {
		return params.Connect.SetReturn(&ControllerReturn{
			Types: MESSAGE_TYPES_ERROR,
			Data:  "close",
		})
	}
	//MessageHeader.Data = 用户的名称
	var MessageHeader = data.(*MessageHeader)

	var user = ConstUsers.Get(MessageHeader.UserUnique)

	if user == nil {
		return params.Connect.SetReturn(&ControllerReturn{
			Types: MESSAGE_TYPES_ERROR,
			Data:  "close",
		})
	}
	var friendUser = ConstBothWay.Friend(user)
	//获取管道以及发送消息的内容
	var C = ConstMessageChannels.Get(user)
	var Data,_ = json.Marshal(struct {
		Data string `json:"data"`
		SourceUnique string `json:"sourceUnique"`
	}{MessageHeader.Data,string(user.Unique)})
	//发送
	if C != nil && friendUser != nil {
		*C <- SendMessage{
			Users:   []UserUnique{friendUser.Unique},
			Message: Message{
				Types: MESSAGE_TYPES_STRING,
				Data:  string(Data),
			},
		}
	}

	return params.Connect.SetReturn(&ControllerReturn{
		Types: "string",
		Data:  "ok",
	})

}

//离开聊天室
func Leave(params RouterParams) Protocol {
	var data interface{}
	var err error
	//返回一个指针类型的结构体
	if data,err = params.StructData(new(MessageHeader)); err != nil {
		return params.Connect.SetReturn(&ControllerReturn{
			Types: MESSAGE_TYPES_ERROR,
			Data:  "close",
		})
	}
	//MessageHeader.Data = 用户的名称
	var MessageHeader = data.(*MessageHeader)
	LeaveClose(ConstUsers.Get(MessageHeader.UserUnique))

	return params.Connect.SetReturn(&ControllerReturn{
		Types: "string",
		Data:  "ok",
	})
}

//系统内容获取控制器
func SystemCtl(params RouterParams) Protocol {
	var data interface{}
	var err error
	//返回一个指针类型的结构体
	if data,err = params.StructData(new(MessageHeader)); err != nil {
		return params.Connect.SetReturn(&ControllerReturn{
			Types: MESSAGE_TYPES_ERROR,
			Data:  "close",
		})
	}
	//MessageHeader.Data = 用户的名称
	var MessageHeader = data.(*MessageHeader)

	var b []byte

	switch MessageHeader.Data {
	case "users": b,_ = json.Marshal(ConstUsers);break
	case "both": b,_ = json.Marshal(ConstBothWay);break
	case "address": b,_ = json.Marshal(ConstAddress);break
	case "group": b,_ = json.Marshal(ConstGroups);break
	case "channel": b = []byte{byte(len(*ConstMessageChannels))};break
	case "goroutine": b,_ = json.Marshal(struct {
		Goroutine int `json:"goroutine"`
	}{Goroutine:runtime.NumGoroutine()});break
	default:
		b,_ = json.Marshal(ConstUsers);break
	}

	return params.Connect.SetReturn(&ControllerReturn{
		Types: "string",
		Data:  string(b),
	})

}