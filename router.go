package imt

import (
	"encoding/json"
	"sync"
)

//controller 结构体参数
type RouterParams struct {
	Data string `json` //原始的json字符串用来解析成需要的数据结构
	Connect Protocol
}

type ControllerReturn struct {
	Types string
	Data string
}

//data json 数据结构解析
func (this *RouterParams) StructData(v interface{}) (interface{},error) {
	var err error

	if err = json.Unmarshal([]byte(this.Data),v); err != nil {
		return nil,err
	}
	return v,nil
}

//限制controller 方法结构体
type ControllerTemplate func(params RouterParams) Protocol

//提供外部注入类型 不允许实例化
type controllers map[string]ControllerTemplate
var controllersLock sync.Mutex

//添加每一个控制器
func (this *controllers) Add(path string,f ControllerTemplate)  {
	(*this)[path] = f
}

//copy 一个控制器组防止并发异常
func (this *controllers) Copy() controllers {
	controllersLock.Lock()
	var c = controllers{}
	for k,v := range *this {
		c.Add(k,v)
	}
	controllersLock.Unlock()
	return c
}


//路由类型
type Router struct {
	Controllers map[string]ControllerTemplate
}

//实例化路由
func (this *Router) Init() *Router {
	this.Controllers = map[string]ControllerTemplate{}
	return this
}

//验证这个路由是否存在有控制器
func (this *Router) Has(path string) bool {
	_,ok := this.Controllers[path]
	return ok
}

//添加每一个控制器
func (this *Router) Add(path string,f ControllerTemplate)  {
	this.Controllers[path] = f
}

//覆盖式的添加路由
func (this *Router) AddAll(f controllers)  {
	this.Controllers = f
}

//执行控制器
func (this *Router) Run(path string,params RouterParams) Protocol {
	if !this.Has(path) {
		return nil
	}

	return (this.Controllers[path])(params)
}