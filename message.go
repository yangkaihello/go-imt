package imt

import (
	"bufio"
	"net"
	"strconv"
	"time"
)

//系统发送通知，聊天内容转发结构
type Message struct {
	Types string `json:"types"`
	Data string `json:"data"`
}

//用户请求资源内容
type MessageHeader struct {
	UserUnique UserUnique `json:"userUnique"`
	Data string `json:"data"`
}

//通信管道需要的message结构
type SendMessage struct {
	Users []UserUnique
	Message Message
}

type UserUnique string

type User struct {
	Id string `json:"id"`
	Sex string `json:"sex"`
	Unique UserUnique `json:"unique"`
	Username string `json:"username"`
	Ip string `json:"ip"`
	Province string `json:"province"`
	City string `json:"city"`
	Send Protocol
}

type Group struct {
	Id string
	Name string
	Users []UserUnique
}

//每个用户的通信管道 不允许实例化
type messageChannels map[UserUnique]chan SendMessage

func (this *messageChannels) Init() *messageChannels {
	*this = map[UserUnique]chan SendMessage{}
	return this
}
func (this *messageChannels) Add(user *User) {
	messageChannelsLock.Lock()
	(*this)[user.Unique] = make(chan SendMessage,0)
	messageChannelsLock.Unlock()
}

func (this *messageChannels) Get(user *User) *chan SendMessage {
	messageChannelsLock.Lock()
	defer messageChannelsLock.Unlock()

	if user == nil {
		return nil
	}
	if _,ok := (*this)[user.Unique]; !ok {
		return nil
	}
	var t = (*this)[user.Unique]
	return &t
}

func (this *messageChannels) Delete(user *User) bool {
	messageChannelsLock.Lock()
	defer messageChannelsLock.Unlock()

	if user == nil {
		return false
	}
	if _,ok := (*this)[user.Unique]; !ok {
		return false
	}
	delete(*this,user.Unique)
	return true
}

//定义组 不允许实例化
type groups map[string]*Group

//初始化操作
func (this *groups) Init() *groups {
	*this = map[string]*Group{}
	return this
}

func (this *groups) GetHash(str string) string {
	return new(Hash).MD5(str)
}

//添加用户组
func (this *groups) Add(g *Group) bool {
	groupLock.Lock()
	defer groupLock.Unlock()
	if g == nil {
		return false
	}
	g.Id = this.GetHash(strconv.Itoa(GroupId))
	(*this)[g.Id] = g
	GroupId++
	return true
}

//获取用户组
func (this *groups) Get(id string) *Group {
	groupLock.Lock()
	defer groupLock.Unlock()
	if _,ok := (*this)[id]; !ok {
		return nil
	}
	var g = (*this)[id]
	return g
}

//删除用户组
func (this *groups) Delete(id string) bool {
	groupLock.Lock()
	defer groupLock.Unlock()
	if _,ok := (*this)[id]; !ok {
		return false
	}
	delete(*this,id)
	return true
}

//对用户组添加用户
func (this *groups) AddUser(id string,u *User) bool {
	groupLock.Lock()
	defer groupLock.Unlock()

	if u == nil {
		return false
	}
	if _,ok := (*this)[id]; !ok {
		return false
	}
	(*this)[id].Users = append((*this)[id].Users, u.Unique)
	return true
}

//剔除用户组成员
func (this *groups) DeleteUser(id string,u *User) bool {
	groupLock.Lock()
	defer groupLock.Unlock()

	if u == nil {
		return false
	}
	if _,ok := (*this)[id]; !ok {
		return false
	}
	var number int
	for k,v := range (*this)[id].Users {
		if v == u.Unique {
			number = k;break
		}
	}
	(*this)[id].Users = append((*this)[id].Users[:number], (*this)[id].Users[number+1:]...)
	return true
}

//定义双人通信 不允许实例化
type bothWay map[UserUnique]*User

//初始化操作
func (this *bothWay) Init() *bothWay {
	*this = map[UserUnique]*User{}
	return this
}

//关联双人通道
func (this *bothWay) Touch(UserOne *User,UserTow *User) bool {
	bothLock.Lock()
	defer bothLock.Unlock()
	if UserOne == nil || UserTow == nil {
		return false
	}
	//建立一个相互通信的组
	(*this)[UserOne.Unique] = UserTow
	return true
}

