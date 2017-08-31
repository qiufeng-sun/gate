package main

import (
	"github.com/golang/protobuf/proto"

	"util/etcd"
	"util/logs"
	"util/run"

	"core/net/dispatcher"
	"core/net/dispatcher/pb"
	"core/net/lan"
)

var _ = logs.Debug

//
var (
	g_lan *lan.Lan
)

// 服务器间相关处理
func InitServers() {
	// init
	g_lan = lan.NewLan(Cfg().LanCfg)

	// reg and watch
	etcd.RegAndWatchs("gate", &Cfg().EtcdCfg, g_lan.Update)

	// recv msg
	go run.Exec(true, procSrvMsg)
}

// recv server msg and send it to client
func procSrvMsg() {
	for {
		raw, e := g_lan.Server.Recv()
		if e != nil {
			logs.Panicln(e)
		}

		var f *pb.PbFrame
		if e := proto.Unmarshal(raw, f); e != nil {
			logs.Panicln(e)
		}

		Dispatch(f)
	}
}

//
func ToServer(c *Client, dstUrl string, d []byte) bool {
	// server
	srvId, _, ok := dispatcher.Url2Part(dstUrl)
	if !ok {
		logs.Warn("invalid url: %v, accId:%v", dstUrl, c.AccId)
		return false
	}

	// message
	f := &pb.PbFrame{
		SrcUrl:  c.Url,
		DstUrls: []string{dstUrl},
		AccId:   int64(c.AccId),
		MsgRaw:  d,
	}

	d, e := proto.Marshal(f)
	if e != nil {
		logs.Warn("accId:%v, error:%v", c.AccId, e)
		return false
	}

	if e := g_lan.Clients.SendMsg(srvId, d); e != nil {
		logs.Warn("send msg failed! accId:%v, error:%v", c.AccId, e)
		return false
	}

	return true
}

//
func SelectRandUrl(srv string) string {
	srvId := g_lan.Clients.SelectRand(srv)
	if "" == srvId {
		return ""
	}
	return dispatcher.Url(srvId, 0)
}

//
func SrvId() string {
	return g_lan.Server.ServerId()
}
