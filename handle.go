package main

import (
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"

	"util/logs"

	"core/net/socket"

	"share/handler"
)

var _ = logs.Debug

// message对象gateway参数
const (
	// key
	GK_Proc_To    = "to"
	GK_Proc_Url   = "url"
	GK_Proc_AccId = "accId"

	// to op
	GK_To_Self   = "self"   // client ping
	GK_To_Client = "client" // from server msg
	GK_To_None   = "none"   // from server msg
	GK_To_Kick   = "kick"   // from server msg
	GK_To_Logon  = "logon"
	GK_To_Server = "xx" // xx为连接到gate的服务器标识。如: match,battle等

	// url op: default=rand

	// url op -- to client msg
	GK_Url_Set = "set"
	GK_Url_Del = "del"

	// url op to server msg
	GK_Url_Auto    = "auto" // fix+rand
	GK_Url_Rand    = "rand"
	GK_Url_Fix     = "fix"
	GK_Url_RandSet = "rand&set"
)

//
type hfunc func(*Client, []byte)

func (this hfunc) Handle(receiver interface{}, msgBuff []byte) {
	this(receiver.(*Client), msgBuff)
}

//
type MsgHandler struct {
	*handler.MsgHandler
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler{handler.NewMsgHandler()}
}

func (this *MsgHandler) RegFunc(msgId int32, f func(c *Client, d []byte)) {
	this.RegHandler(msgId, hfunc(f))
}

//
var (
	h_fromClient = NewMsgHandler()
	h_fromServer = NewMsgHandler()
	g_handlers   = map[string]func(*Client, []byte, string) bool{}
)

//
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
			logs.Panicln("invalid message gateway! msg:", typ.String(), "key:", k)
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
		c.Kick()
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
	logs.Debug("receive client msg!")
	handleMsg(h_fromClient, c, d)
}

//
func HandleServerMsg(c *Client, d []byte) {
	logs.Debug("receive server msg!")
	handleMsg(h_fromServer, c, d)
}

//
func init() {
	g_handlers[GK_Proc_To] = func(c *Client, d []byte, tag string) bool {
		logs.Debug("handle to %v", tag)

		switch tag {
		case GK_To_Self: // to do

		case GK_To_Client:
			c.ProcUrlOp()
			if !c.SendBytes(d) {
				c.Kick()
				return false
			}

		case GK_To_None:
			c.ProcUrlOp()

		case GK_To_Kick:
			c.ProcUrlOp()
			c.Kick()
			return false

		case GK_To_Logon:
			dstUrl := c.SelectUrl(tag)
			if "" == dstUrl {
				logs.Warn("select dstUrl failed! tag:%v", tag)
				c.Kick()
				return false
			}
			return ToServer(c, dstUrl, d)

		default: // to other servers(except logon)
			// check account id 1st
			if !c.IsSetAccId() {
				logs.Warnln("please logon 1st!")
				c.Kick()
				return false
			}

			dstUrl := c.SelectUrl(tag)
			if "" == dstUrl {
				c.Kick()
				return false
			}
			return ToServer(c, dstUrl, d)
		}

		return true
	}

	g_handlers[GK_Proc_Url] = func(c *Client, d []byte, tag string) bool {
		c.SetUrlOp(tag)
		return true
	}

	g_handlers[GK_Proc_AccId] = func(c *Client, d []byte, tag string) bool {
		f := c.CurF
		c.AccId = *f.AccId

		logs.Debug("set accId:%v", c.AccId)

		return true
	}
}
