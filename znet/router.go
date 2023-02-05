package znet

import "websocket/ziface"

//实现router时，先嵌入这个BaseRouter基类，然后根据需要对这个基类的方法进行重写
type BaseRouter struct{}

//在处理conn业务之前的钩子方法hook
func (br *BaseRouter) PreHandle(request ziface.IRequest) {}

//在处理conn业务主方法
func (br *BaseRouter) Handle(request ziface.IRequest) {}

//在处理conn业务之后的钩子方法hook
func (br *BaseRouter) PostHandle(request ziface.IRequest) {}
