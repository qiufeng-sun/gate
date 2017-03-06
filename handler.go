// 消息处理注册管理
package main

//
import (
	"core/net/socket"
)

// 注册信息封装
type HandleInfo struct {
	// to do
}

func NewHandleInfo() *HandleInfo {
	return &HandleInfo{}
}

// to do atomic
func (this *HandleInfo) AddStats() {

}

//
type MsgHandler struct {
	*socket.MsgHandler
}

//
func NewMsgHandler() *MsgHandler {
	return &MsgHandler{socket.NewMsgHandler()}
}

// 注册
func (this *MsgHandler) RegHandler(msgId int32, handler socket.IHandler) {
	this.MsgHandler.RegHandler(msgId, handler, NewHandleInfo())
}

//
type hfunc func(*Client, []byte)

func (this hfunc) Handle(receiver interface{}, msg []byte) {
	this(receiver.(*Client), msg)
}

func (this *MsgHandler) RegFunc(msgId int32, f func(c *Client, d []byte)) {
	this.RegHandler(msgId, hfunc(f))
}

// 获取handler
func (this *MsgHandler) Handler(msgId int32) (socket.IHandler, *HandleInfo, bool) {
	if h, info, ok := this.MsgHandler.Handler(msgId); ok {
		return h, info.(*HandleInfo), true
	}

	return nil, nil, false
}
