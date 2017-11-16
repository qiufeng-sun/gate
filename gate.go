package main

import (
	"util/logs"

	"core/net/dispatcher"
	"core/server"
	"core/time"

	"share/msg"
)

var _ = logs.Debug
var _ = time.Now

//
type Gate struct {
	//
	server.Server

	//
	Url       string // 由ip+port计算得出
	StartTime int64

	//
	MsgPkgName   string
	MsgNameIdMap map[string]int32
}

//
func NewGate() *Gate {
	return &Gate{
		StartTime:    time.Now().Unix(),
		MsgPkgName:   "msg",
		MsgNameIdMap: msg.EMsg_value,
	}
}

//
func (this Gate) String() string {
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

	// init client msg dispatcher
	dispatcher.Init("gate dispatcher", SrvId())

	// recv/send msg among servers
	InitServers()

	// client connect and recv/send msg
	InitClients()

	return true
}
