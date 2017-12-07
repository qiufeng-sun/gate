package main

import (
	"github.com/golang/protobuf/proto"

	"util/logs"

	"core/net/dispatcher"
	"core/net/dispatcher/pb"

	"share/pipe"
)

var _ = logs.Debug

// 服务器间相关处理
func InitServers() {
	//
	pipe.Init(Cfg.LanCfg, Cfg.EtcdCfg, func(f *pb.PbFrame) {
		dispatcher.Dispatch(f, func(dstUrl string) {
			// 通知offline
			NoticeServerOffline(dstUrl, *f.SrcUrl)
		})
	})
}

func ToServer(c *Client, dstUrl string, d []byte) bool {
	// message
	f := &pb.PbFrame{
		SrcUrl:  proto.String(c.Url),
		DstUrls: []string{dstUrl},
		AccId:   proto.Int64(c.AccId),
		MsgRaw:  d,
		Offline: proto.Bool(false),
	}

	return pipe.SendFrame2Server(dstUrl, f)
}

//
func NoticeServerOffline(srcUrl, dstUrl string) bool {
	// message
	f := &pb.PbFrame{
		SrcUrl:  proto.String(srcUrl),
		DstUrls: []string{dstUrl},
		Offline: proto.Bool(true),
	}

	return pipe.SendFrame2Server(dstUrl, f)
}
