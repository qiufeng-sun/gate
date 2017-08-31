package main

import (
	"util/logs"

	"core/time"
)

var _ = logs.Debug
var _ = time.Now

//
type Gate struct {
	Url       string // 由ip+port计算得出
	StartTime int64

	// to do assign value
	MsgPkgName   string
	MsgNameIdMap map[string]int32
}

//
func NewGate() *Gate {
	return &Gate{StartTime: time.Now().Unix()}
}

//
func (this *Gate) String() string {
	return "gate"
}

//
func (this *Gate) Init() bool {
	// config
	if !LoadConfig("conf/") {
		return false
	}

	// 注册消息处理函数
	Register(this.MsgPkgName, this.MsgNameIdMap)

	// recv/send msg among servers
	InitServers()

	// init client msg dispatcher
	InitDispatcher(SrvId())

	// client connect and recv/send msg
	InitClients()

	return true
}

//
func (this *Gate) Update() {
	// do nothing
}

//
func (this *Gate) Destroy() {
	// do nothing
}

//
func (this *Gate) PreQuit() {
	// do nothing
}