//获取聊天对象
func (this *bothWay) Friend(u *User) *User {
	userLock.Lock()
	defer userLock.Unlock()

	if u == nil {
		return nil
	}

	if _,ok := (*this)[u.Unique]; !ok {
		return nil
	}
	var user = (*this)[u.Unique]
	return user
}

//删除双人通信组表示推出
func (this *bothWay) Delete(u *User) bool {
	userLock.Lock()
	defer userLock.Unlock()

	if u == nil {
		return false
	}

	if _,ok := (*this)[u.Unique]; !ok {
		return false
	}
	delete(*this,u.Unique)
	return true
}

//用户组 不允许实例化
type users map[UserUnique]*User

//初始化操作
func (this *users) Init() *users {
	*this = map[UserUnique]*User{}
	return this
}

//用户ID设置
func (this *users) GetHash(str string) string {
	return new(Hash).MD5(str)
}

//添加用户组
func (this *users) Add(u *User) *User {
	userLock.Lock()
	u.Id = this.GetHash(string(time.Now().UnixNano()))
	(*this)[u.Unique] = u
	userLock.Unlock()
	return (*this)[u.Unique]
}

func (this *users) Get(unique UserUnique) *User {
	userLock.Lock()
	defer userLock.Unlock()
	if _,ok := (*this)[unique]; !ok {
		return nil
	}
	var u = (*this)[unique]
	return u
}

func (this *users) GetAll() *users {
	var c = new(users).Init()
	for _,v := range *this {
		c.Add(v)
	}
	return c
}

//删除用户组
func (this *users) Delete(u *User) bool {
	userLock.Lock()
	if _,ok := (*this)[u.Unique]; !ok {
		return false
	}
	delete(*this,u.Unique)
	userLock.Unlock()
	return true
}


type address map[string]map[UserUnique]*User

//初始化操作
func (this *address) Init() *address {
	*this = map[string]map[UserUnique]*User{}
	return this
}

//添加地区组
func (this *address) Add(u *User) {
	addressLock.Lock()
	if _,ok := (*this)[u.Province]; !ok {
		(*this)[u.Province] = map[UserUnique]*User{}
	}
	(*this)[u.Province][u.Unique] = u
	addressLock.Unlock()
}

//添加地区组
func (this *address) Get(u *User) map[UserUnique]*User {
	addressLock.Lock()
	defer addressLock.Unlock()
	if _,ok := (*this)[u.Province]; !ok {
		return nil
	}
	return (*this)[u.Province]
}

//地区组删除
func (this *address) Delete(u *User) bool {
	addressLock.Lock()
	defer addressLock.Unlock()
	if _,ok := (*this)[u.Province][u.Unique]; !ok {
		return false
	}
	delete((*this)[u.Province],u.Unique)
	return true
}

//net.Conn 的包装
type Connect struct {
	Conn *net.Conn
}

func (this *Connect) Init(conn *net.Conn) *Connect {
	this.Conn = conn
	return this
}

func (this *Connect) Has() bool {
	if this.Conn != nil {
		return true
	}else{
		return false
	}
}

//读取缓冲区的2048个字节内容
func (this *Connect) ReadString() (string,error) {
	var b = make([]byte,2048)
	var n,err = (*this.Conn).Read(b)
	return string(b[:n]),err
}

//读取缓冲区某个字节结尾的数据
func (this *Connect) ReadByteEnd(b byte) (string,error) {
	var reader = bufio.NewReader(*this.Conn)
	return reader.ReadString(b)
}

//不同的协议都会有头部需要读取的需求
func (this *Connect) readHeader() (string,error) {
	//需要等待0.01秒来确认header头正常写入
	time.Sleep(time.Millisecond*10)
	var b = make([]byte,2048)
	var n,err = (*this.Conn).Read(b)
	return string(b[:n]),err
}

//发送所有的内容
func (this *Connect) SendString(s string) error {
	var _,err = (*this.Conn).Write([]byte(s))
	return err
}

func (this *Connect) Close()  {
	(*this.Conn).Close()
}


