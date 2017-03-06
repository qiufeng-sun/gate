package main

import (
	"sync"
	"util/logs"
	"util/run"

	"core/net/msg/protobuf"
	"core/net/socket"
)

var _ = logs.Debug

// temp def // to do
type Frame struct {
	SrcUrl string
	Id     int // 需要处理该消息的逻辑id
	MsgRaw []byte
}

func ToFrame(d []byte) *Frame {
	var r *Frame
	socket.ParseMsgData(d, &r)
	return r
}

//
type Client struct {
	Id    int         // client id
	NetId int         // user's msg
	SMsg  chan []byte // servers' msg(frame)

	AccId int // account id
	Load  int // 负载

	//
	urls  map[string]string // serverName=>serverUrl
	urlOp string            // url op

	//
	f *Frame // 当前正在处理的服务端发送过来的消息
}

// global var
var (
	g_clientId int = 0
	g_clients      = map[int]*Client{} // clientId=>client
	g_lock         = &sync.Mutex{}
)

//
func NewClient(netId int) *Client {
	g_clientId++
	logs.Info("new client! clientId:%v, netId:%v", g_clientId, netId)

	return &Client{
		Id:    g_clientId,
		NetId: netId,
		SMsg:  make(chan []byte, 100),
	}
}

func addClient(c *Client) {
	g_lock.Lock()
	g_clients[c.Id] = c
	g_lock.Unlock()
}

func removeClient(c *Client) {
	g_lock.Lock()
	delete(g_clients, c.Id)
	g_lock.Unlock()
}

// run 处理客户端消息
//  - 客户端发送给服务器的消息，及ping
//  - 服务器发给客户端的消息
func (this *Client) run() {
	addClient(this)
	defer removeClient(this)

	// client msg receiver
	receiver := socket.GetMsgReceiver(this.NetId)
	if nil == receiver {
		logs.Panicln("not found client msg receiver! clientId:", this.Id)
	}

	// get receiver msg chan
	rch := receiver.(socket.IMsgChan).GetMsgChan()

	// to do 检查连接(底层给个监听的chan)
	var d []byte
	select {
	case d = <-rch: // client msg
		// check ddos// to do
		HandleClientMsg(this, d)
	case d = <-this.SMsg: // server msg
		this.f = d // to do need unmarshal
		HandleServerMsg(this, d)
	}
}

func (this *Client) SetUrlOp(op string) {
	this.urlOp = op
}

func (this *Client) ResetUrlOp() {
	this.urlOp = ""
}

// to client msg call// to do
func (this *Client) SetUrl() {
	//	op := this.urlOp
	//	if "" == op {
	//		return
	//	}

	//	switch op {
	//	case GK_Url_Set, GK_Url_Del:
	//		url := c.f.GetSrc()
	//		svc, _, _ := nc.UrlToPart(c.f.GetSrc())
	//		if GK_Url_Del == op {
	//			delete(c.urls, svc)
	//		} else {
	//			if nil == c.urls {
	//				c.urls = map[string]string{}
	//			}
	//			c.urls[svc] = url
	//		}
	//	}

	//	c.ResetUrlOp()
}

// SelectUrl to server msg call // to do
// @return server url
func (this *Client) SelectUrl(svc string) string {
	//	defer c.ResetUrlOp()

	//	switch c.urlOp {
	//	case "", GK_Url_Rand: // rand
	//		return c.gs.selectRandUrl(svc)
	//	case GK_Url_Fix: // use cached url
	//		return c.urls[svc]
	//	case GK_Url_RandSet:
	//		url := c.gs.selectRandUrl(svc)
	//		if url != "" {
	//			c.urls[svc] = url
	//		}
	//		return url
	//	}

	//	// cached + rand
	//	url := c.urls[svc]
	//	if url != "" {
	//		return url
	//	}
	//	return c.gs.selectRandUrl(svc)
	return ""
}

// 客户端相关处理
func InitClients() {
	//
	cfg := Cfg()

	// listen to players
	e := socket.Serve(cfg.PlayerLsnAddr, cfg.PlayerMaxConn, &protobuf.PbParser{})
	if e != nil {
		logs.Panicln(e)
	}

	// manage clients
	go run.Exec(true, updateLogonWait)
}

// 更新登录等待状态的客户端
func updateLogonWait() {
	// 待登录的客户端
	logonWait := socket.GetLogonWait()

	// process
	for netId := range logonWait {
		// 检查
		if !checkConnection(netId) {
			socket.KickClient(netId)
			continue
		}

		// new client
		client := NewClient(netId)

		// process client// to do 1st param to false?
		go run.Exec(true, func() {
			defer socket.KickClient(client.NetId)
			client.run()
		})
	}
}

// 检查并登录
func checkConnection(netId int) bool {
	// 底层断开了连接
	if !socket.IsClientConnect(netId) {
		return false
	}

	return true
}
