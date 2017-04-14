package main

import (
	"time"

	"util/logs"
	"util/run"

	"core/net/dispatcher"
	"core/net/dispatcher/pb"
	"core/net/lan"
	"core/net/msg/protobuf"
	"core/net/socket"
)

var _ = logs.Debug

//
var (
	g_dispatcher = dispatcher.New("gate dispatcher", SrvId())
)

func addClient(c *Client) {
	g_dispatcher.Register(c)
}

func removeClient(c *Client) {
	g_dispatcher.Unregister(c)
}

func Dispatch(f *pb.PbFrame) {
	g_dispatcher.Dispatch(f)
}

// 来自玩家的连接管理// to do 心跳
type Client struct {
	*dispatcher.BaseUnit

	NetId int // user's msg

	AccId int // account id
	Load  int // 负载

	//
	urls  map[string]string // serverName=>serverUrl
	urlOp string            // url op

	//
	chKick chan bool
}

//
func NewClient(netId int) *Client {
	return &Client{
		BaseUnit: dispatcher.NewBaseUnit(100),
		NetId:    netId,
		urls:     map[string]string{},
		chKick:   make(chan bool, 1),
	}
}

//
func (this *Client) AddFrame(f *dispatcher.Frame) bool {
	if !this.BaseUnit.AddFrame(f) {
		logs.Warn("too many frames, kick it! accId:%v, id:%v, netId:%v",
			this.AccId, this.Id, this.NetId)
		this.Kick()
		return false
	}

	return true
}

//
func (this *Client) Kick() {
	select {
	case this.chKick <- true:
	default:
	}
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

	ticker := time.NewTicker(time.Millisecond * 1000)
	defer ticker.Stop()

	//
	for {
		select {
		case <-this.chKick:
			logs.Warn("client been kicked! accId:%v, id:%v, netId:%v",
				this.AccId, this.Id, this.NetId)
			return
		case <-ticker.C:
			if !socket.IsClientConnect(this.NetId) {
				return
			}
			this.Load = 0
		case d := <-rch: // client msg
			// check ddos
			this.Load++
			if this.Load > Cfg().PlayerMaxLoad {
				logs.Warn("kick player! accId:%v, id:%v, netId:%v, load:%v",
					this.AccId, this.Id, this.NetId, this.Load)
				return
			}
			HandleClientMsg(this, d)
		case f := <-this.Frames:
			this.CurF = f
			HandleServerMsg(this, f.MsgRaw)
		}
	}
}

func (this *Client) SetUrlOp(op string) {
	this.urlOp = op
}

func (this *Client) ResetUrlOp() {
	this.urlOp = ""
}

// to client msg call
func (this *Client) SetUrl() {
	op := this.urlOp
	if "" == op {
		return
	}

	switch op {
	case GK_Url_Set, GK_Url_Del:
		url := this.CurF.SrcUrl
		srvId, _, ok := dispatcher.Url2Part(url)
		if !ok {
			logs.Warn("invalid src url! url: %v", url)
			break
		}

		srv := lan.SrvName(srvId)
		if GK_Url_Del == op {
			delete(this.urls, srv)
		} else {
			this.urls[srv] = url
		}
	}

	this.ResetUrlOp()
}

// SelectUrl to server msg call
// @return server url
func (this *Client) SelectUrl(srv string) string {
	defer this.ResetUrlOp()

	switch this.urlOp {
	case "", GK_Url_Rand: // rand
		return SelectRandUrl(srv)
	case GK_Url_Fix: // use cached url
		return this.urls[srv]
	case GK_Url_RandSet:
		url := SelectRandUrl(srv)
		if url != "" {
			this.urls[srv] = url
		}
		return url
	}

	// cached + rand
	url := this.urls[srv]
	if url != "" {
		return url
	}
	return SelectRandUrl(srv)
}

//
func (this *Client) SendBytes(d []byte) bool {
	e := socket.SendBytes(this.NetId, d)
	if e != nil {
		logs.Warn("msg send failed! accId:%v, id:%v, netId:%v, error:%v",
			this.AccId, this.Id, this.NetId, e)
		return false
	}
	return true
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
