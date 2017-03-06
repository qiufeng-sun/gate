package main

import (
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"

	"util/logs"

	"core/net/socket"
)

var _ = logs.Debug

//
const (
	// 功能
	GK_Proc_To    = "to"
	GK_Proc_Url   = "url"
	GK_Proc_AccId = "accId"

	// to op
	GK_To_Self   = "self"   // client ping
	GK_To_Client = "client" // server msg
	GK_To_None   = "none"   // server msg
	GK_To_Server = "xx"     // xx为连接到gate的服务器标识。如: logon,battle等

	// url op: default=rand

	// to client msg
	GK_Url_Set = "set"
	GK_Url_Del = "del"

	// to server msg
	GK_Url_Auto    = "auto" // fix+rand
	GK_Url_Rand    = "rand"
	GK_Url_Fix     = "fix"
	GK_Url_RandSet = "rand&set"
)

//
var (
	h_fromClient = NewMsgHandler()
	h_fromServer = NewMsgHandler()
	g_handlers   = map[string]func(*Client, []byte, string) bool{}
)

var getTypeByMsgEnum = func(pkgName, enumName string) reflect.Type {
	name := strings.Replace(enumName, "ID_", pkgName+".", 1)
	return proto.MessageType(name)
}

// msg id example: ID_CSPing
func Register(pkgName string, msgNameIdMap map[string]int32) {
	for name, id := range msgNameIdMap {
		typ := getTypeByMsgEnum(pkgName, name)
		registerHandler(id, typ)
	}
}

func registerHandler(id int32, typ reflect.Type) {
	//
	m := parseGatewayByType(typ)
	if len(m) <= 0 {
		logs.Panicln("message gateway field not set default value!", typ.String())
		return
	}

	//
	var hs []func(c *Client, d []byte, tag string) bool
	var ps []string
	hreg := h_fromClient

	// find handler
	for i := 0; i < len(m); i += 2 {
		k := m[i]
		v := m[i+1]

		h, ok := g_handlers[k]
		if !ok {
			logs.Panicln("invalid message gateway! msg:", typ.String(), ",key:", k)
		}

		if GK_Proc_To == k && (GK_To_Client == v || GK_To_None == v) {
			hreg = h_fromServer
		}

		hs = append(hs, h)
		ps = append(ps, v)
	}

	// reg
	hreg.RegFunc(id, func(c *Client, d []byte) {
		for i, h := range hs {
			if !h(c, d, ps[i]) {
				return
			}
		}
	})
}

//
func handleMsg(mh *MsgHandler, c *Client, d []byte) {
	//
	id, ok := socket.ParseMsgId(d)
	if !ok {
		logs.Warnln("parse msg id failed!")
		return
	}

	h, info, ok := mh.Handler(id)
	if !ok {
		logs.Warn("not found msg handler! accId:%v, msgId:%v", c.AccId, id)
		return
	}

	h.Handle(c, d)
	info.AddStats()
}

//
func HandleClientMsg(c *Client, d []byte) {
	handleMsg(h_fromClient, c, d)
}

//
func HandleServerMsg(c *Client, d []byte) {
	handleMsg(h_fromServer, c, d)
}

// to do
func init() {
	g_handlers[GK_Proc_To] = func(c *Client, d []byte, tag string) bool {
		switch tag {
		case GK_To_Self:
		case GK_To_Client:
		case GK_To_None:
			c.SetUrl()
		default: // to spec server
		}

		return true
	}

	g_handlers[GK_Proc_Url] = nil
}
